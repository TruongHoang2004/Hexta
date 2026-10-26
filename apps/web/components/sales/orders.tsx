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
import {
  sales,
  errorText,
  type Order,
  type Customer,
  type Product,
  type DraftInput,
} from "@/lib/services/sales";

interface Line {
  key: string;
  variant_id: string;
  quantity: string;
  unit_price: string;
}
interface EditorProps {
  order: Order | null;
  customers: Customer[];
  products: Product[];
  members: TenantMember[];
  busy: boolean;
  setBusy: (busy: boolean) => void;
  dirty: boolean;
  setDirty: (dirty: boolean) => void;
  onSaved: (order: Order) => void;
}
function OrderEditor({
  order,
  customers,
  products,
  members,
  busy,
  setBusy,
  dirty,
  setDirty,
  onSaved,
}: EditorProps) {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const [customer, setCustomer] = useState(order?.customer_id || "");
  const [employee, setEmployee] = useState(order?.employee_id || "");
  const [delivery, setDelivery] = useState(order?.requires_delivery || false);
  const [address, setAddress] = useState(order?.delivery_address || "");
  const [notes, setNotes] = useState(order?.notes || "");
  const [lines, setLines] = useState<Line[]>(
    () =>
      order?.items.map((item) => ({
        key: item.id,
        variant_id: item.variant_id,
        quantity: String(item.quantity),
        unit_price: item.price_provided ? String(item.unit_price) : "",
      })) || [],
  );
  const [error, setError] = useState("");
  const readonly = order
    ? order.status !== "draft" ||
      (tenant?.owner_id !== user?.id &&
        !!order.employee_id &&
        order.employee_id !== user?.id)
    : false;
  const owner = tenant?.owner_id === user?.id;
  const variants = products.flatMap((product) =>
    product.variants.map((variant) => ({
      ...variant,
      label: `${product.name} · ${variant.variant_name} (${variant.sku})`,
    })),
  );
  function updateLine(index: number, patch: Partial<Line>) {
    setDirty(true);
    setLines(
      lines.map((line, i) => (i === index ? { ...line, ...patch } : line)),
    );
  }
  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    try {
      const input: DraftInput = {
        version: order?.version || 0,
        customer_id: customer || null,
        employee_id: employee || null,
        requires_delivery: delivery,
        delivery_address: address,
        notes,
        items: lines.map((line) => ({
          variant_id: line.variant_id,
          quantity: Number(line.quantity),
          unit_price: line.unit_price.trim() === "" ? null : line.unit_price,
        })),
      };
      const saved = await sales.saveDraft(tenant.id, order?.id || null, input);
      onSaved(saved);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function confirm() {
    if (!tenant || !order || dirty || busy) return;
    setBusy(true);
    setError("");
    try {
      onSaved(await sales.confirm(tenant.id, order.id, order.version));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function reload() {
    if (
      !tenant ||
      !order ||
      (dirty &&
        !window.confirm(
          "Discard unsaved edits and load the latest saved order?",
        ))
    )
      return;
    setBusy(true);
    setError("");
    try {
      onSaved(await sales.order(tenant.id, order.id));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">
            {order
              ? `${order.status === "draft" ? "Order draft" : "Order details"}`
              : "New order draft"}
          </h2>
          <p className="mt-1 max-w-sm break-all text-xs text-muted">
            {order
              ? `${order.order_number} · Version ${order.version}`
              : "Save now, complete missing details later."}
          </p>
        </div>
        {order && (
          <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-medium uppercase text-primary">
            {order.status}
          </span>
        )}
      </div>
      <Notice error={error} />
      {readonly && (
        <p className="rounded-lg bg-background p-3 text-sm text-muted">
          {order?.status !== "draft"
            ? "This order is confirmed and is read-only in this demo."
            : "Only the assigned sale or workspace owner can edit this draft."}
        </p>
      )}
      {order?.status === "confirmed" && (
        <Link
          href={`/tenants/${tenant?.id}/work`}
          className="inline-block text-sm text-primary underline"
        >
          Create or view linked work
        </Link>
      )}
      <form onSubmit={save} className="space-y-5">
        <fieldset disabled={busy || readonly} className="space-y-5">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field title="Customer">
              <select
                className={inputClass}
                value={customer}
                onChange={(e) => {
                  setCustomer(e.target.value);
                  setDirty(true);
                }}
              >
                <option value="">Select later</option>
                {customer && !customers.some((row) => row.id === customer) && (
                  <option value={customer}>
                    Unavailable customer · choose another
                  </option>
                )}
                {customers.map((row) => (
                  <option key={row.id} value={row.id}>
                    {row.full_name}
                    {row.phone ? ` · ${row.phone}` : ""}
                  </option>
                ))}
              </select>
            </Field>
            <Field title="Assigned sale">
              <select
                className={inputClass}
                value={employee}
                onChange={(e) => {
                  setEmployee(e.target.value);
                  setDirty(true);
                }}
              >
                <option value="" disabled={!owner && !!order?.employee_id}>
                  Unassigned
                </option>
                {employee &&
                  !members.some((row) => row.user_id === employee) && (
                    <option value={employee}>Former member</option>
                  )}
                {members.map((row) => (
                  <option
                    key={row.id}
                    value={row.user_id}
                    disabled={!owner && row.user_id !== user?.id}
                  >
                    {row.user_id === user?.id ? "You" : row.user_id}
                  </option>
                ))}
              </select>
            </Field>
          </div>
          <div className="flex flex-wrap gap-4 text-xs">
            <Link
              href={`/tenants/${tenant?.id}/customers`}
              className="text-primary underline"
            >
              Manage customers
            </Link>
            <Link
              href={`/tenants/${tenant?.id}/products`}
              className="text-primary underline"
            >
              Manage products
            </Link>
            <Link
              href={`/tenants/${tenant?.id}/settings/members`}
              className="text-primary underline"
            >
              Manage team
            </Link>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold">Order items</h3>
              <button
                type="button"
                className={secondaryClass}
                disabled={!variants.length || lines.length >= 100}
                onClick={() => {
                  setDirty(true);
                  setLines([
                    ...lines,
                    {
                      key: crypto.randomUUID(),
                      variant_id: "",
                      quantity: "1",
                      unit_price: "",
                    },
                  ]);
                }}
              >
                Add item
              </button>
            </div>
            {!lines.length && (
              <p className="rounded-lg border border-dashed border-border p-4 text-sm text-muted">
                No items yet. You can save an empty draft.
                {!variants.length &&
                  " Add a product before adding order items."}
              </p>
            )}
            {lines.map((line, index) => (
              <div
                key={line.key}
                className="space-y-3 rounded-lg border border-border bg-background p-3"
              >
                <Field title={`Product ${index + 1}`}>
                  <select
                    className={inputClass}
                    required
                    value={line.variant_id}
                    onChange={(e) => {
                      const variant = variants.find(
                        (row) => row.id === e.target.value,
                      );
                      updateLine(index, {
                        variant_id: e.target.value,
                        unit_price: variant?.price_provided
                          ? String(variant.price)
                          : "",
                      });
                    }}
                  >
                    <option value="">Select product</option>
                    {line.variant_id &&
                      !variants.some((row) => row.id === line.variant_id) && (
                        <option value={line.variant_id}>
                          Unavailable product · {line.variant_id}
                        </option>
                      )}
                    {variants.map((row) => (
                      <option key={row.id} value={row.id}>
                        {row.label}
                      </option>
                    ))}
                  </select>
                </Field>
                <div className="grid grid-cols-[1fr_1fr_auto] items-end gap-3">
                  <Field title="Quantity">
                    <input
                      className={inputClass}
                      type="number"
                      min="1"
                      max="1000000"
                      step="1"
                      required
                      value={line.quantity}
                      onChange={(e) =>
                        updateLine(index, { quantity: e.target.value })
                      }
                    />
                  </Field>
                  <Field title="Unit price">
                    <input
                      className={inputClass}
                      type="number"
                      min="0"
                      max="9999999999999.99"
                      step="0.01"
                      placeholder="Not entered"
                      value={line.unit_price}
                      onChange={(e) =>
                        updateLine(index, { unit_price: e.target.value })
                      }
                    />
                  </Field>
                  <button
                    type="button"
                    aria-label={`Remove item ${index + 1}`}
                    className={secondaryClass}
                    onClick={() => {
                      setDirty(true);
                      setLines(lines.filter((_, i) => i !== index));
                    }}
                  >
                    Remove
                  </button>
                </div>
              </div>
            ))}
            <p className="text-xs text-muted">
              Blank price means unknown. Enter 0 explicitly for a free item.
            </p>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={delivery}
              onChange={(e) => {
                setDelivery(e.target.checked);
                setDirty(true);
              }}
            />
            This order requires delivery
          </label>
          {delivery && (
            <Field title="Delivery address">
              <textarea
                className={inputClass}
                maxLength={5000}
                rows={2}
                value={address}
                onChange={(e) => {
                  setAddress(e.target.value);
                  setDirty(true);
                }}
              />
            </Field>
          )}
          <Field title="Notes">
            <textarea
              className={inputClass}
              rows={3}
              maxLength={10000}
              value={notes}
              onChange={(e) => {
                setNotes(e.target.value);
                setDirty(true);
              }}
            />
          </Field>
          {!readonly && (
            <button type="submit" className={buttonClass}>
              {busy ? "Saving…" : "Save draft"}
            </button>
          )}
        </fieldset>
      </form>
      {order && (
        <div className="space-y-3 border-t border-border pt-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="text-sm">Saved order value</span>
            <strong>{order.total_amount}</strong>
          </div>
          <p className="text-xs text-muted">
            {order.items.some((item) => !item.price_provided)
              ? "Some prices are missing; the saved value is incomplete."
              : "Order value only; no payment is collected."}
          </p>
          {order.status === "draft" && (
            <>
              <p className="text-xs leading-5 text-muted">
                Confirmation requires a customer name and contact, a sale, at
                least one item with an explicit price, and an address if
                delivery is selected.
              </p>
              {dirty && (
                <p role="status" className="text-sm text-amber-700">
                  Unsaved changes. Save the draft before confirming.
                </p>
              )}
              <button
                type="button"
                className={buttonClass}
                disabled={busy || dirty || employee !== user?.id || readonly}
                onClick={() => void confirm()}
              >
                Confirm order
              </button>
              {employee !== user?.id && (
                <p className="text-xs text-muted">
                  Only the assigned sale can confirm.
                </p>
              )}
            </>
          )}
          <button
            type="button"
            className={`${secondaryClass} ml-2`}
            disabled={busy}
            onClick={() => void reload()}
          >
            Reload saved order
          </button>
        </div>
      )}
    </section>
  );
}

export function Orders({
  initialID,
  invalidID = false,
}: {
  initialID?: string;
  invalidID?: boolean;
}) {
  const { tenant } = useWorkspace();
  const detailRequest = useRef(0);
  const [initialLoading, setInitialLoading] = useState(!!initialID);
  const [initialError, setInitialError] = useState(
    invalidID ? "Invalid order link." : "",
  );
  const [initialRetry, setInitialRetry] = useState(0);
  const [rows, setRows] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [selected, setSelected] = useState<Order | null>(null);
  const [editorKey, setEditorKey] = useState(0);
  const [dirty, setDirty] = useState(false);
  const [revision, setRevision] = useState(0);
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
        const [c, p, m] = await Promise.all([
          sales.all<Customer>("customers", tenant!.id, abort.signal),
          sales.all<Product>("products", tenant!.id, abort.signal),
          sdk.identity.getUsers(tenant!.id),
        ]);
        if (!abort.signal.aborted) {
          setCustomers(c);
          setProducts(p);
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
      setError("");
      try {
        const data = await sales.list<Order>(
          "orders",
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
  useEffect(() => {
    if (!tenant || !initialID) return;
    let ignore = false;
    const request = ++detailRequest.current;
    sales
      .order(tenant.id, initialID)
      .then((row) => {
        if (!ignore && detailRequest.current === request) setSelected(row);
      })
      .catch((err: unknown) => {
        if (!ignore && detailRequest.current === request)
          setInitialError(errorText(err));
      })
      .finally(() => {
        if (!ignore) setInitialLoading(false);
      });
    return () => {
      ignore = true;
    };
  }, [tenant, initialID, initialRetry]);
  function saved(row: Order) {
    detailRequest.current++;
    setInitialLoading(false);
    setInitialError("");
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    setRevision((value) => value + 1);
    setSuccess(row.status === "draft" ? "Draft saved." : "Order confirmed.");
  }
  function open(row: Order | null) {
    if (dirty && !window.confirm("Discard unsaved edits and open this order?"))
      return;
    detailRequest.current++;
    setInitialLoading(false);
    setInitialError("");
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    setSuccess("");
  }
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <p className="text-xs font-semibold uppercase tracking-widest text-primary">
            {tenant?.name}
          </p>
          <h1 className="mt-1 text-3xl font-semibold">Orders</h1>
          <p className="mt-2 text-sm text-muted">
            Start with a draft. Confirm when the details are ready.
          </p>
        </div>
        <div className="flex gap-2">
          <button
            className={secondaryClass}
            disabled={busy || loading}
            onClick={() => setRevision((value) => value + 1)}
          >
            Refresh
          </button>
          <button
            className={buttonClass}
            disabled={busy}
            onClick={() => open(null)}
          >
            New order
          </button>
        </div>
      </div>
      <Notice error={error || lookupError} success={success} />
      <div className="grid items-start gap-5 xl:grid-cols-[minmax(240px,0.8fr)_minmax(0,1.2fr)]">
        <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
          <h2 className="font-semibold">All orders</h2>
          {loading ? (
            <p role="status" className="py-8 text-sm text-muted">
              Loading orders…
            </p>
          ) : !rows.length ? (
            <p className="rounded-lg border border-dashed border-border p-6 text-sm text-muted">
              No orders yet. Save your first draft to get started.
            </p>
          ) : (
            <ul className="divide-y divide-border">
              {rows.map((row) => (
                <li key={row.id}>
                  <button
                    disabled={busy}
                    onClick={() => open(row)}
                    className={`w-full rounded-lg p-3 text-left hover:bg-background disabled:opacity-50 ${selected?.id === row.id ? "bg-primary/5" : ""}`}
                  >
                    <div className="flex items-center justify-between gap-3">
                      <span className="truncate text-sm font-medium">
                        {customers.find((c) => c.id === row.customer_id)
                          ?.full_name ||
                          (row.customer_id
                            ? "Unavailable customer"
                            : "Customer not selected")}
                      </span>
                      <span className="text-xs uppercase text-primary">
                        {row.status}
                      </span>
                    </div>
                    <p className="mt-1 truncate font-mono text-xs text-muted">
                      {row.order_number}
                    </p>
                    <p className="mt-2 text-sm">
                      {row.items.length} items · {row.total_amount}
                      {row.items.some((item) => !item.price_provided)
                        ? " (incomplete)"
                        : ""}
                    </p>
                  </button>
                </li>
              ))}
            </ul>
          )}
          <Pager
            page={page}
            total={total}
            onChange={setPage}
            disabled={loading || busy}
          />
        </section>
        {initialError ? (
          <section
            role="alert"
            className="space-y-3 rounded-xl border border-border p-5"
          >
            <h2 className="font-semibold">Order could not be opened</h2>
            <p className="text-sm">{initialError}</p>
            <p className="text-sm text-muted">
              Select another order or explicitly start a new draft.
            </p>
            <div className="flex flex-wrap gap-2">
              {initialID && (
                <button
                  className={secondaryClass}
                  onClick={() => {
                    setInitialError("");
                    setInitialLoading(true);
                    setInitialRetry((value) => value + 1);
                  }}
                >
                  Retry opening order
                </button>
              )}
              <button className={secondaryClass} onClick={() => open(null)}>
                New order
              </button>
            </div>
          </section>
        ) : initialLoading || (lookupLoading && !members.length) ? (
          <p role="status">Loading order form…</p>
        ) : lookupError ? (
          <div className="rounded-xl border border-border p-5">
            <p className="text-sm">
              Unable to load customers, products or members.
            </p>
            <button
              className={`${secondaryClass} mt-3`}
              onClick={() => setRevision((value) => value + 1)}
            >
              Retry
            </button>
          </div>
        ) : (
          <OrderEditor
            key={`${selected?.id || "new"}:${editorKey}`}
            order={selected}
            customers={customers}
            products={products}
            members={members}
            busy={busy}
            setBusy={setBusy}
            dirty={dirty}
            setDirty={setDirty}
            onSaved={saved}
          />
        )}
      </div>
    </div>
  );
}
