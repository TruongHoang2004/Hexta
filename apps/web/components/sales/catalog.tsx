"use client";

import { useEffect, useState } from "react";
import { useWorkspace } from "./workspace";
import {
  buttonClass,
  secondaryClass,
  inputClass,
  EmptyState,
  Field,
  Notice,
  Pager,
  StatusChip,
} from "./primitives";
import {
  sales,
  type Customer,
  type Product,
} from "@/lib/services/sales";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";

const blank = {
  full_name: "",
  phone: "",
  email: "",
  address: "",
  name: "",
  description: "",
  category: "",
  sku: "",
  variant_name: "",
  price: "",
};
export function Catalog({ kind }: { kind: "customers" | "products" }) {
  const { tenant } = useWorkspace();
  const [rows, setRows] = useState<(Customer | Product)[]>([]);
  const [form, setForm] = useState(blank);
  const [editing, setEditing] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const customer = kind === "customers";
  const t = useT();
  const errorText = useErrorText();
  const k = (key: string) => t(`catalog.${kind}.${key}` as MessageKey);
  useEffect(() => {
    if (!tenant) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const data = await sales.list<Customer | Product>(
          kind,
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
  }, [tenant, kind, page, revision, errorText]);
  function edit(row: Customer | Product) {
    setEditing(row.id);
    setError("");
    setSuccess("");
    if ("full_name" in row)
      setForm({
        ...blank,
        full_name: row.full_name,
        phone: row.phone || "",
        email: row.email || "",
        address: row.address || "",
      });
    else
      setForm({
        ...blank,
        name: row.name,
        description: row.description || "",
        category: row.category || "",
        sku: row.variants[0]?.sku || "",
        variant_name: row.variants[0]?.variant_name || "",
        price: row.variants[0]?.price_provided
          ? String(row.variants[0].price)
          : "",
      });
  }
  async function save(event: React.FormEvent) {
    event.preventDefault();
    if (!tenant || busy) return;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      if (customer)
        await sales.saveCustomer(tenant.id, editing, {
          full_name: form.full_name,
          phone: form.phone,
          email: form.email,
          address: form.address,
        });
      else
        await sales.saveProduct(tenant.id, editing, {
          name: form.name,
          description: form.description,
          category: form.category,
          sku: form.sku,
          variant_name: form.variant_name.trim() || t("catalog.products.defaultVariant"),
          price: form.price.trim() === "" ? null : form.price,
        });
      setForm(blank);
      setEditing(null);
      setPage(1);
      setRevision((value) => value + 1);
      setSuccess(k("saved"));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function archive(row: Customer | Product) {
    if (
      !tenant ||
      busy ||
      !window.confirm(t("catalog.archiveConfirm"))
    )
      return;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await sales.archive(kind, tenant.id, row.id);
      if (editing === row.id) {
        setEditing(null);
        setForm(blank);
      }
      setPage(1);
      setRevision((value) => value + 1);
      setSuccess(t("catalog.archived"));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  const field = (
    name: keyof typeof blank,
    label: string,
    required = false,
    type = "text",
    maxLength?: number,
  ) => (
    <Field title={label}>
      <input
        className={inputClass}
        value={form[name]}
        onChange={(e) => setForm({ ...form, [name]: e.target.value })}
        required={required}
        type={type}
        maxLength={maxLength}
        {...(type === "number"
          ? { min: "0", max: "9999999999999.99", step: "0.01" }
          : {})}
      />
    </Field>
  );
  return (
    <div className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold">{k("title")}</h1>
        <p className="mt-1 text-sm text-muted">{k("subtitle")}</p>
      </div>
      <Notice error={error} success={success} />
      <div className="grid items-start gap-5 xl:grid-cols-[1fr_320px]">
        <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
          <div className="flex items-center justify-between">
            <h2 className="font-semibold">{k("list")}</h2>
            <button
              type="button"
              className={secondaryClass}
              onClick={() => {
                setError("");
                setRevision((value) => value + 1);
              }}
              disabled={busy || loading}
            >
              {t("catalog.refresh")}
            </button>
          </div>
          {loading ? (
            <p role="status" className="py-10 text-sm text-muted">
              {t("catalog.loading")}
            </p>
          ) : !rows.length ? (
            <EmptyState title={k("emptyTitle")} description={k("emptyBody")} />
          ) : (
            <ul className="divide-y divide-border">
              {rows.map((row) => (
                <li
                  key={row.id}
                  className={`flex flex-wrap items-start justify-between gap-3 py-4 ${editing === row.id ? "rounded-lg bg-primary/5 px-2" : ""}`}
                >
                  <div className="min-w-0">
                    <p className="font-medium">
                      {"full_name" in row ? row.full_name : row.name}
                    </p>
                    {"full_name" in row ? (
                      <p className="mt-1 break-all text-sm text-muted">
                        {[row.phone, row.email].filter(Boolean).join(" · ") ||
                          t("catalog.customers.noContact")}
                      </p>
                    ) : (
                      <>
                        <p className="mt-1 text-sm text-muted">
                          {row.variants[0]?.sku} ·{" "}
                          {row.variants[0]?.variant_name}
                        </p>
                        <p className="mt-1 text-sm">
                          {!row.variants[0]?.price_provided ? (
                            <StatusChip tone="danger">{t("catalog.products.priceNotSet")}</StatusChip>
                          ) : Number(row.variants[0].price) === 0 ? (
                            t("catalog.products.free")
                          ) : (
                            t("catalog.products.perUnit", { price: String(row.variants[0].price) })
                          )}
                        </p>
                      </>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={secondaryClass}
                      disabled={busy}
                      onClick={() => edit(row)}
                    >
                      {t("catalog.edit")}
                    </button>
                    <button
                      type="button"
                      className={secondaryClass}
                      disabled={busy}
                      onClick={() => void archive(row)}
                    >
                      {t("catalog.archive")}
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
          <Pager
            page={page}
            total={total}
            disabled={loading || busy}
            onChange={setPage}
          />
        </section>
        <form
          onSubmit={save}
          className="space-y-4 rounded-xl border border-border bg-surface p-5"
        >
          <div className="flex items-center justify-between">
            <h2 className="font-semibold">
              {editing ? k("editTitle") : k("add")}
            </h2>
            {editing && (
              <button
                type="button"
                className="text-sm text-primary"
                disabled={busy}
                onClick={() => {
                  setEditing(null);
                  setForm(blank);
                  setError("");
                }}
              >
                {t("catalog.cancel")}
              </button>
            )}
          </div>
          <fieldset disabled={busy} className="space-y-4">
            {customer ? (
              <>
                {field("full_name", t("catalog.customers.fullName"), true, "text", 255)}
                {field("phone", t("catalog.customers.phone"), false, "tel", 32)}
                {field("email", t("catalog.customers.email"), false, "email", 255)}
                {field("address", t("catalog.customers.address"), false, "text", 5000)}
                <p className="text-xs text-muted">{t("catalog.customers.contactHint")}</p>
              </>
            ) : (
              <>
                {field("name", t("catalog.products.name"), true, "text", 255)}
                {field("sku", t("catalog.products.sku"), true, "text", 64)}
                {field("variant_name", t("catalog.products.variant"), false, "text", 128)}
                {field("price", t("catalog.products.price"), false, "number")}
                <p className="text-xs text-muted">{t("catalog.products.priceHint")}</p>
                {field("category", t("catalog.products.category"), false, "text", 100)}
                <Field title={t("catalog.products.description")}>
                  <textarea
                    className={inputClass}
                    rows={3}
                    maxLength={10000}
                    value={form.description}
                    onChange={(e) =>
                      setForm({ ...form, description: e.target.value })
                    }
                  />
                </Field>
              </>
            )}
            <button type="submit" className={`${buttonClass} w-full`}>
              {busy ? t("catalog.saving") : k("save")}
            </button>
          </fieldset>
        </form>
      </div>
    </div>
  );
}
