#!/usr/bin/env python3
"""
Codebase Audit & Technical Debt Scout (audit_codebase.py)

Performs comprehensive audits across the Hexta monorepo for:
1. Hardcoded secrets, API keys, and credential leaks.
2. 5-layer Go architecture boundaries and layer violations (GEMINI.md).
3. Swagger / OpenAPI documentation completeness on HTTP handlers.
4. Technical debt markers (TODO, FIXME, HACK, temporary mock implementations).
5. Unit / integration test coverage gaps in critical packages.
6. Atlas migration integrity and checksum tracking (atlas.sum).
7. English-only compliance in comments and code.

Synthesizes findings into structured GitHub issues with intelligent deduplication,
and archives comprehensive markdown audit reports to `agentic-memory/audits/`.
"""

import argparse
from dataclasses import dataclass, field
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
from typing import Dict, List, Optional, Set, Tuple


@dataclass
class Finding:
    category: str        # 'security', 'architecture', 'swagger', 'debt', 'coverage', 'migration', 'language'
    severity: str        # 'high', 'medium', 'low'
    scope: str           # e.g. 'api', 'web', 'sdk', 'migrations'
    title: str           # Human-readable title
    file_path: str       # Relative file path
    line_number: Optional[int]
    snippet: str
    details: str
    recommendation: str
    fingerprint: str = ""

    def __post_init__(self):
        if not self.fingerprint:
            raw = f"{self.category}:{self.scope}:{self.file_path}:{self.title}"
            self.fingerprint = hashlib.sha256(raw.encode("utf-8")).hexdigest()[:16]


@dataclass
class IssueCandidate:
    title: str
    labels: List[str]
    body: str
    fingerprint: str
    severity: str
    scope: str
    findings: List[Finding] = field(default_factory=list)


# Ignored directories and files for general scanning
IGNORED_DIRS = {
    ".git",
    ".github",
    ".worktrees",
    "node_modules",
    ".next",
    "dist",
    "build",
    "vendor",
    ".turbo",
    "agentic-memory",
    ".gemini",
    "scratch",
    ".agents",
    "coverage",
    "tmp",
}

IGNORED_EXTENSIONS = {
    ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg", ".webp",
    ".pdf", ".zip", ".tar", ".gz", ".lock", ".sum", ".bin",
    ".pyc", ".map", ".woff", ".woff2", ".ttf", ".eot",
}


def is_ignored(path: Path, root: Path) -> bool:
    rel = path.relative_to(root)
    for part in rel.parts:
        if part in IGNORED_DIRS:
            return True
    if path.suffix.lower() in IGNORED_EXTENSIONS:
        return True
    return False


# -----------------------------------------------------------------------------
# Modular Analyzers
# -----------------------------------------------------------------------------

class SecretScanner:
    """Scans for exposed credentials, API keys, and private keys."""

    PATTERNS = [
        ("AWS Access Key", r"\b(AKIA[0-9A-Z]{16})\b", "high"),
        ("Private Key Header", r"-----BEGIN (?:RSA|DSA|EC|OPENSSH|PGP|ENCRYPTED)?\s*PRIVATE KEY-----", "high"),
        ("GitHub Personal Access Token", r"\b(ghp_[a-zA-Z0-9]{36,})\b", "high"),
        ("Slack Token", r"\b(xox[baprs]-[0-9a-zA-Z]{10,48})\b", "high"),
        ("OpenAI API Key", r"\b(sk-[a-zA-Z0-9]{32,})\b", "high"),
        ("Hardcoded Password Assignment", r"""(?i)(?:password|passwd|db_password)\s*[:=]\s*["']([^"'\s]{8,})["']""", "medium"),
        ("Hardcoded JWT Secret Assignment", r"""(?i)(?:jwt_secret|jwtSecret)\s*[:=]\s*["']([^"'\s]{6,})["']""", "high"),
        ("Hardcoded API Secret", r"""(?i)(?:client_secret|clientSecret|api_secret|apiSecret)\s*[:=]\s*["']([^"'\s]{10,})["']""", "high"),
    ]

    EXCLUDED_PATTERNS = [
        r"example", r"dummy", r"placeholder", r"mock", r"test", r"your-",
        r"<.*>", r"\${.*}", r"env\(", r"os\.Getenv", r"getEnv", r"config\.",
        r"changeme", r"password123", r"SECRET_KEY", r"API_KEY",
    ]

    def scan(self, root: Path) -> List[Finding]:
        findings = []
        for p in root.rglob("*"):
            if not p.is_file() or is_ignored(p, root):
                continue

            rel_path = str(p.relative_to(root))
            lower_path = rel_path.lower()

            # Skip test files, mock files, configs, and doc files for secret false positives
            if any(test_marker in lower_path for test_marker in ["test", "spec", "mock", ".env.example", "readme", "config.yaml"]):
                continue

            try:
                content = p.read_text(encoding="utf-8", errors="ignore")
            except Exception:
                continue

            for line_idx, line in enumerate(content.splitlines(), start=1):
                stripped = line.strip()
                if not stripped or stripped.startswith("//") or stripped.startswith("#"):
                    continue

                for label, pattern, sev in self.PATTERNS:
                    match = re.search(pattern, line)
                    if match:
                        matched_val = match.group(0)
                        # Check if matched value is an obvious placeholder
                        if any(re.search(exc, matched_val, re.IGNORECASE) for exc in self.EXCLUDED_PATTERNS):
                            continue
                        if any(re.search(exc, line, re.IGNORECASE) for exc in self.EXCLUDED_PATTERNS):
                            continue

                        snippet = line[:100]
                        findings.append(Finding(
                            category="security",
                            severity=sev,
                            scope="security",
                            title=f"Potential {label.lower()} detected in {p.name}",
                            file_path=rel_path,
                            line_number=line_idx,
                            snippet=snippet,
                            details=f"File contains pattern matching {label}. Hardcoded credentials should never be committed.",
                            recommendation="Move sensitive credentials to environment variables or secret management."
                        ))
        return findings


