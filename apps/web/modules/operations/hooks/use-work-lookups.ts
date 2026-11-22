"use client";

import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import { sdk } from "@/lib/sdk";
import { sales, type Order } from "@/modules/sales";
import { workflows } from "../services/workflows";
import type { WorkflowDefinition } from "../types";

/** Workflows, confirmed orders and members used to start and route work. Reloads on `revision`. */
export function useWorkLookups(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [definitions, setDefinitions] = useState<WorkflowDefinition[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      setError("");
      try {
        const [d, o, m] = await Promise.all([
          workflows.allDefinitions(tenantID, abort.signal),
          sales.all<Order>("orders", tenantID, abort.signal),
          sdk.identity.getUsers(tenantID),
        ]);
        if (!abort.signal.aborted) {
          setDefinitions(d);
          setOrders(o.filter((row) => row.status === "confirmed"));
          setMembers(m);
        }
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, revision, errorText]);

  return { definitions, orders, members, loading, error };
}
