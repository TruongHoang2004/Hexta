#!/usr/bin/env python3
"""
test_get_next_issue.py - Unit tests for get_next_issue.py
"""

import unittest
from unittest.mock import MagicMock, patch

from get_next_issue import (
    claim_issue,
    get_priority_weight,
    matches_domain,
    slugify
)


class TestGetNextIssue(unittest.TestCase):

    def test_slugify(self):
        self.assertEqual(slugify("feat(api): Add Refresh Token"), "feat-api-add-refresh-token")
        self.assertEqual(slugify("Fix Bug #123!"), "fix-bug-123")

    def test_priority_weights(self):
        self.assertEqual(get_priority_weight([{"name": "ready"}, {"name": "priority:high"}]), 1)
        self.assertEqual(get_priority_weight([{"name": "ready"}, {"name": "priority:medium"}]), 2)
        self.assertEqual(get_priority_weight([{"name": "ready"}, {"name": "priority:low"}]), 3)
        self.assertEqual(get_priority_weight([{"name": "ready"}]), 4)

        # Ineligible
        self.assertEqual(get_priority_weight([{"name": "ready"}, {"name": "in-progress"}]), -1)
        self.assertEqual(get_priority_weight([{"name": "ready"}, {"name": "blocked"}]), -1)
        self.assertEqual(get_priority_weight([{"name": "priority:high"}]), -1)  # Missing 'ready'

    def test_matches_domain(self):
        labels_backend = [{"name": "ready"}, {"name": "domain:backend-core"}]
        labels_frontend = [{"name": "ready"}, {"name": "domain:frontend-ui"}]
        labels_untagged = [{"name": "ready"}]

        # All / None matches everything
        self.assertTrue(matches_domain(labels_backend, None))
        self.assertTrue(matches_domain(labels_backend, "all"))

        # Domain specific
        self.assertTrue(matches_domain(labels_backend, "backend-core"))
        self.assertTrue(matches_domain(labels_backend, "domain:backend-core"))
        self.assertFalse(matches_domain(labels_backend, "frontend-ui"))

        self.assertTrue(matches_domain(labels_frontend, "frontend-ui"))
        self.assertFalse(matches_domain(labels_frontend, "backend-core"))

        # Untagged
        self.assertTrue(matches_domain(labels_untagged, "untagged"))
        self.assertFalse(matches_domain(labels_untagged, "backend-core"))

    @patch("subprocess.run")
    def test_claim_issue(self, mock_run):
        mock_run.return_value = MagicMock(returncode=0)

        # Test Dry Run
        success = claim_issue(42, "backend_core", dry_run=True)
        self.assertTrue(success)
        mock_run.assert_not_called()

        # Test Actual Run
        success = claim_issue(42, "backend_core", dry_run=False)
        self.assertTrue(success)
        mock_run.assert_called_once()
        args = mock_run.call_args[0][0]
        self.assertIn("claimed-by:backend_core", args[7])


if __name__ == "__main__":
    unittest.main()
