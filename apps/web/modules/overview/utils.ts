import type { Overview, OverviewPeriod, OverviewStep } from "./types";

export const overviewPeriods: OverviewPeriod[] = ["today", "7d", "30d", "all"];

const periodDays: Record<Exclude<OverviewPeriod, "all">, number> = { today: 1, "7d": 7, "30d": 30 };

/**
 * Half-open creation range [from, to) for a preset, in the browser's local time zone: whole local
 * days ending with today. "all" has no bounds.
 */
export function periodRange(period: OverviewPeriod, now: Date): { from?: string; to?: string } {
  if (period === "all") return {};
  const to = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
  const from = new Date(to.getFullYear(), to.getMonth(), to.getDate() - periodDays[period]);
  return { from: from.toISOString(), to: to.toISOString() };
}

/** Elapsed time since `since`, in the largest whole unit (minimum one minute). */
export function waitedFor(since: string, now: Date): { unit: "days" | "hours" | "minutes"; value: number } {
  const minutes = Math.max(1, Math.floor((now.getTime() - new Date(since).getTime()) / 60000));
  if (minutes >= 24 * 60) return { unit: "days", value: Math.floor(minutes / (24 * 60)) };
  if (minutes >= 60) return { unit: "hours", value: Math.floor(minutes / 60) };
  return { unit: "minutes", value: minutes };
}

/** True when the workspace has data, but none in the selected period/employee. */
export function isEmptySelection(data: Overview): boolean {
  return data.has_activity && data.orders.total === 0 && data.work.total === 0 && data.stuck.total === 0;
}

export interface WorkflowSteps {
  key: string;
  name: string;
  version: number;
  steps: OverviewStep[];
}

/** Groups step counts by pinned workflow version, keeping API order (name, version, step). */
export function groupByWorkflow(steps: OverviewStep[]): WorkflowSteps[] {
  const groups: WorkflowSteps[] = [];
  for (const step of steps) {
    let group = groups.find((row) => row.key === step.workflow_version_id);
    if (!group) {
      group = { key: step.workflow_version_id, name: step.workflow_name, version: step.workflow_version_number, steps: [] };
      groups.push(group);
    }
    group.steps.push(step);
  }
  return groups;
}
