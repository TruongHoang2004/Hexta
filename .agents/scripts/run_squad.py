#!/usr/bin/env python3
"""
run_squad.py - Multi-Agent Squad Orchestrator and Dispatcher for Hexta.

Coordinates the 10 autonomous agents:
  [01] architect_lead   (domain:spec)
  [02] backend_core     (domain:backend-core)
  [03] backend_api      (domain:backend-api)
  [04] backend_infra    (domain:backend-infra)
  [05] frontend_ui      (domain:frontend-ui)
  [06] frontend_logic   (domain:frontend-logic)
  [07] qa_engineer      (domain:qa)
  [08] code_scout       (Audit & Security Scout)
  [09] pr_reviewer      (Gatekeeper & Merge Agent)
  [10] janitor_memory   (Memory Consolidator & Stale Lock Cleaner)

Commands:
  --status:        List all 10 agents, domains, and current queue count.
  --agent <name>:  Execute routine or claim next task for a specific agent.
  --monitors:      Run the 3 automated guardians (Scout + PR Reviewer + Janitor).
  --dry-run:       Simulate execution without modifying git or GitHub state.
"""

import argparse
from dataclasses import dataclass
import json
import logging
import subprocess
import sys
from typing import Dict, List, Optional

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
logger = logging.getLogger("squad_orchestrator")


@dataclass
class AgentConfig:
    id: str
    name: str
    domain: Optional[str]
    category: str  # 'developer', 'monitor', 'lead'
    description: str


SQUAD_ROSTER: Dict[str, AgentConfig] = {
    "architect_lead": AgentConfig(
        id="architect_lead",
        name="Architect & Product Lead",
        domain="spec",
        category="lead",
        description="Backlog specification, architecture designs, atomic decomposition."
    ),
    "backend_core": AgentConfig(
        id="backend_core",
        name="Backend Core Business Engineer",
        domain="backend-core",
        category="developer",
        description="Services, business rules, domain entities in services/api."
    ),
    "backend_api": AgentConfig(
        id="backend_api",
        name="Backend API & Data Engineer",
        domain="backend-api",
        category="developer",
        description="Controllers, DTOs, Gorm repositories, Atlas migrations."
    ),
    "backend_infra": AgentConfig(
        id="backend_infra",
        name="Backend Infrastructure Engineer",
        domain="backend-infra",
        category="developer",
        description="packages/shared, Redis caching, middleware, Uber Fx DI."
    ),
    "frontend_ui": AgentConfig(
        id="frontend_ui",
        name="Frontend UI & Styling Specialist",
        domain="frontend-ui",
        category="developer",
        description="apps/web/app/components, Tailwind CSS, responsive layouts."
    ),
    "frontend_logic": AgentConfig(
        id="frontend_logic",
        name="Frontend Logic & State Engineer",
        domain="frontend-logic",
        category="developer",
        description="apps/web/lib, API clients, form state, authentication."
    ),
    "qa_engineer": AgentConfig(
        id="qa_engineer",
        name="QA & Test Specialist",
        domain="qa",
        category="developer",
        description="Unit/integration tests (*_test.go), mocks, regression tests."
    ),
    "code_scout": AgentConfig(
        id="code_scout",
        name="Security & Technical Debt Scout",
        domain="audit",
        category="monitor",
        description="Static security scans, 5-layer boundaries, debt issue filing."
    ),
    "pr_reviewer": AgentConfig(
        id="pr_reviewer",
        name="Gatekeeper & PR Reviewer",
        domain="review",
        category="monitor",
        description="PR risk evaluation, rebase trivial conflicts, squash auto-merge."
    ),
    "janitor_memory": AgentConfig(
        id="janitor_memory",
        name="Memory Janitor & Lifecycle Reconciler",
        domain="ops",
        category="monitor",
        description="Reconcile issue lifecycle, recover stale tasks (>2h), memory prune."
    ),
}


def print_roster_status():
    print("\n================================================================================")
    print("                      HEXTA AUTONOMOUS AGENT SQUAD (10 AGENTS)                 ")
    print("================================================================================")
    print(f"{'#':<3} {'Agent ID':<16} {'Domain Tag':<22} {'Category':<11} {'Name'}")
    print("-" * 80)
    for idx, (agent_id, cfg) in enumerate(SQUAD_ROSTER.items(), start=1):
        domain_str = f"domain:{cfg.domain}" if cfg.domain else "N/A"
        print(f"[{idx:02d}] {agent_id:<16} {domain_str:<22} {cfg.category:<11} {cfg.name}")
    print("=" * 80 + "\n")


