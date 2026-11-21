"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import type { TenantMember } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { useWorkspace } from "./workspace";
import {
  buttonClass,
  secondaryClass,
  inputClass,
  ChecklistItem,
  ConflictBanner,
  EmptyState,
  Field,
  Notice,
  Pager,
  SegmentedFilter,
  StatusChip,
  type ChipTone,
} from "./primitives";
import {
  sales,
  type Order,
  type Customer,
  type Product,
  type DraftInput,
  type OrderListFilter,
} from "@/lib/services/sales";
import { isConflict, useErrorText } from "@/lib/errors";
import { useLocale, useT } from "@/lib/i18n/provider";
import type { Translate } from "@/lib/i18n/translate";

interface Line {
  key: string;
  variant_id: string;
  quantity: string;
  unit_price: string;
}
const knownStatuses = ["draft", "confirmed", "processing", "fulfilled", "cancelled", "returned"] as const;
type KnownStatus = (typeof knownStatuses)[number];

function statusChip(t: Translate, status: string): { label: string; tone: ChipTone } {
  const known = (knownStatuses as readonly string[]).includes(status)
    ? (status as KnownStatus)
    : null;
  return {
    label: known ? t(`orders.status.${known}`) : t("orders.status.other"),
    tone: status === "confirmed" ? "solid" : "neutral",
  };
}

