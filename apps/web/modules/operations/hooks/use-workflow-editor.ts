"use client";

import { useState } from "react";
import { useErrorText } from "@/lib/errors";
import { workflows } from "../services/workflows";
import type { WorkflowDefinition, WorkflowStepInput } from "../types";
import { blankStep, stepsFrom, swap } from "../utils";
import { useRetryKey } from "./use-retry-key";

/** Name and ordered steps of a workflow; publishing creates a new version. */
export function useWorkflowEditor(
  tenantID: string,
  definition: WorkflowDefinition | null,
  onSaved: (row: WorkflowDefinition) => void,
) {
  const errorText = useErrorText();
  const retryKey = useRetryKey();
  const [name, setName] = useState(definition?.name || "");
  const [steps, setSteps] = useState<WorkflowStepInput[]>(() => stepsFrom(definition));
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const move = (index: number, direction: number) => setSteps(swap(steps, index, direction));
  const update = (index: number, patch: Partial<WorkflowStepInput>) =>
    setSteps(steps.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  const remove = (index: number) => setSteps(steps.filter((_, i) => i !== index));
  const add = () => setSteps([...steps, blankStep()]);

  async function publish() {
    if (!tenantID || busy) return;
    setBusy(true);
    setError("");
    const input = { name, expected_version: definition?.version || 0, steps };
    const key = retryKey(input);
    try {
      onSaved(await workflows.publish(tenantID, definition?.id || null, input, key));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  return { name, setName, steps, move, update, remove, add, busy, error, publish };
}
