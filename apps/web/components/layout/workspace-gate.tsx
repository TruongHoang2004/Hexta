"use client";

import Link from "next/link";
import type { Tenant } from "@hexta/sdk";
import { useT } from "@/lib/i18n/provider";

/** Renders the page only once the workspace is loaded and active; otherwise loading, error or unavailable. */
export function WorkspaceGate({
  tenantID,
  tenant,
  resolving,
  error,
  children,
}: {
  tenantID: string;
  tenant: Tenant | null;
  resolving: boolean;
  error: string;
  children: React.ReactNode;
}) {
  const t = useT();
  if (resolving) return <p role="status">{t("shell.loadingAccount")}</p>;
  if (error)
    return (
      <div role="alert" className="space-y-3 rounded-xl border border-border p-5">
        <p>{error}</p>
        <Link href="/tenants" className="text-primary underline">
          {t("shell.backToWorkspaces")}
        </Link>
      </div>
    );
  if (tenantID && (!tenant || tenant.status !== "active"))
    return (
      <div className="rounded-xl border border-border p-5">
        <h1 className="text-xl font-semibold">{t("shell.unavailableTitle")}</h1>
        <p className="my-3 text-muted">{t("shell.unavailableBody")}</p>
        <Link href="/tenants" className="text-primary underline">
          {t("shell.backToWorkspaces")}
        </Link>
      </div>
    );
  return <div key={tenantID || "account"}>{children}</div>;
}
