#!/usr/bin/env python3
"""
Autonomous PR Review Agent with Auto-Merge and Conflict Escalation Engine.

This agent evaluates open Pull Requests in the repository, classifies their blast radius
and security risk level, performs squash merges on low-risk clean PRs, attempts intelligent
rebase / conflict resolution on dirty PRs, and escalates complex conflicts and high-risk PRs
to human maintainers via automated labeling and descriptive PR review comments.
"""

import argparse
from dataclasses import asdict, dataclass, field
from enum import Enum
import json
import os
from pathlib import Path
import re
import subprocess
import sys
from typing import Any, Dict, List, Optional, Tuple

# ----------------------------------------------------------------------
# Risk & Action Enums
# ----------------------------------------------------------------------

class RiskLevel(str, Enum):
    LOW = "LOW"
    MEDIUM = "MEDIUM"
    HIGH = "HIGH"

class ActionDecision(str, Enum):
    AUTO_MERGE = "AUTO_MERGE"
    REBASE_AND_MERGE = "REBASE_AND_MERGE"
    ESCALATE = "ESCALATE"
    SKIP = "SKIP"

# ----------------------------------------------------------------------
# Path & Classification Regex Patterns
# ----------------------------------------------------------------------

CRITICAL_PATH_PATTERNS = [
    # Auth & Security
    re.compile(r"(?:^|/)auth(?:_service)?\.(?:go|ts|js|py)$", re.IGNORECASE),
    re.compile(r"(?:^|/)(?:auth|oauth|security)/", re.IGNORECASE),
    re.compile(r"(?:^|/)middleware/auth.*", re.IGNORECASE),
    re.compile(r"(?:^|/)security.*", re.IGNORECASE),
    re.compile(r"(?:^|/)oauth.*", re.IGNORECASE),
    re.compile(r"services/api/internal/core/service/auth_service\.go$"),
    # Migrations & Database Schemas
    re.compile(r"^migrations/.*\.sql$"),
    re.compile(r"(?:^|/)migrations/.*"),
    re.compile(r"(?:^|/)atlas\.sum$"),
    # Core Monorepo / Module definitions
    re.compile(r"^services/api/go\.mod$"),
    re.compile(r"^packages/shared/go\.mod$"),
    re.compile(r"^go\.work$"),
]

AUTH_SECURITY_KEYWORD_PATTERNS = [
    re.compile(r"(?:^|/)(?:auth|jwt|oauth|session|token)[^/]*\.(?:go|ts|tsx|js|py)$", re.IGNORECASE),
]

DOC_CONFIG_PATTERNS = [
    re.compile(r"\.md$", re.IGNORECASE),
    re.compile(r"^docs/.*"),
    re.compile(r"^agentic-memory/.*"),
    re.compile(r"^\.gitignore$"),
    re.compile(r"^\.github/.*"),
    re.compile(r"^LICENSE.*"),
]

TEST_FILE_PATTERNS = [
    re.compile(r".*_test\.go$"),
    re.compile(r".*\.test\.(?:ts|tsx|js|jsx)$"),
    re.compile(r".*\.spec\.(?:ts|tsx|js|jsx)$"),
    re.compile(r"(?:^|/)test_.*\.py$"),
    re.compile(r".*_test\.py$"),
]

CORE_SERVICE_PATTERNS = [
    re.compile(r"services/api/internal/core/service/.*\.go$"),
]

TRIVIAL_CONFLICT_FILES = {
    "go.sum",
    "go.work.sum",
    "package-lock.json",
    "pnpm-lock.yaml",
    "docs/docs.go",
    "docs/swagger.json",
    "docs/swagger.yaml",
}

ESCALATION_COMMENT_MARKER = "<!-- pr-review-agent-escalation -->"
ESCALATION_LABEL = "needs-human-review"

# ----------------------------------------------------------------------
# Data Models
# ----------------------------------------------------------------------

