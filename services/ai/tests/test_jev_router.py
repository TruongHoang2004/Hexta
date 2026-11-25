"""Jev routing policy, thresholds, redaction and the real SDK adapter over a local mock transport."""

import asyncio
import json
import logging

import httpx2
import pytest

from hexta_ai.routing.jev import JevConfig, JevRouter, JevUnavailable, _SdkClient, jev_config_from_env, jev_router_from_env
from tests.fakes import JEV, fake_router, jev_response

SCRIPTS = [{"id": "ver-ship", "trigger": "phí ship, ship bao nhiêu", "text": "Phí ship 30.000đ toàn quốc.", "active": True},
           {"id": "ver-hours", "trigger": "giờ mở cửa", "text": "Shop mở cửa 8h-22h.", "active": True}]


def route(router, message="áo polo còn size L không", scripts=SCRIPTS, history=()):
    return asyncio.run(router.route(message, scripts, list(history)))


def test_questions_carry_only_redacted_text_and_local_script_labels():
    router, client = fake_router(jev_response("script", script="s1"))
    decision = route(router, "ship về 12 đường Lê Lợi, Quận 1 bao nhiêu? sdt 0912 345 678",
                     history=[{"role": "user", "text": "mail mình a.b@gmail.com"},
                              {"role": "assistant", "text": "Dạ shop chào bạn"},
                              {"role": "user", "text": "older message that is not sent"}][::-1])
    sent = json.dumps(client.calls[0], ensure_ascii=False)
    assert decision.route == "script" and decision.script_id == "ver-ship" and decision.router == "jev"
    assert client.calls[0]["model"] == "jev-1.14.0"
    for private in ("0912", "a.b@gmail.com", "Lê Lợi", "ver-ship", "ver-hours", "older message"):
        assert private not in sent
    assert "[PHONE]" in sent and "[ADDRESS]" in sent
    questions = client.calls[0]["questions"]
    assert set(questions) == {"intent", "wants_human", "script", "script_fits"}
    assert set(questions["script"]["criteria"]) == {"s1", "s2"}
    assert len(client.calls[0]["state"]["recent_messages"]) == 2


def test_no_scripts_means_no_script_option():
    router, client = fake_router(jev_response("answer"))
    assert route(router, scripts=[]).route == "answer"
    questions = client.calls[0]["questions"]
    assert "script" not in questions and "script" not in questions["intent"]["criteria"]
    router, _ = fake_router(jev_response("script", script="s1"))
    with pytest.raises(JevUnavailable, match="jev_invalid_output"):
        route(router, scripts=[])


@pytest.mark.parametrize("response,expected,reason", [
    (jev_response("answer", 0.9), "answer", "jev_knowledge_question"),
    (jev_response("clarify", 0.8), "clarify", "jev_needs_clarification"),
    (jev_response("human", 0.8), "human", "jev_needs_staff"),
    (jev_response("answer", 0.9, wants_human=0.95), "human", "jev_customer_requested_human"),
    (jev_response("script", 0.9, script="s2"), "script", "jev_script_match"),
])
def test_confident_routes(response, expected, reason):
    router, _ = fake_router(response)
    decision = route(router)
    assert (decision.route, decision.reason) == (expected, reason)
    if expected == "human":
        assert decision.handoff_reason in ("customer_requested_staff", "business_review")


@pytest.mark.parametrize("response", [
    jev_response("answer", 0.5),  # below AI_JEV_MIN_PROBABILITY
    jev_response("answer", 0.9, wants_human=0.5),  # ambiguous companion noul
    jev_response("script", 0.9, script="s1", script_probability=0.4),
    jev_response("script", 0.9, script="s1", script_fits=0.5),
])
def test_uncertain_decisions_return_none_for_fallback(response):
    router, _ = fake_router(response)
    assert route(router) is None


def test_thresholds_are_configuration():
    router, _ = fake_router(jev_response("answer", 0.5), min_probability=0.4)
    assert route(router).route == "answer"
    with pytest.raises(ValueError):
        JevConfig(noul_yes=0.4, noul_no=0.5)
    with pytest.raises(ValueError):
        JevConfig(model="jev-latest")


@pytest.mark.parametrize("response", [
    jev_response("send_message", 0.9), jev_response("answer", float("nan")), jev_response("answer", 1.7),
    jev_response("script", 0.9, script="s9"),
])
def test_invalid_outputs_are_errors(response):
    router, _ = fake_router(response)
    with pytest.raises(JevUnavailable, match="jev_invalid_output"):
        route(router)


