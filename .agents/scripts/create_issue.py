#!/usr/bin/env python3
"""
create_issue.py - Standardized GitHub Issue Creation Tool for Hexta Monorepo.

Ensures all created issues strictly conform to the repository's lifecycle standards,
formatting specifications, and autonomous task runner requirements.
"""

import argparse
from dataclasses import dataclass, field
import json
import logging
import os
import re
import subprocess
import sys
from typing import Any, Dict, List, Optional, Set, Tuple

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("create_issue")

# Allowed types and their GitHub label mappings
TYPE_MAPPING = {
    "feat": "enhancement",
    "feature": "enhancement",
    "enhancement": "enhancement",
    "bug": "bug",
    "fix": "bug",
    "refactor": "enhancement",
    "docs": "documentation",
    "documentation": "documentation",
    "chore": "enhancement",
}

VALID_PRIORITIES = {"high", "medium", "low"}
PRIORITY_LABELS = {
    "high": "priority:high",
    "medium": "priority:medium",
    "low": "priority:low",
}


@dataclass
class IssueDraft:
    title: str
    body: str
    labels: List[str] = field(default_factory=list)

    def validate(self) -> Tuple[bool, List[str]]:
        """Validates that draft adheres to Hexta triage & specification rules."""
        errors = []
        clean_title = self.title.strip()
        clean_body = self.body.strip()

        if len(clean_title) < 10:
            errors.append(f"Title length ({len(clean_title)}) is too short (minimum 10 chars)")

        if len(clean_body) < 50:
            errors.append(f"Body length ({len(clean_body)}) is too short (minimum 50 chars)")

        # Acceptance criteria checkboxes verification
        if not re.search(r"-\s*\[[ xX]\]", clean_body):
            errors.append("Body must contain acceptance criteria checklist with at least one checkbox '- [ ]'")

        # Structural sections check
        if not re.search(r"^#{1,3}\s+", clean_body, re.MULTILINE):
            errors.append("Body must include Markdown section headings (e.g. ## Summary, ## Acceptance Criteria)")

        return len(errors) == 0, errors


def build_issue_body(
    summary: str,
    requirements: List[str],
    acceptance_criteria: List[str],
    technical_notes: Optional[str] = None,
    related_files: Optional[List[str]] = None,
) -> str:
    """Builds standard structured markdown for Hexta GitHub issues."""
    sections = []

    # 1. Summary / Context
    sections.append("## Summary & Context")
    sections.append(summary.strip())

    # 2. Requirements & Scope
    if requirements:
        sections.append("\n## Requirements & Scope")
        for req in requirements:
            clean_req = req.strip()
            if not clean_req.startswith("-"):
                clean_req = f"- {clean_req}"
            sections.append(clean_req)

    # 3. Technical Notes & Guidelines
    if technical_notes or related_files:
        sections.append("\n## Technical Details & Architecture")
        if technical_notes:
            sections.append(technical_notes.strip())
        if related_files:
            sections.append("\n**Key Affected Files / Packages:**")
            for f in related_files:
                sections.append(f"- `{f.strip()}`")

    # 4. Acceptance Criteria (Required checklist)
    sections.append("\n## Acceptance Criteria")
    if acceptance_criteria:
        for ac in acceptance_criteria:
            clean_ac = ac.strip()
            if not clean_ac.startswith("- [ ]") and not clean_ac.startswith("- [x]"):
                if clean_ac.startswith("-"):
                    clean_ac = f"- [ ] {clean_ac.lstrip('- ').strip()}"
                else:
                    clean_ac = f"- [ ] {clean_ac}"
            sections.append(clean_ac)
    else:
        sections.append("- [ ] Implementation completed and verified according to requirements.")
        sections.append("- [ ] Automated unit/integration tests pass.")
        sections.append("- [ ] Code conforms to `GEMINI.md` and repository standards.")

    # 5. Autonomous Runner Workflow Note
    sections.append("\n## Automated Workflow Checklist")
    sections.append("- [ ] `github-task-runner` picks up task from `ready` queue")
    sections.append("- [ ] Implementation plan recorded in `agentic-memory/plans/`")
    sections.append("- [ ] Code review audit recorded in `agentic-memory/reviews/`")
    sections.append("- [ ] Changelog recorded in `agentic-memory/changelogs/`")

    return "\n".join(sections)


def resolve_labels(
    issue_type: str,
    priority: str,
    mark_ready: bool = True,
    extra_labels: Optional[List[str]] = None,
) -> List[str]:
    """Resolves and dedupes label set for the issue."""
    labels: Set[str] = set()

    # Lifecycle status
    if mark_ready:
        labels.add("ready")

    # Priority
    p_norm = priority.lower().strip()
    if p_norm in PRIORITY_LABELS:
        labels.add(PRIORITY_LABELS[p_norm])
    else:
        labels.add("priority:medium")

    # Type mapping
    t_norm = issue_type.lower().strip()
    if t_norm in TYPE_MAPPING:
        labels.add(TYPE_MAPPING[t_norm])

    if extra_labels:
        for el in extra_labels:
            clean_el = el.strip()
            if clean_el:
                labels.add(clean_el)

    return sorted(list(labels))


