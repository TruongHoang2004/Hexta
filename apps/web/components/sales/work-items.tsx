"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import type { TenantMember } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { useWorkspace } from "./workspace";
import {
  buttonClass,
  secondaryClass,
  inputClass,
  ConfirmDialog,
  ConflictBanner,
  EmptyState,
  Field,
  Notice,
  Pager,
  StatusChip,
  Tabs,
  type ChipTone,
} from "./primitives";
import { sales, type Order } from "@/lib/services/sales";
import {
  workflows,
  type WorkflowDefinition,
  type WorkItem,
  type WorkView,
} from "@/lib/services/workflows";
import { isConflict, useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { Translate } from "@/lib/i18n/translate";

type Command = "handoff" | "accept" | "assign" | "complete";
type CommandFields = { recipient_id?: string; responsible_id?: string | null };
const stateTone: Record<WorkItem["state"], ChipTone> = {
  active: "neutral",
  waiting: "warning",
  completed: "success",
};
const eventActions = ["create", "handoff", "accept", "assign", "complete"] as const;

function usePerson() {
  const { user } = useAuthStore();
  const t = useT();
  return (id: string | null) => (!id ? t("work.unassigned") : id === user?.id ? t("work.you") : id);
}

/** Compact step progress: done steps, current step highlighted, later steps muted. */
function StepTrack({ item, t }: { item: WorkItem; t: Translate }) {
  return (
    <ol aria-label={t("work.progress")} className="flex flex-wrap items-center gap-1 text-xs">
      {item.workflow.steps.map((step) => {
        const done = step.ordinal < item.current_step || item.state === "completed";
        const current = step.ordinal === item.current_step && item.state !== "completed";
        return (
          <li
            key={step.id}
            aria-current={current ? "step" : undefined}
            className={`rounded px-2 py-1 ${current ? "bg-foreground font-semibold text-background" : done ? "bg-border text-muted line-through" : "bg-background text-muted"}`}
          >
            {step.ordinal + 1}. {step.name}
          </li>
        );
      })}
    </ol>
  );
}

function WorkCreator({
  definitions,
  orders,
  members,
  busy,
  onCreate,
}: {
  definitions: WorkflowDefinition[];
  orders: Order[];
  members: TenantMember[];
  busy: boolean;
  onCreate: (
    input: { order_id: string; workflow_version_id: string; responsible_id: string },
    key: string,
  ) => Promise<void>;
}) {
  const t = useT();
  const person = usePerson();
  const [order, setOrder] = useState("");
  const [definition, setDefinition] = useState("");
  const [responsible, setResponsible] = useState("");
  const retry = useRef({ payload: "", key: "" });
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (busy) return;
    const input = { order_id: order, workflow_version_id: definition, responsible_id: responsible };
    const payload = JSON.stringify(input);
    if (retry.current.payload !== payload) retry.current = { payload, key: crypto.randomUUID() };
    await onCreate(input, retry.current.key);
  }
  return (
    <details className="rounded-xl border border-border bg-surface p-4">
      <summary className="cursor-pointer font-semibold">{t("work.creator.title")}</summary>
      <p className="my-3 text-sm text-muted">{t("work.creator.body")}</p>
      <form onSubmit={(e) => void submit(e)}>
        <fieldset disabled={busy || !orders.length || !definitions.length} className="grid gap-3 md:grid-cols-3">
          <Field title={t("work.creator.order")}>
            <select className={inputClass} required value={order} onChange={(e) => setOrder(e.target.value)}>
              <option value="">{t("work.creator.selectOrder")}</option>
              {orders.map((row) => (
                <option key={row.id} value={row.id}>
                  {row.order_number}
                </option>
              ))}
            </select>
          </Field>
          <Field title={t("work.creator.workflow")}>
            <select className={inputClass} required value={definition} onChange={(e) => setDefinition(e.target.value)}>
              <option value="">{t("work.creator.selectWorkflow")}</option>
              {definitions.map((row) => (
                <option key={row.id} value={row.versions[0]?.id}>
                  {row.name} · v{row.version}
                </option>
              ))}
            </select>
          </Field>
          <Field title={t("work.creator.responsible")}>
            <select className={inputClass} required value={responsible} onChange={(e) => setResponsible(e.target.value)}>
              <option value="">{t("work.creator.selectEmployee")}</option>
              {members.map((row) => (
                <option key={row.id} value={row.user_id}>
                  {person(row.user_id)}
                </option>
              ))}
            </select>
          </Field>
          <div className="md:col-span-3">
            <button className={buttonClass} type="submit">
              {t("work.creator.submit")}
            </button>
          </div>
        </fieldset>
      </form>
      {!orders.length && <p className="mt-2 text-sm text-muted">{t("work.creator.noOrders")}</p>}
      {!definitions.length && <p className="mt-2 text-sm text-muted">{t("work.creator.noWorkflows")}</p>}
    </details>
  );
}

