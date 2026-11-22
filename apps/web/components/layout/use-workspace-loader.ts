"use client";

import { useEffect, useState } from "react";
import type { Tenant } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { useErrorText } from "@/lib/errors";

/** Loads the user's workspaces and the one named in the URL (if any). */
export function useWorkspaceLoader(tenantID: string) {
  const errorText = useErrorText();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
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

  async function reload() {
    const rows = await sdk.identity.getTenants();
    setTenants(rows);
    if (tenantID) setTenant(await sdk.identity.getTenant(tenantID));
  }

  /** True until the loaded tenant matches the URL, so a previous workspace is never shown. */
  const resolving = loading || (!!tenantID && tenant?.id !== tenantID && !error);
  return { tenants, tenant, error, resolving, reload };
}