class ArchitectureScanner:
    """Enforces 5-layer Go architecture rules (GEMINI.md) in services/api."""

    def scan(self, root: Path) -> List[Finding]:
        findings = []
        api_dir = root / "services" / "api"
        if not api_dir.exists():
            return findings

        # Rule 1: Controllers must not import GORM, database/sql, or repository database instances
        controller_dir = api_dir / "internal" / "present" / "http" / "controller"
        if controller_dir.exists():
            for p in controller_dir.glob("*.go"):
                if p.name.endswith("_test.go") or p.name == "base_controller.go":
                    continue
                rel_path = str(p.relative_to(root))
                content = p.read_text(encoding="utf-8", errors="ignore")

                # Check forbidden imports
                forbidden_controller_imports = [
                    ("gorm.io/gorm", "Direct GORM import in controller violates 5-layer boundary"),
                    ("database/sql", "Direct SQL driver import in controller violates 5-layer boundary"),
                ]
                for imp, detail in forbidden_controller_imports:
                    if f'"{imp}"' in content:
                        findings.append(Finding(
                            category="architecture",
                            severity="high",
                            scope="api",
                            title=f"Layer boundary violation: controller directly imports {imp} in {p.name}",
                            file_path=rel_path,
                            line_number=None,
                            snippet=f'import "{imp}"',
                            details=f"{detail}. Controllers must only communicate with core services or DTOs.",
                            recommendation=f"Move database queries or GORM operations from {p.name} into repository and service layers."
                        ))

                # Check struct fields in controller holding *gorm.DB
                lines = content.splitlines()
                for idx, line in enumerate(lines, start=1):
                    if re.search(r"\bdb\s+\*gorm\.DB\b", line):
                        findings.append(Finding(
                            category="architecture",
                            severity="high",
                            scope="api",
                            title=f"Controller struct contains direct database client reference in {p.name}",
                            file_path=rel_path,
                            line_number=idx,
                            snippet=line.strip(),
                            details="Controllers must not hold direct database connections or perform DB operations directly.",
                            recommendation="Delegate database interactions to an injected service or repository."
                        ))

        # Rule 2: Service layer must not import HTTP frameworks (gin, net/http for handlers)
        service_dir = api_dir / "internal" / "core" / "service"
        if service_dir.exists():
            for p in service_dir.glob("*.go"):
                if p.name.endswith("_test.go") or p.name == "base_service.go":
                    continue
                rel_path = str(p.relative_to(root))
                content = p.read_text(encoding="utf-8", errors="ignore")

                if '"github.com/gin-gonic/gin"' in content:
                    findings.append(Finding(
                        category="architecture",
                        severity="high",
                        scope="api",
                        title=f"Layer boundary violation: service imports Gin transport in {p.name}",
                        file_path=rel_path,
                        line_number=None,
                        snippet='import "github.com/gin-gonic/gin"',
                        details="Core service layer must remain transport-agnostic and free of Gin HTTP framework dependencies.",
                        recommendation="Remove gin.Context and Gin dependencies from the service method signatures."
                    ))

        # Rule 3: Repository layer must not import Gin
        repo_dir = api_dir / "internal" / "repository"
        if repo_dir.exists():
            for p in repo_dir.glob("*.go"):
                if p.name.endswith("_test.go"):
                    continue
                rel_path = str(p.relative_to(root))
                content = p.read_text(encoding="utf-8", errors="ignore")

                if '"github.com/gin-gonic/gin"' in content:
                    findings.append(Finding(
                        category="architecture",
                        severity="high",
                        scope="api",
                        title=f"Layer boundary violation: repository imports Gin in {p.name}",
                        file_path=rel_path,
                        line_number=None,
                        snippet='import "github.com/gin-gonic/gin"',
                        details="Repository layer handles data persistence only and must not depend on HTTP frameworks.",
                        recommendation="Remove Gin references from repository layer."
                    ))

        return findings


