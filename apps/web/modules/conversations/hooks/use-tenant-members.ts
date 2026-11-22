"use client";

import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { useErrorText } from "@/lib/errors";

/** Workspace members, used to name assignees and fill the assign list. */
export function useTenantMembers(tenantID: string) {
  const errorText = useErrorText();
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    let ignore = false;
    if (tenantID) {
      void sdk.identity
        .getUsers(tenantID)
        .then((result) => {
          if (!ignore) setMembers(result);
        })
        .catch((err) => {
          if (!ignore) setError(errorText(err));
        });
    }
    return () => {
      ignore = true;
    };
  }, [tenantID, errorText]);
  return { members, error };
}