@dataclass
class PRDetails:
    number: int
    title: str
    body: str
    head_ref_name: str
    base_ref_name: str
    is_draft: bool
    mergeable: str
    merge_state_status: str
    labels: List[str] = field(default_factory=list)
    files: List[str] = field(default_factory=list)
    additions: int = 0
    deletions: int = 0

@dataclass
class EvaluationResult:
    pr_number: int
    title: str
    head_ref_name: str
    risk_level: RiskLevel
    risk_reasons: List[str]
    mergeable: str
    merge_state_status: str
    decision: ActionDecision
    decision_reason: str
    conflict_files: List[str] = field(default_factory=list)
    auto_resolved: bool = False
    merged: bool = False
    escalated: bool = False
    details: Dict[str, Any] = field(default_factory=dict)

# ----------------------------------------------------------------------
# Risk Classifier
# ----------------------------------------------------------------------

def classify_pr_risk(
    files: List[str],
    title: str = "",
    body: str = "",
    labels: Optional[List[str]] = None,
) -> Tuple[RiskLevel, List[str]]:
    """
    Classifies PR risk into LOW, MEDIUM, or HIGH based on touched files,
    blast radius, security/auth critical paths, and test presence.
    """
    reasons: List[str] = []
    labels = labels or []

    # 1. Check title/label security flags
    title_lower = title.lower()
    if any(l.lower() in ["security", "vulnerability"] for l in labels) or title_lower.startswith("security"):
        reasons.append("PR is explicitly categorized as security-sensitive")

    # 2. Check for critical path files & auth tokens
    critical_matches = []
    auth_matches = []
    migration_matches = []
    module_matches = []

    for f in files:
        if any(p.search(f) for p in CRITICAL_PATH_PATTERNS):
            critical_matches.append(f)
        if any(p.search(f) for p in AUTH_SECURITY_KEYWORD_PATTERNS):
            auth_matches.append(f)
        if re.search(r"(?:^|/)migrations/.*", f) or f.endswith(".sql") or f.endswith("atlas.sum"):
            migration_matches.append(f)
        if f.endswith("go.mod") or f == "go.work":
            module_matches.append(f)

    if critical_matches:
        reasons.append(f"Touches critical path files: {', '.join(critical_matches[:3])}")
    if auth_matches and not any("Touches critical path" in r for r in reasons):
        reasons.append(f"Touches authentication/token paths: {', '.join(auth_matches[:3])}")
    if migration_matches and not any("Touches critical path" in r for r in reasons):
        reasons.append(f"Contains database migrations: {', '.join(migration_matches[:3])}")
    if module_matches and not any("Touches critical path" in r for r in reasons):
        reasons.append(f"Modifies core module definitions: {', '.join(module_matches[:3])}")

    # 3. High blast radius check (> 50 files)
    if len(files) > 50:
        reasons.append(f"High blast radius: {len(files)} files changed (> 50 limit)")

    # 4. Core service without tests check
    core_services = [f for f in files if any(p.search(f) for p in CORE_SERVICE_PATTERNS)]
    test_files = [f for f in files if any(p.search(f) for p in TEST_FILE_PATTERNS)]
    if core_services and not test_files:
        reasons.append(f"Modifies core services ({', '.join(core_services[:2])}) without accompanying unit tests")

    if reasons:
        return RiskLevel.HIGH, reasons

    # Check if purely documentation, agentic-memory, or workflow configs
    all_doc_config = all(any(p.search(f) for p in DOC_CONFIG_PATTERNS) for f in files) if files else True

    # 1. Pure docs, configs, and memory artifacts are always LOW risk
    if all_doc_config:
        return RiskLevel.LOW, ["Documentation, agentic-memory, or workflow configuration changes only"]

    # 2. Feature code changes with tests (or feat title) <= 30 files are MEDIUM risk
    is_feature = title_lower.startswith("feat") or bool(test_files)
    if is_feature and len(files) <= 30:
        return RiskLevel.MEDIUM, [f"Feature code changes with tests/scoped implementation ({len(files)} files <= 30)"]

    # 3. Small scoped patches/refactors <= 10 files
    if len(files) <= 10:
        return RiskLevel.LOW, [f"Small scoped change ({len(files)} files <= 10) with no critical path files"]

    # 4. Medium scoped changes <= 30 files
    if len(files) <= 30:
        return RiskLevel.MEDIUM, [f"Medium scoped changes ({len(files)} files <= 30) with no critical path files"]

    # 5. If > 30 files and not pure docs: HIGH risk
    return RiskLevel.HIGH, [f"Large blast radius: {len(files)} files changed without pure docs classification"]