def execute_gh_create(draft: IssueDraft, repo: Optional[str] = "TruongHoang2004/Hexta", dry_run: bool = False) -> Dict[str, Any]:
    """Calls `gh issue create` or simulates during dry-run."""
    valid, errors = draft.validate()
    if not valid:
        err_msg = "Issue draft failed validation:\n" + "\n".join(f"  - {e}" for e in errors)
        logger.error(err_msg)
        raise ValueError(err_msg)

    cmd = ["gh", "issue", "create", "--title", draft.title, "--body", draft.body]
    if repo:
        cmd.extend(["--repo", repo])
    for lbl in draft.labels:
        cmd.extend(["--label", lbl])

    if dry_run:
        logger.info("[DRY RUN] Would execute command: %s", " ".join(cmd[:6]) + f" ... ({len(draft.labels)} labels)")
        return {
            "status": "dry_run",
            "title": draft.title,
            "labels": draft.labels,
            "body_length": len(draft.body),
            "command": cmd,
        }

    logger.info("Executing gh issue create: '%s' with labels %s", draft.title, draft.labels)
    clean_env = {k: v for k, v in os.environ.items() if not k.lower().endswith('_proxy')}
    try:
        proc = subprocess.run(cmd, capture_output=True, text=True, check=True, env=clean_env)
        output = proc.stdout.strip()
        logger.info("GitHub Issue created successfully: %s", output)

        issue_num = None
        match = re.search(r"/issues/(\d+)", output)
        if match:
            issue_num = int(match.group(1))

        return {
            "status": "created",
            "url": output,
            "issue_number": issue_num,
            "title": draft.title,
            "labels": draft.labels,
        }
    except subprocess.CalledProcessError as exc:
        logger.error("Failed to create GitHub issue: %s", exc.stderr.strip())
        raise RuntimeError(f"gh issue create failed: {exc.stderr.strip()}") from exc


def parse_args():
    parser = argparse.ArgumentParser(
        description="Create standard GitHub issues for Hexta compatible with autonomous task runners"
    )
    parser.add_argument("--title", required=False, help="Issue title (format: 'type(scope): summary')")
    parser.add_argument(
        "--type",
        choices=["feat", "fix", "refactor", "docs", "chore", "enhancement", "bug", "documentation"],
        default="feat",
        help="Type of issue (default: feat)",
    )
    parser.add_argument(
        "--priority",
        choices=["high", "medium", "low"],
        default="medium",
        help="Issue priority (default: medium)",
    )
    parser.add_argument("--summary", help="Short summary / problem description")
    parser.add_argument(
        "--requirement",
        action="append",
        dest="requirements",
        default=[],
        help="Requirement bullet point (can be specified multiple times)",
    )
    parser.add_argument(
        "--criterion",
        action="append",
        dest="criteria",
        default=[],
        help="Acceptance criterion checkbox (can be specified multiple times)",
    )
    parser.add_argument("--notes", help="Technical architecture notes or guidelines")
    parser.add_argument(
        "--file",
        action="append",
        dest="files",
        default=[],
        help="Relevant file path (can be specified multiple times)",
    )
    parser.add_argument("--body", help="Full custom Markdown body (overrides auto-generated body)")
    parser.add_argument("--body-file", help="Path to markdown file containing issue body")
    parser.add_argument(
        "--label",
        action="append",
        dest="extra_labels",
        default=[],
        help="Extra label (can be specified multiple times)",
    )
    parser.add_argument(
        "--no-ready",
        action="store_true",
        help="Do NOT add 'ready' label immediately (keep in backlog for manual triage)",
    )
    parser.add_argument("--repo", default="TruongHoang2004/Hexta", help="Target GitHub repository")
    parser.add_argument("--dry-run", action="store_true", help="Preview output without creating issue")
    parser.add_argument("--json", action="store_true", help="Print JSON output format")
    return parser.parse_args()


def main():
    args = parse_args()

    # Determine body content
    if args.body_file:
        if not os.path.exists(args.body_file):
            logger.error("Body file not found: %s", args.body_file)
            sys.exit(1)
        with open(args.body_file, "r", encoding="utf-8") as f:
            body = f.read()
    elif args.body:
        body = args.body
    else:
        if not args.summary and not args.title:
            logger.error("Either --body, --body-file, or at least --title and --summary must be provided.")
            sys.exit(1)
        summary = args.summary or args.title
        body = build_issue_body(
            summary=summary,
            requirements=args.requirements,
            acceptance_criteria=args.criteria,
            technical_notes=args.notes,
            related_files=args.files,
        )

    title = args.title or (args.summary[:80] if args.summary else "")
    if not title:
        logger.error("Issue title cannot be empty.")
        sys.exit(1)

    labels = resolve_labels(
        issue_type=args.type,
        priority=args.priority,
        mark_ready=not args.no_ready,
        extra_labels=args.extra_labels,
    )

    draft = IssueDraft(title=title, body=body, labels=labels)

    try:
        result = execute_gh_create(draft, repo=args.repo, dry_run=args.dry_run)
        if args.json:
            print(json.dumps(result, indent=2))
        else:
            if result.get("status") == "dry_run":
                print("\n=== [DRY RUN] PREVIEW ISSUE ===")
                print(f"Title:  {draft.title}")
                print(f"Labels: {', '.join(draft.labels)}")
                print("\n--- Body ---")
                print(draft.body)
                print("================================\n")
            else:
                print(f"\n✅ Issue #{result.get('issue_number')} created successfully: {result.get('url')}")
                print(f"Labels: {', '.join(draft.labels)}")
                if "ready" in draft.labels:
                    print("🚀 Ticket is marked 'ready' and eligible for autonomous task runner.")
    except Exception as exc:
        logger.error("Execution failed: %s", exc)
        sys.exit(1)


if __name__ == "__main__":
    main()
