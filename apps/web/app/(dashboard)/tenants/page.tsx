"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useWorkspace } from "@/components/sales/workspace";
import {
  buttonClass,
  inputClass,
  Field,
  Notice,
  secondaryClass,
} from "@/components/sales/primitives";
import { onboarding, type Application } from "@/lib/services/onboarding";
import { errorText } from "@/lib/services/sales";
export default function TenantsPage() {
  const { tenants, reload } = useWorkspace();
  const [requests, setRequests] = useState<Application[]>([]);
  const [operator, setOperator] = useState(false);
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    let ignore = false;
    async function load() {
      setLoading(true);
      try {
        const [r, o] = await Promise.all([
          onboarding.applications(),
          onboarding.operator(),
        ]);
        if (!ignore) {
          setRequests(r);
          setOperator(o);
        }
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
  async function apply(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    setSuccess("");
    try {
      await onboarding.apply(name, slug);
      setName("");
      setSlug("");
      setRevision((v) => v + 1);
      setSuccess(
        "Request submitted. Hexta must approve it before tenant access is granted.",
      );
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function refresh() {
    setBusy(true);
    setError("");
    try {
      await reload();
      setRevision((v) => v + 1);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-7">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-primary">Your account</p>
          <h1 className="mt-2 text-3xl font-semibold">Choose a tenant</h1>
          <p className="mt-3 text-muted">
            Enter an approved business to manage its orders, customers and
            products.
          </p>
        </div>
        <button
          className={secondaryClass}
          disabled={busy || loading}
          onClick={() => void refresh()}
        >
          Refresh access
        </button>
      </div>
      <Notice error={error} success={success} />
      {operator && (
        <Link
          href="/platform"
          className="inline-block text-sm text-primary underline"
        >
          Hexta application review →
        </Link>
      )}
      <section aria-label="Your tenants" className="grid gap-4 sm:grid-cols-2">
        {tenants.map((row) =>
          row.status === "active" ? (
            <Link
              key={row.id}
              href={`/tenants/${row.id}/orders`}
              className="rounded-xl border border-border bg-surface p-6 hover:border-primary"
            >
              <h2 className="text-lg font-semibold">{row.name}</h2>
              <p className="mt-1 text-sm text-muted">{row.slug}</p>
              <p className="mt-5 text-sm font-medium text-primary">
                Enter tenant →
              </p>
            </Link>
          ) : (
            <div key={row.id} className="rounded-xl border border-border p-6">
              <h2 className="font-semibold">{row.name}</h2>
              <p className="mt-2 text-sm text-muted">
                Access unavailable · {row.status}
              </p>
            </div>
          ),
        )}
      </section>
      {!tenants.length && (
        <p className="rounded-xl border border-dashed border-border p-5 text-muted">
          No tenant access yet. Submit a registration request, or open the
          invitation link shared by a tenant owner.
        </p>
      )}
      <section className="space-y-3">
        <h2 className="text-lg font-semibold">Registration requests</h2>
        {loading ? (
          <p role="status">Loading requests…</p>
        ) : requests.length ? (
          <ul className="divide-y divide-border rounded-xl border border-border bg-surface px-5">
            {requests.map((row) => (
              <li key={row.id} className="py-4">
                <div className="flex items-center justify-between gap-3">
                  <span className="font-medium">{row.name}</span>
                  <span className="text-xs uppercase text-primary">
                    {row.status}
                  </span>
                </div>
                <p className="mt-1 text-sm text-muted">{row.slug}</p>
                {row.review_reason && (
                  <p className="mt-2 text-sm">{row.review_reason}</p>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-muted">No registration requests.</p>
        )}
      </section>
      <form
        onSubmit={apply}
        className="max-w-lg space-y-4 rounded-xl border border-border bg-surface p-6"
      >
        <h2 className="text-lg font-semibold">Register a business tenant</h2>
        <p className="text-sm text-muted">
          Submitting this form requests Hexta approval. It does not create an
          active tenant or charge a payment.
        </p>
        <Field title="Business name">
          <input
            className={inputClass}
            required
            minLength={2}
            maxLength={100}
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={busy}
          />
        </Field>
        <Field title="Tenant slug">
          <input
            className={inputClass}
            required
            minLength={2}
            maxLength={100}
            pattern="[a-z0-9]+(-[a-z0-9]+)*"
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            disabled={busy}
          />
        </Field>
        <button className={buttonClass} disabled={busy}>
          {busy ? "Submitting…" : "Submit for approval"}
        </button>
      </form>
    </div>
  );
}
