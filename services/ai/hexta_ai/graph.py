"""Offline graph for exercising the service boundary, not real model inference."""

from typing import Literal, TypedDict

from langgraph.graph import END, START, StateGraph
from pydantic import BaseModel, ConfigDict, Field


class Script(BaseModel):
    model_config = ConfigDict(extra="forbid", str_strip_whitespace=True)
    id: str = Field(min_length=1, max_length=100)
    trigger: str = Field(min_length=1, max_length=200)
    text: str = Field(min_length=1, max_length=4000)
    active: bool = False


class PreviewRequest(BaseModel):
    model_config = ConfigDict(extra="forbid", str_strip_whitespace=True)
    message: str = Field(min_length=1, max_length=8000)
    scripts: list[Script] = Field(default_factory=list, max_length=50)


class PreviewResult(BaseModel):
    simulated: Literal[True] = True
    route: Literal["script", "human"]
    text: str | None = None
    script_id: str | None = None
    reason: str
    steps: list[str]


class State(TypedDict, total=False):
    request: PreviewRequest
    matches: list[Script]
    result: PreviewResult
    steps: list[str]


def prepare(state: State) -> dict:
    request = state["request"]
    normalized = request.message.strip().casefold()
    matches = [
        script for script in request.scripts
        if script.active and script.trigger.strip().casefold() == normalized
    ]
    return {"matches": matches, "steps": ["prepare_context"]}


def route(state: State) -> dict:
    return {"steps": [*state["steps"], "route"]}


def branch(state: State) -> str:
    return "render_script" if len(state["matches"]) == 1 else "request_human"


def render_script(state: State) -> dict:
    script = state["matches"][0]
    steps = [*state["steps"], "render_script"]
    # Variable rendering needs authorized product facts; never emit unresolved templates.
    if "{" in script.text or "}" in script.text:
        result = PreviewResult(route="human", reason="unresolved_script_variables", steps=steps)
    else:
        result = PreviewResult(
            route="script", script_id=script.id, text=script.text,
            reason="simulated_exact_match", steps=steps,
        )
    return {"result": result, "steps": steps}


def request_human(state: State) -> dict:
    steps = [*state["steps"], "request_human"]
    return {
        "result": PreviewResult(
            route="human", reason="no_unique_active_script_fake_model_only", steps=steps,
        ),
        "steps": steps,
    }


def validate_output(state: State) -> dict:
    steps = [*state["steps"], "validate_output"]
    return {"steps": steps, "result": state["result"].model_copy(update={"steps": steps})}


builder = StateGraph(State)
builder.add_node("prepare_context", prepare)
builder.add_node("route", route)
builder.add_node("render_script", render_script)
builder.add_node("request_human", request_human)
builder.add_node("validate_output", validate_output)
builder.add_edge(START, "prepare_context")
builder.add_edge("prepare_context", "route")
builder.add_conditional_edges("route", branch)
builder.add_edge("render_script", "validate_output")
builder.add_edge("request_human", "validate_output")
builder.add_edge("validate_output", END)
preview_graph = builder.compile()
