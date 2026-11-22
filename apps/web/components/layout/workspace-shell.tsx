"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { AuthNav } from "@/components/auth-nav";
import { BrandLogo } from "@/components/brand-logo";
import { useT } from "@/lib/i18n/provider";
import { useAuthStore } from "@/store/useAuthStore";
import { SidebarNav } from "./sidebar-nav";
import { useWorkspaceLoader } from "./use-workspace-loader";
import { WorkspaceContext, type WorkspaceRole } from "./workspace-context";
import { WorkspaceGate } from "./workspace-gate";
import { WorkspaceSwitcher } from "./workspace-switcher";

/**
 * Dashboard layout. Inside a workspace: full-height sidebar plus header. Outside one (workspace list,
 * registration) or while it cannot be shown: a simple header above the page.
 */
export function WorkspaceShell({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const tenantID = typeof params.tenantId === "string" ? params.tenantId : "";
  const t = useT();
  const userID = useAuthStore((state) => state.user?.id);
  const { tenants, tenant, error, resolving, reload } = useWorkspaceLoader(tenantID);
  const isOwner = !!tenant && !!userID && tenant.owner_id === userID;
  const role: WorkspaceRole = isOwner ? "owner" : "member";
  const page = (
    <WorkspaceGate tenantID={tenantID} tenant={tenant} resolving={resolving} error={error}>
      {children}
    </WorkspaceGate>
  );

  if (tenantID && !resolving && !error && tenant) {
    return (
      <WorkspaceContext.Provider value={{ tenant, tenants, role, isOwner, reload }}>
        <div className="flex min-h-screen flex-col bg-background md:flex-row">
          <aside className="flex shrink-0 flex-col border-b border-border bg-surface md:sticky md:top-0 md:h-screen md:w-64 md:border-b-0 md:border-r">
            <Link href="/tenants" className="flex min-h-16 items-center border-b border-border px-5">
              <BrandLogo />
            </Link>
            <WorkspaceSwitcher current={tenant} tenants={tenants} roleLabel={t(`roles.${role}`)} />
            <SidebarNav tenantID={tenantID} />
          </aside>
          <div className="flex min-w-0 flex-1 flex-col">
            <header className="flex min-h-16 flex-wrap items-center justify-between gap-4 border-b border-border bg-surface px-5 py-3">
              <span className="text-sm text-muted">{t("shell.businessWorkspace")}</span>
              <AuthNav />
            </header>
            <main className="min-w-0 flex-1 p-5">{page}</main>
          </div>
        </div>
      </WorkspaceContext.Provider>
    );
  }

  return (
    <WorkspaceContext.Provider value={{ tenant: resolving || tenant?.id !== tenantID ? null : tenant, tenants, role, isOwner, reload }}>
      <div className="min-h-screen bg-background">
        <header className="border-b border-border bg-surface">
          <div className="flex flex-wrap items-center justify-between gap-4 px-5 py-3">
            <Link href="/tenants" className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <BrandLogo />
              <span className="text-sm font-normal text-muted">
                {tenantID ? t("shell.businessWorkspace") : t("shell.yourAccount")}
              </span>
            </Link>
            <AuthNav />
          </div>
        </header>
        <main className={tenantID ? "min-w-0 p-5" : "mx-auto min-w-0 max-w-5xl p-5 md:py-10"}>{page}</main>
      </div>
    </WorkspaceContext.Provider>
  );
}
