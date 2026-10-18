"""Exercise the running simulation service without model or channel credentials."""

import json
import os
import urllib.error
import urllib.request


def main() -> None:
    base = os.environ.get("DEMO_AI_URL", "http://127.0.0.1:58000").rstrip("/")
    token = os.environ["DEMO_AI_INTERNAL_TOKEN"]
    with urllib.request.urlopen(f"{base}/health/ready", timeout=10) as response:
        assert json.load(response)["simulated"] is True
    body = json.dumps({"message": "price", "scripts": [
        {"id": "sample", "trigger": "price", "text": "Select a product.", "active": True},
    ]}).encode()
    request = urllib.request.Request(
        f"{base}/internal/v1/preview", data=body,
        headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"},
    )
    with urllib.request.urlopen(request, timeout=10) as response:
        result = json.load(response)
        assert result["simulated"] is True and result["route"] == "script"
        assert result["script_id"] == "sample"
    request.remove_header("Authorization")
    try:
        urllib.request.urlopen(request, timeout=10)
    except urllib.error.HTTPError as error:
        assert error.code == 401
    else:
        raise AssertionError("Unauthenticated preview unexpectedly accepted")
    print("PASS: simulated readiness, authenticated graph preview, unauthorized rejection")


if __name__ == "__main__":
    main()
