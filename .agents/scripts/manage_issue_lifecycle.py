#!/usr/bin/env python3
"""
manage_issue_lifecycle.py - Automated Issue Lifecycle Manager and State Synchronizer for Hexta.

Provides reconciliation between Pull Requests and Issue states, automated backlog triage,
and stale lock recovery to maintain the 5-state lifecycle:
  [Backlog] -> [ready] -> [in-progress] -> [in-review] -> [done]

Capabilities:
  --sync:          Reconciles open PRs with issue status and strips dev labels from closed issues.
  --triage:        Audits open backlog issues without status and promotes well-specified items to 'ready'.
  --recover-stale: Recovers orphaned 'in-progress' issues inactive for > threshold (default 2h) back to 'ready'.
  --dry-run:       Simulates operations without mutating GitHub repository state.
"""

import argparse
from dataclasses import dataclass, field
from datetime import datetime, timezone
import json
import logging
import re
import subprocess
import sys
from typing import Any, Dict, List, Optional, Set, Tuple

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("issue_lifecycle")

DEV_STATUS_LABELS = {"ready", "in-progress", "in-review"}
ALL_STATUS_LABELS = {"ready", "in-progress", "in-review", "blocked"}
PRIORITY_LABELS = {"priority:high", "priority:medium", "priority:low"}


@dataclass
class ActionRecord:
    issue_number: int
    issue_title: str
    action: str
    labels_added: List[str] = field(default_factory=list)
    labels_removed: List[str] = field(default_factory=list)
    comment: Optional[str] = None
    reason: str = ""

    def to_dict(self) -> Dict[str, Any]:
        return {
            "issue_number": self.issue_number,
            "title": self.issue_title,
            "action": self.action,
            "labels_added": self.labels_added,
            "labels_removed": self.labels_removed,
            "comment": self.comment,
            "reason": self.reason
        }


