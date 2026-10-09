#!/usr/bin/env python3
"""
consolidate_memory.py - Agentic Memory Consolidation and Pruning Engine for Hexta.

Consolidates and prunes `agentic-memory/` to prevent repository bloat over time:
1. Synthesizes individual task changelogs into `agentic-memory/CHANGELOG.md`.
2. Archives completed plans and code reviews for closed issues into `ARCHIVE_INDEX.md`.
3. Prunes obsolete audit reports, keeping only the latest N runs.
4. Strictly preserves core invariant rules (`rules/`), `README.md`, `CHANGELOG.md`, and active issue artifacts.

Usage:
    python3 .agents/scripts/consolidate_memory.py [--dry-run] [--retention-days N] [--keep-audits N] [--all-closed] [--json]
"""

import argparse
from dataclasses import dataclass, field
from datetime import datetime, timezone, timedelta
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
logger = logging.getLogger("memory_consolidator")

PROTECTED_DIRS = {"rules"}
PROTECTED_FILES = {"README.md", "CHANGELOG.md", "ARCHIVE_INDEX.md"}


@dataclass
class PruneCandidate:
    path: str
    category: str
    file_name: str
    size_bytes: int
    date_str: Optional[str]
    issue_number: Optional[int]
    title: str
    reason: str


@dataclass
class ConsolidationReport:
    total_files_scanned: int = 0
    total_bytes_scanned: int = 0
    changelogs_consolidated: int = 0
    plans_pruned: int = 0
    reviews_pruned: int = 0
    audits_pruned: int = 0
    designs_pruned: int = 0
    total_files_pruned: int = 0
    total_bytes_pruned: int = 0
    archived_records: List[Dict[str, Any]] = field(default_factory=list)
    pruned_files: List[str] = field(default_factory=list)
    skipped_files: List[str] = field(default_factory=list)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "total_files_scanned": self.total_files_scanned,
            "total_bytes_scanned": self.total_bytes_scanned,
            "changelogs_consolidated": self.changelogs_consolidated,
            "plans_pruned": self.plans_pruned,
            "reviews_pruned": self.reviews_pruned,
            "audits_pruned": self.audits_pruned,
            "designs_pruned": self.designs_pruned,
            "total_files_pruned": self.total_files_pruned,
            "total_bytes_pruned": self.total_bytes_pruned,
            "savings_kb": round(self.total_bytes_pruned / 1024, 2),
            "archived_records_count": len(self.archived_records),
            "pruned_files": self.pruned_files,
            "skipped_files": self.skipped_files,
        }


def get_repo_root() -> str:
    """Returns absolute path to the repository root."""
    script_dir = os.path.dirname(os.path.abspath(__file__))
    return os.path.abspath(os.path.join(script_dir, "..", ".."))


def fetch_github_issue_states() -> Dict[int, str]:
    """Queries GitHub CLI for issue states (OPEN / CLOSED). Falls back gracefully."""
    issue_states: Dict[int, str] = {}
    try:
        cmd = ["gh", "issue", "list", "--state", "all", "--json", "number,state", "--limit", "300"]
        res = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=10)
        if res.returncode == 0:
            data = json.loads(res.stdout)
            for item in data:
                issue_states[int(item["number"])] = item.get("state", "UNKNOWN").upper()
    except Exception as e:
        logger.warning(f"Could not query GitHub CLI for issue states ({e}); falling back to age/heuristic.")
    return issue_states


def parse_file_metadata(file_path: str) -> Tuple[Optional[str], Optional[int], str]:
    """Extracts date, issue number, and title/heading from markdown file."""
    file_name = os.path.basename(file_path)
    
    # 1. Date extraction (YYYY-MM-DD)
    date_match = re.match(r"^(\d{4}-\d{2}-\d{2})", file_name)
    date_str = date_match.group(1) if date_match else None

    # 2. Issue ID extraction
    issue_match = re.search(r"issue-(\d+)", file_name, re.IGNORECASE)
    if not issue_match:
        issue_match = re.search(r"pr-(\d+)", file_name, re.IGNORECASE)
    issue_number = int(issue_match.group(1)) if issue_match else None

    # 3. Title extraction from first markdown heading
    title = file_name
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            for line in f:
                line_str = line.strip()
                if line_str.startswith("#"):
                    title = re.sub(r"^#+\s*", "", line_str).strip()
                    break
    except Exception:
        pass

    return date_str, issue_number, title


