"use client";
import { useEffect, useState } from "react";
import { Textarea } from "@hexta/ui";
import { onboarding, type Application } from "@/lib/services/onboarding";
import { isConflict, useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import {
  buttonClass,
  ConfirmDialog,
  EmptyState,
  Notice,
  StatusChip,
  Tabs,
  secondaryClass,
} from "@/components/primitives";

type Status = Application["status"];

export default function PlatformPage() {
  const t = useT();
  const errorText = useErrorText();
  const [rows, setRows] = useState<Application[]>([]);
  const [tab, setTab] = useState<Status>("pending");
  const [rejecting, setRejecting] = useState<Application | null>(null);
  const [reason, setReason] = useState("");
  const [touched, setTouched] = useState(false);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    let ignore = false;
    async function load() {
      setLoading(true);
      try {
        const data = await onboarding.applications(true);
        if (!ignore) setRows(data);
      } catch (err) {
        if (!ignore) setError(errorText(err));
      } finally {
        if (!ignore) setLoading(false);
      }
    }
    void load();
    return () => {
      ignore = true;
    };
  }, [revision, errorText]);
  async function review(row: Application, decision: "approved" | "rejected", text: string) {
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await onboarding.review(row.id, decision, text);
      setRejecting(null);
      setSuccess(t(decision === "approved" ? "platform.approved" : "platform.rejected", { name: row.name }));
    } catch (err) {
      setError(isConflict(err) ? t("platform.conflict") : errorText(err));
      if (isConflict(err)) setRejecting(null);
    } finally {
      setBusy(false);
      setRevision((v) => v + 1);
    }
  }
  const shown = rows.filter((row) => row.status === tab);
  const tabs = (["pending", "approved", "rejected"] as const).map((value) => ({
    value,
    label: t(`platform.tabs.${value}`),
    badge: <StatusChip>{rows.filter((row) => row.status === value).length}</StatusChip>,
  }));
  const reasonMissing = touched && !reason.trim();
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("platform.title")}</h1>
          <p className="mt-1 max-w-2xl text-sm text-muted">{t("platform.intro")}</p>
        </div>
        <button
          type="button"
          className={secondaryClass}
          disabled={busy || loading}
          onClick={() => {
            setError("");
            setRevision((v) => v + 1);
          }}
        >
          {t("platform.refresh")}
        </button>
      </div>
      <Notice error={error} success={success} />
      <Tabs label={t("platform.tabsLabel")} items={tabs} value={tab} onChange={setTab}>
        {loading ? (
          <p role="status">{t("platform.loading")}</p>
        ) : !shown.length ? (
          <EmptyState title={t(`platform.empty.${tab}`)} />
        ) : (
          <ul className="space-y-3">
            {shown.map((row) => (
              <li
                key={row.id}
                className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-surface p-4"
              >
                <div className="min-w-0">
                  <h2 className="font-semibold">{row.name}</h2>
                  <p className="text-sm text-muted">
                    {t("platform.slug", { slug: row.slug })} ·{" "}
                    {t("platform.submitted", { date: new Date(row.created_at).toLocaleString() })}
                  </p>
                  {row.status !== "pending" && row.review_reason && (
                    <p className="mt-1 text-sm">{t("platform.reason", { reason: row.review_reason })}</p>
                  )}
                </div>
                {row.status === "pending" && (
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={buttonClass}
                      disabled={busy}
                      onClick={() => void review(row, "approved", "")}
                    >
                      {t("platform.approve")}
                    </button>
                    <button
                      type="button"
                      className={secondaryClass}
                      disabled={busy}
                      onClick={() => {
                        setRejecting(row);
                        setReason("");
                        setTouched(false);
                      }}
                    >
                      {t("platform.reject")}
                    </button>
                  </div>
                )}
              </li>
            ))}
          </ul>
        )}
      </Tabs>
      <ConfirmDialog
        open={!!rejecting}
        title={t("platform.rejectTitle", { name: rejecting?.name ?? "" })}
        confirmLabel={t("platform.rejectConfirm")}
        cancelLabel={t("common.cancel")}
        busy={busy}
        onCancel={() => setRejecting(null)}
        onConfirm={() => {
          setTouched(true);
          if (rejecting && reason.trim()) void review(rejecting, "rejected", reason.trim());
        }}
      >
        <label className="block font-medium" htmlFor="reject-reason">
          {t("platform.rejectLabel")}
        </label>
        <Textarea
          id="reject-reason"
          className={reasonMissing ? "border-2 border-error" : undefined}
          rows={4}
          maxLength={2000}
          value={reason}
          placeholder={t("platform.rejectPlaceholder")}
          aria-invalid={reasonMissing}
          aria-describedby={reasonMissing ? "reject-error" : undefined}
          onChange={(e) => setReason(e.target.value)}
        />
        {reasonMissing && (
          <p id="reject-error" role="alert" className="text-error">
            {t("platform.rejectRequired")}
          </p>
        )}
      </ConfirmDialog>
    </div>
  );
}
