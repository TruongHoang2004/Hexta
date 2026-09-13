#!/usr/bin/env python3
"""
Unit tests for the Autonomous PR Review Agent (pr_review_agent.py).
Tests risk classification, evaluation logic, critical path detection, and CLI helpers.
"""

import unittest
from unittest.mock import MagicMock, patch

from pr_review_agent import (
    ActionDecision,
    PRDetails,
    PRReviewAgent,
    RiskLevel,
    TRIVIAL_CONFLICT_FILES,
    classify_pr_risk,
)


class TestPRReviewAgentRiskClassification(unittest.TestCase):
    def test_pr_19_docs_only_low_risk(self):
        """PR #19: Documentation and agentic-memory files only."""
        files = [
            "README.md",
            "agentic-memory/plans/2026-09-12_issue-18-local-dev-guide_plan.md",
            "agentic-memory/designs/2026-09-12_issue-18-local-dev-guide_design.md",
            "agentic-memory/reviews/2026-09-12_issue-18-local-dev-guide_review.md",
            "agentic-memory/changelogs/2026-09-12_issue-18-local-dev-guide_changelog.md",
        ]
        risk, reasons = classify_pr_risk(files, title="docs: create comprehensive local development guide")
        self.assertEqual(risk, RiskLevel.LOW)
        self.assertTrue(any("Documentation" in r for r in reasons))

    def test_pr_25_readmes_and_docs_low_risk(self):
        """PR #25: Monorepo documentation and template replacements."""
        files = [
            "README.md",
            "GEMINI.md",
            "packages/shared/README.md",
            "infrastructure/README.md",
            "test/README.md",
            "agentic-memory/rules/development-rules.md",
        ]
        risk, _ = classify_pr_risk(files, title="docs: replace legacy gitlab references and template readmes")
        self.assertEqual(risk, RiskLevel.LOW)

    def test_pr_9_ci_workflows_low_risk(self):
        """PR #9: GitHub Actions workflow files and docs <= 10 files."""
        files = [
            ".github/workflows/backend-ci.yml",
            ".github/workflows/docker-ci-cd.yml",
            ".github/workflows/frontend-ci.yml",
            "GEMINI.md",
            "agentic-memory/plans/2026-09-12_issue-3-ci-cd-workflows_plan.md",
        ]
        risk, _ = classify_pr_risk(files, title="ci: implement GitHub Actions CI/CD workflows")
        self.assertEqual(risk, RiskLevel.LOW)

    def test_pr_12_security_auth_critical_path_high_risk(self):
        """PR #12: Touches auth_service.go and security tokens."""
        files = [
            "services/api/internal/core/service/auth_service.go",
            "services/api/config/config.go",
            "apps/web/app/auth/callback/page.tsx",
        ]
        risk, reasons = classify_pr_risk(
            files,
            title="security(auth): patch hardcoded JWT secret, OAuth2 CSRF vulnerability",
            labels=["security"],
        )
        self.assertEqual(risk, RiskLevel.HIGH)
        self.assertTrue(any("critical path" in r or "security" in r for r in reasons))

    def test_pr_13_web_auth_tokens_high_risk(self):
        """PR #13: Reconcile auth token storage and JWT claims."""
        files = [
            "apps/web/lib/auth_token.ts",
            "apps/web/app/auth/callback/page.tsx",
            "agentic-memory/plans/2026-09-12_issue-6_plan.md",
        ]
        risk, reasons = classify_pr_risk(
            files,
            title="fix(web): reconcile auth token storage, synchronize JWT claims",
        )
        self.assertEqual(risk, RiskLevel.HIGH)
        self.assertTrue(any("critical path" in r or "authentication/token" in r for r in reasons))

    def test_pr_14_multi_tenant_db_models_high_risk(self):
        """PR #14: Touches core services without tests and migrations."""
        files = [
            "services/api/internal/core/service/tenant_service.go",
            "migrations/api/0001_tenants.sql",
            "services/api/internal/repository/tenant_repository.go",
        ]
        risk, reasons = classify_pr_risk(
            files,
            title="feat(api): implement multi-tenant database models, service domain, and REST API",
        )
        self.assertEqual(risk, RiskLevel.HIGH)
        self.assertTrue(any("migrations" in r or "core services" in r for r in reasons))

    def test_pr_15_sdk_token_refresh_high_risk(self):
        """PR #15: Auth token refresh interceptor in SDK."""
        files = [
            "packages/sdk/src/auth_token_refresh.ts",
            "packages/sdk/package.json",
            "agentic-memory/plans/2026-09-12_issue-8-sdk-token-refresh_plan.md",
        ]
        risk, reasons = classify_pr_risk(
            files,
            title="feat(sdk): add silent token refresh interceptor and modernize SDK package branding",
        )
        self.assertEqual(risk, RiskLevel.HIGH)
        self.assertTrue(any("authentication/token" in r for r in reasons))

    def test_pr_23_module_migration_large_blast_radius_high_risk(self):
        """PR #23: Touches core go.mod files and has > 50 changed files."""
        files = [f"pkg/sub{i}/file.go" for i in range(55)] + [
            "services/api/go.mod",
            "packages/shared/go.mod",
        ]
        risk, reasons = classify_pr_risk(
            files,
            title="refactor(go): migrate Go module paths and internal imports from gitlab to github",
        )
        self.assertEqual(risk, RiskLevel.HIGH)
        self.assertTrue(any("blast radius" in r or "module definitions" in r for r in reasons))

    def test_pr_24_automation_with_tests_medium_risk(self):
        """PR #24: Automation scripts and workflows with unit tests."""
        files = [
            ".agents/scripts/manage_issue_lifecycle.py",
            ".agents/scripts/manage_issue_lifecycle.sh",
            ".agents/scripts/test_manage_issue_lifecycle.py",
            ".agents/skills/github-task-runner/SKILL.md",
            ".github/workflows/issue-lifecycle-sync.yml",
        ]
        risk, _ = classify_pr_risk(
            files,
            title="feat(automation): implement issue lifecycle manager and parallel state synchronization workflow",
        )
        self.assertEqual(risk, RiskLevel.MEDIUM)