class SwaggerScanner:
    """Verifies Swagger / OpenAPI annotations on Go controllers."""

    def scan(self, root: Path) -> List[Finding]:
        findings = []
        controller_dir = root / "services" / "api" / "internal" / "present" / "http" / "controller"
        if not controller_dir.exists():
            return findings

        handler_regex = re.compile(r"^func\s+\(\w+\s+\*\w+Controller\)\s+([A-Z]\w+)\s*\((?:c\s+\*gin\.Context|ctx\s+\*gin\.Context)\)")

        for p in controller_dir.glob("*.go"):
            if p.name == "base_controller.go" or p.name.endswith("_test.go"):
                continue

            rel_path = str(p.relative_to(root))
            lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()

            for idx, line in enumerate(lines):
                match = handler_regex.match(line.strip())
                if match:
                    handler_name = match.group(1)
                    # Inspect preceding comments (up to 25 lines backwards)
                    comment_block = []
                    back_idx = idx - 1
                    while back_idx >= 0 and (lines[back_idx].strip().startswith("//") or not lines[back_idx].strip()):
                        if lines[back_idx].strip().startswith("//"):
                            comment_block.insert(0, lines[back_idx].strip())
                        back_idx -= 1

                    comment_text = "\n".join(comment_block)
                    has_summary = "@Summary" in comment_text
                    has_router = "@Router" in comment_text
                    has_tags = "@Tags" in comment_text
                    has_success = "@Success" in comment_text

                    if not (has_summary and has_router):
                        findings.append(Finding(
                            category="swagger",
                            severity="medium",
                            scope="api",
                            title=f"Missing Swagger annotations on controller handler {handler_name} in {p.name}",
                            file_path=rel_path,
                            line_number=idx + 1,
                            snippet=line.strip(),
                            details=f"Endpoint handler '{handler_name}' is missing required @Summary or @Router annotations.",
                            recommendation="Add standard Swagger doc annotations (@Summary, @Description, @Tags, @Produce, @Success, @Router)."
                        ))
                    elif has_success and "response.Response" not in comment_text and "map[string]" not in comment_text:
                        # Check response wrapper compliance
                        findings.append(Finding(
                            category="swagger",
                            severity="low",
                            scope="api",
                            title=f"Swagger @Success in {p.name} should use response.Response[T] wrapper for {handler_name}",
                            file_path=rel_path,
                            line_number=idx + 1,
                            snippet=line.strip(),
                            details=f"Handler '{handler_name}' uses raw DTO in @Success instead of standard response.Response[T].",
                            recommendation="Update Swagger @Success tag to follow `response.Response[dto.YourResponse]` format."
                        ))

        return findings


class DebtScanner:
    """Scans for TODOs, FIXMEs, and temporary mock implementations."""

    DEBT_PATTERNS = [
        ("TODO comment", r"//\s*(?:TODO|todo)\b:?(.*)", "low"),
        ("FIXME comment", r"//\s*(?:FIXME|fixme)\b:?(.*)", "medium"),
        ("HACK comment", r"//\s*(?:HACK|hack)\b:?(.*)", "medium"),
        ("Hardcoded Mock Tenant", r"""(?:["']mock-tenant(?:-id)?["']|["']tenant-123["'])""", "medium"),
        ("Mock Tenant Comment", r"""//\s*Mock hardcoded tenant""", "medium"),
    ]

    def scan(self, root: Path) -> List[Finding]:
        findings = []
        scan_dirs = ["services", "apps", "packages", "migrations"]

        for base_name in scan_dirs:
            base_dir = root / base_name
            if not base_dir.exists():
                continue

            for p in base_dir.rglob("*"):
                if not p.is_file() or is_ignored(p, root):
                    continue

                rel_path = str(p.relative_to(root))
                # Skip test files and protobuf generated files
                if ".pb.go" in p.name or p.name.endswith("_test.go") or ".spec." in p.name or ".test." in p.name:
                    continue

                try:
                    lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()
                except Exception:
                    continue

                for idx, line in enumerate(lines, start=1):
                    stripped = line.strip()
                    for label, pattern, sev in self.DEBT_PATTERNS:
                        match = re.search(pattern, stripped)
                        if match:
                            comment_text = match.group(1).strip() if match.groups() else stripped
                            findings.append(Finding(
                                category="debt",
                                severity=sev,
                                scope=base_name,
                                title=f"Technical debt: {label} in {p.name}",
                                file_path=rel_path,
                                line_number=idx,
                                snippet=stripped[:120],
                                details=f"{label} found: '{comment_text}'.",
                                recommendation="Address the pending task or replace temporary mock with production implementation."
                            ))
        return findings


