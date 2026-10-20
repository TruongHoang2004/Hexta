"use client";

import { BrandLogo } from "@/components/brand-logo";
import { createContext, useContext, useEffect, useState } from "react";
import Link from "next/link";
import { useParams, usePathname } from "next/navigation";
import type { Tenant } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { AuthNav } from "@/components/auth-nav";
import { errorText } from "@/lib/services/sales";

interface WorkspaceState {
  tenant: Tenant | null;
  tenants: Tenant[];
  reload: () => Promise<void>;
}
const Context = createContext<WorkspaceState | null>(null);
export function useWorkspace() {
  const state = useContext(Context);
  if (!state) throw new Error("Tenant context is required");
  return state;
}
export function Workspace({ children }: { children: React.ReactNode }) {
  const params = useParams();
  const tenantID = typeof params.tenantId === "string" ? params.tenantId : "";
  const pathname = usePathname();
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
  }, [tenantID]);
  const resolving =
    loading || (!!tenantID && tenant?.id !== tenantID && !error);
  const base = `/tenants/${tenantID}`;
  const links = [
    [`${base}/orders`, "Orders"],
    [`${base}/customers`, "Customers"],
    [`${base}/products`, "Products"],
    [`${base}/work`, "Work"],
    [`${base}/workflows`, "Workflows"],
    [`${base}/settings/members`, "Members & invitations"],
    [`${base}/settings/subscription`, "Service plan"],
  ];
  return (
    <Context.Provider
      value={{
        tenant: resolving || tenant?.id !== tenantID ? null : tenant,
        tenants,
        reload,
      }}
    >
      <div className="min-h-screen bg-background">
        <header className="border-b border-border bg-surface">
          <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4 px-5 py-4">
            <Link href="/tenants" className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <BrandLogo />
              <span className="text-sm font-normal text-muted">
                {tenantID ? "Business workspace" : "Your account"}
              </span>
            </Link>
            <AuthNav />
          </div>
        </header>
        <div
          className={
            tenantID
              ? "mx-auto grid max-w-7xl gap-6 p-5 md:grid-cols-[210px_1fr]"
              : "mx-auto max-w-5xl p-5 md:py-10"
          }
        >
          {tenantID && !resolving && !error && tenant && (
            <aside className="space-y-5">
              <div className="rounded-lg border border-border bg-surface p-3">
                <p className="text-xs text-muted">Tenant</p>
                <p className="mt-1 font-semibold">{tenant.name}</p>
                <Link
                  href="/tenants"
                  className="mt-2 inline-block text-sm text-primary underline"
                >
                  Change tenant
                </Link>
              </div>
              <nav
                aria-label="Tenant resources"
                className="flex flex-wrap gap-1 md:flex-col"
              >
                {links.map(([href, title]) => (
                  <Link
                    key={href}
                    href={href}
                    aria-current={pathname === href ? "page" : undefined}
                    className={`rounded-lg px-3 py-2 text-sm font-medium ${pathname === href ? "bg-primary text-white" : "text-muted hover:bg-surface"}`}
                  >
                    {title}
                  </Link>
                ))}
              </nav>
            </aside>
          )}
          <main className="min-w-0">
            {resolving ? (
              <p role="status">Loading account…</p>
            ) : error ? (
              <div
                role="alert"
                className="space-y-3 rounded-xl border border-border p-5"
              >
                <p>{error}</p>
                <Link href="/tenants" className="text-primary underline">
                  Back to your tenants
                </Link>
              </div>
            ) : tenantID && (!tenant || tenant.status !== "active") ? (
              <div className="rounded-xl border border-border p-5">
                <h1 className="text-xl font-semibold">
                  Tenant access is unavailable
                </h1>
                <p className="my-3 text-muted">
                  Contact Hexta or the tenant owner for assistance.
                </p>
                <Link href="/tenants" className="text-primary underline">
                  Back to your tenants
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
