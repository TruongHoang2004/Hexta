"""Exercise authenticated preview; real model calls require explicit --live."""

import argparse
import json
import os
import urllib.error
import urllib.request

import hexta_ai  # noqa: F401 - loads service-local .env


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--live", action="store_true", help="allow up to two paid calls with synthetic domain facts")
    args = parser.parse_args()
    base = os.environ.get("DEMO_AI_URL", "http://127.0.0.1:58000").rstrip("/")
    token = os.environ.get("AI_INTERNAL_TOKEN") or os.environ.get("DEMO_AI_INTERNAL_TOKEN", "")
    with urllib.request.urlopen(f"{base}/health/ready", timeout=10) as response:
        health = json.load(response)
    if not health["simulated"] and not args.live:
        raise SystemExit("Real provider active; use --live explicitly to allow model calls.")
    if health["simulated"] and args.live:
        raise SystemExit("Live smoke requires an active real provider; check its private key configuration.")
    body = ({"message": "What are the store opening hours?", "allow_generation": True,
             "facts": [{"id": "hours", "text": "The store is open Monday to Friday, 9 AM to 5 PM."}]}
            if args.live else {"message": "price", "scripts": [
                {"id": "sample", "trigger": "price", "text": "Select a product.", "active": True}]})
    request = urllib.request.Request(
        f"{base}/internal/v1/preview", data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"},
    )
    with urllib.request.urlopen(request, timeout=70) as response:
        result = json.load(response)
    assert result["simulated"] == health["simulated"]
    assert result["route"] in ("script", "generate", "clarify", "human")
    if not args.live:
        assert result["route"] == "script" and result["script_id"] == "sample"
    request.remove_header("Authorization")
    try:
        urllib.request.urlopen(request, timeout=10)
    except urllib.error.HTTPError as error:
        assert error.code == 401
    else:
        raise AssertionError("Unauthenticated preview unexpectedly accepted")
    print("PASS: readiness, authenticated candidate routing, unauthorized rejection; no customer send")


if __name__ == "__main__":
    try:
        main()
    except urllib.error.HTTPError as error:
        raise SystemExit(f"AI smoke failed with HTTP {error.code}; inspect safe runtime status.") from None
    except urllib.error.URLError:
        raise SystemExit("AI smoke failed: service unavailable.") from None