function WorkDetail({
  item,
  members,
  busy,
  onCommand,
  onRefresh,
}: {
  item: WorkItem;
  members: TenantMember[];
  busy: boolean;
  onCommand: (action: Command, fields?: CommandFields) => Promise<boolean>;
  onRefresh: () => void;
}) {
  const { isOwner } = useWorkspace();
  const { user } = useAuthStore();
  const t = useT();
  const person = usePerson();
  const [recipient, setRecipient] = useState("");
  const [handoffOpen, setHandoffOpen] = useState(false);
  const [responsible, setResponsible] = useState(item.responsible_id || "");
  const canProgress = item.responsible_id === user?.id && item.state === "active";
  const finalStep = item.current_step === item.workflow.steps.length - 1;
  const currentStep = item.workflow.steps[item.current_step];
  const nextStep = item.workflow.steps[item.current_step + 1];
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex flex-wrap justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{item.workflow.name}</h2>
          <p className="text-xs text-muted">
            {t("work.workflowVersion", { number: item.workflow.number, version: item.version })}
          </p>
          <p className="break-all text-xs text-muted">{t("work.order", { id: item.order_id })}</p>
        </div>
        <StatusChip tone={stateTone[item.state]}>{t(`work.states.${item.state}`)}</StatusChip>
      </div>
      <StepTrack item={item} t={t} />
      <p className="break-all text-sm">
        {t("work.responsible")}: <strong>{person(item.responsible_id)}</strong>
      </p>
      {item.state === "waiting" && (
        <p className="break-all rounded-lg border border-warning bg-warning/15 p-3 text-sm">
          {t("work.waitingFor", { name: person(item.recipient_id), responsible: person(item.responsible_id) })}
        </p>
      )}
      <div className="flex flex-wrap gap-2">
        {canProgress && !finalStep && (
          <button type="button" className={buttonClass} disabled={busy} onClick={() => setHandoffOpen(true)}>
            {t("work.actions.handoff")}
          </button>
        )}
        {item.state === "waiting" && item.recipient_id === user?.id && (
          <button type="button" className={buttonClass} disabled={busy} onClick={() => void onCommand("accept")}>
            {t("work.actions.accept")}
          </button>
        )}
        {canProgress && finalStep && (
          <button type="button" className={buttonClass} disabled={busy} onClick={() => void onCommand("complete")}>
            {t("work.actions.complete")}
          </button>
        )}
        {isOwner && item.state === "waiting" && (
          <button
            type="button"
            className={secondaryClass}
            disabled={busy}
            onClick={() => void onCommand("assign", { responsible_id: item.responsible_id })}
          >
            {t("work.actions.cancelWait")}
          </button>
        )}
        <button type="button" className={secondaryClass} disabled={busy} onClick={onRefresh}>
          {t("work.actions.reload")}
        </button>
      </div>
      {isOwner && item.state !== "completed" && (
        <fieldset disabled={busy} className="space-y-3 border-t border-border pt-4">
          <legend className="sr-only">{t("work.owner.title")}</legend>
          <Field title={t("work.owner.reassignLabel")}>
            <select className={inputClass} value={responsible} onChange={(e) => setResponsible(e.target.value)}>
              <option value="">{t("work.owner.queue")}</option>
              {item.responsible_id && !members.some((row) => row.user_id === item.responsible_id) && (
                <option value={item.responsible_id}>{t("work.formerMember")}</option>
              )}
              {members.map((row) => (
                <option key={row.id} value={row.user_id}>
                  {person(row.user_id)}
                </option>
              ))}
            </select>
          </Field>
          <p className="text-xs text-muted">{t("work.owner.reassignNote")}</p>
          <button
            type="button"
            className={secondaryClass}
            onClick={() => void onCommand("assign", { responsible_id: responsible || null })}
          >
            {t("work.actions.reassign")}
          </button>
        </fieldset>
      )}
      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="font-semibold">{t("work.history.title")}</h3>
        <p className="text-xs text-muted">{t("work.history.note")}</p>
        <ol className="space-y-3">
          {item.events?.map((event) => (
            <li key={event.id} className="space-y-1 break-all rounded-lg bg-background p-3 text-xs">
              <p className="font-medium">
                v{event.version} ·{" "}
                {(eventActions as readonly string[]).includes(event.action)
                  ? t(`work.history.actions.${event.action as (typeof eventActions)[number]}`)
                  : t("work.history.actions.other")}
              </p>
              <p>
                {t("work.history.actor", {
                  name: person(event.actor_id),
                  from: event.from_step + 1,
                  to: event.current_step + 1,
                })}
              </p>
              <p>
                {t("work.history.responsible", {
                  from: person(event.from_responsible_id),
                  to: person(event.responsible_id),
                })}
                {event.recipient_id ? t("work.history.nominated", { name: person(event.recipient_id) }) : ""}
              </p>
              <p className="text-muted">{new Date(event.created_at).toLocaleString()}</p>
            </li>
          ))}
        </ol>
      </div>
      <ConfirmDialog
        open={handoffOpen}
        title={t("work.handoffDialog.title")}
        confirmLabel={t("work.handoffDialog.confirm")}
        cancelLabel={t("common.cancel")}
        busy={busy}
        confirmDisabled={!recipient}
        onCancel={() => setHandoffOpen(false)}
        onConfirm={async () => {
          if (await onCommand("handoff", { recipient_id: recipient })) setHandoffOpen(false);
        }}
      >
        {nextStep && <p className="font-medium">{t("work.handoffDialog.next", { name: nextStep.name })}</p>}
        <Field title={t("work.handoffDialog.recipient")}>
          <select className={inputClass} value={recipient} onChange={(e) => setRecipient(e.target.value)}>
            <option value="">{t("work.handoffDialog.selectRecipient")}</option>
            {members
              .filter((row) => row.user_id !== user?.id)
              .map((row) => (
                <option key={row.id} value={row.user_id}>
                  {person(row.user_id)}
                </option>
              ))}
          </select>
        </Field>
        <p className="rounded-lg bg-background p-3">
          {currentStep?.handoff_mode === "acceptance"
            ? t("work.handoffDialog.acceptance")
            : t("work.handoffDialog.direct")}
        </p>
      </ConfirmDialog>
    </section>
  );
}

