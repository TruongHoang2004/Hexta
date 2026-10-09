#!/usr/bin/env python3
"""
send_telegram_report.py - Automated Daily Progress Report Generator & Telegram Dispatcher.

Collects:
  - Git commits in the last 24 hours
  - Git branch & uncommitted status
  - Open/Merged PRs and Issues (via gh CLI if authenticated)
  - Codebase Health & Audit summary (from agentic-memory/audits/)
  - Task queue & lifecycle status

Sends to Telegram using Telegram Bot API (or previews in dry-run mode).
Credentials can be set via:
  - TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID environment variables
  - Or in .env file at repo root (TELEGRAM_BOT_TOKEN=..., TELEGRAM_CHAT_ID=...)
"""

import argparse
from datetime import datetime, timezone, timedelta
import json
import os
from pathlib import Path
import subprocess
import sys
import urllib.parse
import urllib.request

REPO_ROOT = Path(__file__).resolve().parent.parent.parent


def load_env_file():
    """Load TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID from .env if present."""
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


def get_git_commits_24h() -> list[str]:
    """Retrieve commits created in the last 24 hours."""
    try:
        res = subprocess.run(
            ["git", "log", "--since=24 hours ago", "--pretty=format:%h - %s (%an)"],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if res.returncode == 0 and res.stdout.strip():
            return [line.strip() for line in res.stdout.strip().split("\n") if line.strip()]
    except Exception:
        pass
    return []


def get_git_branch() -> str:
    """Get current active git branch."""
    try:
        res = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if res.returncode == 0:
            return res.stdout.strip()
    except Exception:
        pass
    return "unknown"


def get_latest_audit_summary() -> str:
    """Extract executive summary from the latest audit report if available."""
    audits_dir = REPO_ROOT / "agentic-memory" / "audits"
    if not audits_dir.exists():
        return "No audit reports found."

    audit_files = sorted(audits_dir.glob("*_codebase-audit.md"), reverse=True)
    if not audit_files:
        return "No audit files found."

    latest = audit_files[0]
    try:
        content = latest.read_text(encoding="utf-8")
        lines = content.splitlines()
        summary_lines = []
        capture = False
        for line in lines:
            if "## 1. Executive Summary" in line:
                capture = True
                continue
            if capture:
                if line.startswith("## "):
                    break
                if line.strip():
                    summary_lines.append(line.strip())
        if summary_lines:
            return "\n".join(summary_lines[:8])
        return f"Latest audit: {latest.name}"
    except Exception as e:
        return f"Error reading audit: {e}"


def get_github_prs() -> tuple[list[str], list[str]]:
    """Fetch open and recently merged PRs via gh CLI if available."""
    open_prs = []
    merged_prs = []
    try:
        # Check open PRs
        res_open = subprocess.run(
            ["gh", "pr", "list", "--state", "open", "--limit", "5", "--json", "number,title,author"],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if res_open.returncode == 0 and res_open.stdout.strip():
            data = json.loads(res_open.stdout)
            for pr in data:
                open_prs.append(f"#{pr.get('number')} {pr.get('title')} (@{pr.get('author', {}).get('login', '')})")

        # Check merged PRs in last 24h
        res_merged = subprocess.run(
            ["gh", "pr", "list", "--state", "merged", "--limit", "5", "--json", "number,title,mergedAt"],
            cwd=REPO_ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        if res_merged.returncode == 0 and res_merged.stdout.strip():
            data = json.loads(res_merged.stdout)
            cutoff = datetime.now(timezone.utc) - timedelta(hours=24)
            for pr in data:
                merged_at = pr.get("mergedAt")
                if merged_at:
                    try:
                        dt = datetime.fromisoformat(merged_at.replace("Z", "+00:00"))
                        if dt >= cutoff:
                            merged_prs.append(f"#{pr.get('number')} {pr.get('title')}")
                    except Exception:
                        pass
    except Exception:
        pass
    return open_prs, merged_prs


def generate_report_markdown(report_date: str) -> str:
    """Generate comprehensive Markdown report."""
    branch = get_git_branch()
    commits = get_git_commits_24h()
    open_prs, merged_prs = get_github_prs()
    audit_summary = get_latest_audit_summary()

    lines = []
    lines.append(f"📊 *HEXTA DAILY STANDUP DIGEST*")
    lines.append(f"📅 *Date*: `{report_date}` (Asia/Ho\\_Chi\\_Minh)")
    lines.append(f"🌿 *Current Branch*: `{branch}`")
    lines.append("")

    # Commits section
    lines.append(f"🚀 *Commits (Past 24 Hours)*: `{len(commits)} commit(s)`")
    if commits:
        for c in commits[:10]:
            safe_c = c.replace("_", "\\_").replace("*", "\\*").replace("[", "\\[")
            lines.append(f"• `{safe_c}`")
        if len(commits) > 10:
            lines.append(f"• _...và {len(commits) - 10} commit khác_")
    else:
        lines.append("• _Không có commit mới trong 24 giờ qua_")
    lines.append("")

    # PRs section
    lines.append(f"🔀 *Pull Requests*:")
    if merged_prs:
        lines.append(f"*Merged (24h)*: {len(merged_prs)}")
        for pr in merged_prs:
            lines.append(f"  ✅ `{pr}`")
    else:
        lines.append("• Merged: _0 PR trong 24h qua_")

    if open_prs:
        lines.append(f"*Open PRs*: {len(open_prs)}")
        for pr in open_prs:
            lines.append(f"  ⏳ `{pr}`")
    else:
        lines.append("• Open PRs: _Không có PR nào đang chờ review_")
    lines.append("")

    # Health & Audit
    lines.append(f"🛡️ *Codebase Health & Audit*:")
    safe_audit = audit_summary.replace("_", "\\_").replace("*", "\\*").replace("[", "\\[")
    lines.append(f"```\n{safe_audit}\n```")
    lines.append("")

    lines.append("🤖 _Báo cáo được tự động tạo bởi Hexta Autonomous Automation._")
    return "\n".join(lines)


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
            if res_json.get("ok"):
                return True
            print(f"[ERROR] Telegram API response not OK: {res_body}", file=sys.stderr)
            return False
    except Exception as e:
        print(f"[ERROR] Failed to send Telegram message: {e}", file=sys.stderr)
        return False


def main():
    parser = argparse.ArgumentParser(description="Hexta Daily Standup Telegram Reporter")
    parser.add_argument("--dry-run", action="store_true", help="Print digest without sending to Telegram")
    parser.add_argument("--output-file", help="Optional path to save generated report")
    args = parser.parse_args()

    load_env_file()
    token = os.environ.get("TELEGRAM_BOT_TOKEN", "").strip()
    chat_id = os.environ.get("TELEGRAM_CHAT_ID", "").strip()

    now_vn = datetime.now(timezone(timedelta(hours=7)))
    date_str = now_vn.strftime("%Y-%m-%d %H:%M:%S")

    report_md = generate_report_markdown(date_str)

    if args.output_file:
        out_p = Path(args.output_file)
        out_p.parent.mkdir(parents=True, exist_ok=True)
        out_p.write_text(report_md, encoding="utf-8")
        print(f"Report saved to {out_p}")

    print("\n--- [GENERATED REPORT PREVIEW] ---")
    print(report_md)
    print("----------------------------------\n")

    if args.dry_run:
        print("[DRY-RUN] Message preview displayed above. Not sending to Telegram.")
        return

    if not token or not chat_id:
        print(
            "[WARN] TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not configured in environment or .env!\n"
            "Please configure them to enable automated Telegram dispatch:\n"
            "  1. Chat with @BotFather to create a bot and get TELEGRAM_BOT_TOKEN\n"
            "  2. Chat with @userinfobot to get your TELEGRAM_CHAT_ID\n"
            "  3. Add them to .env at repository root:\n"
            "     TELEGRAM_BOT_TOKEN=<your_token>\n"
            "     TELEGRAM_CHAT_ID=<your_chat_id>",
            file=sys.stderr,
        )
        sys.exit(1)

    print(f"Sending daily report to Telegram chat_id: {chat_id}...")
    success = send_to_telegram(token, chat_id, report_md)
    if success:
        print("✅ Daily report sent successfully to Telegram!")
    else:
        sys.exit(1)


if __name__ == "__main__":
    main()
