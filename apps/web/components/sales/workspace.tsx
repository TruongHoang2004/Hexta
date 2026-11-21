"use client";

import { BrandLogo } from "@/components/brand-logo";
import { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, usePathname } from "next/navigation";
import type { Tenant } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { AuthNav } from "@/components/auth-nav";
import { useAuthStore } from "@/store/useAuthStore";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";

/** Tenant role as seen by the UI; the API still authorizes every action. */
export type WorkspaceRole = "owner" | "member";

interface WorkspaceState {
  tenant: Tenant | null;
  tenants: Tenant[];
  role: WorkspaceRole;
  isOwner: boolean;
  reload: () => Promise<void>;
}
const Context = createContext<WorkspaceState | null>(null);
export function useWorkspace() {
  const state = useContext(Context);
  if (!state) throw new Error("Tenant context is required");
  return state;
}

interface NavItem {
  path: string;
  label: MessageKey;
}
const mainNav: NavItem[] = [
  { path: "inbox", label: "nav.inbox" },
  { path: "storefront", label: "nav.consultant" },
  { path: "orders", label: "nav.orders" },
  { path: "work", label: "nav.work" },
  { path: "customers", label: "nav.customers" },
  { path: "products", label: "nav.products" },
  { path: "workflows", label: "nav.workflows" },
];
export const settingsNav: NavItem[] = [
  { path: "settings/members", label: "nav.members" },
  { path: "settings/channels", label: "nav.channels" },
  { path: "settings/subscription", label: "nav.plan" },
];

function NavLink({ href, active, children }: { href: string; active: boolean; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      aria-current={active ? "page" : undefined}
      className={`flex min-h-10 items-center rounded-lg px-3 text-sm ${active ? "bg-primary/10 font-semibold text-foreground" : "text-muted hover:bg-background hover:text-foreground"}`}
    >
      {children}
    </Link>
  );
}

export function Workspace({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const tenantID = typeof params.tenantId === "string" ? params.tenantId : "";
  const pathname = usePathname();
  const t = useT();
  const errorText = useErrorText();
  const userID = useAuthStore((state) => state.user?.id);
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  async function reload() {
    const rows = await sdk.identity.getTenants();
    setTenants(rows);
    if (tenantID) setTenant(await sdk.identity.getTenant(tenantID));
  }
  useEffect(() => {
    let ignore = false;
    async function load() {
      setLoading(true);
      setError("");
      try {
        const [rows, current] = await Promise.all([
          sdk.identity.getTenants(),
          tenantID ? sdk.identity.getTenant(tenantID) : Promise.resolve(null),
        ]);
        if (!ignore) {
          setTenants(rows);
          setTenant(current);
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
  }, [tenantID, errorText]);
  const resolving = loading || (!!tenantID && tenant?.id !== tenantID && !error);
  const base = `/tenants/${tenantID}`;
  const isOwner = !!tenant && !!userID && tenant.owner_id === userID;
  const role: WorkspaceRole = isOwner ? "owner" : "member";
  const isActive = (path: string) => pathname === `${base}/${path}` || pathname.startsWith(`${base}/${path}/`);
  const showShell = !!tenantID && !resolving && !error && !!tenant;
  return (
    <Context.Provider
      value={{
        tenant: resolving || tenant?.id !== tenantID ? null : tenant,
        tenants,
        role,
        isOwner,
        reload,
      }}
    >
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
        <div className={tenantID ? "flex flex-wrap" : "mx-auto max-w-5xl p-5 md:py-10"}>
          {showShell && (
            <aside className="flex-[1_1_220px] border-border bg-surface md:max-w-[240px] md:border-r">
              <div className="flex flex-col gap-1 p-3 md:sticky md:top-0">
                <div className="mb-2 border-b border-border px-2 pb-3">
                  <p className="font-semibold">{tenant.name}</p>
                  <p className="text-xs text-muted">{t(`roles.${role}`)}</p>
                  <Link href="/tenants" className="mt-1 inline-block text-xs text-primary underline">
                    {t("shell.changeWorkspace")}
                  </Link>
                </div>
                <nav aria-label={t("shell.navLabel")} className="flex flex-col gap-0.5">
                  {mainNav.map((item) => (
                    <NavLink key={item.path} href={`${base}/${item.path}`} active={isActive(item.path)}>
                      {t(item.label)}
                    </NavLink>
                  ))}
                  <p className="mt-3 px-3 pb-1 text-xs font-medium uppercase tracking-wide text-muted">
                    {t("shell.settings")}
                  </p>
                  {settingsNav.map((item) => (
                    <NavLink key={item.path} href={`${base}/${item.path}`} active={isActive(item.path)}>
                      {t(item.label)}
                    </NavLink>
                  ))}
                </nav>
              </div>
            </aside>
          )}
          <main className={tenantID ? "min-w-0 flex-[999_1_560px] p-5" : "min-w-0"}>
            {resolving ? (
              <p role="status">{t("shell.loadingAccount")}</p>
            ) : error ? (
              <div role="alert" className="space-y-3 rounded-xl border border-border p-5">
                <p>{error}</p>
                <Link href="/tenants" className="text-primary underline">
                  {t("shell.backToWorkspaces")}
                </Link>
              </div>
            ) : tenantID && (!tenant || tenant.status !== "active") ? (
              <div className="rounded-xl border border-border p-5">
                <h1 className="text-xl font-semibold">{t("shell.unavailableTitle")}</h1>
                <p className="my-3 text-muted">{t("shell.unavailableBody")}</p>
                <Link href="/tenants" className="text-primary underline">
                  {t("shell.backToWorkspaces")}
                </Link>
              </div>
            ) : (
              <div key={tenantID || "account"}>{children}</div>
            )}
          </main>
        </div>
      </div>
    </Context.Provider>
  );
}
