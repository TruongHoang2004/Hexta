"use client";

import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import { sdk } from "@/lib/sdk";
import { sales } from "../services/sales";
import type { Customer, Product } from "../types";

/** Customers, products and members the order editor picks from. Reloads when `revision` changes. */
export function useOrderLookups(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
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
        const [c, p, m] = await Promise.all([
          sales.all<Customer>("customers", tenantID, abort.signal),
          sales.all<Product>("products", tenantID, abort.signal),
          sdk.identity.getUsers(tenantID),
        ]);
        if (!abort.signal.aborted) {
          setCustomers(c);
          setProducts(p);
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

  return { customers, products, members, loading, error };
}
