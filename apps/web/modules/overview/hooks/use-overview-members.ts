"use client";

import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import { sdk } from "@/lib/sdk";

/** Current workspace members: employee filter options and display names for responsible members. */
export function useOverviewMembers(tenantID: string, enabled: boolean) {
  const errorText = useErrorText();
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!tenantID || !enabled) return;
    let ignore = false;
    sdk.identity
      .getUsers(tenantID)
      .then((rows) => {
        if (!ignore) setMembers(rows);
      })
      .catch((err: unknown) => {
        if (!ignore) setError(errorText(err));
      });
    return () => {
      ignore = true;
    };
  }, [tenantID, enabled, errorText]);
  return { members, error };
}