def run_monitors(dry_run: bool = False):
    logger.info("Executing Autonomous Guardians Triad (Janitor + PR Reviewer + Code Scout)...")

    # 1. Janitor: sync lifecycle and recover stale tasks
    logger.info("[1/3] Running Janitor Memory (manage_issue_lifecycle.py)...")
    cmd_janitor = ["python3", ".agents/scripts/manage_issue_lifecycle.py", "--sync", "--recover-stale"]
    if dry_run:
        cmd_janitor.append("--dry-run")
    subprocess.run(cmd_janitor)

    # 2. PR Reviewer: scan and process PRs
    logger.info("[2/3] Running PR Reviewer (pr_review_agent.py)...")
    cmd_pr = ["python3", ".agents/scripts/pr_review_agent.py", "--review-and-merge"]
    if dry_run:
        cmd_pr.append("--dry-run")
    subprocess.run(cmd_pr)

    # 3. Code Scout: static audit
    logger.info("[3/3] Running Code Scout (audit_codebase.py)...")
    cmd_scout = ["python3", ".agents/scripts/audit_codebase.py"]
    subprocess.run(cmd_scout)

    logger.info("Guardian Triad execution cycle complete.")


def run_agent_task(agent_id: str, dry_run: bool = False):
    if agent_id not in SQUAD_ROSTER:
        logger.error("Unknown agent '%s'. Available agents: %s", agent_id, ", ".join(SQUAD_ROSTER.keys()))
        sys.exit(1)

    cfg = SQUAD_ROSTER[agent_id]
    logger.info("Running execution cycle for [%s] (%s)...", cfg.id, cfg.name)

    if cfg.id == "janitor_memory":
        cmd = ["python3", ".agents/scripts/manage_issue_lifecycle.py", "--sync", "--recover-stale"]
        if dry_run:
            cmd.append("--dry-run")
        subprocess.run(cmd)
        cmd2 = ["python3", ".agents/scripts/consolidate_memory.py"]
        if dry_run:
            cmd2.append("--dry-run")
        subprocess.run(cmd2)
        return

    if cfg.id == "pr_reviewer":
        cmd = ["python3", ".agents/scripts/pr_review_agent.py", "--review-and-merge"]
        if dry_run:
            cmd.append("--dry-run")
        subprocess.run(cmd)
        return

    if cfg.id == "code_scout":
        cmd = ["python3", ".agents/scripts/audit_codebase.py"]
        subprocess.run(cmd)
        return

    # Developer agents claim next task in their domain
    logger.info("Querying next eligible task for domain '%s'...", cfg.domain)
    cmd = ["python3", ".agents/scripts/get_next_issue.py", "--domain", cfg.domain, "--claim-by", cfg.id]
    if dry_run:
        cmd.append("--dry-run")

    res = subprocess.run(cmd, capture_output=True, text=True)
    try:
        data = json.loads(res.stdout or "{}")
        if data.get("status") == "no_tasks":
            logger.info("No tasks waiting in queue for %s (%s).", cfg.name, data.get("message"))
        elif data.get("status") in ("claimed", "ready"):
            issue = data.get("issue", {})
            logger.info("Agent [%s] assigned to Issue #%s: %s (Branch: %s)",
                        cfg.id, issue.get("number"), issue.get("title"), issue.get("branch_name"))
        else:
            print(res.stdout or res.stderr)
    except Exception:
        print(res.stdout or res.stderr)


def main():
    parser = argparse.ArgumentParser(description="Hexta Multi-Agent Squad Orchestrator")
    parser.add_argument("--status", action="store_true", help="Print status of all 10 agents")
    parser.add_argument("--monitors", action="store_true", help="Run the 3 guardian monitor agents")
    parser.add_argument("--agent", help="Run a specific agent cycle by ID")
    parser.add_argument("--dry-run", action="store_true", help="Preview mode without mutating state")
    args = parser.parse_args()

    if args.status or (not args.monitors and not args.agent):
        print_roster_status()
        if not args.monitors and not args.agent:
            return

    if args.monitors:
        run_monitors(dry_run=args.dry_run)

    if args.agent:
        run_agent_task(args.agent, dry_run=args.dry_run)


if __name__ == "__main__":
    main()