class CoverageScanner:
    """Flags critical business packages lacking unit tests."""

    def scan(self, root: Path) -> List[Finding]:
        findings = []

        # Check Go service packages
        service_dir = root / "services" / "api" / "internal" / "core" / "service"
        if service_dir.exists():
            go_files = [f for f in service_dir.glob("*.go") if not f.name.endswith("_test.go") and f.name != "base_service.go"]
            test_files = list(service_dir.glob("*_test.go"))

            for gf in go_files:
                expected_test = service_dir / f"{gf.stem}_test.go"
                if not expected_test.exists():
                    findings.append(Finding(
                        category="coverage",
                        severity="medium",
                        scope="api",
                        title=f"Missing unit test suite for service {gf.name}",
                        file_path=str(gf.relative_to(root)),
                        line_number=None,
                        snippet=f"package {gf.stem}",
                        details=f"Core business logic in {gf.name} lacks corresponding test file {expected_test.name}.",
                        recommendation=f"Create {expected_test.name} covering core service methods and error paths."
                    ))

        # Check Go repository packages
        repo_dir = root / "services" / "api" / "internal" / "repository"
        if repo_dir.exists():
            repo_files = [f for f in repo_dir.glob("*.go") if not f.name.endswith("_test.go") and f.name != "base_repository.go"]
            for rf in repo_files:
                expected_test = repo_dir / f"{rf.stem}_test.go"
                if not expected_test.exists():
                    findings.append(Finding(
                        category="coverage",
                        severity="low",
                        scope="api",
                        title=f"Missing repository test coverage for {rf.name}",
                        file_path=str(rf.relative_to(root)),
                        line_number=None,
                        snippet=f"package repository",
                        details=f"Data persistence wrapper in {rf.name} lacks unit or integration test verification.",
                        recommendation=f"Add unit or mock repository tests for {rf.name}."
                    ))

        # Check TypeScript packages
        sdk_src = root / "packages" / "sdk" / "src"
        if sdk_src.exists():
            ts_files = [f for f in sdk_src.glob("*.ts") if not f.name.endswith(".test.ts") and not f.name.endswith(".d.ts")]
            tests = list(sdk_src.glob("*.test.ts"))
            if ts_files and not tests:
                findings.append(Finding(
                    category="coverage",
                    severity="medium",
                    scope="sdk",
                    title="Missing test suite for packages/sdk",
                    file_path="packages/sdk",
                    line_number=None,
                    snippet="packages/sdk/src",
                    details="SDK package contains source code but no unit or integration tests in packages/sdk/src.",
                    recommendation="Add Vitest / Jest test suites to verify client methods."
                ))

        return findings


class MigrationScanner:
    """Verifies Atlas schema migration checksum tracking."""

    def scan(self, root: Path) -> List[Finding]:
        findings = []
        migration_dir = root / "migrations" / "api"
        if not migration_dir.exists():
            return findings

        atlas_sum = migration_dir / "atlas.sum"
        if not atlas_sum.exists():
            findings.append(Finding(
                category="migration",
                severity="high",
                scope="migrations",
                title="Missing atlas.sum checksum file in migrations/api",
                file_path="migrations/api/atlas.sum",
                line_number=None,
                snippet="migrations/api/atlas.sum",
                details="Atlas requires atlas.sum to enforce migration directory integrity and avoid drift.",
                recommendation="Run `atlas migrate hash` or `make migrate-hash` to generate atlas.sum."
            ))
            return findings

        sum_content = atlas_sum.read_text(encoding="utf-8", errors="ignore")
        sql_files = list(migration_dir.glob("*.sql"))

        for sql_f in sql_files:
            if sql_f.name not in sum_content:
                findings.append(Finding(
                    category="migration",
                    severity="high",
                    scope="migrations",
                    title=f"Migration {sql_f.name} missing from atlas.sum checksums",
                    file_path=str(sql_f.relative_to(root)),
                    line_number=None,
                    snippet=sql_f.name,
                    details=f"SQL migration file {sql_f.name} is not tracked in atlas.sum, causing migration verification failure.",
                    recommendation="Rehash migrations using `atlas migrate hash`."
                ))

        return findings


