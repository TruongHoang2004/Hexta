"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";
import { sales } from "../services/sales";
import type { CatalogKind, CatalogRow } from "../types";
import { blankCatalogForm, catalogFormFrom, type CatalogForm } from "../utils";

/** Paged customer/product list plus the add/edit form and archive action. */
export function useCatalog(kind: CatalogKind, tenantID: string) {
  const t = useT();
  const errorText = useErrorText();
  const [rows, setRows] = useState<CatalogRow[]>([]);
  const [form, setForm] = useState<CatalogForm>(blankCatalogForm);
  const [editing, setEditing] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const data = await sales.list<CatalogRow>(kind, tenantID, page, abort.signal);
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
  }, [tenantID, kind, page, revision, errorText]);

  function changed(message: string) {
    setPage(1);
    setRevision((value) => value + 1);
    setSuccess(message);
  }

  function edit(row: CatalogRow) {
    setEditing(row.id);
    setError("");
    setSuccess("");
    setForm(catalogFormFrom(row));
  }

  function cancelEdit() {
    setEditing(null);
    setForm(blankCatalogForm);
    setError("");
  }

  async function save() {
    if (!tenantID || busy) return;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      if (kind === "customers")
        await sales.saveCustomer(tenantID, editing, {
          full_name: form.full_name,
          phone: form.phone,
          email: form.email,
          address: form.address,
        });
      else
        await sales.saveProduct(tenantID, editing, {
          name: form.name,
          description: form.description,
          category: form.category,
          sku: form.sku,
          variant_name: form.variant_name.trim() || t("catalog.products.defaultVariant"),
          price: form.price.trim() === "" ? null : form.price,
        });
      setForm(blankCatalogForm);
      setEditing(null);
      changed(t(`catalog.${kind}.saved` as MessageKey));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  async function archive(row: CatalogRow) {
    if (!tenantID || busy || !window.confirm(t("catalog.archiveConfirm"))) return;
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await sales.archive(kind, tenantID, row.id);
      if (editing === row.id) {
        setEditing(null);
        setForm(blankCatalogForm);
      }
      changed(t("catalog.archived"));
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  function refresh() {
    setError("");
    setRevision((value) => value + 1);
  }

  return {
    rows,
    total,
    page,
    setPage,
    loading,
    busy,
    error,
    success,
    form,
    setForm,
    editing,
    edit,
    cancelEdit,
    save,
    archive,
    refresh,
  };
}
