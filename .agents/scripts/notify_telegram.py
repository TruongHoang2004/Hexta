#!/usr/bin/env python3
"""
notify_telegram.py - Dispatch task pickup and completion notifications to Telegram.

Usage:
  python3 .agents/scripts/notify_telegram.py --pickup --issue <num> --title "<title>" --branch "<branch>"
  python3 .agents/scripts/notify_telegram.py --complete --issue <num> --title "<title>" --pr <pr_num> --pr-url "<url>"
  python3 .agents/scripts/notify_telegram.py --message "<custom_text>"
"""

import argparse
from datetime import datetime, timezone, timedelta
import json
import os
from pathlib import Path
import sys
import urllib.request

REPO_ROOT = Path(__file__).resolve().parent.parent.parent


def load_env_file():
    """Load TELEGRAM credentials from .env."""
    env_file = REPO_ROOT / ".env"
    if env_file.exists():
        with open(env_file, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith("#") and "=" in line:
                    k, v = line.split("=", 1)
                    k = k.strip()
                    v = v.strip().strip("'\"")
                    if k not in os.environ:
                        os.environ[k] = v


def send_to_telegram(token: str, chat_id: str, text: str) -> bool:
    """Send message to Telegram via Bot API."""
    url = f"https://api.telegram.org/bot{token}/sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": text,
        "parse_mode": "Markdown",
        "disable_web_page_preview": True,
    }
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(
        url,
        data=data,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            res_body = resp.read().decode("utf-8")
            res_json = json.loads(res_body)
            return bool(res_json.get("ok"))
    except Exception as e:
        print(f"[ERROR] Failed to send Telegram notification: {e}", file=sys.stderr)
        return False


def main():
    parser = argparse.ArgumentParser(description="Dispatch task notification to Telegram")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--pickup", action="store_true", help="Task pickup notification")
    group.add_argument("--complete", action="store_true", help="Task completion notification")
    group.add_argument("--message", help="Custom markdown message to dispatch")

    parser.add_argument("--issue", type=int, help="Issue number")
    parser.add_argument("--title", help="Issue title")
    parser.add_argument("--branch", help="Git branch name")
    parser.add_argument("--pr", type=int, help="Pull Request number")
    parser.add_argument("--pr-url", help="Pull Request URL")
    parser.add_argument("--dry-run", action="store_true", help="Preview message without sending")

    args = parser.parse_args()

    load_env_file()
    token = os.environ.get("TELEGRAM_BOT_TOKEN", "").strip()
    chat_id = os.environ.get("TELEGRAM_CHAT_ID", "").strip()

    now_vn = datetime.now(timezone(timedelta(hours=7)))
    date_str = now_vn.strftime("%Y-%m-%d %H:%M:%S")

    if args.message:
        text = args.message
    elif args.pickup:
        safe_title = (args.title or "Unknown").replace("_", "\\_").replace("*", "\\*")
        lines = [
            "🚀 *TASK PICKED UP*",
            f"⏰ *Time*: `{date_str}` (Asia/Ho\\_Chi\\_Minh)",
            f"📌 *Issue*: [#{args.issue}](https://github.com/TruongHoang2004/Hexta/issues/{args.issue}) - {safe_title}",
        ]
        if args.branch:
            lines.append(f"🌿 *Branch*: `{args.branch}`")
        lines.append("⚡ *Status*: Transitioned to `in-progress`. Autonomous development started.")
        text = "\n".join(lines)
    elif args.complete:
        safe_title = (args.title or "Unknown").replace("_", "\\_").replace("*", "\\*")
        lines = [
            "✅ *TASK COMPLETED & PR OPENED*",
            f"⏰ *Time*: `{date_str}` (Asia/Ho\\_Chi\\_Minh)",
            f"📌 *Issue*: [#{args.issue}](https://github.com/TruongHoang2004/Hexta/issues/{args.issue}) - {safe_title}",
        ]
        if args.pr:
            pr_link = args.pr_url or f"https://github.com/TruongHoang2004/Hexta/pull/{args.pr}"
            lines.append(f"🔀 *Pull Request*: [#{args.pr}]({pr_link})")
        lines.append("📝 *Status*: Transitioned to `in-review`. Memory artifacts recorded.")
        text = "\n".join(lines)
    else:
        print("[ERROR] No valid notification mode specified.", file=sys.stderr)
        sys.exit(1)

    print("\n--- [TELEGRAM NOTIFICATION] ---")
    print(text)
    print("-------------------------------\n")

    if args.dry_run:
        print("[DRY-RUN] Preview shown above. Not sent.")
        return

    if not token or not chat_id:
        print("[WARN] Telegram credentials missing. Skipping dispatch.", file=sys.stderr)
        return

    success = send_to_telegram(token, chat_id, text)
    if success:
        print("✅ Telegram notification sent successfully!")
    else:
        sys.exit(1)


if __name__ == "__main__":
    main()