class LanguageScanner:
    """Verifies English-only rule compliance across source code and comments."""

    NON_ASCII_COMMENT = re.compile(r"//.*[^\x00-\x7F]+|/\*.*[^\x00-\x7F]+.*\*/")

    def scan(self, root: Path) -> List[Finding]:
        findings = []
        target_dirs = ["services", "apps", "packages"]

        for tdir in target_dirs:
            base = root / tdir
            if not base.exists():
                continue

            for p in base.rglob("*"):
                if not p.is_file() or is_ignored(p, root):
                    continue
                if p.suffix not in {".go", ".ts", ".tsx", ".js", ".jsx"}:
                    continue
                if ".pb.go" in p.name:
                    continue

                try:
                    lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()
                except Exception:
                    continue

                for idx, line in enumerate(lines, start=1):
                    match = self.NON_ASCII_COMMENT.search(line)
                    if match:
                        snippet = match.group(0).strip()[:100]
                        # Filter out innocuous unicode like smart quotes or emojis
                        if any(ord(c) > 127 and ord(c) > 255 for c in snippet):
                            findings.append(Finding(
                                category="language",
                                severity="low",
                                scope=tdir,
                                title=f"Non-English comment detected in {p.name}",
                                file_path=str(p.relative_to(root)),
                                line_number=idx,
                                snippet=snippet,
                                details="GEMINI.md Rule 1 mandates English exclusively for all code, comments, and docs.",
                                recommendation="Translate non-English comments into clear English."
                            ))
        return findings


# -----------------------------------------------------------------------------
# Issue Synthesizer & Deduplication Engine
# -----------------------------------------------------------------------------

class DeduplicationEngine:
    """Queries existing GitHub issues to prevent creating duplicate tasks."""

    def __init__(self):
        self.existing_issues: List[Dict] = []
        self.fingerprints: Set[str] = set()
        self.titles: Set[str] = set()
        self.loaded = False

    def load_existing(self) -> bool:
        cmd = [
            "gh", "issue", "list",
            "--state", "all",
            "--limit", "300",
            "--json", "number,title,body,labels,state"
        ]
        try:
            res = subprocess.run(cmd, capture_output=True, text=True, check=True)
            self.existing_issues = json.loads(res.stdout or "[]")
            for issue in self.existing_issues:
                title = issue.get("title", "").strip().lower()
                self.titles.add(title)

                body = issue.get("body", "")
                fp_matches = re.findall(r"<!--\s*audit-fingerprint:\s*([a-zA-Z0-9_-]+)\s*-->", body)
                for fp in fp_matches:
                    self.fingerprints.add(fp)

            self.loaded = True
            return True
        except Exception as e:
            print(f"[WARN] Failed to query existing GitHub issues: {e}", file=sys.stderr)
            self.loaded = False
            return False

    def is_duplicate(self, candidate: IssueCandidate) -> Tuple[bool, str]:
        if not self.loaded:
            return False, ""

        # Check fingerprint
        if candidate.fingerprint in self.fingerprints:
            return True, f"Fingerprint '{candidate.fingerprint}' matches existing issue"

        # Check exact or normalized title
        cand_title_norm = candidate.title.strip().lower()
        if cand_title_norm in self.titles:
            return True, f"Exact title matches existing issue: '{candidate.title}'"

        # Check semantic title overlap
        for ex in self.existing_issues:
            ex_title = ex.get("title", "").strip().lower()
            ex_num = ex.get("number")

            # Check if existing issue already covers the exact scope and subject
            if self._titles_conflict(cand_title_norm, ex_title):
                return True, f"Semantic conflict with issue #{ex_num}: '{ex.get('title')}'"

        return False, ""

    def _titles_conflict(self, cand: str, existing: str) -> bool:
        # Common significant phrase pairs or keywords
        key_phrases = [
            "jwt secret",
            "csrf",
            "tenant",
            "architecture",
            "swagger",
            "openapi",
            "atlas.sum",
            "token refresh",
            "english language",
            "unit test",
            "test coverage",
        ]
        for phrase in key_phrases:
            if phrase in cand and phrase in existing:
                return True
        return False


def synthesize_issues(findings: List[Finding]) -> List[IssueCandidate]:
    """Groups findings by topic/component into coherent issue candidates."""
    candidates: List[IssueCandidate] = []

    # 1. Architecture Violations in services/api
    arch_findings = [f for f in findings if f.category == "architecture"]
    if arch_findings:
        files = sorted(list(set(f.file_path for f in arch_findings)))
        details_list = "\n".join([f"- **`{f.file_path}`** (line {f.line_number or 'N/A'}): {f.details} (`{f.snippet}`)" for f in arch_findings])
        
        fingerprint = hashlib.sha256(f"architecture-violations:{','.join(files)}".encode()).hexdigest()[:16]
        body = f"""### Summary & Problem Statement
During automated codebase inspection, architectural boundary violations were detected violating the 5-layer Go architecture rules in `GEMINI.md`:
Controllers or services import disallowed dependencies (e.g. direct GORM / database clients in controllers, or transport frameworks in service logic).

### Affected Files & References
{details_list}

### Remediation & Acceptance Criteria
- [ ] Remove `*gorm.DB` or SQL drivers from controller structs and move data access to repository layer.
- [ ] Ensure all controller handlers invoke service layer methods rather than performing DB ping/queries.
- [ ] Run `go build ./...` and `go test ./...` in `services/api` to verify boundary compliance.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="fix(arch): resolve 5-layer architecture boundary violations in API controllers",
            labels=["bug", "priority:high"],
            body=body,
            fingerprint=fingerprint,
            severity="high",
            scope="api",
            findings=arch_findings,
        ))

    # 2. Swagger / OpenAPI Documentation Gaps
    swagger_findings = [f for f in findings if f.category == "swagger" and f.severity == "medium"]
    if swagger_findings:
        details_list = "\n".join([f"- **`{f.file_path}:{f.line_number}`**: {f.details}" for f in swagger_findings])
        fingerprint = hashlib.sha256(b"swagger-missing-annotations").hexdigest()[:16]
        body = f"""### Summary & Problem Statement