function useMoney() {
  const locale = useLocale();
  return (value: string | number) => {
    const amount = Number(value);
    return Number.isFinite(amount)
      ? `${new Intl.NumberFormat(locale === "vi" ? "vi-VN" : "en-US", { maximumFractionDigits: 2 }).format(amount)} đ`
      : String(value);
  };
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
  const { tenant, isOwner: owner } = useWorkspace();
  const { user } = useAuthStore();
  const t = useT();
  const errorText = useErrorText();
  const money = useMoney();
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
  const [conflict, setConflict] = useState(false);
  const readonly = order
    ? order.status !== "draft" ||
      (!owner && !!order.employee_id && order.employee_id !== user?.id)
    : false;
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
  function fail(err: unknown) {
    setConflict(isConflict(err));
    setError(errorText(err));
  }
  async function persist(employeeID: string) {
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    setConflict(false);
    try {
      const input: DraftInput = {
        version: order?.version || 0,
        customer_id: customer || null,
        employee_id: employeeID || null,
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
      fail(err);
    } finally {
      setBusy(false);
    }
  }
  function save(event: React.FormEvent) {
    event.preventDefault();
    void persist(employee);
  }
  async function confirm() {
    if (!tenant || !order || dirty || busy) return;
    setBusy(true);
    setError("");
    setConflict(false);
    try {
      onSaved(await sales.confirm(tenant.id, order.id, order.version));
    } catch (err) {
      fail(err);
    } finally {
      setBusy(false);
    }
  }
  async function reload(askFirst: boolean) {
    if (!tenant || !order || (askFirst && dirty && !window.confirm(t("orders.discardReload"))))
      return;
    setBusy(true);
    setError("");
    setConflict(false);
    try {
      onSaved(await sales.order(tenant.id, order.id));
    } catch (err) {
      fail(err);
    } finally {
      setBusy(false);
    }
  }
  // Guidance mirrored from the server's confirmation gate; the server re-checks on confirm.
  const chosen = customers.find((row) => row.id === customer);
  const linePrices = lines.filter((line) => line.unit_price.trim() === "" || Number(line.unit_price) < 0);
  const checks = [
    { key: "saved", passed: !!order && !dirty, label: t("orders.checklist.saved") },
    {
      key: "assignee",
      passed: !!user?.id && employee === user.id,
      label: t("orders.checklist.assignee"),
      hint: employee && employee !== user?.id ? t("orders.checklist.assigneeOther") : undefined,
    },
    { key: "name", passed: !!(chosen?.full_name ?? "").trim(), label: t("orders.checklist.customerName") },
    {
      key: "contact",
      passed: !!((chosen?.phone ?? "").trim() || (chosen?.email ?? "").trim()),
      label: t("orders.checklist.customerContact"),
    },
    {
      key: "items",
      passed:
        lines.length > 0 &&
        lines.every((line) => line.variant_id && Number.isInteger(Number(line.quantity)) && Number(line.quantity) > 0),
      label: t("orders.checklist.items"),
    },
    {
      key: "prices",
      passed: lines.length > 0 && linePrices.length === 0,
      label: t("orders.checklist.prices"),
      hint: linePrices.length ? t("orders.checklist.pricesMissing", { count: linePrices.length }) : undefined,
    },
    {
      key: "address",
      passed: !delivery || address.trim() !== "",
      label: t("orders.checklist.address"),
    },
  ];
  const failing = checks.filter((check) => !check.passed).length;
  const claimable = !!order && order.status === "draft" && !order.employee_id && !!user?.id && !readonly;
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">
            {order ? order.order_number : t("orders.editor.newTitle")}
          </h2>
          <p className="mt-1 text-xs text-muted">
            {order ? t("orders.editor.version", { version: order.version }) : t("orders.editor.newHint")}
          </p>
        </div>
        {order && <StatusChip tone={statusChip(t, order.status).tone}>{statusChip(t, order.status).label}</StatusChip>}
      </div>
      {conflict ? (
        <ConflictBanner
          message={error}
          actionLabel={t("orders.editor.conflictReload")}
          onReload={() => void reload(false)}
        />
      ) : (
        <Notice error={error} />
      )}
      {readonly && (
        <p className="rounded-lg bg-background p-3 text-sm text-muted">
          {order?.status !== "draft" ? t("orders.editor.readonlyConfirmed") : t("orders.editor.readonlyOther")}
        </p>
      )}
      {order?.status === "confirmed" && (
        <Link href={`/tenants/${tenant?.id}/work`} className="inline-block text-sm text-primary underline">
          {t("orders.editor.linkedWork")}
        </Link>
      )}
      <div className="flex flex-wrap items-start gap-5">
        <form onSubmit={save} className="min-w-0 flex-[999_1_420px] space-y-5">
          <fieldset disabled={busy || readonly} className="space-y-5">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field title={t("orders.editor.customer")}>
                <select
                  className={inputClass}
                  value={customer}
                  onChange={(e) => {
                    setCustomer(e.target.value);
                    setDirty(true);
                  }}
                >
                  <option value="">{t("orders.editor.selectLater")}</option>
                  {customer && !customers.some((row) => row.id === customer) && (
                    <option value={customer}>{t("orders.unavailableCustomer")}</option>
                  )}
                  {customers.map((row) => (
                    <option key={row.id} value={row.id}>
                      {row.full_name}
                      {row.phone ? ` · ${row.phone}` : ""}
                    </option>
                  ))}
                </select>
              </Field>
              <Field title={t("orders.editor.assignee")}>
                <select
                  className={inputClass}
                  value={employee}
                  onChange={(e) => {
                    setEmployee(e.target.value);
                    setDirty(true);
                  }}
                >
                  <option value="" disabled={!owner && !!order?.employee_id}>
                    {t("orders.editor.unassigned")}
                  </option>
                  {employee && !members.some((row) => row.user_id === employee) && (
                    <option value={employee}>{t("orders.editor.formerMember")}</option>
                  )}
                  {members.map((row) => (
                    <option key={row.id} value={row.user_id} disabled={!owner && row.user_id !== user?.id}>
                      {row.user_id === user?.id ? t("orders.editor.you") : row.user_id}
                    </option>
                  ))}
                </select>
              </Field>
            </div>
            <div className="flex flex-wrap gap-4 text-xs">
              <Link href={`/tenants/${tenant?.id}/customers`} className="text-primary underline">
                {t("orders.editor.manageCustomers")}
              </Link>
              <Link href={`/tenants/${tenant?.id}/products`} className="text-primary underline">
                {t("orders.editor.manageProducts")}
              </Link>
              <Link href={`/tenants/${tenant?.id}/settings/members`} className="text-primary underline">
                {t("orders.editor.manageTeam")}
              </Link>
            </div>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold">{t("orders.editor.items")}</h3>
                <button
                  type="button"
                  className={secondaryClass}
                  disabled={!variants.length || lines.length >= 100}
                  onClick={() => {
                    setDirty(true);
                    setLines([...lines, { key: crypto.randomUUID(), variant_id: "", quantity: "1", unit_price: "" }]);
                  }}
                >
                  {t("orders.editor.addItem")}
                </button>
              </div>
              {!lines.length && (
                <p className="rounded-lg border border-dashed border-border p-4 text-sm text-muted">
                  {t("orders.editor.noItems")} {!variants.length && t("orders.editor.noProducts")}
                </p>
              )}
              {lines.map((line, index) => {
                const missingPrice = line.unit_price.trim() === "";
                return (
                  <div
                    key={line.key}
                    className={`space-y-3 rounded-lg border bg-background p-3 ${missingPrice ? "border-error/60" : "border-border"}`}
                  >
                    <Field title={t("orders.editor.product", { index: index + 1 })}>
                      <select
                        className={inputClass}
                        required
                        value={line.variant_id}
                        onChange={(e) => {
                          const variant = variants.find((row) => row.id === e.target.value);
                          updateLine(index, {
                            variant_id: e.target.value,
                            unit_price: variant?.price_provided ? String(variant.price) : "",
                          });
                        }}
                      >
                        <option value="">{t("orders.editor.selectProduct")}</option>
                        {line.variant_id && !variants.some((row) => row.id === line.variant_id) && (
                          <option value={line.variant_id}>
                            {t("orders.editor.unavailableProduct", { id: line.variant_id })}
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
                      <Field title={t("orders.editor.quantity")}>
                        <input
                          className={inputClass}
                          type="number"
                          min="1"
                          max="1000000"
                          step="1"
                          required
                          value={line.quantity}
                          onChange={(e) => updateLine(index, { quantity: e.target.value })}
                        />
                      </Field>
                      <Field title={t("orders.editor.unitPrice")}>
                        <input
                          className={inputClass}
                          type="number"
                          min="0"
                          max="9999999999999.99"
                          step="0.01"
                          placeholder={t("orders.editor.priceNotSet")}
                          value={line.unit_price}
                          onChange={(e) => updateLine(index, { unit_price: e.target.value })}
                        />
                      </Field>
                      <button
                        type="button"
                        aria-label={t("orders.editor.removeItem", { index: index + 1 })}
                        className={secondaryClass}
                        onClick={() => {
                          setDirty(true);
                          setLines(lines.filter((_, i) => i !== index));
                        }}
                      >
                        {t("orders.editor.remove")}
                      </button>
                    </div>
                  </div>
                );
              })}
              <p className="text-xs text-muted">{t("orders.editor.priceHint")}</p>
            </div>
            <label className="flex min-h-11 items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={delivery}
                onChange={(e) => {
                  setDelivery(e.target.checked);
                  setDirty(true);
                }}
              />
              {t("orders.editor.delivery")}
            </label>
            {delivery && (
              <Field title={t("orders.editor.address")}>
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
            <Field title={t("orders.editor.notes")}>
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
              <div className="flex flex-wrap gap-2">
                <button type="submit" className={buttonClass}>
                  {busy ? t("orders.editor.saving") : t("orders.editor.saveDraft")}
                </button>
                {claimable && (
                  <button type="button" className={secondaryClass} onClick={() => user?.id && void persist(user.id)}>
                    {t("orders.editor.claim")}
                  </button>
                )}
              </div>
            )}
          </fieldset>
        </form>
        {(!order || order.status === "draft") && (
          <aside className="flex-[1_1_260px] space-y-3 rounded-lg border border-border bg-background p-4">
            <h3 className="font-semibold">{t("orders.checklist.title")}</h3>
            <p id="confirm-why" className="text-sm text-muted" aria-live="polite">
              {failing ? t("orders.checklist.missing", { count: failing }) : t("orders.checklist.ready")}
            </p>
            <ul>
              {checks.map((check) => (
                <ChecklistItem
                  key={check.key}
                  passed={check.passed}
                  label={check.label}
                  hint={check.hint}
                  passedText={t("orders.checklist.passed")}
                  failedText={t("orders.checklist.failed")}
                />
              ))}
            </ul>
            {order && (
              <button
                type="button"
                className={`${buttonClass} w-full`}
                aria-describedby="confirm-why"
                disabled={busy || readonly || failing > 0}
                onClick={() => void confirm()}
              >
                {t("orders.editor.confirm")}
              </button>
            )}
            <p className="text-xs text-muted">{t("orders.checklist.serverNote")}</p>
          </aside>
        )}
      </div>
      {order && (
        <div className="space-y-2 border-t border-border pt-4">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <span className="text-sm">{t("orders.editor.savedValue")}</span>
            <strong>{money(order.total_amount)}</strong>
          </div>
          <p className="text-xs text-muted">
            {order.items.some((item) => !item.price_provided)
              ? t("orders.editor.valueIncomplete")
              : t("orders.editor.valueNote")}
          </p>
          {dirty && order.status === "draft" && (
            <p role="status" className="text-sm">
              {t("orders.editor.unsaved")}
            </p>
          )}
          <button type="button" className={secondaryClass} disabled={busy} onClick={() => void reload(true)}>
            {t("orders.editor.reload")}
          </button>
        </div>
      )}
    </section>
  );
}

type ListView = "all" | "draft" | "confirmed" | "unassigned" | "mine";
const viewFilters: Record<ListView, OrderListFilter> = {
  all: {},
  draft: { status: "draft" },
  confirmed: { status: "confirmed" },
  unassigned: { assignee: "unassigned" },
  mine: { assignee: "me" },
};

export function Orders() {
  const { tenant } = useWorkspace();
  const t = useT();
  const errorText = useErrorText();
  const money = useMoney();
  const [view, setView] = useState<ListView>("all");
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
  }, [tenant, revision, errorText]);
  useEffect(() => {
    if (!tenant) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const data = await sales.orders(tenant!.id, page, viewFilters[view], abort.signal);
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
  function saved(row: Order) {
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    setRevision((value) => value + 1);
    setSuccess(row.status === "draft" ? t("orders.savedDraft") : t("orders.confirmedOrder"));
  }
  function open(row: Order | null) {
    if (dirty && !window.confirm(t("orders.discardOpen"))) return;
    setDirty(false);
    setSelected(row);
    setEditorKey((value) => value + 1);
    setSuccess("");
  }
  const filterOptions: { value: ListView; label: string }[] = (
    ["all", "draft", "confirmed", "unassigned", "mine"] as const
  ).map((value) => ({ value, label: t(`orders.filters.${value}`) }));
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("orders.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("orders.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            className={secondaryClass}
            disabled={busy || loading}
            onClick={() => setRevision((value) => value + 1)}
          >
            {t("orders.refresh")}
          </button>
          <button type="button" className={buttonClass} disabled={busy} onClick={() => open(null)}>
            {t("orders.newDraft")}
          </button>
        </div>
      </div>
      <Notice error={error || lookupError} success={success} />
      <div className="flex flex-wrap items-start gap-5">
        <section className="min-w-0 flex-[1_1_300px] space-y-4 rounded-xl border border-border bg-surface p-4 xl:max-w-[420px]">
          <h2 className="font-semibold">{t("orders.listTitle")}</h2>
          <SegmentedFilter
            label={t("orders.filterLabel")}
            options={filterOptions}
            value={view}
            onChange={(value) => {
              setView(value);
              setPage(1);
            }}
          />
          {loading ? (
            <p role="status" className="py-8 text-sm text-muted">
              {t("orders.loading")}
            </p>
          ) : !rows.length ? (
            view === "all" ? (
              <EmptyState title={t("orders.emptyTitle")} description={t("orders.emptyBody")} />
            ) : (
              <EmptyState title={t("orders.emptyFiltered")} />
            )
          ) : (
            <ul className="-mx-2 divide-y divide-border">
              {rows.map((row) => {
                const chip = statusChip(t, row.status);
                const incomplete = row.items.some((item) => !item.price_provided);
                return (
                  <li key={row.id}>
                    <button
                      type="button"
                      disabled={busy}
                      aria-current={selected?.id === row.id ? "true" : undefined}
                      onClick={() => open(row)}
                      className={`w-full rounded-lg p-3 text-left hover:bg-background disabled:opacity-50 ${selected?.id === row.id ? "bg-primary/10" : ""}`}
                    >
                      <div className="flex items-center justify-between gap-3">
                        <span className="truncate text-sm font-medium">
                          {customers.find((c) => c.id === row.customer_id)?.full_name ||
                            (row.customer_id ? t("orders.unavailableCustomer") : t("orders.noCustomer"))}
                        </span>
                        <StatusChip tone={chip.tone}>{chip.label}</StatusChip>
                      </div>
                      <p className="mt-1 truncate font-mono text-xs text-muted">{row.order_number}</p>
                      <p className="mt-2 text-sm">
                        {t("orders.itemsCount", { count: row.items.length })} · {t("orders.orderValue")}{" "}
                        {money(row.total_amount)}
                        {incomplete && <span className="text-muted"> ({t("orders.incomplete")})</span>}
                      </p>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
          <Pager page={page} total={total} onChange={setPage} disabled={loading || busy} />
        </section>
        <div className="min-w-0 flex-[999_1_520px]">
          {lookupLoading && !members.length ? (
            <p role="status">{t("orders.loadingForm")}</p>
          ) : lookupError ? (
            <div className="rounded-xl border border-border p-5">
              <p className="text-sm">{t("orders.lookupFailed")}</p>
              <button
                type="button"
                className={`${secondaryClass} mt-3`}
                onClick={() => setRevision((value) => value + 1)}
              >
                {t("orders.retry")}
              </button>
            </div>
          ) : (
            <OrderEditor
              key={editorKey}
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
    </div>
  );
}
