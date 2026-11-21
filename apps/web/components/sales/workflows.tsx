"use client";

import { useEffect, useRef, useState } from "react";
import { useWorkspace } from "./workspace";
import {
  buttonClass,
  secondaryClass,
  inputClass,
  EmptyState,
  Field,
  Notice,
  Pager,
  SegmentedFilter,
  StatusChip,
} from "./primitives";
import {
  workflows,
  type WorkflowDefinition,
  type WorkflowStepInput,
} from "@/lib/services/workflows";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";

type Mode = WorkflowStepInput["handoff_mode"];

function WorkflowEditor({
  definition,
  onSaved,
}: {
  definition: WorkflowDefinition | null;
  onSaved: (row: WorkflowDefinition) => void;
}) {
  const { tenant } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
  const [name, setName] = useState(definition?.name || "");
  const [steps, setSteps] = useState<WorkflowStepInput[]>(
    definition?.versions[0]?.steps.map(({ name, handoff_mode }) => ({ name, handoff_mode })) || [
      { name: "", handoff_mode: "acceptance" },
    ],
  );
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const retry = useRef({ payload: "", key: "" });
  function move(index: number, direction: number) {
    const next = [...steps];
    [next[index], next[index + direction]] = [next[index + direction], next[index]];
    setSteps(next);
  }
  function update(index: number, patch: Partial<WorkflowStepInput>) {
    setSteps(steps.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  }
  async function publish(event: React.FormEvent) {
    event.preventDefault();
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    const input = { name, expected_version: definition?.version || 0, steps };
    const payload = JSON.stringify(input);
    if (retry.current.payload !== payload) retry.current = { payload, key: crypto.randomUUID() };
    try {
      onSaved(await workflows.publish(tenant.id, definition?.id || null, input, retry.current.key));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  const modes: { value: Mode; label: string }[] = [
    { value: "direct", label: t("workflows.editor.modes.direct") },
    { value: "acceptance", label: t("workflows.editor.modes.acceptance") },
  ];
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <h2 className="text-lg font-semibold">
        {definition
          ? t("workflows.editor.editTitle", { version: definition.version })
          : t("workflows.editor.newTitle")}
      </h2>
      <p className="rounded-lg bg-background p-3 text-sm">{t("workflows.editor.body")}</p>
      <Notice error={error} />
      <form onSubmit={publish} className="space-y-4">
        <fieldset disabled={busy} className="space-y-4">
          <Field title={t("workflows.editor.name")}>
            <input className={inputClass} required maxLength={120} value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <ol className="space-y-3">
            {steps.map((step, index) => (
              <li key={index} className="flex flex-wrap items-end gap-3 rounded-lg border border-border bg-background p-3">
                <span
                  aria-hidden="true"
                  className="inline-flex h-8 w-8 flex-none items-center justify-center rounded-full bg-surface font-semibold"
                >
                  {index + 1}
                </span>
                <div className="min-w-0 flex-[1_1_200px]">
                  <Field title={t("workflows.editor.step", { index: index + 1 })}>
                    <input
                      className={inputClass}
                      required
                      maxLength={120}
                      value={step.name}
                      onChange={(e) => update(index, { name: e.target.value })}
                    />
                  </Field>
                </div>
                <SegmentedFilter
                  label={t("workflows.editor.mode", { index: index + 1 })}
                  options={modes}
                  value={step.handoff_mode}
                  onChange={(value) => update(index, { handoff_mode: value })}
                />
                <div className="flex gap-1">
                  <button
                    type="button"
                    className={secondaryClass}
                    aria-label={t("workflows.editor.moveUpLabel", { index: index + 1 })}
                    disabled={index === 0}
                    onClick={() => move(index, -1)}
                  >
                    {t("workflows.editor.moveUp")}
                  </button>
                  <button
                    type="button"
                    className={secondaryClass}
                    aria-label={t("workflows.editor.moveDownLabel", { index: index + 1 })}
                    disabled={index === steps.length - 1}
                    onClick={() => move(index, 1)}
                  >
                    {t("workflows.editor.moveDown")}
                  </button>
                  <button
                    type="button"
                    className={secondaryClass}
                    aria-label={t("workflows.editor.remove", { index: index + 1 })}
                    disabled={steps.length === 1}
                    onClick={() => setSteps(steps.filter((_, i) => i !== index))}
                  >
                    ×
                  </button>
                </div>
                {index === steps.length - 1 && (
                  <p className="w-full text-xs text-muted">{t("workflows.editor.lastStepNote")}</p>
                )}
              </li>
            ))}
          </ol>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <button
              type="button"
              className={secondaryClass}
              disabled={steps.length >= 30}
              onClick={() => setSteps([...steps, { name: "", handoff_mode: "acceptance" }])}
            >
              {t("workflows.editor.add")}
            </button>
            <span className="text-sm text-muted">{t("workflows.editor.count", { count: steps.length })}</span>
          </div>
          <button className={buttonClass} type="submit">
            {busy
              ? t("workflows.editor.publishing")
              : definition
                ? t("workflows.editor.publishNew", { version: definition.version + 1 })
                : t("workflows.editor.publish")}
          </button>
        </fieldset>
      </form>
      <div className="space-y-1 border-t border-border pt-4 text-sm">
        <h3 className="font-semibold">{t("workflows.editor.explainTitle")}</h3>
        <p>{t("workflows.editor.explainDirect")}</p>
        <p>{t("workflows.editor.explainAcceptance")}</p>
      </div>
    </section>
  );
}

export function Workflows() {
  const { tenant, isOwner: owner } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
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
        const data = await workflows.definitions(tenant!.id, page, abort.signal);
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
  }, [tenant, page, revision, errorText]);
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("workflows.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("workflows.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className={secondaryClass}
            onClick={() => {
              setSelected(null);
              setRevision(revision + 1);
            }}
          >
            {t("workflows.refresh")}
          </button>
          {owner && (
            <button
              type="button"
              className={buttonClass}
              onClick={() => {
                setSelected(null);
                setSuccess("");
              }}
            >
              {t("workflows.new")}
            </button>
          )}
        </div>
      </div>
      <Notice error={error} success={success} />
      {!owner && <p className="text-sm text-muted">{t("workflows.ownerOnly")}</p>}
      <div className="flex flex-wrap items-start gap-5">
        <section className="min-w-0 flex-[1_1_320px] space-y-4 rounded-xl border border-border bg-surface p-5">
          <h2 className="font-semibold">{t("workflows.list")}</h2>
          {loading ? (
            <p role="status">{t("workflows.loading")}</p>
          ) : !rows.length ? (
            <EmptyState title={t("workflows.emptyTitle")} description={t("workflows.emptyBody")} />
          ) : (
            rows.map((row) => (
              <article
                key={row.id}
                className={`space-y-2 rounded-lg border p-3 ${selected?.id === row.id ? "border-foreground" : "border-border"}`}
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <h3 className="font-medium">
                    {row.name} <StatusChip>{t("workflows.versionLabel", { number: row.version })}</StatusChip>
                  </h3>
                  {owner && (
                    <button
                      type="button"
                      className={secondaryClass}
                      onClick={() => {
                        setSelected(row);
                        setSuccess("");
                      }}
                    >
                      {t("workflows.edit")}
                    </button>
                  )}
                </div>
                <ol className="list-decimal pl-5 text-sm text-muted">
                  {row.versions[0]?.steps.map((step, i, all) => (
                    <li key={step.id}>
                      {step.name} ·{" "}
                      {i < all.length - 1 ? t(`workflows.modeShort.${step.handoff_mode}`) : t("workflows.finalStep")}
                    </li>
                  ))}
                </ol>
              </article>
            ))
          )}
          <Pager page={page} total={total} onChange={setPage} disabled={loading} />
        </section>
        {owner && (
          <div className="min-w-0 flex-[999_1_480px]">
            <WorkflowEditor
              key={selected ? `${selected.id}:${selected.version}` : `new:${revision}`}
              definition={selected}
              onSaved={(row) => {
                setSelected(row);
                setRevision((v) => v + 1);
                setSuccess(t("workflows.published", { name: row.name, version: row.version }));
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