HTTP controller methods in `services/api` lack standard Swagger / OpenAPI annotations (`@Summary`, `@Router`, `@Tags`, `@Success`), leading to incomplete API documentation and client SDK generation gaps.

### Affected Handlers
{details_list}

### Remediation & Acceptance Criteria
- [ ] Add `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Success`, and `@Router` annotations to all listed controller methods.
- [ ] Ensure `@Success` uses standard `response.Response[dto.YourResponse]` wrapper.
- [ ] Run `make swagger` to regenerate docs and verify with `go build ./...`.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="docs(swagger): add missing OpenAPI annotations to API controller endpoints",
            labels=["enhancement", "priority:medium"],
            body=body,
            fingerprint=fingerprint,
            severity="medium",
            scope="api",
            findings=swagger_findings,
        ))

    # 3. Hardcoded Mock Tenant Implementations
    mock_findings = [f for f in findings if "mock" in f.title.lower() or "tenant" in f.title.lower()]
    if mock_findings:
        files = sorted(list(set(f.file_path for f in mock_findings)))
        details_list = "\n".join([f"- **`{f.file_path}:{f.line_number}`**: `{f.snippet}`" for f in mock_findings])
        fingerprint = hashlib.sha256(b"debt-mock-tenant-cleanup").hexdigest()[:16]
        body = f"""### Summary & Problem Statement
Temporary hardcoded mock tenant IDs (e.g. `'mock-tenant-id'`) were detected in user-facing components. Tenant context should be resolved dynamically from the authenticated user token or session context.

### Affected Locations
{details_list}

### Remediation & Acceptance Criteria
- [ ] Replace static mock tenant IDs with authenticated context from session/token.
- [ ] Add fallback error handling for unauthenticated or missing tenant scenarios.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="refactor(tenant): replace temporary mock tenant IDs with dynamic token session context",
            labels=["enhancement", "priority:medium"],
            body=body,
            fingerprint=fingerprint,
            severity="medium",
            scope="web",
            findings=mock_findings,
        ))

    # 4. Service Unit Test Coverage Gaps
    coverage_findings = [f for f in findings if f.category == "coverage" and f.severity == "medium"]
    if coverage_findings:
        details_list = "\n".join([f"- **`{f.file_path}`**: {f.details}" for f in coverage_findings])
        fingerprint = hashlib.sha256(b"coverage-missing-unit-tests").hexdigest()[:16]
        body = f"""### Summary & Problem Statement
Critical business logic packages and services lack corresponding automated test suites (`*_test.go` or `*.test.ts`), risking untested edge cases and regressions during refactors.

### Packages Lacking Tests
{details_list}

### Remediation & Acceptance Criteria
- [ ] Add unit test suites with mock dependencies for uncovered services.
- [ ] Verify test suite passes with `go test ./...` / `pnpm test`.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="test(coverage): implement unit test suites for critical core services",
            labels=["enhancement", "priority:medium"],
            body=body,
            fingerprint=fingerprint,
            severity="medium",
            scope="api",
            findings=coverage_findings,
        ))

    # 5. Security & Leaked Secret Findings
    sec_findings = [f for f in findings if f.category == "security"]
    for sf in sec_findings:
        fingerprint = sf.fingerprint
        body = f"""### Summary & Problem Statement
A potential credential exposure or sensitive secret was flagged during automated security scanning:
{sf.details}

### Affected Location
- **File**: `{sf.file_path}`
- **Line**: `{sf.line_number or 'N/A'}`
- **Snippet**: `{sf.snippet}`

### Remediation & Acceptance Criteria
- [ ] Remove hardcoded credentials from source code.
- [ ] Migrate credential configuration to environment variables and secret stores.
- [ ] Invalidate and rotate any exposed keys or tokens.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title=f"security: resolve potential hardcoded credential in {Path(sf.file_path).name}",
            labels=["bug", "priority:high"],
            body=body,
            fingerprint=fingerprint,
            severity="high",
            scope="security",
            findings=[sf],
        ))

    # 6. Atlas Migration Integrity Findings
    mig_findings = [f for f in findings if f.category == "migration"]
    if mig_findings:
        details_list = "\n".join([f"- **`{f.file_path}`**: {f.details}" for f in mig_findings])
        fingerprint = hashlib.sha256(b"migration-atlas-sum-integrity").hexdigest()[:16]
        body = f"""### Summary & Problem Statement
