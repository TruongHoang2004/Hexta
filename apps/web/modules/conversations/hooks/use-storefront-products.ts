"use client";

import { useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { sales, type Product } from "@/modules/sales";

/** Every product of the workspace for the storefront simulation. */
export function useStorefrontProducts(tenantID: string) {
  const errorText = useErrorText();
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    if (!tenantID) return;
    const controller = new AbortController();
    sales
      .all<Product>("products", tenantID, controller.signal)
      .then(setProducts)
      .catch((err) => {
        if (!controller.signal.aborted) setError(errorText(err));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [tenantID, errorText]);
  return { products, loading, error, setError };
}