def test_env_flag_defaults_off_and_requires_key(monkeypatch):
    for name in ("AI_JEV_ENABLED", "TYPESAFE_API_KEY", "AI_JEV_MODEL"):
        monkeypatch.delenv(name, raising=False)
    assert jev_router_from_env() is None
    monkeypatch.setenv("AI_JEV_ENABLED", "true")
    assert jev_router_from_env() is None  # no key, stays off
    monkeypatch.setenv("TYPESAFE_API_KEY", "test-only-key")
    monkeypatch.setenv("AI_JEV_MIN_PROBABILITY", "0.75")
    router = jev_router_from_env()
    assert router.config.min_probability == 0.75 and router.model == "jev-1.14.0"
    asyncio.run(router.close())
    monkeypatch.setenv("AI_JEV_NOUL_YES", "secret-invalid")
    with pytest.raises(RuntimeError, match="^Invalid Jev router configuration$"):
        jev_config_from_env()
    with pytest.raises(RuntimeError):
        JevRouter(JevConfig(enabled=False, api_key="x"))


def sdk_body(intent="answer"):
    return {"model": "jev-1.14.0", "usage": {"input_tokens": 50, "output_tokens": 3}, "answers": {
        "intent": {"type": "choice", "choice": intent, "confidence": 0.8,
                   "probabilities": {"answer": 0.85, "clarify": 0.1, "human": 0.05}},
        "wants_human": {"type": "noul", "noul": 0.05}}}


def sdk_router(handler, **config):
    config = JEV.model_copy(update={"max_retries": 2, "timeout": 2, "total_timeout": 5} | config)
    return JevRouter(config, _SdkClient(config, transport=httpx2.MockTransport(handler)))


def test_real_sdk_request_shape_and_response_access():
    requests = []

    def handler(request):
        requests.append(request)
        return httpx2.Response(200, json=sdk_body())

    async def scenario():
        router = sdk_router(handler)
        try:
            return await router.route("áo polo giá bao nhiêu", [], [])
        finally:
            await router.close()
    decision = asyncio.run(scenario())
    assert decision.route == "answer" and decision.probability == 0.85
    request = requests[0]
    assert request.url.path == "/v1/systemone" and request.headers["authorization"] == "Bearer test-only-typesafe-key"
    content = request.content
    if request.headers.get("content-encoding") == "zstd":  # pragma: no cover - only with zstd available
        pytest.skip("compressed body")
    body = json.loads(content)
    assert body["model"] == "jev-1.14.0" and body["questions"]["intent"]["type"] == "choice"
    assert body["questions"]["wants_human"]["type"] == "noul"


@pytest.mark.parametrize("status,code,calls", [
    (429, "jev_rate_limited", 3), (529, "jev_unavailable", 3), (500, "jev_unavailable", 1),
    (401, "jev_auth", 1), (422, "jev_bad_request", 1), (413, "jev_bad_request", 1),
])
def test_sdk_errors_retry_only_rate_limits_and_stay_sanitized(status, code, calls, caplog):
    seen = []

    def handler(request):
        seen.append(request)
        return httpx2.Response(status, json={"error": {"message": "secret body echo"}}, headers={"retry-after-ms": "1"})

    async def scenario():
        router = sdk_router(handler)
        try:
            with pytest.raises(JevUnavailable) as error:
                await router.route("hello", [], [])
            return str(error.value)
        finally:
            await router.close()
    with caplog.at_level(logging.DEBUG):
        assert asyncio.run(scenario()) == code
    assert len(seen) == calls
    assert "secret body echo" not in caplog.text and "test-only-typesafe-key" not in caplog.text


def test_sdk_timeout_and_malformed_response():
    async def slow(request):
        await asyncio.sleep(5)
        return httpx2.Response(200, json=sdk_body())

    async def scenario(handler, **config):
        router = sdk_router(handler, **config)
        try:
            with pytest.raises(JevUnavailable) as error:
                await router.route("hello", [], [])
            return str(error.value)
        finally:
            await router.close()
    assert asyncio.run(scenario(slow, timeout=1, total_timeout=1)) == "jev_timeout"
    assert asyncio.run(scenario(lambda r: httpx2.Response(200, json={"model": "x"}))) == "jev_invalid_output"
