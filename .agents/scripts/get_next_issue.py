#!/usr/bin/env python3
"""
get_next_issue.py - Atomic Task Selector & Claimer for Hexta Multi-Agent Swarm.

Capabilities:
  --domain <name>: Filter issues tagged with domain:<name> (e.g. backend-core, frontend-ui).
  --claim-by <agent_id>: Atomically claims the selected issue by removing 'ready' and
                         adding 'in-progress' and 'claimed-by:<agent_id>'.
  --repo <owner/repo>: Explicit repository slug (defaults to TruongHoang2004/Hexta if needed).
  --dry-run: Simulates task selection and claim without mutating GitHub state.
"""

import argparse
import json
import logging
import re
import subprocess
import sys
from typing import Any, Dict, List, Optional, Tuple

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("get_next_issue")


def slugify(text: str) -> str:
    text = text.lower()
    text = re.sub(r'[^a-z0-9]+', '-', text).strip('-')
    return text[:40]


def get_priority_weight(labels: List[Dict[str, Any]]) -> int:
    label_names = [l.get("name", "").lower() for l in labels]

    # Exclude issues already in-progress, in-review, or blocked
    if any(name in ["in-progress", "in progress", "wip", "blocked", "in-review", "review", "pr-opened"] for name in label_names):
        return -1

    # Issue MUST have 'ready' label to be eligible for pickup
    if not any(name in ["ready", "status:ready"] for name in label_names):
        return -1

    if any(name in ["priority:high", "high", "priority-high"] for name in label_names):
        return 1
    if any(name in ["priority:medium", "medium", "priority-medium"] for name in label_names):
        return 2
    if any(name in ["priority:low", "low", "priority-low"] for name in label_names):
        return 3
    return 4  # Default / no priority label


def matches_domain(labels: List[Dict[str, Any]], target_domain: Optional[str]) -> bool:
    if not target_domain or target_domain.lower() in ("all", "any"):
        return True

    target = target_domain.lower().replace("domain:", "").strip()
    label_names = [l.get("name", "").lower() for l in labels]

    domain_labels = [name.replace("domain:", "").strip() for name in label_names if name.startswith("domain:")]
    if not domain_labels:
        # Untagged domain tasks can be picked if target_domain is None or 'general'
        return target in ("general", "untagged")

    return any(d == target for d in domain_labels)


def get_clean_env() -> Dict[str, str]:
    import os
    return {k: v for k, v in os.environ.items() if not k.lower().endswith('_proxy')}


def fetch_open_issues(repo: Optional[str] = "TruongHoang2004/Hexta") -> List[Dict[str, Any]]:
    cmd = ["gh", "issue", "list", "--state", "open", "--json", "number,title,body,labels,createdAt", "--limit", "100"]
    if repo:
        cmd.extend(["--repo", repo])

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True, env=get_clean_env())
        return json.loads(result.stdout or "[]")
    except subprocess.CalledProcessError as e:
        logger.error("Failed to run gh command: %s", e.stderr)
        raise RuntimeError(f"GitHub CLI command failed: {e.stderr.strip()}") from e
    except FileNotFoundError:
        logger.error("gh CLI not found on PATH.")
        raise RuntimeError("gh CLI not found on PATH.")


def claim_issue(issue_number: int, agent_id: str, repo: Optional[str] = "TruongHoang2004/Hexta", dry_run: bool = False) -> bool:
    cmd = [
        "gh", "issue", "edit", str(issue_number),
        "--remove-label", "ready",
        "--add-label", f"in-progress,claimed-by:{agent_id}"
    ]
    if repo:
        cmd.extend(["--repo", repo])

    if dry_run:
        logger.info("[DRY-RUN] Would claim issue #%d for agent '%s'", issue_number, agent_id)
        return True

    try:
        subprocess.run(cmd, capture_output=True, text=True, check=True, env=get_clean_env())
        logger.info("Successfully claimed issue #%d for agent '%s'", issue_number, agent_id)
        return True
    except subprocess.CalledProcessError as e:
        logger.error("Failed to claim issue #%d: %s", issue_number, e.stderr)
        return False


def main():
    parser = argparse.ArgumentParser(description="Fetch and optionally claim the next eligible issue.")
    parser.add_argument("--domain", help="Target domain (e.g. backend-core, frontend-ui, qa)")
    parser.add_argument("--claim-by", help="Agent identifier to atomically claim the task")
    parser.add_argument("--repo", default="TruongHoang2004/Hexta", help="GitHub repository slug (e.g. TruongHoang2004/Hexta)")
    parser.add_argument("--dry-run", action="store_true", help="Simulate without claiming")
    args = parser.parse_args()

    try:
        issues = fetch_open_issues(repo=args.repo)
    except Exception as e:
        print(json.dumps({"status": "error", "message": str(e)}), file=sys.stderr)
        sys.exit(1)

    if not issues:
        print(json.dumps({"status": "no_tasks", "message": "No open issues found in repository."}))
        sys.exit(0)

    # Filter by readiness and domain
    eligible: List[Tuple[int, int, Dict[str, Any]]] = []
    for issue in issues:
        weight = get_priority_weight(issue.get("labels", []))
        if weight > 0 and matches_domain(issue.get("labels", []), args.domain):
            eligible.append((weight, issue.get("number", 0), issue))

    if not eligible:
        msg = f"No ready tasks available for domain '{args.domain}'." if args.domain else "All open issues are either in-progress, in-review, or blocked."
        print(json.dumps({"status": "no_tasks", "message": msg}))
        sys.exit(0)

    # Sort by priority ascending, then by number ascending (oldest first)
    eligible.sort(key=lambda item: (item[0], item[1]))
    selected = eligible[0][2]
    branch_name = f"task/issue-{selected['number']}-{slugify(selected['title'])}"

    claimed = False
    if args.claim_by:
        claimed = claim_issue(selected["number"], args.claim_by, repo=args.repo, dry_run=args.dry_run)

    output = {
        "status": "claimed" if (args.claim_by and claimed) else "ready",
        "agent": args.claim_by,
        "domain": args.domain,
        "issue": {
            "number": selected["number"],
            "title": selected["title"],
            "body": selected.get("body", ""),
            "labels": [l.get("name") for l in selected.get("labels", [])],
            "branch_name": branch_name
        }
    }
    print(json.dumps(output, indent=2))


if __name__ == "__main__":
    main()