Atlas database migration integrity checks identified missing or unverified migration files in `migrations/api`. This can cause schema drift or CI migration failures.

### Affected Files
{details_list}

### Remediation & Acceptance Criteria
- [ ] Rehash Atlas migrations using `atlas migrate hash` or `make migrate-hash`.
- [ ] Verify migration directory status with `atlas migrate status`.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="fix(db): reconcile atlas.sum migration checksum integrity",
            labels=["bug", "priority:high"],
            body=body,
            fingerprint=fingerprint,
            severity="high",
            scope="migrations",
            findings=mig_findings,
        ))

    # 7. Unaddressed TODO / FIXME Technical Debt
    debt_findings = [f for f in findings if f.category == "debt" and f not in mock_findings]
    if len(debt_findings) >= 3:
        details_list = "\n".join([f"- **`{f.file_path}:{f.line_number}`**: `{f.snippet}`" for f in debt_findings[:10]])
        fingerprint = hashlib.sha256(b"debt-todo-fixme-backlog").hexdigest()[:16]
        body = f"""### Summary & Problem Statement
Multiple unresolved `TODO`, `FIXME`, or `HACK` technical debt annotations were identified across active packages.

### Sample Debt Items
{details_list}

### Remediation & Acceptance Criteria
- [ ] Review identified debt items, resolve actionable implementations, and remove obsolete comments.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="refactor(debt): resolve pending TODO and FIXME debt annotations across services",
            labels=["enhancement", "priority:low"],
            body=body,
            fingerprint=fingerprint,
            severity="low",
            scope="monorepo",
            findings=debt_findings,
        ))

    # 8. English Language Rule Violations
    lang_findings = [f for f in findings if f.category == "language"]
    if lang_findings:
        details_list = "\n".join([f"- **`{f.file_path}:{f.line_number}`**: `{f.snippet}`" for f in lang_findings[:10]])
        fingerprint = hashlib.sha256(b"language-english-only-compliance").hexdigest()[:16]
        body = f"""### Summary & Problem Statement
Non-English comments were detected in source files, violating GEMINI.md Rule 1 (English exclusively for all code, comments, documentation).

### Sample Locations
{details_list}

### Remediation & Acceptance Criteria
- [ ] Translate all non-English comments into clear, concise English.

