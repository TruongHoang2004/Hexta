#!/usr/bin/env python3
"""Unit tests for manage_issue_lifecycle.py"""

import unittest
from datetime import datetime, timedelta, timezone
from unittest.mock import MagicMock, patch

from manage_issue_lifecycle import IssueLifecycleManager, ActionRecord


class TestManageIssueLifecycle(unittest.TestCase):
    def setUp(self):
        self.manager = IssueLifecycleManager(dry_run=True, verbose=False, stale_hours=2.0)

    def test_extract_referenced_issues(self):
        # 1. From head branch
        pr1 = {
            "headRefName": "task/issue-22-feat-automation",
            "title": "feat(automation): issue lifecycle",
            "body": "Nothing special"
        }
        self.assertEqual(self.manager.extract_referenced_issues(pr1), {22})

        # 2. From body closing keyword
        pr2 = {
            "headRefName": "some-branch",
            "title": "Fix bug",
            "body": "Closes #15 and resolves #18"
        }
        self.assertEqual(self.manager.extract_referenced_issues(pr2), {15, 18})

        # 3. From title and refs
        pr3 = {
            "headRefName": "patch-1",
            "title": "security(auth): patch vulnerability (refs #4)",
            "body": "See #7 for details"
        }
        self.assertEqual(self.manager.extract_referenced_issues(pr3), {4, 7})

    def test_infer_priority(self):
        self.assertEqual(
            self.manager._infer_priority("security(auth): patch CVE-2026-001", "High risk vulnerability"),
            "priority:high"
        )
        self.assertEqual(
            self.manager._infer_priority("docs: update README", "Improve documentation"),
            "priority:low"
        )
        self.assertEqual(
            self.manager._infer_priority("feat(api): add multi-tenant support", "Standard feature"),
            "priority:medium"
        )

    def test_sync_closed_issues_strips_stale_labels(self):
        mock_issues = [
            {
                "number": 10,
                "title": "feat(audit): audit",
                "state": "CLOSED",
                "labels": [{"name": "in-review"}, {"name": "priority:medium"}],
                "body": "Done"
            },
            {
                "number": 1,
                "title": "feat(config): ai config",
                "state": "CLOSED",
                "labels": [{"name": "in-progress"}],
                "body": "Done"
            }
        ]
        mock_prs = []

        with patch.object(self.manager, "fetch_all_issues", return_value=mock_issues), \
             patch.object(self.manager, "fetch_all_prs", return_value=mock_prs):
            mutations = self.manager.sync()
            self.assertEqual(mutations, 2)
            self.assertEqual(len(self.manager.actions), 2)
            
            # Action 1: issue 10 removes in-review
            self.assertEqual(self.manager.actions[0].issue_number, 10)
            self.assertEqual(self.manager.actions[0].labels_removed, ["in-review"])
            
            # Action 2: issue 1 removes in-progress
            self.assertEqual(self.manager.actions[1].issue_number, 1)
            self.assertEqual(self.manager.actions[1].labels_removed, ["in-progress"])

    def test_sync_open_issue_with_open_pr(self):
        mock_issues = [
            {
                "number": 20,
                "title": "refactor(go): migrate paths",
                "state": "OPEN",
                "labels": [{"name": "in-progress"}, {"name": "priority:high"}],
                "body": "Refactoring"
            }
        ]
        mock_prs = [
            {
                "number": 23,
                "headRefName": "task/issue-20-refactor-go",
                "state": "OPEN",
                "title": "refactor(go): migrate paths",
                "body": "Closes #20"
            }
        ]

        with patch.object(self.manager, "fetch_all_issues", return_value=mock_issues), \
             patch.object(self.manager, "fetch_all_prs", return_value=mock_prs):
            mutations = self.manager.sync()
            self.assertEqual(mutations, 1)
            self.assertEqual(self.manager.actions[0].issue_number, 20)
            self.assertIn("in-review", self.manager.actions[0].labels_added)
            self.assertIn("in-progress", self.manager.actions[0].labels_removed)

    def test_triage_promotes_qualified_issue(self):
        mock_issues = [
            {
                "number": 99,
                "title": "feat(search): implement semantic product search",
                "state": "OPEN",
                "labels": [],
                "body": "### Summary\nImplement vector search for products.\n\n### Acceptance Criteria\n- [ ] Search endpoint returns results"
            },
            {
                "number": 100,
                "title": "too short",
                "state": "OPEN",
                "labels": [],
                "body": "tiny"
            }
        ]

        with patch.object(self.manager, "fetch_all_issues", return_value=mock_issues):
            promoted = self.manager.triage()
            self.assertEqual(promoted, 1)
            self.assertEqual(self.manager.actions[0].issue_number, 99)
            self.assertIn("ready", self.manager.actions[0].labels_added)
            self.assertIn("priority:medium", self.manager.actions[0].labels_added)

    def test_recover_stale_unlocks_inactive_in_progress(self):
        three_hours_ago = (datetime.now(timezone.utc) - timedelta(hours=3.5)).isoformat()
        mock_issues = [
            {
                "number": 50,
                "title": "feat(abandoned): abandoned task",
                "state": "OPEN",
                "labels": [{"name": "in-progress"}],
                "updatedAt": three_hours_ago
            }
        ]
        mock_prs = []  # No active PR

        with patch.object(self.manager, "fetch_all_issues", return_value=mock_issues), \
             patch.object(self.manager, "fetch_all_prs", return_value=mock_prs):
            recovered = self.manager.recover_stale()
            self.assertEqual(recovered, 1)
            self.assertEqual(self.manager.actions[0].issue_number, 50)
            self.assertIn("ready", self.manager.actions[0].labels_added)
            self.assertIn("in-progress", self.manager.actions[0].labels_removed)
            self.assertIn("Autonomous Task Runner State Recovery", self.manager.actions[0].comment)


if __name__ == "__main__":
    unittest.main()