# ----------------------------------------------------------------------
# PR Review Agent Engine
# ----------------------------------------------------------------------

class PRReviewAgent:
    def __init__(
        self,
        repo_dir: Optional[str] = None,
        dry_run: bool = False,
        max_prs: int = 10,
    ):
        self.repo_dir = repo_dir or os.getcwd()
        self.dry_run = dry_run
        self.max_prs = max_prs

    def run_command(
        self,
        cmd: List[str],
        cwd: Optional[str] = None,
        check: bool = True,
    ) -> subprocess.CompletedProcess:
        """Executes a subprocess command safely."""
        work_dir = cwd or self.repo_dir
        return subprocess.run(
            cmd,
            cwd=work_dir,
            capture_output=True,
            text=True,
            check=check,
        )

    def ensure_label(self) -> None:
        """Ensures the needs-human-review label exists in the repository."""
        if self.dry_run:
            print(f"[DRY-RUN] Would ensure label '{ESCALATION_LABEL}' exists.")
            return

        cmd = [
            "gh", "label", "create", ESCALATION_LABEL,
            "--color", "d93f0b",
            "--description", "PR requires manual human review - agent cannot auto-merge",
            "--force",
        ]
        try:
            self.run_command(cmd, check=False)
        except Exception as e:
            print(f"Warning: Failed to ensure label '{ESCALATION_LABEL}': {e}", file=sys.stderr)

    def fetch_open_prs(self, pr_number: Optional[int] = None) -> List[PRDetails]:
        """Fetches open Pull Requests with full metadata using GitHub CLI."""
        if pr_number is not None:
            cmd = [
                "gh", "pr", "view", str(pr_number),
                "--json", "number,title,body,headRefName,baseRefName,isDraft,mergeable,mergeStateStatus,labels,files,additions,deletions",
            ]
            res = self.run_command(cmd)
            data = [json.loads(res.stdout)]
        else:
            cmd = [
                "gh", "pr", "list", "--state", "open", "--limit", str(self.max_prs),
                "--json", "number,title,body,headRefName,baseRefName,isDraft,mergeable,mergeStateStatus,labels,files,additions,deletions",
            ]
            res = self.run_command(cmd)
            data = json.loads(res.stdout or "[]")

        prs = []
        for item in data:
            files = [f.get("path") for f in item.get("files", []) if f.get("path")]
            labels = [l.get("name") for l in item.get("labels", []) if l.get("name")]
            prs.append(
                PRDetails(
                    number=item.get("number"),
                    title=item.get("title", ""),
                    body=item.get("body", ""),
                    head_ref_name=item.get("headRefName", ""),
                    base_ref_name=item.get("baseRefName", "main"),
                    is_draft=bool(item.get("isDraft", False)),
                    mergeable=item.get("mergeable", "UNKNOWN"),
                    merge_state_status=item.get("mergeStateStatus", "UNKNOWN"),
                    labels=labels,
                    files=files,
                    additions=item.get("additions", 0),
                    deletions=item.get("deletions", 0),
                )
            )
        return prs

    def inspect_conflicts_in_memory(self, head_ref: str, base_ref: str = "main") -> List[str]:
        """
        Uses git plumbing (git merge-tree) to determine conflicting files
        without mutating the current worktree.
        """
        try:
            # Fetch remote refs first
            self.run_command(["git", "fetch", "origin", f"{base_ref}:{base_ref}"], check=False)
            self.run_command(["git", "fetch", "origin", f"{head_ref}:{head_ref}"], check=False)

            # Get merge base
            base_res = self.run_command(["git", "merge-base", f"origin/{base_ref}", f"origin/{head_ref}"], check=False)
            if base_res.returncode != 0 or not base_res.stdout.strip():
                return []
            merge_base = base_res.stdout.strip()

            # Execute git merge-tree
            tree_res = self.run_command(
                ["git", "merge-tree", merge_base, f"origin/{base_ref}", f"origin/{head_ref}"],
                check=False,
            )
            conflicts = []
            for line in tree_res.stdout.splitlines():
                line = line.strip()
                if line.startswith("our") or line.startswith("their"):
                    parts = line.split()
                    if len(parts) >= 4:
                        conflicts.append(parts[3])
                elif "<<<<<<<" in line:
                    pass
            return sorted(list(set(conflicts)))
        except Exception:
            return []

    def attempt_rebase_and_conflict_resolution(
        self,
        pr: PRDetails,
    ) -> Tuple[bool, str, List[str]]:
        """
        Attempts to checkout the PR branch in an isolated temporary worktree,
        run git rebase origin/main, and auto-resolve trivial conflicts (lockfiles, generated files).
        """
        conflicting_files: List[str] = []
        temp_worktree_dir = os.path.join(self.repo_dir, f".worktrees_tmp_rebase_{pr.number}")

        if self.dry_run:
            conflicts = self.inspect_conflicts_in_memory(pr.head_ref_name, pr.base_ref_name)
            if not conflicts:
                conflicts = [f for f in pr.files if f in TRIVIAL_CONFLICT_FILES or "go.mod" in f or "GEMINI.md" in f][:3]
            is_trivial = bool(conflicts and all(os.path.basename(c) in TRIVIAL_CONFLICT_FILES for c in conflicts))
            if is_trivial:
                return True, "Lockfiles/generated files trivially resolvable in dry-run preview", conflicts
            return False, "Complex source code conflicts detected", conflicts

        # Live isolated worktree execution
        try:
            self.run_command(["git", "fetch", "origin", pr.head_ref_name])
            self.run_command(["git", "fetch", "origin", pr.base_ref_name])

            # Create ephemeral detached worktree
            add_res = self.run_command(
                ["git", "worktree", "add", "--detach", temp_worktree_dir, f"origin/{pr.head_ref_name}"],
                check=False,
            )
            if add_res.returncode != 0:
                return False, f"Failed to create isolated worktree: {add_res.stderr.strip()}", []

            # Attempt rebase
            rebase_res = self.run_command(
                ["git", "rebase", f"origin/{pr.base_ref_name}"],
                cwd=temp_worktree_dir,
                check=False,
            )

            if rebase_res.returncode == 0:
                # Rebased cleanly! Push changes
                push_res = self.run_command(
                    ["git", "push", "--force-with-lease", "origin", f"HEAD:{pr.head_ref_name}"],
                    cwd=temp_worktree_dir,
                    check=False,
                )
                if push_res.returncode == 0:
                    return True, "Rebased cleanly against main and updated remote branch", []
                return False, f"Failed to push rebased branch: {push_res.stderr.strip()}", []

            # Rebase paused with conflicts. Check unmerged files
            status_res = self.run_command(
                ["git", "diff", "--name-only", "--diff-filter=U"],
                cwd=temp_worktree_dir,
                check=False,
            )
            conflicting_files = [f.strip() for f in status_res.stdout.splitlines() if f.strip()]

            # Check if all conflicts are trivial lockfiles or generated files
            non_trivial = [f for f in conflicting_files if os.path.basename(f) not in TRIVIAL_CONFLICT_FILES]
            if non_trivial:
                # Complex source conflicts! Abort rebase
                self.run_command(["git", "rebase", "--abort"], cwd=temp_worktree_dir, check=False)
                return False, f"Complex conflicts in source files: {', '.join(non_trivial[:3])}", conflicting_files

            # Auto-resolve trivial files
            for f in conflicting_files:
                basename = os.path.basename(f)
                if basename in ["go.sum", "go.work.sum", "package-lock.json", "pnpm-lock.yaml"]:
                    # Accept ours (origin/main) or run sync
                    self.run_command(["git", "checkout", "--ours", f], cwd=temp_worktree_dir, check=False)
                    self.run_command(["git", "add", f], cwd=temp_worktree_dir, check=False)
                elif basename in ["docs.go", "swagger.json", "swagger.yaml"]:
                    self.run_command(["git", "checkout", "--theirs", f], cwd=temp_worktree_dir, check=False)
                    self.run_command(["git", "add", f], cwd=temp_worktree_dir, check=False)

            # Continue rebase
            cont_res = self.run_command(
                ["git", "-c", "core.editor=true", "rebase", "--continue"],
                cwd=temp_worktree_dir,
                check=False,
            )
            if cont_res.returncode == 0:
                push_res = self.run_command(
                    ["git", "push", "--force-with-lease", "origin", f"HEAD:{pr.head_ref_name}"],
                    cwd=temp_worktree_dir,
                    check=False,
                )
                if push_res.returncode == 0:
                    return True, "Auto-resolved trivial lockfile/generated conflicts and rebased", conflicting_files
                return False, f"Failed to push resolved branch: {push_res.stderr.strip()}", conflicting_files

            self.run_command(["git", "rebase", "--abort"], cwd=temp_worktree_dir, check=False)
            return False, "Failed to resolve conflicts automatically", conflicting_files

        finally:
            # Clean up ephemeral worktree
            if os.path.exists(temp_worktree_dir):
                self.run_command(["git", "worktree", "remove", "--force", temp_worktree_dir], check=False)

    def evaluate_pr(self, pr: PRDetails) -> EvaluationResult:
        """Evaluates PR risk and decides action: AUTO_MERGE, REBASE_AND_MERGE, ESCALATE, or SKIP."""
        if pr.is_draft:
            return EvaluationResult(
                pr_number=pr.number,
                title=pr.title,
                head_ref_name=pr.head_ref_name,
                risk_level=RiskLevel.LOW,
                risk_reasons=["PR is currently marked as draft"],
                mergeable=pr.mergeable,
                merge_state_status=pr.merge_state_status,
                decision=ActionDecision.SKIP,
                decision_reason="Skipping draft PR",
            )

        risk_level, risk_reasons = classify_pr_risk(
            files=pr.files,
            title=pr.title,
            body=pr.body,
            labels=pr.labels,
        )

        is_conflicting = (
            pr.mergeable == "CONFLICTING"
            or pr.merge_state_status in ["DIRTY", "CONFLICTING"]
        )

        # If GitHub reports UNKNOWN, check conflicts directly using git plumbing
        if (pr.mergeable == "UNKNOWN" or pr.merge_state_status == "UNKNOWN") and not is_conflicting:
            detected_conflicts = self.inspect_conflicts_in_memory(pr.head_ref_name, pr.base_ref_name)
            if detected_conflicts:
                is_conflicting = True
                pr.mergeable = "CONFLICTING"
                pr.merge_state_status = "DIRTY"
        is_clean = (
            pr.mergeable == "MERGEABLE"
            and pr.merge_state_status in ["CLEAN", "UNKNOWN", "HAS_HOOKS"]
        ) or (pr.mergeStateStatus == "CLEAN" if hasattr(pr, "mergeStateStatus") else False)

        # High Risk PRs are always escalated
        if risk_level == RiskLevel.HIGH:
            return EvaluationResult(
                pr_number=pr.number,
                title=pr.title,
                head_ref_name=pr.head_ref_name,
                risk_level=risk_level,
                risk_reasons=risk_reasons,
                mergeable=pr.mergeable,
                merge_state_status=pr.merge_state_status,
                decision=ActionDecision.ESCALATE,
                decision_reason="High risk change requires human review and sign-off",
                conflict_files=[] if not is_conflicting else ["(conflicting state detected)"],
            )

        # Conflicting Low or Medium Risk PRs
        if is_conflicting:
            return EvaluationResult(
                pr_number=pr.number,
                title=pr.title,
                head_ref_name=pr.head_ref_name,
                risk_level=risk_level,
                risk_reasons=risk_reasons,
                mergeable=pr.mergeable,
                merge_state_status=pr.merge_state_status,
                decision=ActionDecision.REBASE_AND_MERGE,
                decision_reason="Conflicts detected; requires rebase or trivial conflict resolution",
            )

        # Clean Low or Medium Risk PRs -> Auto-merge!
        return EvaluationResult(
            pr_number=pr.number,
            title=pr.title,
            head_ref_name=pr.head_ref_name,
            risk_level=risk_level,
            risk_reasons=risk_reasons,
            mergeable=pr.mergeable,
            merge_state_status=pr.merge_state_status,
            decision=ActionDecision.AUTO_MERGE,
            decision_reason=f"Clean merge state and {risk_level.value} risk rating",
        )

    def escalate_pr(self, result: EvaluationResult, pr: PRDetails) -> bool:
        """Escalates a PR by adding the needs-human-review label and posting a detailed comment."""
        if self.dry_run:
            print(f"[DRY-RUN] Would add label '{ESCALATION_LABEL}' to PR #{pr.number}.")
            print(f"[DRY-RUN] Would post escalation comment to PR #{pr.number} (Reason: {result.decision_reason}).")
            result.escalated = True
            return True

        # Check if already labelled
        if ESCALATION_LABEL not in pr.labels:
            try:
                self.run_command(["gh", "pr", "edit", str(pr.number), "--add-label", ESCALATION_LABEL], check=False)
            except Exception as e:
                print(f"Warning: Could not add label to PR #{pr.number}: {e}", file=sys.stderr)

        # Check existing comments to prevent spam
        try:
            view_res = self.run_command(["gh", "pr", "view", str(pr.number), "--json", "comments"], check=False)
            if view_res.returncode == 0:
                comments = json.loads(view_res.stdout or "{}").get("comments", [])
                for c in comments:
                    if ESCALATION_COMMENT_MARKER in c.get("body", ""):
                        # Already commented
                        result.escalated = True
                        return True
        except Exception:
            pass

        # Build comment
        risk_bullets = "\n".join([f"- {r}" for r in result.risk_reasons])
        conflict_section = ""
        if result.conflict_files:
            files_bullets = "\n".join([f"- `{f}`" for f in result.conflict_files[:5]])
            conflict_section = f"\n#### Conflicting / Affected Files\n{files_bullets}\n"

        comment_body = (
            f"{ESCALATION_COMMENT_MARKER}\n"
            f"### ⚠️ Autonomous PR Review Agent: Manual Review Required\n\n"
            f"This Pull Request has been evaluated by the **PR Review Agent** and flagged for human intervention.\n\n"
            f"#### Evaluation Summary\n"
            f"- **Risk Level**: `{result.risk_level.value}`\n"
            f"- **Merge State**: `{result.merge_state_status}` ({result.mergeable})\n"
            f"- **Files Changed**: {len(pr.files)}\n"
            f"- **Evaluation Decision**: {result.decision_reason}\n\n"
            f"#### Risk Factors & Rationale\n"
            f"{risk_bullets}\n"
            f"{conflict_section}\n"
            f"**Next Steps**: Maintainers, please inspect the changes and resolve any conflicts before proceeding.\n"
        )

        try:
            self.run_command(["gh", "pr", "comment", str(pr.number), "--body", comment_body], check=True)
            result.escalated = True
            return True
        except Exception as e:
            print(f"Warning: Failed to comment on PR #{pr.number}: {e}", file=sys.stderr)
            return False

    def merge_pr(self, result: EvaluationResult, pr: PRDetails) -> bool:
        """Merges an auto-mergeable PR using squash merge and cleans up issue status."""
        if self.dry_run:
            print(f"[DRY-RUN] Would squash merge PR #{pr.number} (Branch: {pr.head_ref_name}).")
            result.merged = True
            return True

        merge_cmd = ["gh", "pr", "merge", str(pr.number), "--squash", "--delete-branch"]
        try:
            self.run_command(merge_cmd, check=True)
            result.merged = True
            print(f"Successfully merged PR #{pr.number} ({pr.head_ref_name}) via squash merge.")

            # Post-merge issue state cleanup: remove in-review label if linked issue exists
            linked_issues = re.findall(r"(?:closes|fixes|resolves|refs)\s+#(\d+)", pr.body, re.IGNORECASE)
            for issue_num in set(linked_issues):
                try:
                    self.run_command(
                        ["gh", "issue", "edit", issue_num, "--remove-label", "in-review"],
                        check=False,
                    )
                except Exception:
                    pass
            return True
        except Exception as e:
            print(f"Failed to merge PR #{pr.number}: {e}", file=sys.stderr)
            return False

    def process_pr(
        self,
        pr: PRDetails,
        resolve_conflicts_only: bool = False,
    ) -> EvaluationResult:
        """Processes a single PR through the evaluation, resolution, merge, and escalation pipeline."""
        res = self.evaluate_pr(pr)

        if resolve_conflicts_only:
            if res.decision == ActionDecision.REBASE_AND_MERGE:
                resolved, msg, conflicts = self.attempt_rebase_and_conflict_resolution(pr)
                res.conflict_files = conflicts
                if resolved:
                    res.auto_resolved = True
                    res.decision_reason = msg
                    res.decision = ActionDecision.AUTO_MERGE
                else:
                    res.decision = ActionDecision.ESCALATE
                    res.decision_reason = msg
                    self.escalate_pr(res, pr)
            return res

        # Standard Review & Merge Flow
        if res.decision == ActionDecision.AUTO_MERGE:
            self.merge_pr(res, pr)

        elif res.decision == ActionDecision.REBASE_AND_MERGE:
            resolved, msg, conflicts = self.attempt_rebase_and_conflict_resolution(pr)
            res.conflict_files = conflicts
            if resolved:
                res.auto_resolved = True
                res.decision_reason = msg
                res.decision = ActionDecision.AUTO_MERGE
                self.merge_pr(res, pr)
            else:
                res.decision = ActionDecision.ESCALATE
                res.decision_reason = msg
                self.escalate_pr(res, pr)

        elif res.decision == ActionDecision.ESCALATE:
            self.escalate_pr(res, pr)

        return res

    def run(
        self,
        mode: str = "review-and-merge",
        target_pr: Optional[int] = None,
    ) -> List[EvaluationResult]:
        """Main entry point orchestrating PR review execution."""
        self.ensure_label()
        prs = self.fetch_open_prs(target_pr)
        results = []

        resolve_conflicts_only = (mode == "resolve-conflicts")

        for pr in prs:
            res = self.process_pr(pr, resolve_conflicts_only=resolve_conflicts_only)
            results.append(res)

        return results

    def format_markdown_report(self, results: List[EvaluationResult]) -> str:
        """Generates a structured markdown report of all processed PRs."""
        lines = [
            "# Autonomous PR Review Agent Report",
            f"**Mode**: {'DRY RUN PREVIEW' if self.dry_run else 'ACTIVE EXECUTION'}",
            f"**PRs Processed**: {len(results)}",
            "",
            "| PR | Title | Risk | Merge State | Decision | Action Taken |",
            "|---|---|---|---|---|---|",
        ]

        for r in results:
            action_str = "None"
            if self.dry_run:
                if r.decision == ActionDecision.AUTO_MERGE:
                    action_str = "Would Merge"
                elif r.decision == ActionDecision.ESCALATE:
                    action_str = "Would Escalate"
                elif r.decision == ActionDecision.REBASE_AND_MERGE:
                    action_str = "Would Rebase"
                else:
                    action_str = "Skip"
            else:
                if r.merged:
                    action_str = "Merged"
                elif r.escalated:
                    action_str = "Escalated"
                elif r.auto_resolved:
                    action_str = "Rebased"
                else:
                    action_str = "Pending"

            risk_badge = f"🟢 `{r.risk_level.value}`" if r.risk_level == RiskLevel.LOW else (
                f"🟡 `{r.risk_level.value}`" if r.risk_level == RiskLevel.MEDIUM else f"🔴 `{r.risk_level.value}`"
            )
            title_clean = r.title.replace("|", "/")
            lines.append(
                f"| #{r.pr_number} | {title_clean[:40]} | {risk_badge} | `{r.merge_state_status}` | `{r.decision.value}` | {action_str} |"
            )

        lines.append("")
        lines.append("### Detailed Evaluation & Rationale")
        for r in results:
            lines.append(f"#### PR #{r.pr_number}: {r.title}")
            lines.append(f"- **Risk Level**: `{r.risk_level.value}`")
            lines.append(f"- **Decision**: `{r.decision.value}` ({r.decision_reason})")
            lines.append(f"- **Risk Factors**:")
            for reason in r.risk_reasons:
                lines.append(f"  - {reason}")
            if r.conflict_files:
                lines.append(f"- **Conflicting Files**: {', '.join(r.conflict_files[:4])}")
            lines.append("")

        return "\n".join(lines)


