#!/usr/bin/env python3
"""
notify_telegram.py - Dispatch task notifications to Telegram via Bot API.

Usage:
    python3 .agents/scripts/notify_telegram.py --pickup --issue-id 30 --title "Task title"
    python3 .agents/scripts/notify_telegram.py --complete --issue-id 30 --title "Task title" --pr-url "https://github.com/..."
    python3 .agents/scripts/notify_telegram.py --message "Custom text message"
"""

import argparse
from datetime import datetime, timezone, timedelta
import json
import os
from pathlib import Path
import sys
import urllib.parse
import urllib.request

REPO_ROOT = Path(__file__).resolve().parent.parent.parent


def load_env_file():
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
            if res_json.get("ok"):
                return True
            print(f"[ERROR] Telegram API error: {res_body}", file=sys.stderr)
            return False
    except Exception as e:
        print(f"[ERROR] Failed to send Telegram message: {e}", file=sys.stderr)
        return False


def main():
    parser = argparse.ArgumentParser(description="Dispatch task notification to Telegram")
    parser.add_argument("--pickup", action="store_true", help="Notify task pickup")
    parser.add_argument("--complete", action="store_true", help="Notify task completion")
    parser.add_argument("--issue-id", type=str, help="GitHub Issue ID")
    parser.add_argument("--title", type=str, help="Issue Title")
    parser.add_argument("--pr-url", type=str, help="Pull Request URL")
    parser.add_argument("--message", type=str, help="Custom message")
    parser.add_argument("--dry-run", action="store_true", help="Preview without sending")
    args = parser.parse_args()

    load_env_file()
    token = os.environ.get("TELEGRAM_BOT_TOKEN", "").strip()
    chat_id = os.environ.get("TELEGRAM_CHAT_ID", "").strip()

    now_vn = datetime.now(timezone(timedelta(hours=7)))
    date_str = now_vn.strftime("%Y-%m-%d %H:%M:%S")

    if args.message:
        text = args.message
    elif args.pickup:
        issue_id = args.issue_id or "N/A"
        title = (args.title or "").replace("_", "\\_").replace("*", "\\*")
        text = (
            f"🚀 *TASK PICKED UP*\n"
            f"⏰ *Time*: `{date_str}` (Asia/Ho\\_Chi\\_Minh)\n"
            f"📋 *Issue*: #{issue_id} - `{title}`\n"
            f"🔄 *Status*: Transitioned to `in-progress`\n"
            f"🛠️ *Branch*: Checked out. Implementation started."
        )
    elif args.complete:
        issue_id = args.issue_id or "N/A"
        title = (args.title or "").replace("_", "\\_").replace("*", "\\*")
        pr_link = args.pr_url or "Opened"
        text = (
            f"✅ *TASK COMPLETED*\n"
            f"⏰ *Time*: `{date_str}` (Asia/Ho\\_Chi\\_Minh)\n"
            f"📋 *Issue*: #{issue_id} - `{title}`\n"
            f"🔀 *Pull Request*: {pr_link}\n"
            f"📝 *Review*: Verified with `go test -race`, memory artifacts recorded.\n"
            f"🔄 *Status*: Transitioned to `in-review`"
        )
    else:
        parser.print_help()
        sys.exit(1)

    print("\n--- Telegram Notification Preview ---")
    print(text)
    print("-------------------------------------\n")

    if args.dry_run:
        print("[DRY-RUN] Notification not dispatched.")
        return

    if not token or not chat_id:
        print("[WARN] TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not found in environment or .env. Skipping dispatch.")
        return

    success = send_to_telegram(token, chat_id, text)
    if success:
        print("✅ Telegram notification dispatched successfully.")
    else:
        print("[WARN] Could not dispatch to Telegram.")


if __name__ == "__main__":
    main()
