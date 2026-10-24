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
  Field,
  Notice,
  Pager,
} from "./primitives";
import { sales, errorText, type Order } from "@/lib/services/sales";
import {
  workflows,
  type WorkflowDefinition,
  type WorkItem,
} from "@/lib/services/workflows";

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
    input: {
      order_id: string;
      workflow_version_id: string;
      responsible_id: string;
    },
    key: string,
  ) => Promise<void>;
}) {
  const [order, setOrder] = useState("");
  const [definition, setDefinition] = useState("");
  const [responsible, setResponsible] = useState("");
  const retry = useRef({ payload: "", key: "" });
  const { user } = useAuthStore();
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (busy) return;
    const input = {
      order_id: order,
      workflow_version_id: definition,
      responsible_id: responsible,
    };
    const payload = JSON.stringify(input);
    if (retry.current.payload !== payload)
      retry.current = { payload, key: crypto.randomUUID() };
    await onCreate(input, retry.current.key);
  }
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <h2 className="font-semibold">Create work from an order</h2>
      <p className="text-sm text-muted">
        Choose a confirmed order and its own workflow. This is separate from
        order confirmation.
      </p>
      <form onSubmit={(e) => void submit(e)}>
        <fieldset
          disabled={busy || !orders.length || !definitions.length}
          className="space-y-3"
        >
          <Field title="Confirmed order">
            <select
              className={inputClass}
              required
              value={order}
              onChange={(e) => setOrder(e.target.value)}
            >
              <option value="">Select order</option>
              {orders.map((row) => (
                <option key={row.id} value={row.id}>
                  {row.order_number} · {row.total_amount}
                </option>
              ))}
            </select>
          </Field>
          <Field title="Workflow">
            <select
              className={inputClass}
              required
              value={definition}
              onChange={(e) => setDefinition(e.target.value)}
            >
              <option value="">Select workflow</option>
              {definitions.map((row) => (
                <option key={row.id} value={row.versions[0]?.id}>
                  {row.name} · v{row.version}
                </option>
              ))}
            </select>
          </Field>
          <Field title="Initial responsible employee">
            <select
              className={inputClass}
              required
              value={responsible}
              onChange={(e) => setResponsible(e.target.value)}
            >
              <option value="">Select employee</option>
              {members.map((row) => (
                <option key={row.id} value={row.user_id}>
                  {row.user_id === user?.id ? "You" : row.user_id}
                </option>
              ))}
            </select>
          </Field>
          <button className={buttonClass} type="submit">
            Create work
          </button>
        </fieldset>
      </form>
      {!orders.length && (
        <p className="text-sm text-muted">Confirm an order first.</p>
      )}
      {!definitions.length && (
        <p className="text-sm text-muted">Publish a workflow first.</p>
      )}
    </section>
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
  onCommand: (
    action: "handoff" | "accept" | "assign" | "complete",
    fields?: { recipient_id?: string; responsible_id?: string | null },
  ) => void;
  onRefresh: () => void;
}) {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const [recipient, setRecipient] = useState("");
  const [responsible, setResponsible] = useState(item.responsible_id || "");
  const label = (id: string | null) =>
    !id ? "Unassigned" : id === user?.id ? "You" : id;
  const canProgress =
    item.responsible_id === user?.id && item.state === "active";
  const finalStep = item.current_step === item.workflow.steps.length - 1;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{item.workflow.name}</h2>
          <p className="text-xs text-muted">
            Workflow v{item.workflow.number} · Work version {item.version}
          </p>
        </div>
        <span className="text-sm font-medium uppercase text-primary">
          {item.state}
        </span>
      </div>
      <p className="break-all text-xs text-muted">Order {item.order_id}</p>
      <ol className="space-y-2">
        {item.workflow.steps.map((step) => (
          <li
            key={step.id}
            className={`rounded-lg border p-3 text-sm ${step.ordinal === item.current_step ? "border-primary bg-primary/5" : "border-border text-muted"}`}
          >
            {step.ordinal + 1}. {step.name}
            {step.ordinal < item.current_step
              ? " · completed"
              : step.ordinal === item.current_step
                ? item.state === "completed"
                  ? " · completed"
                  : " · current step"
                : ""}
          </li>
        ))}
      </ol>
      <p className="break-all text-sm">
        Responsible: <strong>{label(item.responsible_id)}</strong>
      </p>
      {item.state === "waiting" && (
        <p className="rounded-lg bg-background p-3 text-sm break-all">
          Waiting for {label(item.recipient_id)} to accept.{" "}
          {label(item.responsible_id)} remains responsible for the current step.
        </p>
      )}
      <fieldset disabled={busy} className="space-y-3">
        {canProgress && !finalStep && (
          <div className="space-y-3">
            <Field title="Next responsible employee">
              <select
                className={inputClass}
                value={recipient}
                onChange={(e) => setRecipient(e.target.value)}
              >
                <option value="">Select recipient</option>
                {members
                  .filter((row) => row.user_id !== user?.id)
                  .map((row) => (
                    <option key={row.id} value={row.user_id}>
                      {row.user_id}
                    </option>
                  ))}
              </select>
            </Field>
            <p className="text-xs text-muted">
              {item.workflow.steps[item.current_step].handoff_mode ===
              "acceptance"
                ? "The recipient must accept before responsibility changes."
                : "Responsibility transfers immediately to the next step."}
            </p>
            <button
              className={buttonClass}
              disabled={!recipient}
              onClick={() => onCommand("handoff", { recipient_id: recipient })}
            >
              Hand over to next step
            </button>
          </div>
        )}
        {item.state === "waiting" && item.recipient_id === user?.id && (
          <button className={buttonClass} onClick={() => onCommand("accept")}>
            Accept handoff
          </button>
        )}
        {canProgress && finalStep && (
          <button className={buttonClass} onClick={() => onCommand("complete")}>
            Complete work
          </button>
        )}
        {tenant?.owner_id === user?.id && item.state !== "completed" && (
          <div className="space-y-3 border-t border-border pt-4">
            <Field title="Owner reassignment">
              <select
                className={inputClass}
                value={responsible}
                onChange={(e) => setResponsible(e.target.value)}
              >
                <option value="">Unassigned queue</option>
                {item.responsible_id &&
                  !members.some(
                    (row) => row.user_id === item.responsible_id,
                  ) && (
                    <option value={item.responsible_id}>Former member</option>
                  )}
                {members.map((row) => (
                  <option key={row.id} value={row.user_id}>
                    {label(row.user_id)}
                  </option>
                ))}
              </select>
            </Field>
            <p className="text-xs text-muted">
              Reassignment keeps this step and cancels any pending handoff.
            </p>
            <button
              className={secondaryClass}
              onClick={() =>
                onCommand("assign", { responsible_id: responsible || null })
              }
            >
              Apply reassignment
            </button>
          </div>
        )}
        <button className={secondaryClass} onClick={onRefresh}>
          Reload saved work
        </button>
      </fieldset>
      <div className="space-y-3 border-t border-border pt-4">
        <h3 className="font-semibold">Recent history</h3>
        <p className="text-xs text-muted">
          Latest 100 events; earlier history is retained.
        </p>
        <ol className="space-y-3">
          {item.events?.map((event) => (
            <li
              key={event.id}
              className="rounded-lg bg-background p-3 text-xs space-y-1 break-all"
            >
              <p className="font-medium">
                v{event.version} · {event.action} · {event.state}
              </p>
              <p>
                Actor: {label(event.actor_id)} · Step {event.from_step + 1} →{" "}
                {event.current_step + 1}
              </p>
              <p>
                Responsible: {label(event.from_responsible_id)} →{" "}
                {label(event.responsible_id)}
                {event.recipient_id
                  ? ` · Nominated: ${label(event.recipient_id)}`
                  : ""}
              </p>
              <p className="text-muted">
                {new Date(event.created_at).toLocaleString()}
              </p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}

export function WorkItems({
  initialID,
  invalidID = false,
}: {
  initialID?: string;
  invalidID?: boolean;
}) {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const detailRequest = useRef(0);
  const [initialError, setInitialError] = useState(
    invalidID ? "Invalid work link." : "",
  );
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
  }, [tenant, revision]);
  useEffect(() => {
    if (!tenant) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const data = await workflows.items(tenant!.id, page, abort.signal);
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
  useEffect(() => {
    if (!tenant || !initialID) return;
    let ignore = false;
    const request = ++detailRequest.current;
    workflows
      .item(tenant.id, initialID)
      .then((row) => {
        if (!ignore && detailRequest.current === request) setSelected(row);
      })
      .catch((err: unknown) => {
        if (!ignore && detailRequest.current === request)
          setInitialError(errorText(err));
      });
    return () => {
      ignore = true;
    };
  }, [tenant, initialID, invalidID]);
  async function open(id: string) {
    if (!tenant || busy) return;
    detailRequest.current++;
    setInitialError("");
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      setSelected(await workflows.item(tenant.id, id));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function create(
    input: {
      order_id: string;
      workflow_version_id: string;
      responsible_id: string;
    },
    key: string,
  ) {
    if (!tenant || busy) return;
    detailRequest.current++;
    setInitialError("");
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      setSelected(await workflows.create(tenant.id, input, key));
      setCreatorKey((v) => v + 1);
      setRevision((v) => v + 1);
      setSuccess(
        "Work created. The selected employee is responsible for the first step.",
      );
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function command(
    action: "handoff" | "accept" | "assign" | "complete",
    fields?: { recipient_id?: string; responsible_id?: string | null },
  ) {
    if (!tenant || !selected || busy) return;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      setSelected(await workflows.command(tenant.id, selected, action, fields));
      setRevision((v) => v + 1);
      setSuccess("Saved. Responsibility and history have been updated.");
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap justify-between gap-3">
        <div>
          <h1 className="text-3xl font-semibold">Work</h1>
          <p className="mt-2 text-sm text-muted">
            Follow each step and see who is responsible.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link
            className={secondaryClass}
            href={`/tenants/${tenant?.id}/workflows`}
          >
            Manage workflows
          </Link>
          <button
            className={secondaryClass}
            disabled={busy || loading}
            onClick={() => {
              setError("");
              setRevision((v) => v + 1);
              if (selected) void open(selected.id);
            }}
          >
            Refresh
          </button>
        </div>
      </div>
      <Notice error={initialError || error || lookupError} success={success} />
      <div className="grid items-start gap-5 xl:grid-cols-2">
        <div className="space-y-5">
          {lookupLoading ? (
            <p role="status">Loading team and workflows…</p>
          ) : (
            !lookupError &&
            tenant?.owner_id === user?.id && (
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
          <section className="rounded-xl border border-border bg-surface p-5 space-y-4">
            <h2 className="font-semibold">All work</h2>
            {loading ? (
              <p role="status">Loading work…</p>
            ) : !rows.length ? (
              <p className="text-sm text-muted">
                No work yet. The owner can create work from a confirmed order.
              </p>
            ) : (
              rows.map((row) => (
                <button
                  key={row.id}
                  disabled={busy}
                  className={`block w-full rounded-lg border border-border p-3 text-left space-y-1 hover:bg-background ${selected?.id === row.id ? "bg-primary/5" : ""}`}
                  onClick={() => void open(row.id)}
                >
                  <span className="block font-medium">
                    {row.workflow.name} · {row.state}
                  </span>
                  <span className="block text-sm text-muted">
                    Step {row.current_step + 1}:{" "}
                    {row.workflow.steps[row.current_step]?.name}
                  </span>
                  <span className="block text-xs break-all">
                    Responsible:{" "}
                    {row.responsible_id === user?.id
                      ? "You"
                      : row.responsible_id || "Unassigned queue"}
                  </span>
                  {row.recipient_id && (
                    <span className="block text-xs break-all">
                      Waiting for:{" "}
                      {row.recipient_id === user?.id ? "You" : row.recipient_id}
                    </span>
                  )}
                </button>
              ))
            )}
            <Pager
              page={page}
              total={total}
              onChange={setPage}
              disabled={busy || loading}
            />
          </section>
        </div>
        {selected ? (
          <WorkDetail
            key={`${selected.id}:${selected.version}`}
            item={selected}
            members={members}
            busy={busy}
            onCommand={(action, fields) => void command(action, fields)}
            onRefresh={() => void open(selected.id)}
          />
        ) : (
          <p className="rounded-xl border border-dashed border-border p-6 text-sm text-muted">
            Select work to see its steps, responsibility and history.
          </p>
        )}
      </div>
    </div>
  );
}