# ----------------------------------------------------------------------
# CLI Interface
# ----------------------------------------------------------------------

def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Autonomous PR Review Agent with auto-merge and conflict escalation."
    )
    group = parser.add_mutually_exclusive_group()
    group.add_argument(
        "--review-and-merge",
        action="store_true",
        default=True,
        help="Review open PRs, auto-merge low-risk PRs, and escalate high-risk PRs (default mode)",
    )
    group.add_argument(
        "--resolve-conflicts",
        action="store_true",
        help="Target conflicting PRs to attempt automated rebase & trivial conflict resolution",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview all decisions without mutating remote PRs or merging branches",
    )
    parser.add_argument(
        "--pr",
        type=int,
        default=None,
        help="Target a specific PR number instead of all open PRs",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Output structured JSON results instead of human-readable report",
    )
    parser.add_argument(
        "--ensure-label",
        action="store_true",
        help="Ensure 'needs-human-review' label exists on repository and exit",
    )
    parser.add_argument(
        "--max-prs",
        type=int,
        default=10,
        help="Maximum open PRs to process in a single sweep (default 10)",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    agent = PRReviewAgent(dry_run=args.dry_run, max_prs=args.max_prs)

    if args.ensure_label:
        agent.ensure_label()
        print(f"Ensured '{ESCALATION_LABEL}' label exists.")
        sys.exit(0)

    mode = "resolve-conflicts" if args.resolve_conflicts else "review-and-merge"
    results = agent.run(mode=mode, target_pr=args.pr)

    if args.json:
        output_data = [
            {
                "pr_number": r.pr_number,
                "title": r.title,
                "head_ref_name": r.head_ref_name,
                "risk_level": r.risk_level.value,
                "risk_reasons": r.risk_reasons,
                "mergeable": r.mergeable,
                "merge_state_status": r.merge_state_status,
                "decision": r.decision.value,
                "decision_reason": r.decision_reason,
                "conflict_files": r.conflict_files,
                "auto_resolved": r.auto_resolved,
                "merged": r.merged,
                "escalated": r.escalated,
            }
            for r in results
        ]
        print(json.dumps(output_data, indent=2))
    else:
        report = agent.format_markdown_report(results)
        print(report)


if __name__ == "__main__":
    main()
