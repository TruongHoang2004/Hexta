"""Offline Vietnamese eval harness: fixture integrity, redaction, replay scoring and the live guard."""

import json

from hexta_ai.routing import evaluate


def test_fixture_is_valid_and_offline_run_passes_without_pii(capsys):
    data = evaluate.load()
    assert len(data["cases"]) >= 25 and {c["expected"] for c in data["cases"]} == {"script", "answer", "clarify", "human"}
    assert evaluate.main([]) == 0
    report = json.loads(capsys.readouterr().out)
    assert report["accuracy"] == 1 and report["pii_leaks"] == []
    assert evaluate.pii_leaks({"cases": [{"id": "x", "pii": ["0912"]}]}, ["phone 0912"]) == ["x: 0912"]


def test_replay_scores_recorded_probabilities_and_sweeps_thresholds(tmp_path, capsys):
    data = evaluate.load()
    recordings = {}
    for case in data["cases"]:
        expected = case["expected"]
        recordings[case["id"]] = {
            "choices": {"intent": {"choice": expected, "probabilities": {expected: 0.65}},
                        "script": {"choice": "s1", "probabilities": {"s1": 0.9}}},
            "nouls": {"wants_human": 0.9 if expected == "human" else 0.1,
                      "script_fits": 0.9 if expected == "script" else 0.1}}
    path = tmp_path / "recordings.json"
    path.write_text(json.dumps(recordings), encoding="utf-8")
    assert evaluate.main(["--replay", str(path), "--sweep"]) == 0
    lines = [json.loads(line) for line in capsys.readouterr().out.splitlines()]
    by_threshold = {line["min_probability"]: line for line in lines}
    # Above the recorded 0.65 probability every non-human case falls back to the deterministic matcher.
    assert by_threshold[0.6]["fallbacks"] < by_threshold[0.7]["fallbacks"]
    assert by_threshold[0.6]["accuracy"] > by_threshold[0.7]["accuracy"]


def test_live_mode_requires_explicit_confirmation_and_key(monkeypatch):
    monkeypatch.delenv("TYPESAFE_API_KEY", raising=False)
    assert evaluate.main(["--live"]) == 2
    assert evaluate.main(["--live", "--confirm-paid-calls"]) == 2