def parse_date(date_str: str) -> Optional[datetime]:
    try:
        return datetime.strptime(date_str, "%Y-%m-%d").replace(tzinfo=timezone.utc)
    except Exception:
        return None


def consolidate_task_changelog(
    changelog_path: str,
    master_changelog_path: str,
    dry_run: bool = False
) -> bool:
    """
    Checks if task changelog is already recorded in master CHANGELOG.md.
    If missing, synthesizes and appends the entry.
    """
    if not os.path.exists(changelog_path) or not os.path.exists(master_changelog_path):
        return False

    with open(changelog_path, "r", encoding="utf-8") as f:
        task_content = f.read()

    with open(master_changelog_path, "r", encoding="utf-8") as f:
        master_content = f.read()

    # Extract issue reference or title
    issue_match = re.search(r"(?:Issue|Task|PR)[^\n\d]*#?(\d+)", task_content, re.IGNORECASE)
    issue_num = int(issue_match.group(1)) if issue_match else None

    # Check if already present
    if issue_num and f"[#{issue_num}]" in master_content:
        logger.debug(f"Changelog for #{issue_num} already exists in master CHANGELOG.md")
        return True

    # If missing, extract structured fields and append
    summary_match = re.search(r"##\s*1\.\s*Summary[^\n]*\n([\s\S]*?)(?=\n##|\Z)", task_content)
    decisions_match = re.search(r"##\s*2\.\s*Key Technical Decisions[^\n]*\n([\s\S]*?)(?=\n##|\Z)", task_content)
    impact_match = re.search(r"##\s*3\.\s*Impacted Files[^\n]*\n([\s\S]*?)(?=\n##|\Z)", task_content)
    verify_match = re.search(r"##\s*4\.\s*Verification[^\n]*\n([\s\S]*?)(?=\n##|\Z)", task_content)

    title_match = re.search(r"^#\s*(?:Task Changelog:\s*)?([^\n]+)", task_content, re.MULTILINE)
    title = title_match.group(1).strip() if title_match else "Feature Update"

    date_str, _, _ = parse_file_metadata(changelog_path)
    entry_date = date_str or datetime.now(timezone.utc).strftime("%Y-%m-%d")

    tag = f"[#{issue_num}]" if issue_num else "[Task]"
    entry_lines = [
        f"\n### {tag} {title}",
        f"- **Date**: {entry_date} | **Scope**: `packages/`, `apps/`"
    ]

    if summary_match:
        summary_clean = summary_match.group(1).strip().replace("\n", " ")
        entry_lines.append(f"- **Summary**: {summary_clean}")

    if decisions_match:
        entry_lines.append("- **Key Technical Decisions**:")
        for line in decisions_match.group(1).strip().splitlines():
            line_s = line.strip()
            if line_s:
                entry_lines.append(f"  {line_s}")

    if impact_match:
        files = [re.sub(r"^[-*]\s*", "", line.strip()) for line in impact_match.group(1).strip().splitlines() if line.strip()]
        if files:
            entry_lines.append(f"- **Impacted Files**: {', '.join(files[:6])}")

    if verify_match:
        cmd_lines = [line.strip() for line in verify_match.group(1).strip().splitlines() if line.strip()]
        if cmd_lines:
            entry_lines.append(f"- **Verification**: {'; '.join(cmd_lines[:3])}")

    new_entry = "\n".join(entry_lines) + "\n"

    logger.info(f"Consolidating unrecorded changelog into master CHANGELOG.md: {tag} {title}")
    if not dry_run:
        with open(master_changelog_path, "a", encoding="utf-8") as f:
            f.write(new_entry)

    return True


