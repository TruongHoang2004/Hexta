"use client";
import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useWorkspace } from "@/components/sales/workspace";
import { useAuthStore } from "@/store/useAuthStore";
import { sdk } from "@/lib/sdk";
import { onboarding, type Invitation } from "@/lib/services/onboarding";
import { errorText } from "@/lib/services/sales";
import {
  buttonClass,
  inputClass,
  Field,
  Notice,
  secondaryClass,
} from "@/components/sales/primitives";
export default function MembersPage() {
  const { tenant } = useWorkspace();
  const { user } = useAuthStore();
  const owner = tenant?.owner_id === user?.id;
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [invites, setInvites] = useState<Invitation[]>([]);
  const [email, setEmail] = useState("");
  const [link, setLink] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    if (!tenant) return;
    let ignore = false;
    async function load() {
      setLoading(true);
      try {
        const [m, i] = await Promise.all([
          sdk.identity.getUsers(tenant!.id),
          owner ? onboarding.invitations(tenant!.id) : Promise.resolve([]),
        ]);
        if (!ignore) {
          setMembers(m);
          setInvites(i);
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
  }, [tenant, owner, revision]);
  async function invite(event: React.FormEvent) {
    event.preventDefault();
    if (!tenant) return;
    setBusy(true);
    setError("");
    setLink("");
    try {
      const result = await onboarding.invite(tenant.id, email);
      setLink(`${window.location.origin}/invitations#${result.token}`);
      setEmail("");
      setRevision((v) => v + 1);
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function revoke(id: string) {
    if (!tenant) return;
    setBusy(true);
    setError("");
    try {
      await onboarding.revoke(tenant.id, id);
      setRevision((v) => v + 1);
      setLink("");
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-semibold">Members & invitations</h1>
      <p className="text-muted">
        Members must accept an owner invitation before they can access this
        tenant.
      </p>
      <Notice error={error} />
      <button
        className={secondaryClass}
        disabled={loading || busy}
        onClick={() => {
          setError("");
          setRevision((v) => v + 1);
        }}
      >
        Refresh
      </button>
      <section className="rounded-xl border border-border bg-surface p-5">
        <h2 className="font-semibold">Active members</h2>
        {loading ? (
          <p role="status" className="mt-3">
            Loading members…
          </p>
        ) : (
          <ul className="divide-y divide-border">
            {members.map((m) => (
              <li key={m.id} className="flex justify-between gap-3 py-3">
                <span className="break-all font-mono text-xs">
                  {m.user_id}
                  {m.user_id === user?.id ? " (you)" : ""}
                </span>
                <span className="text-xs uppercase text-primary">{m.role}</span>
              </li>
            ))}
          </ul>
        )}
      </section>
      {owner && (
        <>
          <form
            onSubmit={invite}
            className="max-w-lg space-y-4 rounded-xl border border-border bg-surface p-5"
          >
            <h2 className="font-semibold">Invite an employee</h2>
            <Field title="Recipient email">
              <input
                className={inputClass}
                type="email"
                required
                maxLength={255}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={busy}
              />
            </Field>
            <p className="text-sm text-muted">
              Email sending is not connected. Create a private invitation link
              and share it directly with this recipient. The link expires after
              7 days. A new invitation to the same email revokes its previous
              pending link.
            </p>
            <button className={buttonClass} disabled={busy}>
              {busy ? "Creating…" : "Create invitation link"}
            </button>
          </form>
          {link && (
            <div className="space-y-3 rounded-xl border border-primary bg-surface p-5">
              <label className="block text-sm font-medium">
                Private invitation link
                <input
                  className={`${inputClass} mt-2`}
                  readOnly
                  value={link}
                  onFocus={(e) => e.target.select()}
                />
              </label>
              <p className="text-sm text-muted">
                No email has been sent. Copy this link now; it is shown only
                once. Share it only with the intended recipient.
              </p>
            </div>
          )}
          <section className="space-y-3">
            <h2 className="font-semibold">Invitations</h2>
            {!invites.length ? (
              <p className="text-sm text-muted">No invitations yet.</p>
            ) : (
              <ul className="divide-y divide-border rounded-xl border border-border bg-surface px-5">
                {invites.map((i) => (
                  <li
                    key={i.id}
                    className="flex flex-wrap items-center justify-between gap-3 py-4"
                  >
                    <div>
                      <p className="text-sm font-medium">{i.email}</p>
                      <p className="mt-1 text-xs text-muted">
                        {i.status} · Expires{" "}
                        {new Date(i.expires_at).toLocaleString()}
                      </p>
                    </div>
                    {i.status === "pending" && (
                      <button
                        className={secondaryClass}
                        disabled={busy}
                        onClick={() => void revoke(i.id)}
                      >
                        Revoke
                      </button>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </section>
        </>
      )}
    </div>
  );
}
