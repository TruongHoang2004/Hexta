"""Bounded candidate routing. Only Go can authorize business actions or sending."""

import json
import logging
import os
import re
import unicodedata
from typing import Literal, TypedDict

from langgraph.graph import END, START, StateGraph
from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator

from hexta_ai.providers.base import BaseAIProvider, Classification, ProviderError, Usage
from hexta_ai.providers.fake import FakeProvider
from hexta_ai.redaction import contact_values, redact_pii
from hexta_ai.routing.decision import HANDOFF_REASONS, RouteDecision, deterministic_decision
from hexta_ai.routing.jev import JevUnavailable

log = logging.getLogger(__name__)

# Channel text limit enforced by Go for any proposal.
ANSWER_MAX_CHARS = 2000


def answer_min_confidence_from_env() -> float:
    try:
        value = float(os.getenv("AI_ANSWER_MIN_CONFIDENCE", "0.7"))
    except ValueError:
        raise RuntimeError("Invalid AI_ANSWER_MIN_CONFIDENCE") from None
    if not 0.5 <= value <= 1:
        raise RuntimeError("Invalid AI_ANSWER_MIN_CONFIDENCE")
    return value


# Static clarifying question (not model-authored, so it needs no review; Go may auto-send it). AI_CLARIFY_TEXT
# overrides it per deployment.
DEFAULT_CLARIFY_TEXT = ("Dạ, anh/chị đang quan tâm sản phẩm hoặc dịch vụ nào ạ? Anh/chị nói rõ thêm giúp shop "
                        "để shop hỗ trợ chính xác nhé.")


def checked_clarify_text(value: str) -> str:
    value = sanitize(value)
    if not value or len(value) > 500 or "{" in value or "}" in value:
        raise RuntimeError("Invalid AI_CLARIFY_TEXT")
    return value


def clarify_text_from_env() -> str:
    value = os.getenv("AI_CLARIFY_TEXT", "")
    return checked_clarify_text(value) if value.strip() else DEFAULT_CLARIFY_TEXT


def sanitize(text: str) -> str:
    return "".join(c for c in text if c in "\n\t" or unicodedata.category(c) not in ("Cc", "Cf", "Cs")).strip()


def redact(text: str) -> str:
    """Legacy heuristic redaction for the provider classifier path (emails and long digit runs)."""
    text = re.sub(r"[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}", "[REDACTED_EMAIL]", text)
    return re.sub(r"(?<!\w)\+?\d[\d ().-]{6,}\d(?!\w)", "[REDACTED_NUMBER]", text)


def numbers(text: str) -> set[str]:
    """Digit sequences with thousand/decimal separators removed ("1.250.000" -> "1250000")."""
    return {re.sub(r"\D", "", m) for m in re.findall(r"\d(?:[\d.,\s]*\d)?", text)} - {""}


class InputModel(BaseModel):
    model_config = ConfigDict(extra="forbid", str_strip_whitespace=True)

    @field_validator("*", mode="before")
    @classmethod
    def clean_strings(cls, value):
        return sanitize(value) if isinstance(value, str) else value


class Script(InputModel):
    id: str = Field(min_length=1, max_length=100)
    trigger: str = Field(min_length=1, max_length=200)
    text: str = Field(min_length=1, max_length=4000)
    active: bool = False


class ConversationMessage(InputModel):
    role: Literal["user", "assistant"]
    text: str = Field(min_length=1, max_length=2000)


class DomainFact(InputModel):
    id: str = Field(min_length=1, max_length=100)
    text: str = Field(min_length=1, max_length=2000)


class PreviewRequest(InputModel):
    message: str = Field(min_length=1, max_length=8000)
    scripts: list[Script] = Field(default_factory=list, max_length=50)
    history: list[ConversationMessage] = Field(default_factory=list, max_length=20)
    facts: list[DomainFact] = Field(default_factory=list, max_length=50)
    constraints: list[str] = Field(default_factory=list, max_length=20)
    allow_generation: bool = False
    allow_knowledge_answer: bool = False
    tenant_id: str | None = Field(default=None, pattern=r"^[A-Za-z0-9_-]{1,100}$")
    context_revision: int = Field(default=0, ge=0)
    control_generation: int = Field(default=0, ge=0)

    @field_validator("constraints")
    @classmethod
    def clean_constraints(cls, values):
        result = [sanitize(v) for v in values]
        if any(not v or len(v) > 1000 for v in result):
            raise ValueError("Invalid domain constraints")
        return result

    @model_validator(mode="after")
    def unique_ids(self):
        for values in (self.scripts, self.facts):
            if len({item.id for item in values}) != len(values):
                raise ValueError("Duplicate snapshot identifiers")
        return self


