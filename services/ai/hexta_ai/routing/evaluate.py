"""Offline Vietnamese routing evaluation for the Jev router; live mode needs explicit owner approval.

    uv run python -m hexta_ai.routing.evaluate                       # offline plumbing + redaction check
    uv run python -m hexta_ai.routing.evaluate --replay rec.json     # score recorded Jev answers, sweep thresholds
    uv run python -m hexta_ai.routing.evaluate --live --confirm-paid-calls --record rec.json

Offline mode uses an oracle client (it returns each case's expected route) so it checks request
building, redaction and the decision policy, not Jev accuracy. Recordings contain only case IDs and
probabilities, never message text. Live mode makes one paid System One call per case.
"""

import argparse
import asyncio
import json
import os
import sys
from collections import Counter
from pathlib import Path
from types import SimpleNamespace

from pydantic import SecretStr

from hexta_ai.routing.decision import deterministic_decision
from hexta_ai.routing.jev import JevConfig, JevRouter, JevUnavailable, jev_config_from_env

FIXTURE = Path(__file__).resolve().parents[2] / "evals" / "jev_routing_vi.json"


def load(path: Path = FIXTURE) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    ids = [case["id"] for case in data["cases"]]
    scripts = {s["id"] for s in data["scripts"]}
    if len(set(ids)) != len(ids):
        raise ValueError("Duplicate case IDs")
    for case in data["cases"]:
        if case["expected"] not in ("script", "answer", "clarify", "human"):
            raise ValueError(f"Invalid expected route in {case['id']}")
        if (case["expected"] == "script") != (case.get("script_id") in scripts):
            raise ValueError(f"Script expectation mismatch in {case['id']}")
    return data


def to_response(recorded: dict):
    choices = {name: SimpleNamespace(choice=v["choice"], confidence=v.get("confidence", 0),
                                     probabilities=v["probabilities"]) for name, v in recorded["choices"].items()}
    nouls = {name: SimpleNamespace(noul=v) for name, v in recorded["nouls"].items()}
    return SimpleNamespace(choices=choices, nouls=nouls)


def from_response(response) -> dict:
    return {"choices": {name: {"choice": a.choice, "confidence": a.confidence, "probabilities": dict(a.probabilities)}
                        for name, a in response.choices.items()},
            "nouls": {name: a.noul for name, a in response.nouls.items()}}


class OracleClient:
    """Returns each case's labelled route with a fixed probability; records what would be sent."""

    def __init__(self, labels: dict[str, str], probability: float = 0.9):
        self.case = None
        self.labels = labels
        self.probability = probability
        self.sent = []

    async def system_one(self, state, questions, *, model=None):
        self.sent.append(json.dumps({"state": state, "questions": questions}, ensure_ascii=False))
        expected, script = self.case["expected"], self.case.get("script_id")
        choices = {"intent": {"choice": expected, "probabilities": {expected: self.probability}}}
        nouls = {"wants_human": 0.9 if expected == "human" else 0.05}
        if "script" in questions:
            label = next((k for k, v in self.labels.items() if v == script), next(iter(questions["script"]["criteria"])))
            choices["script"] = {"choice": label, "probabilities": {label: self.probability}}
            nouls["script_fits"] = 0.9 if expected == "script" else 0.1
        return to_response({"choices": choices, "nouls": nouls})

    async def aclose(self):
        pass


class ReplayClient:
    def __init__(self, recordings: dict):
        self.recordings = recordings
        self.case = None

    async def system_one(self, state, questions, *, model=None):
        recorded = self.recordings.get(self.case["id"])
        if recorded is None:
            raise JevUnavailable("jev_error")
        return to_response(recorded)

    async def aclose(self):
        pass