export function WorkItems() {
  const { tenant, isOwner } = useWorkspace();
  const { user } = useAuthStore();
  const t = useT();
  const errorText = useErrorText();
  const person = usePerson();
  const [view, setView] = useState<WorkView>("mine");
  const [rows, setRows] = useState<WorkItem[]>([]);
  const [selected, setSelected] = useState<WorkItem | null>(null);
  const [definitions, setDefinitions] = useState<WorkflowDefinition[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [revision, setRevision] = useState(0);
  const [creatorKey, setCreatorKey] = useState(0);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [lookupLoading, setLookupLoading] = useState(true);
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [lookupError, setLookupError] = useState("");
  const [success, setSuccess] = useState("");
  useEffect(() => {
    if (!tenant) return;
    const abort = new AbortController();
    async function load() {
      setLookupLoading(true);
      setLookupError("");
      try {
        const [d, o, m] = await Promise.all([
          workflows.allDefinitions(tenant!.id, abort.signal),
          sales.all<Order>("orders", tenant!.id, abort.signal),
          sdk.identity.getUsers(tenant!.id),
        ]);
        if (!abort.signal.aborted) {
          setDefinitions(d);
          setOrders(o.filter((row) => row.status === "confirmed"));
          setMembers(m);
        }
      } catch (err) {
        if (!abort.signal.aborted) setLookupError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLookupLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenant, revision, errorText]);
  useEffect(() => {
    if (!tenant) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const data = await workflows.items(tenant!.id, page, abort.signal, view);
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
  }, [tenant, page, view, revision, errorText]);
  function fail(err: unknown) {
    setConflict(isConflict(err));
    setError(errorText(err));
  }
  async function open(id: string) {
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    setConflict(false);
    setSuccess("");
    try {
      setSelected(await workflows.item(tenant.id, id));
    } catch (err) {
      fail(err);
    } finally {
      setBusy(false);
    }
  }
  async function create(
    input: { order_id: string; workflow_version_id: string; responsible_id: string },
    key: string,
  ) {
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    setConflict(false);
    setSuccess("");
    try {
      setSelected(await workflows.create(tenant.id, input, key));
      setCreatorKey((v) => v + 1);
      setRevision((v) => v + 1);
      setSuccess(t("work.created"));
    } catch (err) {
      fail(err);
    } finally {
      setBusy(false);
    }
  }
  async function command(action: Command, fields?: CommandFields): Promise<boolean> {
    if (!tenant || !selected || busy) return false;
    setBusy(true);
    setError("");
    setConflict(false);
    setSuccess("");
    try {
      setSelected(await workflows.command(tenant.id, selected, action, fields));
      setRevision((v) => v + 1);
      setSuccess(t("work.saved"));
      return true;
    } catch (err) {
      fail(err);
      return false;
    } finally {
      setBusy(false);
    }
  }
  const tabs = (["mine", "awaiting_me", "all"] as const).map((value) => ({
    value,
    label: t(`work.tabs.${value}`),
  }));
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("work.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("work.subtitle")}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link className={secondaryClass} href={`/tenants/${tenant?.id}/workflows`}>
            {t("work.manageWorkflows")}
          </Link>
          <button
            type="button"
            className={secondaryClass}
            disabled={busy || loading}
            onClick={() => {
              setError("");
              setRevision((v) => v + 1);
              if (selected) void open(selected.id);
            }}
          >
            {t("work.refresh")}
          </button>
        </div>
      </div>
      {conflict && selected ? (
        <ConflictBanner message={error} actionLabel={t("work.actions.reload")} onReload={() => void open(selected.id)} />
      ) : (
        <Notice error={error || lookupError} success={success} />
      )}
      {lookupLoading ? (
        <p role="status">{t("work.loadingLookups")}</p>
      ) : (
        !lookupError &&
        isOwner && (
          <WorkCreator
            key={creatorKey}
            definitions={definitions}
            orders={orders}
            members={members}
            busy={busy}
            onCreate={create}
          />
        )
      )}
      <div className="flex flex-wrap items-start gap-5">
        <section className="min-w-0 flex-[1_1_320px] space-y-4 rounded-xl border border-border bg-surface p-4 xl:max-w-[460px]">
          <Tabs
            label={t("work.tabsLabel")}
            idPrefix="work"
            items={tabs}
            value={view}
            onChange={(value) => {
              setView(value);
              setPage(1);
            }}
          />
          <div id="work-panel" role="tabpanel" aria-labelledby={`work-tab-${view}`} className="space-y-2">
            {loading ? (
              <p role="status">{t("work.loading")}</p>
            ) : !rows.length ? (
              <EmptyState title={t(`work.empty.${view}`)} />
            ) : (
              rows.map((row) => {
                const step = row.workflow.steps[row.current_step];
                return (
                  <button
                    type="button"
                    key={row.id}
                    disabled={busy}
                    aria-current={selected?.id === row.id ? "true" : undefined}
                    className={`block w-full space-y-2 rounded-lg border p-3 text-left hover:bg-background ${selected?.id === row.id ? "border-foreground" : "border-border"}`}
                    onClick={() => void open(row.id)}
                  >
                    <span className="flex flex-wrap items-center justify-between gap-2">
                      <span className="font-medium">{row.workflow.name}</span>
                      <StatusChip tone={stateTone[row.state]}>{t(`work.states.${row.state}`)}</StatusChip>
                    </span>
                    <span className="block text-sm text-muted">
                      {t("work.step", {
                        current: row.current_step + 1,
                        total: row.workflow.steps.length,
                        name: step?.name ?? "",
                      })}
                    </span>
                    <span className="block break-all text-xs">
                      {t("work.responsible")}: {person(row.responsible_id)}
                      {row.recipient_id && row.state === "waiting"
                        ? ` · ${t("work.waitingShort", { name: row.recipient_id === user?.id ? t("work.you") : row.recipient_id })}`
                        : ""}
                    </span>
                  </button>
                );
              })
            )}
          </div>
          <Pager page={page} total={total} onChange={setPage} disabled={busy || loading} />
        </section>
        <div className="min-w-0 flex-[999_1_480px]">
          {selected ? (
            <WorkDetail
              key={`${selected.id}:${selected.version}`}
              item={selected}
              members={members}
              busy={busy}
              onCommand={command}
              onRefresh={() => void open(selected.id)}
            />
          ) : (
            <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted">
              {t("work.selectPrompt")}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
