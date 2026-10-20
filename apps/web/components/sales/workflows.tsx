"use client";

import { useEffect, useRef, useState } from "react";
import { useAuthStore } from "@/store/useAuthStore";
import { useWorkspace } from "./workspace";
import {
  buttonClass,
  secondaryClass,
  inputClass,
  Field,
  Notice,
  Pager,
} from "./primitives";
import { errorText } from "@/lib/services/sales";
import {
  workflows,
  type WorkflowDefinition,
  type WorkflowStepInput,
} from "@/lib/services/workflows";

function WorkflowEditor({
  definition,
  onSaved,
}: {
  definition: WorkflowDefinition | null;
  onSaved: (row: WorkflowDefinition) => void;
}) {
  const { tenant } = useWorkspace();
  const [name, setName] = useState(definition?.name || "");
  const [steps, setSteps] = useState<WorkflowStepInput[]>(
    definition?.versions[0]?.steps.map(({ name, handoff_mode }) => ({
      name,
      handoff_mode,
    })) || [{ name: "", handoff_mode: "acceptance" }],
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const retry = useRef({ payload: "", key: "" });
  function move(index: number, direction: number) {
    const next = [...steps];
    [next[index], next[index + direction]] = [
      next[index + direction],
      next[index],
    ];
    setSteps(next);
  }
  async function publish(event: React.FormEvent) {
    event.preventDefault();
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    const input = { name, expected_version: definition?.version || 0, steps };
    const payload = JSON.stringify(input);
    if (retry.current.payload !== payload)
      retry.current = { payload, key: crypto.randomUUID() };
    try {
      onSaved(
        await workflows.publish(
          tenant.id,
          definition?.id || null,
          input,
          retry.current.key,
        ),
      );
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="rounded-xl border border-border bg-surface p-5 space-y-4">
      <h2 className="text-lg font-semibold">
        {definition
          ? `Edit workflow · version ${definition.version}`
          : "New workflow"}
      </h2>
      <p className="text-sm text-muted">
        Publish ordered steps for your business. Existing work keeps its
        original steps.
      </p>
      <Notice error={error} />
      <form onSubmit={publish} className="space-y-4">
        <fieldset disabled={busy} className="space-y-4">
          <Field title="Workflow name">
            <input
              className={inputClass}
              required
              maxLength={120}
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </Field>
          {steps.map((step, index) => (
            <div
              key={index}
              className="rounded-lg border border-border p-3 space-y-3"
            >
              <Field title={`Step ${index + 1}`}>
                <input
                  className={inputClass}
                  required
                  maxLength={120}
                  value={step.name}
                  onChange={(e) =>
                    setSteps(
                      steps.map((row, i) =>
                        i === index ? { ...row, name: e.target.value } : row,
                      ),
                    )
                  }
                />
              </Field>
              <Field title="Handoff to the next step">
                <select
                  className={inputClass}
                  value={step.handoff_mode}
                  onChange={(e) =>
                    setSteps(
                      steps.map((row, i) =>
                        i === index
                          ? {
                              ...row,
                              handoff_mode: e.target
                                .value as WorkflowStepInput["handoff_mode"],
                            }
                          : row,
                      ),
                    )
                  }
                >
                  <option value="acceptance">Recipient must accept</option>
                  <option value="direct">Transfer directly</option>
                </select>
              </Field>
              {index === steps.length - 1 && (
                <p className="text-xs text-muted">
                  The last step is completed directly; this policy applies if
                  another step is added later.
                </p>
              )}
              <div className="flex flex-wrap gap-2">
                <button
                  type="button"
                  className={secondaryClass}
                  disabled={index === 0}
                  onClick={() => move(index, -1)}
                >
                  Move up
                </button>
                <button
                  type="button"
                  className={secondaryClass}
                  disabled={index === steps.length - 1}
                  onClick={() => move(index, 1)}
                >
                  Move down
                </button>
                <button
                  type="button"
                  className={secondaryClass}
                  disabled={steps.length === 1}
                  onClick={() => setSteps(steps.filter((_, i) => i !== index))}
                >
                  Remove step
                </button>
              </div>
            </div>
          ))}
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              className={secondaryClass}
              disabled={steps.length >= 30}
              onClick={() =>
                setSteps([...steps, { name: "", handoff_mode: "acceptance" }])
              }
            >
              Add step
            </button>
            <button className={buttonClass} type="submit">
              {busy
                ? "Publishing…"
                : definition
                  ? "Publish new version"
                  : "Publish workflow"}
            </button>
          </div>
        </fieldset>
      </form>
    </section>
  );
}

export function Workflows() {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const [rows, setRows] = useState<WorkflowDefinition[]>([]);
  const [selected, setSelected] = useState<WorkflowDefinition | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  useEffect(() => {
    if (!tenant) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const data = await workflows.definitions(
          tenant!.id,
          page,
          abort.signal,
        );
        if (!abort.signal.aborted) {
          setRows(data.items);
          setTotal(data.total);
        }
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenant, page, revision]);
  const owner = tenant?.owner_id === user?.id;
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap justify-between gap-3">
        <div>
          <h1 className="text-3xl font-semibold">Workflows</h1>
          <p className="mt-2 text-sm text-muted">
            Choose the steps your business needs.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            className={secondaryClass}
            onClick={() => {
              setSelected(null);
              setRevision(revision + 1);
            }}
          >
            Refresh
          </button>
          {owner && (
            <button
              className={buttonClass}
              onClick={() => {
                setSelected(null);
                setSuccess("");
              }}
            >
              New workflow
            </button>
          )}
        </div>
      </div>
      <Notice error={error} success={success} />
      {!owner && (
        <p className="text-sm text-muted">
          The tenant owner manages workflow definitions.
        </p>
      )}
      <div className="grid items-start gap-5 xl:grid-cols-2">
        <section className="rounded-xl border border-border bg-surface p-5 space-y-4">
          <h2 className="font-semibold">Published workflows</h2>
          {loading ? (
            <p role="status">Loading workflows…</p>
          ) : !rows.length ? (
            <p className="text-sm text-muted">
              No workflows yet. The owner can publish the first one.
            </p>
          ) : (
            rows.map((row) => (
              <div
                key={row.id}
                className="border-b border-border py-3 space-y-2"
              >
                <div className="flex justify-between gap-2">
                  <h3 className="font-medium">
                    {row.name} · v{row.version}
                  </h3>
                  {owner && (
                    <button
                      className={secondaryClass}
                      onClick={() => {
                        setSelected(row);
                        setSuccess("");
                      }}
                    >
                      Edit
                    </button>
                  )}
                </div>
                <ol className="list-decimal pl-5 text-sm text-muted">
                  {row.versions[0]?.steps.map((step, i, all) => (
                    <li key={step.id}>
                      {step.name}
                      {i < all.length - 1
                        ? ` · ${step.handoff_mode === "acceptance" ? "recipient accepts" : "direct handoff"}`
                        : " · final step"}
                    </li>
                  ))}
                </ol>
              </div>
            ))
          )}
          <Pager
            page={page}
            total={total}
            onChange={setPage}
            disabled={loading}
          />
        </section>
        {owner && (
          <WorkflowEditor
            key={
              selected
                ? `${selected.id}:${selected.version}`
                : `new:${revision}`
            }
            definition={selected}
            onSaved={(row) => {
              setSelected(row);
              setRevision((v) => v + 1);
              setSuccess(`Published ${row.name}, version ${row.version}.`);
            }}
          />
        )}
      </div>
    </div>
  );
}