class TestPRReviewAgentEvaluation(unittest.TestCase):
    def setUp(self):
        self.agent = PRReviewAgent(dry_run=True)

    def test_evaluate_clean_low_risk_auto_merges(self):
        pr = PRDetails(
            number=19,
            title="docs: update guide",
            body="Closes #18",
            head_ref_name="task/issue-18",
            base_ref_name="main",
            is_draft=False,
            mergeable="MERGEABLE",
            merge_state_status="CLEAN",
            files=["README.md", "agentic-memory/plans/plan.md"],
        )
        result = self.agent.evaluate_pr(pr)
        self.assertEqual(result.risk_level, RiskLevel.LOW)
        self.assertEqual(result.decision, ActionDecision.AUTO_MERGE)

    def test_evaluate_conflicting_low_risk_rebases(self):
        pr = PRDetails(
            number=9,
            title="ci: update actions",
            body="Closes #3",
            head_ref_name="task/issue-3",
            base_ref_name="main",
            is_draft=False,
            mergeable="CONFLICTING",
            merge_state_status="DIRTY",
            files=[".github/workflows/backend-ci.yml", "GEMINI.md"],
        )
        result = self.agent.evaluate_pr(pr)
        self.assertEqual(result.risk_level, RiskLevel.LOW)
        self.assertEqual(result.decision, ActionDecision.REBASE_AND_MERGE)

    def test_evaluate_high_risk_escalates_regardless_of_mergeable(self):
        pr = PRDetails(
            number=12,
            title="security(auth): patch security",
            body="Closes #4",
            head_ref_name="task/issue-4",
            base_ref_name="main",
            is_draft=False,
            mergeable="MERGEABLE",
            merge_state_status="CLEAN",
            files=["services/api/internal/core/service/auth_service.go"],
        )
        result = self.agent.evaluate_pr(pr)
        self.assertEqual(result.risk_level, RiskLevel.HIGH)
        self.assertEqual(result.decision, ActionDecision.ESCALATE)

    def test_evaluate_draft_pr_skips(self):
        pr = PRDetails(
            number=99,
            title="wip: experimental",
            body="",
            head_ref_name="task/issue-99",
            base_ref_name="main",
            is_draft=True,
            mergeable="MERGEABLE",
            merge_state_status="CLEAN",
            files=["README.md"],
        )
        result = self.agent.evaluate_pr(pr)
        self.assertEqual(result.decision, ActionDecision.SKIP)


class TestTrivialConflictFiles(unittest.TestCase):
    def test_trivial_files_present(self):
        self.assertIn("go.sum", TRIVIAL_CONFLICT_FILES)
        self.assertIn("go.work.sum", TRIVIAL_CONFLICT_FILES)
        self.assertIn("package-lock.json", TRIVIAL_CONFLICT_FILES)
        self.assertIn("docs/docs.go", TRIVIAL_CONFLICT_FILES)


if __name__ == "__main__":
    unittest.main()
