"use client";

import { useT } from "@/lib/i18n/provider";
import type { Overview } from "../types";
import { groupByWorkflow } from "../utils";

/** Work counts per step of each pinned workflow version, with the cohort totals on top. */
export function WorkByStepCard({ work }: { work: Overview["work"] }) {
  const t = useT();
  const workflows = groupByWorkflow(work.by_steps);
  return (
    <section aria-labelledby="overview-work" className="space-y-3 rounded-xl border border-border bg-surface p-4">
      <div>
        <h2 id="overview-work" className="font-semibold">
          {t("overview.work.title")}
        </h2>
        <p className="mt-1 text-sm text-muted">
          {t("overview.work.summary", { active: work.active, waiting: work.waiting, completed: work.completed })}
          {work.unassigned > 0 && ` · ${t("overview.work.unassigned", { count: work.unassigned })}`}
        </p>
      </div>
      {workflows.length === 0 ? (
        <p className="text-sm text-muted">{t("overview.work.empty")}</p>
      ) : (
        workflows.map((workflow) => (
          <div key={workflow.key} className="space-y-1">
            <h3 className="text-sm font-medium">{t("overview.work.workflow", { name: workflow.name, version: workflow.version })}</h3>
            <ul className="divide-y divide-border text-sm">
              {workflow.steps.map((step) => (
                <li key={step.step_ordinal} className="flex flex-wrap justify-between gap-x-3 gap-y-1 py-2">
                  <span>
                    {step.step_ordinal + 1}. {step.step_name}
                  </span>
                  <span className="text-muted tabular-nums">
                    {[
                      step.active || step.waiting ? t("overview.work.stepCounts", { active: step.active, waiting: step.waiting }) : "",
                      step.unassigned ? t("overview.work.stepUnassigned", { count: step.unassigned }) : "",
                      step.completed ? t("overview.work.completed", { count: step.completed }) : "",
                    ]
                      .filter(Boolean)
                      .join(" · ")}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        ))
      )}
      <p className="text-xs text-muted">{t("overview.work.note")}</p>
    </section>
  );
}