<!-- audit-fingerprint: {fingerprint} -->
"""
        candidates.append(IssueCandidate(
            title="chore(i18n): enforce English language rule across source code and comments",
            labels=["enhancement", "priority:low"],
            body=body,
            fingerprint=fingerprint,
            severity="low",
            scope="monorepo",
            findings=lang_findings,
        ))

    return candidates



# -----------------------------------------------------------------------------
# Report Generator
# -----------------------------------------------------------------------------

def generate_markdown_report(
    date_str: str,
    root: Path,
    findings: List[Finding],
    candidates: List[IssueCandidate],
    action_log: List[str]
) -> str:
    category_counts = {}
    severity_counts = {"high": 0, "medium": 0, "low": 0}

    for f in findings:
        category_counts[f.category] = category_counts.get(f.category, 0) + 1
        severity_counts[f.severity] = severity_counts.get(f.severity, 0) + 1

    report = [
        f"# Codebase Audit Report: {date_str}",
        "",
        f"- **Inspection Date**: {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M:%S UTC')}",
        f"- **Repository Root**: `{root.resolve()}`",
        f"- **Total Findings**: {len(findings)}",
        f"- **Issue Candidates Synthesized**: {len(candidates)}",
        "",
        "---",
        "",
        "## 1. Executive Summary & Health Metrics",
        "",
        "| Category | Count | High | Medium | Low |",
        "|---|---|---|---|---|",
    ]

    for cat in sorted(category_counts.keys()):
        cat_findings = [f for f in findings if f.category == cat]
        h = sum(1 for f in cat_findings if f.severity == "high")
        m = sum(1 for f in cat_findings if f.severity == "medium")
        l = sum(1 for f in cat_findings if f.severity == "low")
        report.append(f"| `{cat}` | {len(cat_findings)} | {h} | {m} | {l} |")

    report.extend([
        "",
        "### Severity Breakdown",
        f"- **High Severity**: {severity_counts['high']}",
        f"- **Medium Severity**: {severity_counts['medium']}",
        f"- **Low Severity**: {severity_counts['low']}",
        "",
        "---",
        "",
        "## 2. Issue Generation & Deduplication Actions",
        "",
    ])

    if action_log:
        for action in action_log:
            report.append(f"- {action}")
    else:
        report.append("No automated issue actions taken.")

    report.extend([
        "",
        "---",
        "",
        "## 3. Detailed Audit Findings",
        "",
    ])

    for cat in sorted(category_counts.keys()):
        cat_findings = [f for f in findings if f.category == cat]
        report.append(f"### Category: `{cat}` ({len(cat_findings)} findings)")
        report.append("")
        report.append("| Severity | Location | Title | Details |")
        report.append("|---|---|---|---|")
        for f in cat_findings:
            loc = f"`{f.file_path}`"
            if f.line_number:
                loc += f":{f.line_number}"
            snippet_clean = f.snippet.replace("|", "\\|")
            report.append(f"| **{f.severity.upper()}** | {loc} | {f.title} | {f.details} |")
        report.append("")

    return "\n".join(report)


# -----------------------------------------------------------------------------
# Main CLI Entry Point
# -----------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description="Hexta Codebase Audit & Technical Debt Scout")
    parser.add_argument("--root", default=".", help="Root directory of repository (default: current directory)")
    parser.add_argument("--dry-run", action="store_true", help="Run audit and generate report without creating GitHub issues")
    parser.add_argument("--create-issues", action="store_true", help="Create GitHub issues for non-duplicate candidates")
    parser.add_argument("--max-issues", type=int, default=5, help="Maximum number of GitHub issues to create per run (default: 5)")
    parser.add_argument("--output", default="", help="Custom output path for markdown audit report")
    parser.add_argument("--fail-on-high", action="store_true", help="Exit with code 1 if high-severity findings are discovered")
    args = parser.parse_args()

    root = Path(args.root).resolve()
    print(f"[*] Starting Codebase Audit on {root}...")

    # Run modular scanners
    scanners = [
        ("Secrets", SecretScanner()),
        ("Architecture", ArchitectureScanner()),
        ("Swagger", SwaggerScanner()),
        ("Technical Debt", DebtScanner()),
        ("Test Coverage", CoverageScanner()),
        ("Atlas Migrations", MigrationScanner()),
        ("Language Rules", LanguageScanner()),
    ]

    all_findings: List[Finding] = []
    for name, scanner in scanners:
        print(f"  -> Running {name} scanner...")
        results = scanner.scan(root)
        print(f"     Found {len(results)} items.")
        all_findings.extend(results)

    print(f"[*] Total findings discovered: {len(all_findings)}")

    # Synthesize issue candidates
    candidates = synthesize_issues(all_findings)
    print(f"[*] Synthesized {len(candidates)} actionable issue candidates.")

    # Deduplication & Issue Creation
    dedup = DeduplicationEngine()
    dedup.load_existing()

    action_log: List[str] = []
    issues_created_count = 0

    for cand in candidates:
        is_dup, reason = dedup.is_duplicate(cand)
        if is_dup:
            msg = f"SKIPPED (Duplicate): '{cand.title}' - Reason: {reason}"
            print(f"  [DEDUP] {msg}")
            action_log.append(msg)
            continue

        if args.create_issues and not args.dry_run:
            if issues_created_count >= args.max_issues:
                msg = f"DEFERRED (Cap Reached): '{cand.title}' (Max limit of {args.max_issues} issues reached)"
                print(f"  [CAP] {msg}")
                action_log.append(msg)
                continue

            # Create GitHub issue via gh cli
            labels_arg = ",".join(cand.labels)
            create_cmd = [
                "gh", "issue", "create",
                "--title", cand.title,
                "--body", cand.body,
                "--label", labels_arg,
            ]
            try:
                res = subprocess.run(create_cmd, capture_output=True, text=True, check=True, cwd=str(root))
                created_url = res.stdout.strip()
                issues_created_count += 1
                msg = f"CREATED: [{cand.title}]({created_url}) with labels `{labels_arg}`"
                print(f"  [CREATE] {msg}")
                action_log.append(msg)
            except Exception as e:
                msg = f"FAILED: Could not create issue '{cand.title}': {e}"
                print(f"  [ERROR] {msg}", file=sys.stderr)
                action_log.append(msg)
        else:
            msg = f"PENDING (Dry Run): Candidate '{cand.title}' with severity `{cand.severity}` and labels {cand.labels}"
            action_log.append(msg)

    # Generate and save audit report
    date_str = datetime.now().strftime("%Y-%m-%d")
    if args.output:
        report_path = Path(args.output).resolve()
    else:
        audits_dir = root / "agentic-memory" / "audits"
        audits_dir.mkdir(parents=True, exist_ok=True)
        report_path = audits_dir / f"{date_str}_codebase-audit.md"

    report_content = generate_markdown_report(date_str, root, all_findings, candidates, action_log)
    report_path.write_text(report_content, encoding="utf-8")
    print(f"[+] Audit report successfully archived to: {report_path}")

    # High severity check
    high_count = sum(1 for f in all_findings if f.severity == "high")
    if args.fail_on_high and high_count > 0:
        print(f"[!] Failing run due to {high_count} high-severity findings.", file=sys.stderr)
        sys.exit(1)

    print("[*] Codebase audit completed successfully.")


if __name__ == "__main__":
    main()
