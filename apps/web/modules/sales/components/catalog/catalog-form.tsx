"use client";

import { Textarea } from "@hexta/ui";
import { Field, buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";
import type { useCatalog } from "../../hooks/use-catalog";
import type { CatalogKind } from "../../types";
import type { CatalogForm as CatalogFormValues } from "../../utils";
import { CatalogTextField } from "./catalog-text-field";

/** Add or edit one customer or product. */
export function CatalogForm({ kind, catalog }: { kind: CatalogKind; catalog: ReturnType<typeof useCatalog> }) {
  const t = useT();
  const k = (key: string) => t(`catalog.${kind}.${key}` as MessageKey);
  const { form, setForm, editing, busy } = catalog;
  const field = (name: keyof CatalogFormValues, label: string, required = false, type = "text", maxLength?: number) => (
    <CatalogTextField
      label={label}
      value={form[name]}
      onChange={(value) => setForm({ ...form, [name]: value })}
      required={required}
      type={type}
      maxLength={maxLength}
    />
  );
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        void catalog.save();
      }}
      className="space-y-4 rounded-xl border border-border bg-surface p-5"
    >
      <div className="flex items-center justify-between">
        <h2 className="font-semibold">{editing ? k("editTitle") : k("add")}</h2>
        {editing && (
          <button type="button" className="text-sm text-primary" disabled={busy} onClick={catalog.cancelEdit}>
            {t("catalog.cancel")}
          </button>
        )}
      </div>
      <fieldset disabled={busy} className="space-y-4">
        {kind === "customers" ? (
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
              <Textarea
                rows={3}
                maxLength={10000}
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
              />
            </Field>
          </>
        )}
        <button type="submit" className={`${buttonClass} w-full`}>
          {busy ? t("catalog.saving") : k("save")}
        </button>
      </fieldset>
    </form>
  );
}
