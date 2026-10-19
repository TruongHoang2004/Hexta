"use client";
import { useEffect, useState } from "react";
import { onboarding, type Application } from "@/lib/services/onboarding";
import { errorText } from "@/lib/services/sales";
import {
  buttonClass,
  inputClass,
  Notice,
  secondaryClass,
} from "@/components/sales/primitives";
export default function PlatformPage() {
  const [rows, setRows] = useState<Application[]>([]);
  const [reasons, setReasons] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
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
  }, [revision]);
  async function review(row: Application, decision: string) {
    setBusy(true);
    setError("");
    try {
      await onboarding.review(row.id, decision, reasons[row.id] || "");
      setRevision((v) => v + 1);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-5">
      <h1 className="text-3xl font-semibold">Hexta tenant review</h1>
      <p className="text-muted">
        Platform operator access is separate from tenant ownership. Approval
        provisions the business and its owner; it does not record payment.
      </p>
      <Notice error={error} />
      <button
        className={secondaryClass}
        disabled={busy || loading}
        onClick={() => {
          setError("");
          setRevision((v) => v + 1);
        }}
      >
        Refresh
      </button>
      {loading ? (
        <p role="status">Loading applications…</p>
      ) : (
        <ul className="space-y-4">
          {rows.map((row) => (
            <li
              key={row.id}
              className="space-y-3 rounded-xl border border-border bg-surface p-5"
            >
              <div className="flex justify-between">
                <h2 className="font-semibold">{row.name}</h2>
                <span className="text-xs uppercase text-primary">
                  {row.status}
                </span>
              </div>
              <p className="text-sm text-muted">Slug: {row.slug}</p>
              {row.status === "pending" ? (
                <>
                  <label className="block text-sm">
                    Review reason (required for rejection)
                    <textarea
                      className={`${inputClass} mt-2`}
                      maxLength={2000}
                      value={reasons[row.id] || ""}
                      onChange={(e) =>
                        setReasons({ ...reasons, [row.id]: e.target.value })
                      }
                      disabled={busy}
                    />
                  </label>
                  <div className="flex gap-3">
                    <button
                      className={buttonClass}
                      disabled={busy}
                      onClick={() => void review(row, "approved")}
                    >
                      Approve tenant
                    </button>
                    <button
                      className={secondaryClass}
                      disabled={busy || !reasons[row.id]?.trim()}
                      onClick={() => void review(row, "rejected")}
                    >
                      Reject
                    </button>
                  </div>
                </>
              ) : (
                <p className="text-sm">{row.review_reason}</p>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