async def evaluate(data: dict, router: JevRouter, client) -> dict:
    scripts = data["scripts"]
    labels = {f"s{i + 1}": s["id"] for i, s in enumerate(scripts)}
    results, recordings = [], {}
    for case in data["cases"]:
        client.case = case
        if isinstance(client, OracleClient):
            client.labels = labels
        error = ""
        try:
            decision = await router.route(case["message"], scripts, case.get("history", []))
        except JevUnavailable as failure:
            decision, error = None, failure.code
        recorded = getattr(client, "last", None)
        if recorded is not None:
            recordings[case["id"]] = recorded
        if decision is None:
            decision = deterministic_decision(case["message"], scripts, error or "jev_low_probability")
        actual = decision.route
        ok = actual == case["expected"] and (actual != "script" or decision.script_id == case.get("script_id"))
        results.append({"id": case["id"], "expected": case["expected"], "actual": actual, "ok": ok,
                        "router": decision.router, "probability": round(decision.probability, 3),
                        "fallback": decision.fallback_reason})
    correct = sum(r["ok"] for r in results)
    confusion = Counter(f"{r['expected']}->{r['actual']}" for r in results if not r["ok"])
    return {"cases": len(results), "correct": correct, "accuracy": round(correct / max(1, len(results)), 3),
            "fallbacks": sum(r["router"] != "jev" for r in results), "errors": dict(confusion),
            "results": results, "recordings": recordings}


def pii_leaks(data: dict, sent: list[str]) -> list[str]:
    leaks = []
    for case, payload in zip(data["cases"], sent):
        leaks += [f"{case['id']}: {value}" for value in case.get("pii", []) if value in payload]
    return leaks


class RecordingClient:
    """Wraps the live SDK adapter and keeps only probabilities for replay."""

    def __init__(self, inner):
        self.inner = inner
        self.case = None
        self.last = None

    async def system_one(self, state, questions, *, model=None):
        self.last = None
        response = await self.inner.system_one(state, questions, model=model)
        self.last = from_response(response)
        return response

    async def aclose(self):
        await self.inner.aclose()


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--fixture", type=Path, default=FIXTURE)
    parser.add_argument("--replay", type=Path)
    parser.add_argument("--live", action="store_true")
    parser.add_argument("--confirm-paid-calls", action="store_true")
    parser.add_argument("--record", type=Path)
    parser.add_argument("--sweep", action="store_true", help="With --replay, score several min-probability values")
    args = parser.parse_args(argv)
    data = load(args.fixture)

    async def run(config: JevConfig, client):
        router = JevRouter(config, client)
        try:
            return await evaluate(data, router, client)
        finally:
            await router.close()

    base = jev_config_from_env().model_copy(update={"enabled": True})
    if not base.api_key.get_secret_value():
        base = base.model_copy(update={"api_key": SecretStr("offline-placeholder")})
    if args.live:
        if not args.confirm_paid_calls or not os.getenv("TYPESAFE_API_KEY", "").strip():
            print("Live evaluation needs --confirm-paid-calls and TYPESAFE_API_KEY (owner approval).", file=sys.stderr)
            return 2
        from hexta_ai.routing.jev import _SdkClient
        report = asyncio.run(run(base, RecordingClient(_SdkClient(base))))
        if args.record:
            args.record.write_text(json.dumps(report["recordings"], indent=2), encoding="utf-8")
    elif args.replay:
        recordings = json.loads(args.replay.read_text(encoding="utf-8"))
        values = [0.4, 0.5, 0.6, 0.7, 0.8, 0.9] if args.sweep else [base.min_probability]
        for value in values:
            report = asyncio.run(run(base.model_copy(update={"min_probability": value}), ReplayClient(recordings)))
            print(json.dumps({"min_probability": value, "accuracy": report["accuracy"],
                              "fallbacks": report["fallbacks"], "errors": report["errors"]}, ensure_ascii=False))
        return 0
    else:
        client = OracleClient({})
        report = asyncio.run(run(base, client))
        leaks = pii_leaks(data, client.sent)
        report["pii_leaks"] = leaks
        if leaks or report["accuracy"] != 1:
            print(json.dumps({k: v for k, v in report.items() if k != "recordings"}, ensure_ascii=False, indent=2))
            return 1
    print(json.dumps({k: v for k, v in report.items() if k not in ("recordings", "results")}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
