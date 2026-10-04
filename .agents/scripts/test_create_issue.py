#!/usr/bin/env python3
"""Unit tests for create_issue.py"""

import unittest
from unittest.mock import MagicMock, patch

from create_issue import (
    IssueDraft,
    build_issue_body,
    resolve_labels,
    execute_gh_create,
)


class TestCreateIssue(unittest.TestCase):
    def test_issue_draft_validation_success(self):
        draft = IssueDraft(
            title="feat(api): implement email verification",
            body="## Summary\nImplement email verification endpoint.\n\n## Acceptance Criteria\n- [ ] Send token email\n- [ ] Verify endpoint returns 200",
            labels=["ready", "priority:medium", "enhancement"]
        )
        valid, errors = draft.validate()
        self.assertTrue(valid)
        self.assertEqual(len(errors), 0)

    def test_issue_draft_validation_short_title(self):
        draft = IssueDraft(
            title="short",
            body="## Summary\nLong enough body with acceptance criteria.\n\n## Acceptance Criteria\n- [ ] Criteria",
            labels=["ready"]
        )
        valid, errors = draft.validate()
        self.assertFalse(valid)
        self.assertTrue(any("Title length" in e for e in errors))

    def test_issue_draft_validation_missing_checkbox(self):
        draft = IssueDraft(
            title="feat(api): valid title for feature",
            body="## Summary\nLong enough body without acceptance criteria checkboxes in markdown format.",
            labels=["ready"]
        )
        valid, errors = draft.validate()
        self.assertFalse(valid)
        self.assertTrue(any("acceptance criteria checklist" in e for e in errors))

    def test_build_issue_body(self):
        body = build_issue_body(
            summary="Add refresh token rotation to enhance security.",
            requirements=["Generate new refresh token on rotation", "Invalidate old family"],
            acceptance_criteria=["Return 401 if compromised", "Pass test cases"],
            technical_notes="Modify internal/core/service/auth_service.go",
            related_files=["services/api/internal/core/service/auth_service.go"]
        )
        self.assertIn("## Summary & Context", body)
        self.assertIn("## Requirements & Scope", body)
        self.assertIn("## Acceptance Criteria", body)
        self.assertIn("- [ ] Return 401 if compromised", body)
        self.assertIn("services/api/internal/core/service/auth_service.go", body)

    def test_resolve_labels(self):
        labels = resolve_labels("feat", "high", mark_ready=True, extra_labels=["backend"])
        self.assertEqual(labels, ["backend", "enhancement", "priority:high", "ready"])

        labels_no_ready = resolve_labels("fix", "low", mark_ready=False)
        self.assertEqual(labels_no_ready, ["bug", "priority:low"])

    @patch("subprocess.run")
    def test_execute_gh_create_mock(self, mock_run):
        mock_proc = MagicMock()
        mock_proc.stdout = "https://github.com/TruongHoang2004/Hexta/issues/99\n"
        mock_run.return_value = mock_proc

        draft = IssueDraft(
            title="feat(auth): token revocation endpoint",
            body="## Summary\nEndpoint for revoking access tokens.\n\n## Acceptance Criteria\n- [ ] Revoke token\n- [ ] Return 200",
            labels=["ready", "priority:high"]
        )
        res = execute_gh_create(draft, dry_run=False)
        self.assertEqual(res["status"], "created")
        self.assertEqual(res["issue_number"], 99)
        self.assertEqual(res["url"], "https://github.com/TruongHoang2004/Hexta/issues/99")


if __name__ == "__main__":
    unittest.main()