class IssueLifecycleManager:
    def __init__(self, dry_run: bool = False, verbose: bool = False, stale_hours: float = 2.0):
        self.dry_run = dry_run
        self.verbose = verbose
        self.stale_hours = stale_hours
        self.actions: List[ActionRecord] = []

        if verbose:
            logger.setLevel(logging.DEBUG)

    def _run_gh_json(self, cmd: List[str]) -> Any:
        """Executes a gh CLI command and returns parsed JSON output."""
        try:
            logger.debug("Executing command: %s", " ".join(cmd))
            res = subprocess.run(cmd, capture_output=True, text=True, check=True)
            return json.loads(res.stdout or "[]")
        except subprocess.CalledProcessError as e:
            logger.error("gh CLI command failed: %s\nStderr: %s", " ".join(cmd), e.stderr)
            raise RuntimeError(f"GitHub CLI command failed: {e.stderr.strip()}") from e
        except json.JSONDecodeError as e:
            logger.error("Failed to parse JSON output: %s", e)
            return []

    def _mutate_issue(
        self,
        issue_number: int,
        title: str,
        add_labels: Optional[List[str]] = None,
        remove_labels: Optional[List[str]] = None,
        comment: Optional[str] = None,
        reason: str = ""
    ) -> bool:
        """Mutates an issue's labels and/or posts an audit comment."""
        add_labels = add_labels or []
        remove_labels = remove_labels or []

        if not add_labels and not remove_labels and not comment:
            return False

        action_name = "reconcile"
        if add_labels and not remove_labels:
            action_name = "add_labels"
        elif remove_labels and not add_labels:
            action_name = "remove_labels"
        elif comment and not add_labels and not remove_labels:
            action_name = "comment"

        record = ActionRecord(
            issue_number=issue_number,
            issue_title=title,
            action=action_name,
            labels_added=add_labels,
            labels_removed=remove_labels,
            comment=comment,
            reason=reason
        )
        self.actions.append(record)

        log_prefix = "[DRY-RUN] " if self.dry_run else ""
        logger.info(
            "%sIssue #%d: %s | Add: %s | Remove: %s | Reason: %s",
            log_prefix, issue_number, action_name, add_labels, remove_labels, reason
        )

        if self.dry_run:
            return True

        # Execute labels mutation
        if add_labels or remove_labels:
            edit_cmd = ["gh", "issue", "edit", str(issue_number)]
            if add_labels:
                edit_cmd.extend(["--add-label", ",".join(add_labels)])
            if remove_labels:
                edit_cmd.extend(["--remove-label", ",".join(remove_labels)])
            try:
                subprocess.run(edit_cmd, capture_output=True, text=True, check=True)
            except subprocess.CalledProcessError as e:
                logger.error("Failed to edit issue #%d: %s", issue_number, e.stderr)
                return False

        # Execute comment
        if comment:
            comment_cmd = ["gh", "issue", "comment", str(issue_number), "--body", comment]
            try:
                subprocess.run(comment_cmd, capture_output=True, text=True, check=True)
            except subprocess.CalledProcessError as e:
                logger.error("Failed to comment on issue #%d: %s", issue_number, e.stderr)
                return False

        return True

    def fetch_all_issues(self) -> List[Dict[str, Any]]:
        """Fetches all repository issues across states."""
        cmd = [
            "gh", "issue", "list",
            "--state", "all",
            "--limit", "200",
            "--json", "number,title,body,labels,state,createdAt,updatedAt"
        ]
        return self._run_gh_json(cmd)

    def fetch_all_prs(self) -> List[Dict[str, Any]]:
        """Fetches all repository pull requests across states."""
        cmd = [
            "gh", "pr", "list",
            "--state", "all",
            "--limit", "200",
            "--json", "number,title,body,headRefName,state,mergedAt,closedAt,createdAt,updatedAt"
        ]
        return self._run_gh_json(cmd)

    def extract_referenced_issues(self, pr: Dict[str, Any]) -> Set[int]:
        """Identifies all issue numbers linked or referenced by a Pull Request."""
        refs: Set[int] = set()

        # 1. Check head branch name convention: task/issue-<number>-<slug>
        head_branch = pr.get("headRefName") or ""
        branch_match = re.search(r"task/issue-(\d+)", head_branch)
        if branch_match:
            refs.add(int(branch_match.group(1)))

        # 2. Check PR Title
        title = pr.get("title") or ""
        for match in re.finditer(r"(?:closes|fixes|resolves|refs)?\s*#(\d+)", title, re.IGNORECASE):
            refs.add(int(match.group(1)))

        # 3. Check PR Body for explicit closing directives and references
        body = pr.get("body") or ""
        for match in re.finditer(r"(?:closes|close|closed|fixes|fix|fixed|resolves|resolve|resolved)\s+#(\d+)", body, re.IGNORECASE):
            refs.add(int(match.group(1)))

        for match in re.finditer(r"(?:refs|ref|references|see)\s+#(\d+)", body, re.IGNORECASE):
            refs.add(int(match.group(1)))

        return refs

    def sync(self) -> int:
        """
        Reconciles PR and issue states:
        1. Strips development labels from CLOSED issues.
        2. Ensures OPEN issues with an active OPEN PR have 'in-review' and no 'in-progress' or 'ready'.
        3. Reconciles OPEN issues referencing merged or rejected PRs.
        """
        logger.info("Starting issue lifecycle synchronization...")
        issues = self.fetch_all_issues()
        prs = self.fetch_all_prs()

        # Map issue_number -> list of associated PRs
        issue_to_prs: Dict[int, List[Dict[str, Any]]] = {}
        for pr in prs:
            referenced = self.extract_referenced_issues(pr)
            for iss_num in referenced:
                issue_to_prs.setdefault(iss_num, []).append(pr)

        changes_count = 0

        for issue in issues:
            num = issue.get("number", 0)
            title = issue.get("title", "")
            state = issue.get("state", "").upper()
            label_names = {l.get("name") for l in issue.get("labels", []) if l.get("name")}
            associated_prs = issue_to_prs.get(num, [])

            # --- Rule A: Closed Issue Cleanup ---
            if state == "CLOSED":
                stale_labels = [lbl for lbl in ["in-progress", "in-review", "ready"] if lbl in label_names]
                if stale_labels:
                    reason = f"Issue is CLOSED; stripped stale {', '.join(stale_labels)} label(s)"
                    if self._mutate_issue(num, title, remove_labels=stale_labels, reason=reason):
                        changes_count += 1
                continue

            # --- Rule B: Open Issue with Pull Requests ---
            if state == "OPEN":
                open_prs = [pr for pr in associated_prs if pr.get("state") == "OPEN"]
                merged_prs = [pr for pr in associated_prs if pr.get("state") == "MERGED"]
                closed_prs = [pr for pr in associated_prs if pr.get("state") == "CLOSED"]

                if open_prs:
                    pr_nums = [str(p.get("number")) for p in open_prs]
                    add_labels: List[str] = []
                    remove_labels: List[str] = []

                    if "in-review" not in label_names:
                        add_labels.append("in-review")
                    if "in-progress" in label_names:
                        remove_labels.append("in-progress")
                    if "ready" in label_names:
                        remove_labels.append("ready")

                    if add_labels or remove_labels:
                        reason = f"Active open PR(s) #{','.join(pr_nums)} found; ensuring 'in-review' state"
                        if self._mutate_issue(
                            num, title, add_labels=add_labels, remove_labels=remove_labels, reason=reason
                        ):
                            changes_count += 1
                elif merged_prs and not open_prs:
                    # All PRs merged, but issue still open: remove in-progress / in-review
                    removals = [lbl for lbl in ["in-progress", "in-review"] if lbl in label_names]
                    if removals:
                        reason = f"All associated PR(s) merged; removing {', '.join(removals)}"
                        if self._mutate_issue(num, title, remove_labels=removals, reason=reason):
                            changes_count += 1
                elif closed_prs and not open_prs and not merged_prs:
                    # PRs closed without merge: if stuck in 'in-review', revert to 'in-progress' or 'ready'
                    if "in-review" in label_names:
                        reason = "Associated PR closed without merge; reverting 'in-review' to 'in-progress'"
                        if self._mutate_issue(
                            num, title, add_labels=["in-progress"], remove_labels=["in-review"], reason=reason
                        ):
                            changes_count += 1

        logger.info("Lifecycle sync completed. Total mutations: %d", changes_count)
        return changes_count

    def triage(self) -> int:
        """
        Audits open backlog issues lacking status labels.
        Promotes well-specified issues to 'ready' and assigns an estimated priority.
        """
        logger.info("Starting backlog specification triage...")
        issues = self.fetch_all_issues()
        triaged_count = 0

        for issue in issues:
            if issue.get("state", "").upper() != "OPEN":
                continue

            num = issue.get("number", 0)
            title = issue.get("title", "")
            body = issue.get("body", "") or ""
            label_names = {l.get("name") for l in issue.get("labels", []) if l.get("name")}

            # Skip issues that already have a status label
            if any(status in label_names for status in ALL_STATUS_LABELS):
                continue

            # Evaluate specification quality criteria
            has_sufficient_title = len(title.strip()) >= 10
            has_sufficient_body = len(body.strip()) >= 50
            has_structure = bool(
                re.search(r"^#{1,3}\s+", body, re.MULTILINE) or
                re.search(r"-\s*\[[ xX]\]", body) or
                re.search(r"(?:acceptance\s+criteria|requirements|scope|objective)", body, re.IGNORECASE)
            )

            is_qualified = has_sufficient_title and has_sufficient_body and has_structure

            if is_qualified:
                add_labels = ["ready"]
                # Infer priority if missing
                has_priority = any(p in label_names for p in PRIORITY_LABELS)
                if not has_priority:
                    inferred_priority = self._infer_priority(title, body)
                    add_labels.append(inferred_priority)

                reason = "Backlog issue meets specification criteria; promoted to 'ready'"
                if self._mutate_issue(num, title, add_labels=add_labels, reason=reason):
                    triaged_count += 1
            else:
                logger.debug(
                    "Issue #%d '%s' did not pass triage quality check (title_len=%d, body_len=%d, structured=%s)",
                    num, title, len(title), len(body), has_structure
                )

        logger.info("Triage audit completed. Promoted issues: %d", triaged_count)
        return triaged_count

    def recover_stale(self) -> int:
        """
        Identifies 'in-progress' issues inactive for > stale_hours with no open PR,
        reverting them to 'ready' to unlock the task queue.
        """
        logger.info("Scanning for stale 'in-progress' task locks (threshold: %.1fh)...", self.stale_hours)
        issues = self.fetch_all_issues()
        prs = self.fetch_all_prs()

        # Collect issue numbers with active open PRs
        open_pr_issues: Set[int] = set()
        for pr in prs:
            if pr.get("state") == "OPEN":
                open_pr_issues.update(self.extract_referenced_issues(pr))

        now = datetime.now(timezone.utc)
        recovered_count = 0

        for issue in issues:
            if issue.get("state", "").upper() != "OPEN":
                continue

            num = issue.get("number", 0)
            title = issue.get("title", "")
            label_names = {l.get("name") for l in issue.get("labels", []) if l.get("name")}

            if "in-progress" not in label_names:
                continue

            # If there is an active open PR, this task is progressing towards review, skip recovery
            if num in open_pr_issues:
                continue

            # Check inactivity duration
            updated_at_str = issue.get("updatedAt")
            if not updated_at_str:
                continue

            try:
                # Handle ISO 8601 timestamps like "2026-09-12T10:20:42Z"
                updated_at = datetime.fromisoformat(updated_at_str.replace("Z", "+00:00"))
            except ValueError:
                logger.warning("Could not parse updatedAt '%s' for issue #%d", updated_at_str, num)
                continue

            elapsed = (now - updated_at).total_seconds() / 3600.0
            if elapsed >= self.stale_hours:
                reason = f"Inactive in 'in-progress' for {elapsed:.1f}h (threshold: {self.stale_hours:.1f}h) with no open PR"
                comment_body = (
                    f"🤖 **Autonomous Task Runner State Recovery**\n\n"
                    f"This task was locked in `in-progress` for **{elapsed:.1f} hours** with no active Pull Request or update. "
                    f"Automatically unlocking and restoring state to `ready` for the queue runner."
                )
                if self._mutate_issue(
                    num,
                    title,
                    add_labels=["ready"],
                    remove_labels=["in-progress"],
                    comment=comment_body,
                    reason=reason
                ):
                    recovered_count += 1

        logger.info("Stale recovery completed. Recovered tasks: %d", recovered_count)
        return recovered_count

    def _infer_priority(self, title: str, body: str) -> str:
        """Infers an issue priority from keywords in title and description."""
        combined = f"{title} {body}".lower()
        if any(kw in combined for kw in ["cve", "vulnerability", "security", "crash", "critical", "blocker"]):
            return "priority:high"
        if any(kw in combined for kw in ["docs", "documentation", "readme", "chore", "style"]):
            return "priority:low"
        return "priority:medium"

    def get_summary(self) -> Dict[str, Any]:
        """Builds structured summary dictionary of execution results."""
        return {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "dry_run": self.dry_run,
            "stale_hours_threshold": self.stale_hours,
            "total_actions": len(self.actions),
            "actions": [a.to_dict() for a in self.actions]
        }


