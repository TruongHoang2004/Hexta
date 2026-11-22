"use client";

import type { TenantMember } from "@hexta/sdk";
import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import { roleLabel } from "../../utils";

export function MemberList({ members, loading }: { members: TenantMember[]; loading: boolean }) {
  const t = useT();
  const userID = useAuthStore((state) => state.user?.id);
  return (
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
                {m.user_id === userID && <span className="text-muted">{t("members.you")}</span>}
              </span>
              <StatusChip tone={m.role === "owner" ? "solid" : "neutral"}>{roleLabel(t, m.role)}</StatusChip>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
