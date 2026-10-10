#!/usr/bin/env python3
"""
send_telegram_alerts.py - Automated Critical Alert Monitor & Telegram Dispatcher.

Checks hourly for:
  1. High-Risk Pull Requests (marked needs-human-review or high blast radius)
  2. Stale in-progress task locks (>2h inactive without an open PR)
  3. High-severity codebase audit findings (security leaks or architectural breaches)

Deduplicates alerts using .agents/.alert_history.json to avoid spamming.
Dispatches directly to Telegram via Bot API using TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID.
"""

import argparse
from datetime import datetime, timezone, timedelta
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import urllib.parse
import urllib.request

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
HISTORY_FILE = REPO_ROOT / ".agents" / ".alert_history.json"


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


def load_history() -> dict:
    if HISTORY_FILE.exists():
        try:
            return json.loads(HISTORY_FILE.read_text(encoding="utf-8"))
        except Exception:
            return {}
    return {}


def save_history(history: dict):
    HISTORY_FILE.parent.mkdir(parents=True, exist_ok=True)
    HISTORY_FILE.write_text(json.dumps(history, indent=2), encoding="utf-8")


def check_stale_tasks() -> list[str]:
    """Check for tasks locked in-progress >2h without an active PR."""
    stale_alerts = []
    cmd = ["python3", ".agents/scripts/manage_issue_lifecycle.py", "--sync", "--recover-stale", "--dry-run"]
    try:
        res = subprocess.run(cmd, cwd=REPO_ROOT, capture_output=True, text=True, check=False)
        output = res.stdout + res.stderr
        for line in output.splitlines():
            if "RECOVER" in line or "Stale lock" in line or "Recovering stale" in line:
                stale_alerts.append(line.strip())
    except Exception:
        pass
    return stale_alerts


def check_high_risk_prs() -> list[str]:
    """Check for open PRs with high risk or needs-human-review label."""
    pr_alerts = []
    try:
        res = subprocess.run(
            ["gh", "pr", "list", "--state", "open", "--json", "number,title,labels,headRefName"],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if res.returncode == 0 and res.stdout.strip():
            prs = json.loads(res.stdout)
            for pr in prs:
                label_names = [lbl.get("name", "") for lbl in pr.get("labels", [])]
                if "needs-human-review" in label_names or "risk:high" in label_names:
                    pr_alerts.append(f"PR #{pr.get('number')} '{pr.get('title')}' yêu cầu người duyệt (needs-human-review)")
    except Exception:
        pass
    return pr_alerts


def check_audit_high_severity() -> list[str]:
    """Check latest audit report for High Severity issues."""
    audit_alerts = []
    audits_dir = REPO_ROOT / "agentic-memory" / "audits"
    if not audits_dir.exists():
        return []

    audit_files = sorted(audits_dir.glob("*_codebase-audit.md"), reverse=True)
    if not audit_files:
        return []

    latest = audit_files[0]
    try:
        content = latest.read_text(encoding="utf-8")
        if "**High Severity**: 0" not in content and "High Severity" in content:
            lines = content.splitlines()
            capture = False
            for line in lines:
                if "**HIGH**" in line or "| HIGH |" in line:
                    clean = line.replace("|", " ").strip()
                    audit_alerts.append(f"Phát hiện lỗi nghiêm trọng: {clean[:120]}")
    except Exception:
        pass
    return audit_alerts


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
        print(f"[ERROR] Failed to send Telegram alert: {e}", file=sys.stderr)
        return False


def main():
    parser = argparse.ArgumentParser(description="Hexta Hourly Telegram Alert Monitor")
    parser.add_argument("--dry-run", action="store_true", help="Print alerts without sending")
    parser.add_argument("--force", "-f", action="store_true", help="Bypass throttling and send alert immediately")
    parser.add_argument("--force-test", action="store_true", help="Send a test alert immediately")
    args = parser.parse_args()

    load_env_file()
    token = os.environ.get("TELEGRAM_BOT_TOKEN", "").strip()
    chat_id = os.environ.get("TELEGRAM_CHAT_ID", "").strip()

    now_vn = datetime.now(timezone(timedelta(hours=7)))
    date_str = now_vn.strftime("%Y-%m-%d %H:%M:%S")

    stale_tasks = check_stale_tasks()
    high_risk_prs = check_high_risk_prs()
    audit_findings = check_audit_high_severity()

    all_alerts = []
    if stale_tasks:
        all_alerts.append("⏳ *Task Bị Treo (>2 giờ)*:")
        for st in stale_tasks:
            all_alerts.append(f"  • {st}")
    if high_risk_prs:
        all_alerts.append("⚠️ *Pull Requests Cần Duyệt Gấp (High-Risk)*:")
        for pr in high_risk_prs:
            all_alerts.append(f"  • {pr}")
    if audit_findings:
        all_alerts.append("🛡️ *Cảnh Báo Bảo Mật / Kiến Trúc Nghiêm Trọng*:")
        for af in audit_findings:
            all_alerts.append(f"  • {af}")

    force_send = args.force or args.force_test
    if args.force_test and not all_alerts:
        all_alerts.append("🔔 *Test Thông Báo Cảnh Báo*: Kênh cảnh báo Telegram đang hoạt động bình thường và sẵn sàng giám sát 24/7.")

    if not all_alerts:
        print(f"[{date_str}] OK: Hệ thống ổn định. Không có cảnh báo khẩn cấp.")
        return

    # Check deduplication
    alert_text = "\n".join(all_alerts)
    alert_hash = hashlib.sha256(alert_text.encode("utf-8")).hexdigest()
    history = load_history()
    last_sent = history.get(alert_hash)

    # Throttle: don't resend identical alert if sent within last 6 hours (unless force or dry-run)
    if not force_send and not args.dry_run and last_sent:
        try:
            last_dt = datetime.fromisoformat(last_sent)
            if (now_vn - last_dt) < timedelta(hours=6):
                print(f"[{date_str}] Alert already sent recently ({last_sent}). Throttling duplicate.")
                return
        except Exception:
            pass

    message_lines = [
        "🚨 *HEXTA URGENT GUARDIAN ALERT*",
        f"⏰ *Thời gian*: `{date_str}` (Asia/Ho\\_Chi\\_Minh)",
        "",
        alert_text,
        "",
        "👉 _Vui lòng kiểm tra repository hoặc can thiệp khi cần thiết._"
    ]
    full_message = "\n".join(message_lines)

    print("\n--- [TELEGRAM ALERT MESSAGE] ---")
    print(full_message)
    print("--------------------------------\n")

    if args.dry_run:
        print("[DRY-RUN] Alert preview displayed above. Not sending to Telegram.")
        return

    if not token or not chat_id:
        print("[WARN] Telegram credentials missing. Skipping dispatch.", file=sys.stderr)
        return

    print("Dispatching alert to Telegram...")
    success = send_to_telegram(token, chat_id, full_message)
    if success:
        print("✅ Telegram alert sent successfully!")
        history[alert_hash] = now_vn.isoformat()
        save_history(history)
    else:
        sys.exit(1)


if __name__ == "__main__":
    main()
