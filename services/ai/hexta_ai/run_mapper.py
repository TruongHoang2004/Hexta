"""Pure run aggregate boundary transformations; no storage or provider calls."""

from hexta.ai.v1 import service_pb2 as pb
from hexta_ai.graph import PreviewRequest


def preview_request(request: pb.StartRunRequest) -> PreviewRequest:
    return PreviewRequest(
        message=request.message,
        scripts=[{"id": s.id, "trigger": s.trigger, "text": s.text, "active": s.active} for s in request.scripts],
        history=[{"role": h.role, "text": h.text} for h in request.history],
        facts=[{"id": f.id, "text": f.text} for f in request.facts],
        constraints=list(request.constraints), allow_generation=request.allow_generation,
        context_revision=request.context_revision, control_generation=request.control_generation,
    )


def completed_run(run: pb.Run, result) -> pb.Run:
    candidate = pb.Run()
    candidate.CopyFrom(run)
    candidate.route = result.route
    candidate.text = result.text or ""
    candidate.script_id = result.script_id or ""
    candidate.reason = result.reason
    del candidate.steps[:]
    candidate.steps.extend(result.steps)
    candidate.simulated = result.simulated
    candidate.provider = result.provider
    candidate.model = result.model
    candidate.confidence = result.confidence
    candidate.input_tokens = result.usage.input_tokens
    candidate.output_tokens = result.usage.output_tokens
    candidate.usage_estimated = result.usage.estimated
    candidate.error_code = ""
    candidate.status = pb.RUN_STATUS_SUCCEEDED
    return candidate