class PreviewResult(BaseModel):
    simulated: bool = True
    route: Literal["script", "generate", "clarify", "human"]
    text: str | None = None
    script_id: str | None = None
    reason: str
    steps: list[str]
    provider: str = "fake"
    model: str = "deterministic"
    confidence: float = Field(default=0, ge=0, le=1)
    usage: Usage = Field(default_factory=Usage)
    candidate_kind: Literal["", "script", "generated", "clarify", "handoff"] = ""
    cited_source_ids: list[str] = Field(default_factory=list, max_length=20)
    needs_human: bool = False
    handoff_reason: Literal["", *HANDOFF_REASONS] = ""
    requires_review: bool = False
    router: str = ""
    router_model: str = ""
    route_probability: float = Field(default=0, ge=0, le=1)
    router_fallback_reason: str = ""


class State(TypedDict, total=False):
    request: PreviewRequest
    context: dict
    classification: Classification
    decision: RouteDecision
    knowledge: list
    knowledge_error: str
    usage: Usage
    result: PreviewResult
    steps: list[str]


def add_usage(first: Usage, second: Usage) -> Usage:
    return Usage(input_tokens=first.input_tokens + second.input_tokens,
                 output_tokens=first.output_tokens + second.output_tokens,
                 estimated=first.estimated or second.estimated)


