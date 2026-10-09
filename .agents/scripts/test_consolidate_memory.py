#!/usr/bin/env python3
"""Unit tests for consolidate_memory.py"""

import os
import shutil
import tempfile
import unittest
from unittest.mock import patch

from consolidate_memory import (
    parse_file_metadata,
    parse_date,
    consolidate_task_changelog,
    update_archive_index,
    scan_and_consolidate,
    PruneCandidate,
)


class TestConsolidateMemory(unittest.TestCase):
    def setUp(self):
        self.test_dir = tempfile.mkdtemp()
        self.memory_dir = os.path.join(self.test_dir, "agentic-memory")
        os.makedirs(self.memory_dir, exist_ok=True)
        os.makedirs(os.path.join(self.memory_dir, "rules"), exist_ok=True)
        os.makedirs(os.path.join(self.memory_dir, "plans"), exist_ok=True)
        os.makedirs(os.path.join(self.memory_dir, "reviews"), exist_ok=True)
        os.makedirs(os.path.join(self.memory_dir, "audits"), exist_ok=True)
        os.makedirs(os.path.join(self.memory_dir, "changelogs"), exist_ok=True)

        # Create master CHANGELOG.md and README.md
        self.master_changelog = os.path.join(self.memory_dir, "CHANGELOG.md")
        with open(self.master_changelog, "w", encoding="utf-8") as f:
            f.write("# Master Changelog\n\n### [#1] Initial Setup\n- Summary: Done\n")

        self.master_readme = os.path.join(self.memory_dir, "README.md")
        with open(self.master_readme, "w", encoding="utf-8") as f:
            f.write("# Agentic Memory\n")

    def tearDown(self):
        shutil.rmtree(self.test_dir)

    def test_parse_file_metadata(self):
        sample_path = os.path.join(self.memory_dir, "plans", "2026-09-12_issue-4-security-auth_plan.md")
        with open(sample_path, "w", encoding="utf-8") as f:
            f.write("# Implementation Plan: Security Authentication\nSome details\n")

        dt_str, issue_num, title = parse_file_metadata(sample_path)
        self.assertEqual(dt_str, "2026-09-12")
        self.assertEqual(issue_num, 4)
        self.assertEqual(title, "Implementation Plan: Security Authentication")

    def test_consolidate_task_changelog_unrecorded(self):
        task_cl = os.path.join(self.memory_dir, "changelogs", "2026-10-01_issue-99_changelog.md")
        with open(task_cl, "w", encoding="utf-8") as f:
            f.write("# Task Changelog: Add Rate Limiting\n\nIssue: #99\n\n## 1. Summary\nAdded rate limiter.\n")

        consolidated = consolidate_task_changelog(task_cl, self.master_changelog, dry_run=False)
        self.assertTrue(consolidated)

        with open(self.master_changelog, "r", encoding="utf-8") as f:
            content = f.read()
        self.assertIn("[#99]", content)
        self.assertIn("Add Rate Limiting", content)

    def test_consolidate_task_changelog_already_recorded(self):
        task_cl = os.path.join(self.memory_dir, "changelogs", "2026-09-10_issue-1_changelog.md")
        with open(task_cl, "w", encoding="utf-8") as f:
            f.write("# Task Changelog\nIssue: #1\n")

        consolidated = consolidate_task_changelog(task_cl, self.master_changelog, dry_run=False)
        self.assertTrue(consolidated)

        # Confirm not duplicated
        with open(self.master_changelog, "r", encoding="utf-8") as f:
            content = f.read()
        self.assertEqual(content.count("[#1]"), 1)

    def test_update_archive_index(self):
        index_file = os.path.join(self.memory_dir, "ARCHIVE_INDEX.md")
        candidates = [
            PruneCandidate(
                path="dummy",
                category="plans",
                file_name="2026-09-12_issue-5_plan.md",
                size_bytes=1000,
                date_str="2026-09-12",
                issue_number=5,
                title="Plan 5",
                reason="Test"
            )
        ]
        update_archive_index(index_file, candidates, dry_run=False)
        self.assertTrue(os.path.exists(index_file))

        with open(index_file, "r", encoding="utf-8") as f:
            content = f.read()
        self.assertIn("| 2026-09-12 | `plans` | #5 | Plan 5 |", content)

    @patch("consolidate_memory.fetch_github_issue_states")
    def test_scan_and_consolidate_pruning_behavior(self, mock_issues):
        mock_issues.return_value = {4: "CLOSED", 28: "OPEN"}

        # 1. Closed plan older than 14 days -> should be pruned
        old_closed_plan = os.path.join(self.memory_dir, "plans", "2026-09-12_issue-4-auth_plan.md")
        with open(old_closed_plan, "w", encoding="utf-8") as f:
            f.write("# Plan 4\n")

        # 2. Open issue plan -> MUST BE PROTECTED (skipped)
        open_plan = os.path.join(self.memory_dir, "plans", "2026-09-12_issue-28-arch_plan.md")
        with open(open_plan, "w", encoding="utf-8") as f:
            f.write("# Plan 28\n")

        # 3. Multiple audits -> only keep latest 2
        for d in ["2026-09-01", "2026-09-08", "2026-09-15"]:
            aud_path = os.path.join(self.memory_dir, "audits", f"{d}_codebase-audit.md")
            with open(aud_path, "w", encoding="utf-8") as f:
                f.write(f"# Audit {d}\n")

        # Run consolidation
        report = scan_and_consolidate(
            memory_root=self.memory_dir,
            retention_days=14,
            keep_audits=2,
            all_closed=False,
            dry_run=False
        )

        self.assertEqual(report.plans_pruned, 1)
        self.assertFalse(os.path.exists(old_closed_plan))
        # Open issue must still exist
        self.assertTrue(os.path.exists(open_plan))

        # Audits: 2026-09-01 should be pruned (oldest of 3), 2026-09-08 and 2026-09-15 kept
        self.assertEqual(report.audits_pruned, 1)
        self.assertFalse(os.path.exists(os.path.join(self.memory_dir, "audits", "2026-09-01_codebase-audit.md")))
        self.assertTrue(os.path.exists(os.path.join(self.memory_dir, "audits", "2026-09-15_codebase-audit.md")))
        self.assertTrue(os.path.exists(os.path.join(self.memory_dir, "audits", "2026-09-08_codebase-audit.md")))

    @patch("consolidate_memory.fetch_github_issue_states")
    def test_dry_run_does_not_delete(self, mock_issues):
        mock_issues.return_value = {4: "CLOSED"}
        old_plan = os.path.join(self.memory_dir, "plans", "2026-09-12_issue-4-auth_plan.md")
        with open(old_plan, "w", encoding="utf-8") as f:
            f.write("# Plan 4\n")

        report = scan_and_consolidate(
            memory_root=self.memory_dir,
            retention_days=14,
            keep_audits=2,
            all_closed=False,
            dry_run=True
        )

        self.assertEqual(report.plans_pruned, 1)
        # In dry run, file must NOT be deleted
        self.assertTrue(os.path.exists(old_plan))


if __name__ == "__main__":
    unittest.main()
