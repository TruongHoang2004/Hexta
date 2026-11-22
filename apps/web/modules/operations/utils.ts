import type { ChipTone } from "@/components/primitives";
import type { Translate } from "@/lib/i18n/translate";
import type { WorkEvent, WorkflowDefinition, WorkflowStepInput, WorkItem, WorkView } from "./types";

export const stateTone: Record<WorkItem["state"], ChipTone> = {
  active: "neutral",
  waiting: "warning",
  completed: "success",
};

export const workViews = ["mine", "awaiting_me", "all"] as const satisfies readonly WorkView[];

const eventActions = ["create", "handoff", "accept", "assign", "complete"] as const;

export function eventActionLabel(t: Translate, action: WorkEvent["action"]) {
  return (eventActions as readonly string[]).includes(action)
    ? t(`work.history.actions.${action as (typeof eventActions)[number]}`)
    : t("work.history.actions.other");
}

export const blankStep = (): WorkflowStepInput => ({ name: "", handoff_mode: "acceptance" });

/** Editable steps of the definition's latest version, or one blank step for a new workflow. */
export function stepsFrom(definition: WorkflowDefinition | null): WorkflowStepInput[] {
  return definition?.versions[0]?.steps.map(({ name, handoff_mode }) => ({ name, handoff_mode })) || [blankStep()];
}

export function swap<T>(rows: T[], index: number, direction: number): T[] {
  const next = [...rows];
  [next[index], next[index + direction]] = [next[index + direction], next[index]];
  return next;
}