def update_archive_index(
    index_path: str,
    archived_items: List[PruneCandidate],
    dry_run: bool = False
) -> None:
    """Appends pruned plans and reviews to ARCHIVE_INDEX.md for perpetual traceability."""
    if not archived_items:
        return

    existing_content = ""
    if os.path.exists(index_path):
        with open(index_path, "r", encoding="utf-8") as f:
            existing_content = f.read()
    else:
        existing_content = (
            "# Agentic Memory Archive Index\n\n"
            "This index maintains a compact, permanent reference log of completed task plans and reviews "
            "that were synthesized into `agentic-memory/CHANGELOG.md` and safely pruned from the active tree "
            "to prevent repository bloat.\n\n"
            "| Date | Category | Issue / Ref | Title / Objective | Archived On |\n"
            "|---|---|---|---|---|\n"
        )

    new_rows = []
    now_str = datetime.now(timezone.utc).strftime("%Y-%m-%d")

    for item in archived_items:
        issue_ref = f"#{item.issue_number}" if item.issue_number else "-"
        date_display = item.date_str or "-"
        clean_title = item.title.replace("|", "/").strip()
        row_signature = f"| {date_display} | {item.category} | {issue_ref} |"
        
        if row_signature not in existing_content:
            new_rows.append(f"| {date_display} | `{item.category}` | {issue_ref} | {clean_title} | {now_str} |")

    if new_rows:
        logger.info(f"Adding {len(new_rows)} records to {os.path.basename(index_path)}")
        if not dry_run:
            with open(index_path, "w", encoding="utf-8") as f:
                if not existing_content.endswith("\n"):
                    existing_content += "\n"
                f.write(existing_content + "\n".join(new_rows) + "\n")