def build_graph(provider: BaseAIProvider, *, router=None, knowledge=None, simulated: bool | None = None,
                answer_min_confidence: float | None = None, clarify_text: str | None = None):
    """router: optional primary router (JevRouter or the offline KeywordRouter; None = provider classifier).
    knowledge: optional KnowledgeService."""
    run_simulated = provider.simulated if simulated is None else simulated
    min_confidence = answer_min_confidence_from_env() if answer_min_confidence is None else answer_min_confidence
    clarification = clarify_text_from_env() if clarify_text is None else checked_clarify_text(clarify_text)

    def validate_input(state):
        # Revalidate even when an internal caller bypassed normal Pydantic creation.
        request = PreviewRequest.model_validate(state["request"].model_dump())
        return {"request": request, "steps": ["validate_input"], "usage": Usage()}

    async def prepare_context(state):
        request = state["request"]
        context = request.model_dump(exclude={"tenant_id", "allow_knowledge_answer"})
        context["scripts"] = [s.model_dump() for s in request.scripts if s.active]
        # Redact only content, preserving snapshot identities for exact validation.
        context["message"] = redact(context["message"])
        for s in context["scripts"]:
            s["trigger"], s["text"] = redact(s["trigger"]), redact(s["text"])
        for item in [*context["history"], *context["facts"]]:
            item["text"] = redact(item["text"])
        context["constraints"] = [redact(c) for c in context["constraints"]]
        if await provider.estimate_tokens(json.dumps(context, ensure_ascii=False)) > provider.max_input_tokens:
            raise ProviderError("INPUT_BUDGET_EXCEEDED")
        return {"context": context, "steps": [*state["steps"], "prepare_context"]}

    async def provider_decision(state):
        decision, usage = await provider.classify(state["context"])
        # Provider-selected IDs and confidence are untrusted; enforce eligibility here.
        active_ids = {s["id"] for s in state["context"]["scripts"]}
        if (decision.confidence < provider.confidence_threshold
                or (decision.route == "script" and decision.script_id not in active_ids)
                or (decision.route != "script" and decision.script_id is not None)
                or (decision.route == "generate" and (
                    not state["request"].allow_generation or not state["context"]["facts"]))):
            decision = Classification(route="human", script_id=None, confidence=decision.confidence)
        routed = RouteDecision(route=decision.route, script_id=decision.script_id, probability=decision.confidence,
                               router="deterministic" if provider.simulated else "provider",
                               router_model=provider.model, fallback_reason="jev_disabled")
        return decision, routed, usage

    async def jev_decision(state):
        request = state["request"]
        scripts = [s.model_dump() for s in request.scripts if s.active]
        history = [h.model_dump() for h in request.history]
        try:
            decision = await router.route(request.message, scripts, history)
            reason = getattr(router, "uncertain_reason", "jev_low_probability")
        except JevUnavailable as error:
            decision, reason = None, error.code
        if decision is None:
            # Never fall back to a paid model call: unique exact match or human.
            decision = deterministic_decision(request.message, scripts, reason)
            decision = decision.model_copy(update={"router_model": router.model})
        elif decision.route == "answer" and not (request.allow_knowledge_answer and knowledge is not None):
            decision = decision.model_copy(update={
                "route": "human", "handoff_reason": "business_review",
                "reason": "knowledge_answer_not_enabled" if not request.allow_knowledge_answer else "knowledge_unavailable"})
        classification = Classification(route="script" if decision.route == "script" else "human",
                                        script_id=decision.script_id, confidence=decision.probability)
        return classification, decision, Usage()

    async def classify_intent(state):
        classification, decision, usage = await (jev_decision(state) if router is not None else provider_decision(state))
        return {"classification": classification, "decision": decision, "usage": usage,
                "steps": [*state["steps"], "classify_intent"]}

    def result(state, route, reason, text=None, script_id=None, **fields):
        decision = state["decision"]
        values = {"candidate_kind": {"script": "script", "generate": "generated", "clarify": "clarify",
                                     "human": "handoff"}[route],
                  "router": decision.router, "router_model": decision.router_model,
                  "route_probability": decision.probability, "router_fallback_reason": decision.fallback_reason}
        if route == "human":
            values.update(needs_human=True, handoff_reason=decision.handoff_reason or "insufficient_information")
        if route == "generate":
            values["requires_review"] = True
        values.update(fields)
        values.setdefault("confidence", state["classification"].confidence)
        return PreviewResult(
            simulated=run_simulated, route=route, reason=reason, text=text,
            script_id=script_id, steps=[], provider=provider.name, model=provider.model,
            usage=state["usage"], **values,
        )

    def handoff(state, handoff_reason, reason, **fields):
        output = result(state, "human", reason, **fields)
        return output.model_copy(update={"handoff_reason": handoff_reason})

    def render_script(state):
        script = next(s for s in state["context"]["scripts"] if s["id"] == state["decision"].script_id)
        reason = state["decision"].reason or (
            "simulated_exact_match" if provider.simulated else "active_script_match")
        return {"result": result(state, "script", reason, script["text"], script["id"]),
                "steps": [*state["steps"], "render_script"]}

    async def generate_candidate(state):
        candidate, usage = await provider.generate_candidate(state["context"])
        facts = {f["id"]: f["text"] for f in state["context"]["facts"]}
        valid = (bool(candidate.fact_ids) and len(set(candidate.fact_ids)) == len(candidate.fact_ids)
                 and all(i in facts for i in candidate.fact_ids))
        grounded = "\n".join(facts[i] for i in candidate.fact_ids) if valid else ""
        total = add_usage(state["usage"], usage)
        state = {**state, "usage": total}
        output = result(state, "generate", "grounded_domain_candidate", candidate.text,
                        cited_source_ids=[f"fact:{i}" for i in candidate.fact_ids] if valid else [])
        if not valid or candidate.text != grounded:
            output = handoff(state, "insufficient_information", "ungrounded_candidate")
        return {"result": output, "usage": total, "steps": [*state["steps"], "generate_candidate"]}

    async def retrieve_knowledge(state):
        request = state["request"]
        steps = [*state["steps"], "retrieve_knowledge"]
        items = [{"id": f"fact:{f.id}", "title": "", "text": f.text, "attributes": {}} for f in request.facts]
        if knowledge is None or not request.tenant_id:
            return {"knowledge": items, "knowledge_error": "" if items else "knowledge_unavailable", "steps": steps}
        active = {s.id for s in request.scripts if s.active}
        try:
            found = await knowledge.retrieve(request.tenant_id, redact_pii(request.message), active_script_ids=active)
        except Exception:
            # Storage/embedding failure: no answer without grounding; bodies never leave this process.
            log.warning("Knowledge retrieval failed; handing off without logging payloads")
            return {"knowledge": [], "knowledge_error": "knowledge_retrieval_failed", "steps": steps}
        items = [{"id": i.id, "title": i.title, "text": i.text, "attributes": i.attributes} for i in found] + items
        return {"knowledge": items, "knowledge_error": "", "steps": steps}

    async def answer_candidate(state):
        steps = [*state["steps"], "answer_candidate"]
        items = state.get("knowledge") or []
        if state.get("knowledge_error") in ("knowledge_retrieval_failed", "knowledge_unavailable") and not items:
            return {"result": handoff(state, "ai_error", state["knowledge_error"]), "steps": steps}
        if not items:
            return {"result": handoff(state, "insufficient_information", "knowledge_insufficient"), "steps": steps}
        request = state["request"]
        # Claude needs the question (addresses kept for shipping questions) but not customer phones/emails.
        context = {
            "knowledge": items,
            "conversation": [{"role": h.role, "text": redact_pii(h.text, addresses=False)} for h in request.history],
            "customer_message": redact_pii(request.message, addresses=False),
            "constraints": list(request.constraints),
        }
        try:
            answer, usage = await provider.answer(context)
        except ProviderError as error:
            return {"result": handoff(state, "ai_error", f"answer_{error.code.lower()}"), "steps": steps}
        total = add_usage(state["usage"], usage)
        state = {**state, "usage": total}
        known = {item["id"]: item for item in items}
        cited = answer.cited_source_ids
        text = sanitize(answer.text)
        if answer.needs_human:
            output = handoff(state, "insufficient_information", "model_requested_human")
        elif answer.kind == "clarify":
            output = (result(state, "clarify", "model_clarification", text, requires_review=True,
                             confidence=answer.confidence)
                      if text and len(text) <= ANSWER_MAX_CHARS else
                      handoff(state, "ai_error", "invalid_answer_output"))
        else:
            sources = "\n".join(f"{known[i]['title']}\n{known[i]['text']}\n{json.dumps(known[i]['attributes'], ensure_ascii=False)}"
                                for i in cited if i in known)
            if (not cited or len(set(cited)) != len(cited) or any(i not in known for i in cited)):
                output = handoff(state, "insufficient_information", "ungrounded_answer")
            elif answer.confidence < min_confidence:
                output = handoff(state, "insufficient_information", "low_answer_confidence")
            elif not text or len(text) > ANSWER_MAX_CHARS:
                output = handoff(state, "ai_error", "invalid_answer_output")
            elif not numbers(text) <= numbers(sources) | numbers(request.message):
                # Every price, stock count or other number must come from the cited records.
                output = handoff(state, "insufficient_information", "ungrounded_number")
            else:
                output = result(state, "generate", "knowledge_answer", text, cited_source_ids=cited,
                                confidence=answer.confidence)
        return {"result": output.model_copy(update={"usage": total}), "usage": total, "steps": steps}

    def ask_clarification(state):
        return {"result": result(state, "clarify", state["decision"].reason or "needs_clarification", clarification),
                "steps": [*state["steps"], "ask_clarification"]}

    def request_human(state):
        decision = state["decision"]
        reason = decision.reason or ("ai_disabled" if provider.name == "disabled" else
                                     "no_unique_active_script_fake_model_only" if provider.simulated else
                                     "human_review_required")
        return {"result": result(state, "human", reason), "steps": [*state["steps"], "request_human"]}

    def validate_output(state):
        output = state["result"]
        text = output.text
        invalid = {"route": "human", "text": None, "script_id": None, "candidate_kind": "handoff",
                   "cited_source_ids": [], "requires_review": False, "needs_human": True}
        if text and ("{" in text or "}" in text):
            output = output.model_copy(update={**invalid, "reason": "unresolved_script_variables",
                                               "handoff_reason": "ai_error"})
        elif text and output.requires_review:
            # Model text keeps tenant numbers (prices, hotline) but never contacts absent from the knowledge.
            allowed = contact_values("\n".join(i["text"] for i in state.get("knowledge") or []))
            output = output.model_copy(update={"text": redact_pii(sanitize(text), addresses=False, keep=allowed)})
        elif text:
            text = redact(sanitize(text))
            if not text or len(text) > 4000:
                output = output.model_copy(update={**invalid, "reason": "invalid_candidate_output",
                                                   "handoff_reason": "ai_error"})
            else:
                output = output.model_copy(update={"text": text})
        steps = [*state["steps"], "validate_output"]
        output = PreviewResult.model_validate(output.model_copy(update={"steps": steps}).model_dump())
        json.dumps(output.model_dump(), allow_nan=False)
        return {"steps": steps, "result": output}

    builder = StateGraph(State)
    nodes = {"validate_input": validate_input, "prepare_context": prepare_context,
             "classify_intent": classify_intent, "render_script": render_script,
             "generate_candidate": generate_candidate, "retrieve_knowledge": retrieve_knowledge,
             "answer_candidate": answer_candidate, "ask_clarification": ask_clarification,
             "request_human": request_human, "validate_output": validate_output}
    for name, function in nodes.items():
        builder.add_node(name, function)
    builder.add_edge(START, "validate_input")
    builder.add_edge("validate_input", "prepare_context")
    builder.add_edge("prepare_context", "classify_intent")
    builder.add_conditional_edges("classify_intent", lambda s: {
        "script": "render_script", "generate": "generate_candidate", "answer": "retrieve_knowledge",
        "clarify": "ask_clarification", "human": "request_human",
    }[s["decision"].route])
    builder.add_edge("retrieve_knowledge", "answer_candidate")
    for node in ("render_script", "generate_candidate", "answer_candidate", "ask_clarification", "request_human"):
        builder.add_edge(node, "validate_output")
    builder.add_edge("validate_output", END)
    return builder.compile()


preview_graph = build_graph(FakeProvider())
