"use client";

import { useCallback, useEffect, useState } from "react";
import { useErrorText } from "@/lib/errors";
import { scripts } from "../services/scripts";
import type { MyScriptPermissions, ScriptPermission } from "../types";

/**
 * The caller's current script permissions, used only to show or hide actions; the API authorizes every command
 * again. Reloads on `revision` so a revoked permission disappears after the next refused action.
 */
export function useScriptAccess(tenantID: string, revision: number) {
  const errorText = useErrorText();
  const [access, setAccess] = useState<MyScriptPermissions | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!tenantID) return;
    const abort = new AbortController();
    scripts
      .myPermissions(tenantID, abort.signal)
      .then((result) => {
        if (!abort.signal.aborted) {
          setAccess(result);
          setError("");
        }
      })
      .catch((err) => {
        if (!abort.signal.aborted) setError(errorText(err));
      });
    return () => abort.abort();
  }, [tenantID, revision, errorText]);

  const can = useCallback((permission: ScriptPermission) => !!access?.permissions.includes(permission), [access]);
  const readOnly = access !== null && access.permissions.length === 0;
  return { access, can, readOnly, error };
}
