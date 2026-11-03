"""Bounded candidate routing. Only Go can authorize business actions or sending."""

import json
import re
import unicodedata
from typing import Literal, TypedDict

from langgraph.graph import END, START, StateGraph
from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator

from hexta_ai.providers.base import BaseAIProvider, Classification, ProviderError, Usage
from hexta_ai.providers.fake import FakeProvider


def sanitize(text: str) -> str:
    return "".join(c for c in text if c in "\n\t" or unicodedata.category(c) not in ("Cc", "Cf", "Cs")).strip()


def redact(text: str) -> str:
    text = re.sub(r"[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}", "[REDACTED_EMAIL]", text)
    return re.sub(r"(?<!\w)\+?\d[\d ().-]{6,}\d(?!\w)", "[REDACTED_NUMBER]", text)


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


class State(TypedDict, total=False):
    request: PreviewRequest
    context: dict
    classification: Classification
    usage: Usage
    result: PreviewResult
    steps: list[str]


def build_graph(provider: BaseAIProvider):
    def validate_input(state):
        # Revalidate even when an internal caller bypassed normal Pydantic creation.
        request = PreviewRequest.model_validate(state["request"].model_dump())
        return {"request": request, "steps": ["validate_input"], "usage": Usage()}

    async def prepare_context(state):
        request = state["request"]
        context = request.model_dump()
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

    async def classify_intent(state):
        decision, usage = await provider.classify(state["context"])
        # Provider-selected IDs and confidence are untrusted; enforce eligibility here.
        active_ids = {s["id"] for s in state["context"]["scripts"]}
        if (decision.confidence < provider.confidence_threshold
                or (decision.route == "script" and decision.script_id not in active_ids)
                or (decision.route != "script" and decision.script_id is not None)
                or (decision.route == "generate" and (
                    not state["request"].allow_generation or not state["context"]["facts"]))):
            decision = Classification(route="human", script_id=None, confidence=decision.confidence)
        return {"classification": decision, "usage": usage,
                "steps": [*state["steps"], "classify_intent"]}

    def result(state, route, reason, text=None, script_id=None):
        return PreviewResult(
            simulated=provider.simulated, route=route, reason=reason, text=text,
            script_id=script_id, steps=[], provider=provider.name, model=provider.model,
            confidence=state["classification"].confidence, usage=state["usage"],
        )

    def render_script(state):
        script = next(s for s in state["context"]["scripts"] if s["id"] == state["classification"].script_id)
        return {"result": result(state, "script", "simulated_exact_match" if provider.simulated else "active_script_match",
                                 script["text"], script["id"]),
                "steps": [*state["steps"], "render_script"]}

    async def generate_candidate(state):
        candidate, usage = await provider.generate_candidate(state["context"])
        facts = {f["id"]: f["text"] for f in state["context"]["facts"]}
        valid = (bool(candidate.fact_ids) and len(set(candidate.fact_ids)) == len(candidate.fact_ids)
                 and all(i in facts for i in candidate.fact_ids))
        grounded = "\n".join(facts[i] for i in candidate.fact_ids) if valid else ""
        output = result(state, "generate", "grounded_domain_candidate", candidate.text)
        if not valid or candidate.text != grounded:
            output = result(state, "human", "ungrounded_candidate")
        total = Usage(input_tokens=state["usage"].input_tokens + usage.input_tokens,
                      output_tokens=state["usage"].output_tokens + usage.output_tokens,
                      estimated=state["usage"].estimated or usage.estimated)
        return {"result": output.model_copy(update={"usage": total}), "usage": total,
                "steps": [*state["steps"], "generate_candidate"]}

    def ask_clarification(state):
        return {"result": result(state, "clarify", "needs_clarification",
                                 "Could you clarify which product or service you mean?"),
                "steps": [*state["steps"], "ask_clarification"]}

    def request_human(state):
        reason = "no_unique_active_script_fake_model_only" if provider.simulated else "human_review_required"
        return {"result": result(state, "human", reason), "steps": [*state["steps"], "request_human"]}

    def validate_output(state):
        output = state["result"]
        text = output.text
        if text and ("{" in text or "}" in text):
            output = output.model_copy(update={"route": "human", "text": None, "script_id": None,
                                               "reason": "unresolved_script_variables"})
        elif text:
            text = redact(sanitize(text))
            if not text or len(text) > 4000:
                output = output.model_copy(update={"route": "human", "text": None, "script_id": None,
                                                   "reason": "invalid_candidate_output"})
            else:
                output = output.model_copy(update={"text": text})
        steps = [*state["steps"], "validate_output"]
        output = PreviewResult.model_validate(output.model_copy(update={"steps": steps}).model_dump())
        json.dumps(output.model_dump(), allow_nan=False)
        return {"steps": steps, "result": output}

    builder = StateGraph(State)
    nodes = {"validate_input": validate_input, "prepare_context": prepare_context,
             "classify_intent": classify_intent, "render_script": render_script,
             "generate_candidate": generate_candidate, "ask_clarification": ask_clarification,
             "request_human": request_human, "validate_output": validate_output}
    for name, function in nodes.items():
        builder.add_node(name, function)
    builder.add_edge(START, "validate_input")
    builder.add_edge("validate_input", "prepare_context")
    builder.add_edge("prepare_context", "classify_intent")
    builder.add_conditional_edges("classify_intent", lambda s: {
        "script": "render_script", "generate": "generate_candidate",
        "clarify": "ask_clarification", "human": "request_human",
    }[s["classification"].route])
    for node in ("render_script", "generate_candidate", "ask_clarification", "request_human"):
        builder.add_edge(node, "validate_output")
    builder.add_edge("validate_output", END)
    return builder.compile()


preview_graph = build_graph(FakeProvider())