def scan_and_consolidate(
    memory_root: str,
    retention_days: int = 14,
    keep_audits: int = 2,
    all_closed: bool = False,
    prune_designs: bool = False,
    dry_run: bool = False
) -> ConsolidationReport:
    """Core engine evaluating agentic-memory files, consolidating changelogs, and pruning stale files."""
    report = ConsolidationReport()
    now = datetime.now(timezone.utc)
    retention_cutoff = now - timedelta(days=retention_days)

    issue_states = fetch_github_issue_states()
    master_changelog = os.path.join(memory_root, "CHANGELOG.md")
    archive_index = os.path.join(memory_root, "ARCHIVE_INDEX.md")

    candidates_to_prune: List[PruneCandidate] = []
    archived_for_index: List[PruneCandidate] = []

    # 1. Process changelogs/
    changelogs_dir = os.path.join(memory_root, "changelogs")
    if os.path.isdir(changelogs_dir):
        for fname in sorted(os.listdir(changelogs_dir)):
            if fname in PROTECTED_FILES or fname.startswith("."):
                continue
            fpath = os.path.join(changelogs_dir, fname)
            if not os.path.isfile(fpath):
                continue

            sz = os.path.getsize(fpath)
            report.total_files_scanned += 1
            report.total_bytes_scanned += sz

            date_str, issue_num, title = parse_file_metadata(fpath)
            # Ensure it is in master CHANGELOG.md
            consolidated = consolidate_task_changelog(fpath, master_changelog, dry_run=dry_run)
            if consolidated:
                report.changelogs_consolidated += 1
                cand = PruneCandidate(
                    path=fpath,
                    category="changelogs",
                    file_name=fname,
                    size_bytes=sz,
                    date_str=date_str,
                    issue_number=issue_num,
                    title=title,
                    reason="Consolidated into master CHANGELOG.md"
                )
                candidates_to_prune.append(cand)

    # 2. Process audits/
    audits_dir = os.path.join(memory_root, "audits")
    if os.path.isdir(audits_dir):
        audit_files = []
        for fname in os.listdir(audits_dir):
            if fname in PROTECTED_FILES or fname.startswith("."):
                continue
            fpath = os.path.join(audits_dir, fname)
            if os.path.isfile(fpath):
                sz = os.path.getsize(fpath)
                report.total_files_scanned += 1
                report.total_bytes_scanned += sz
                date_str, _, title = parse_file_metadata(fpath)
                audit_files.append((date_str or "0000-00-00", fpath, fname, sz, title))

        # Sort newest to oldest
        audit_files.sort(key=lambda x: x[0], reverse=True)
        # Keep latest `keep_audits`, prune older ones
        for idx, (dt_str, fpath, fname, sz, title) in enumerate(audit_files):
            if idx >= keep_audits:
                cand = PruneCandidate(
                    path=fpath,
                    category="audits",
                    file_name=fname,
                    size_bytes=sz,
                    date_str=dt_str,
                    issue_number=None,
                    title=title,
                    reason=f"Routine audit superseded by {keep_audits} newer runs"
                )
                candidates_to_prune.append(cand)
                report.audits_pruned += 1

    # 3. Process plans/ and reviews/
    for category in ["plans", "reviews"]:
        cat_dir = os.path.join(memory_root, category)
        if not os.path.isdir(cat_dir):
            continue

        for fname in sorted(os.listdir(cat_dir)):
            if fname in PROTECTED_FILES or fname.startswith("."):
                continue
            fpath = os.path.join(cat_dir, fname)
            if not os.path.isfile(fpath):
                continue

            sz = os.path.getsize(fpath)
            report.total_files_scanned += 1
            report.total_bytes_scanned += sz

            date_str, issue_num, title = parse_file_metadata(fpath)
            file_date = parse_date(date_str) if date_str else None

            # Protection checks
            is_open = False
            is_closed = False
            if issue_num and issue_num in issue_states:
                st = issue_states[issue_num]
                if st == "OPEN":
                    is_open = True
                elif st == "CLOSED":
                    is_closed = True

            # If issue is known to be open, never prune
            if is_open:
                report.skipped_files.append(f"{category}/{fname} (active issue #{issue_num})")
                continue

            # Pruning eligibility
            should_prune = False
            reason = ""

            if all_closed and is_closed:
                should_prune = True
                reason = f"Closed issue #{issue_num} (--all-closed)"
            elif file_date and file_date < retention_cutoff:
                if is_closed or not issue_num:
                    should_prune = True
                    reason = f"Exceeded retention ({retention_days}d cutoff: {retention_cutoff.strftime('%Y-%m-%d')})"
                else:
                    report.skipped_files.append(f"{category}/{fname} (issue status unconfirmed)")
            elif is_closed and (not file_date or file_date < retention_cutoff):
                should_prune = True
                reason = f"Closed issue #{issue_num} older than retention"

            if should_prune:
                cand = PruneCandidate(
                    path=fpath,
                    category=category,
                    file_name=fname,
                    size_bytes=sz,
                    date_str=date_str,
                    issue_number=issue_num,
                    title=title,
                    reason=reason
                )
                candidates_to_prune.append(cand)
                archived_for_index.append(cand)
                if category == "plans":
                    report.plans_pruned += 1
                elif category == "reviews":
                    report.reviews_pruned += 1
            else:
                report.skipped_files.append(f"{category}/{fname} (within retention window)")

    # 4. Process designs/ if prune_designs enabled
    designs_dir = os.path.join(memory_root, "designs")
    if prune_designs and os.path.isdir(designs_dir):
        for fname in sorted(os.listdir(designs_dir)):
            if fname in PROTECTED_FILES or fname.startswith("."):
                continue
            fpath = os.path.join(designs_dir, fname)
            if not os.path.isfile(fpath):
                continue
            sz = os.path.getsize(fpath)
            report.total_files_scanned += 1
            report.total_bytes_scanned += sz
            date_str, issue_num, title = parse_file_metadata(fpath)
            file_date = parse_date(date_str) if date_str else None
            is_closed = (issue_num in issue_states and issue_states[issue_num] == "CLOSED")

            if (all_closed and is_closed) or (file_date and file_date < retention_cutoff):
                cand = PruneCandidate(
                    path=fpath,
                    category="designs",
                    file_name=fname,
                    size_bytes=sz,
                    date_str=date_str,
                    issue_number=issue_num,
                    title=title,
                    reason="Design pruned by explicit flag"
                )
                candidates_to_prune.append(cand)
                archived_for_index.append(cand)
                report.designs_pruned += 1

    # 5. Record Archive Index
    update_archive_index(archive_index, archived_for_index, dry_run=dry_run)
    report.archived_records = [
        {
            "category": c.category,
            "file": c.file_name,
            "issue": c.issue_number,
            "title": c.title,
            "size": c.size_bytes
        } for c in archived_for_index
    ]

    # 6. Perform file deletions
    for cand in candidates_to_prune:
        report.total_files_pruned += 1
        report.total_bytes_pruned += cand.size_bytes
        report.pruned_files.append(f"{cand.category}/{cand.file_name}")

        if dry_run:
            logger.info(f"[DRY RUN] Would delete: {cand.category}/{cand.file_name} ({cand.size_bytes}B) - {cand.reason}")
        else:
            try:
                os.remove(cand.path)
                logger.info(f"Deleted: {cand.category}/{cand.file_name} ({cand.size_bytes}B)")
            except Exception as e:
                logger.error(f"Failed to delete {cand.path}: {e}")

    return report


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Agentic Memory Consolidation and Pruning Engine for Hexta."
    )
    parser.add_argument(
        "--retention-days",
        type=int,
        default=14,
        help="Retention window in days for completed plans and reviews (default: 14)."
    )
    parser.add_argument(
        "--keep-audits",
        type=int,
        default=2,
        help="Number of latest audit reports to preserve in audits/ (default: 2)."
    )
    parser.add_argument(
        "--all-closed",
        action="store_true",
        help="Prune all artifacts tied to closed GitHub issues regardless of retention days."
    )
    parser.add_argument(
        "--prune-designs",
        action="store_true",
        help="Include stale designs in pruning (default: false, designs are kept)."
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Simulate operations without deleting files or modifying documents."
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="Output consolidation results as formatted JSON."
    )

    args = parser.parse_args()
    repo_root = get_repo_root()
    memory_root = os.path.join(repo_root, "agentic-memory")

    if not os.path.isdir(memory_root):
        logger.error(f"agentic-memory directory not found at: {memory_root}")
        return 1

    logger.info(
        f"Starting Agentic Memory Consolidation (dry_run={args.dry_run}, "
        f"retention={args.retention_days}d, keep_audits={args.keep_audits}, all_closed={args.all_closed})..."
    )

    report = scan_and_consolidate(
        memory_root=memory_root,
        retention_days=args.retention_days,
        keep_audits=args.keep_audits,
        all_closed=args.all_closed,
        prune_designs=args.prune_designs,
        dry_run=args.dry_run
    )

    if args.json:
        print(json.dumps(report.to_dict(), indent=2))
        return 0

    print("\n" + "=" * 60)
    print(f" AGENTIC MEMORY CONSOLIDATION REPORT {'[DRY RUN]' if args.dry_run else '[EXECUTED]'}")
    print("=" * 60)
    print(f" Files Scanned:           {report.total_files_scanned} ({report.total_bytes_scanned / 1024:.1f} KB)")
    print(f" Changelogs Consolidated: {report.changelogs_consolidated}")
    print(f" Audits Pruned:           {report.audits_pruned}")
    print(f" Plans Pruned:            {report.plans_pruned}")
    print(f" Reviews Pruned:          {report.reviews_pruned}")
    if args.prune_designs:
        print(f" Designs Pruned:          {report.designs_pruned}")
    print("-" * 60)
    print(f" Total Files Pruned:      {report.total_files_pruned}")
    print(f" Storage Reclaimed:       {report.total_bytes_pruned / 1024:.1f} KB")
    print(f" Archived Records Logged: {len(report.archived_records)} -> ARCHIVE_INDEX.md")
    print(f" Active Files Retained:   {report.total_files_scanned - report.total_files_pruned}")
    print("=" * 60 + "\n")

    return 0


if __name__ == "__main__":
    sys.exit(main())
