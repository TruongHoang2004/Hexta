"use client";
import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useWorkspace } from "@/components/sales/workspace";
import { useAuthStore } from "@/store/useAuthStore";
import { sdk } from "@/lib/sdk";
import { onboarding, type Invitation } from "@/lib/services/onboarding";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import {
  buttonClass,
  inputClass,
  EmptyState,
  Field,
  Notice,
  StatusChip,
  secondaryClass,
  type ChipTone,
} from "@/components/sales/primitives";

const roles = ["owner", "admin", "member"] as const;
const statuses = ["pending", "accepted", "declined", "revoked", "expired"] as const;
const statusTone: Record<(typeof statuses)[number], ChipTone> = {
  pending: "warning",
  accepted: "success",
  declined: "neutral",
  revoked: "neutral",
  expired: "neutral",
};

export default function MembersPage() {
  const { tenant, isOwner: owner } = useWorkspace();
  const { user } = useAuthStore();
  const t = useT();
  const errorText = useErrorText();
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [invites, setInvites] = useState<Invitation[]>([]);
  const [email, setEmail] = useState("");
  const [link, setLink] = useState("");
  const [copied, setCopied] = useState(false);
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
  }, [tenant, owner, revision, errorText]);
  async function invite(event: React.FormEvent) {
    event.preventDefault();
    if (!tenant) return;
    setBusy(true);
    setError("");
    setLink("");
    setCopied(false);
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
  async function copy() {
    try {
      await navigator.clipboard.writeText(link);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }
  const roleLabel = (role: string) =>
    (roles as readonly string[]).includes(role) ? t(`members.roles.${role as (typeof roles)[number]}`) : role;
  const statusOf = (status: string) =>
    (statuses as readonly string[]).includes(status) ? (status as (typeof statuses)[number]) : null;
  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-muted">{t("members.intro")}</p>
        <button
          type="button"
          className={secondaryClass}
          disabled={loading || busy}
          onClick={() => {
            setError("");
            setRevision((v) => v + 1);
          }}
        >
          {t("members.refresh")}
        </button>
      </div>
      <Notice error={error} />
      {owner && (
        <form onSubmit={invite} className="space-y-3 rounded-xl border border-border bg-surface p-5">
          <h2 className="font-semibold">{t("members.inviteTitle")}</h2>
          <div className="flex flex-wrap items-end gap-3">
            <div className="min-w-0 flex-[1_1_280px]">
              <Field title={t("members.email")}>
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
            </div>
            <StatusChip>{t("members.roleNote")}</StatusChip>
            <button className={buttonClass} disabled={busy}>
              {busy ? t("members.creating") : t("members.create")}
            </button>
          </div>
          <p className="text-sm text-muted">{t("members.inviteHelp")}</p>
          {link && (
            <div className="space-y-2 rounded-lg border border-primary bg-background p-4">
              <label className="block text-sm font-medium">
                {t("members.linkLabel")}
                <span className="mt-2 flex gap-2">
                  <input className={inputClass} readOnly value={link} onFocus={(e) => e.target.select()} />
                  <button type="button" className={secondaryClass} onClick={() => void copy()}>
                    {copied ? t("members.copied") : t("members.copy")}
                  </button>
                </span>
              </label>
              <p className="text-sm text-muted">{t("members.linkNote")}</p>
            </div>
          )}
        </form>
      )}
      <section className="rounded-xl border border-border bg-surface p-5">
        <h2 className="font-semibold">{t("members.active", { count: members.length })}</h2>
        {loading ? (
          <p role="status" className="mt-3">
            {t("members.loading")}
          </p>
        ) : (
          <ul className="divide-y divide-border">
            {members.map((m) => (
              <li key={m.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <span className="break-all text-sm">
                  {m.email || `#${m.user_id.slice(-6)}`}{" "}
                  {m.user_id === user?.id && <span className="text-muted">{t("members.you")}</span>}
                </span>
                <StatusChip tone={m.role === "owner" ? "solid" : "neutral"}>{roleLabel(m.role)}</StatusChip>
              </li>
            ))}
          </ul>
        )}
      </section>
      {owner && (
        <section className="space-y-3 rounded-xl border border-border bg-surface p-5">
          <h2 className="font-semibold">{t("members.invitations")}</h2>
          {!invites.length ? (
            <EmptyState title={t("members.noInvitations")} />
          ) : (
            <ul className="divide-y divide-border">
              {invites.map((i) => {
                const status = statusOf(i.status);
                return (
                  <li key={i.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                    <div>
                      <p className="text-sm font-medium">{i.email}</p>
                      <p className="mt-1 flex flex-wrap items-center gap-2 text-xs text-muted">
                        <StatusChip tone={status ? statusTone[status] : "neutral"}>
                          {status ? t(`members.status.${status}`) : t("members.status.other")}
                        </StatusChip>
                        {t("members.expires", { date: new Date(i.expires_at).toLocaleString() })}
                      </p>
                    </div>
                    {i.status === "pending" && (
                      <button type="button" className={secondaryClass} disabled={busy} onClick={() => void revoke(i.id)}>
                        {t("members.revoke")}
                      </button>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </section>
      )}
    </div>
  );
}