def main():
    parser = argparse.ArgumentParser(
        description="Automated Issue Lifecycle Manager and State Synchronizer for Hexta"
    )
    parser.add_argument(
        "--sync", action="store_true",
        help="Reconcile PR and issue states and clean up closed issues"
    )
    parser.add_argument(
        "--triage", action="store_true",
        help="Audit open backlog issues and promote qualified issues to 'ready'"
    )
    parser.add_argument(
        "--recover-stale", action="store_true",
        help="Unlock stale 'in-progress' issues inactive for > threshold"
    )
    parser.add_argument(
        "--stale-hours", type=float, default=2.0,
        help="Inactivity threshold in hours for stale recovery (default: 2.0)"
    )
    parser.add_argument(
        "--dry-run", action="store_true",
        help="Simulate label changes and comments without mutating GitHub state"
    )
    parser.add_argument(
        "--verbose", action="store_true",
        help="Enable verbose logging output"
    )
    parser.add_argument(
        "--json", action="store_true",
        help="Emit output in structured JSON format"
    )

    args = parser.parse_args()

    # Default to --sync if no operational mode selected
    if not (args.sync or args.triage or args.recover_stale):
        args.sync = True

    manager = IssueLifecycleManager(
        dry_run=args.dry_run,
        verbose=args.verbose,
        stale_hours=args.stale_hours
    )

    if args.sync:
        manager.sync()

    if args.recover_stale:
        manager.recover_stale()

    if args.triage:
        manager.triage()

    if args.json:
        print(json.dumps(manager.get_summary(), indent=2))
    else:
        summary = manager.get_summary()
        print(f"\nExecution Summary: {summary['total_actions']} action(s) performed (dry_run={args.dry_run}).")


if __name__ == "__main__":
    main()
