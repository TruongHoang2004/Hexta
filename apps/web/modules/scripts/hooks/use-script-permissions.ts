"use client";

import { useEffect, useState } from "react";
import { isConflict, useErrorText } from "@/lib/errors";
import { onboarding, type Invitation } from "@/lib/services/onboarding";
import { scripts } from "../services/scripts";
import type { MemberScriptPermissions, ScriptPermission } from "../types";

export interface PermissionChange {
  member: MemberScriptPermissions;
  permission: ScriptPermission;
  granted: boolean;
}

/** Owner view of members' script permissions and pending invitations, with grant and revoke. */
export function useScriptPermissions(tenantID: string, owner: boolean) {
  const errorText = useErrorText();
  const [members, setMembers] = useState<MemberScriptPermissions[]>([]);
  const [invites, setInvites] = useState<Invitation[]>([]);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState("");
  const [error, setError] = useState("");
  const [conflict, setConflict] = useState(false);
  const [changed, setChanged] = useState<PermissionChange | null>(null);
  const [revision, setRevision] = useState(0);

  useEffect(() => {
    if (!tenantID || !owner) return;
    const abort = new AbortController();
    async function load() {
      setLoading(true);
      try {
        const [rows, invitations] = await Promise.all([
          scripts.memberPermissions(tenantID, abort.signal),
          onboarding.invitations(tenantID),
        ]);
        if (abort.signal.aborted) return;
        setMembers(rows);
        setInvites(invitations.filter((i) => i.status === "pending"));
        setError("");
        setConflict(false);
      } catch (err) {
        if (!abort.signal.aborted) setError(errorText(err));
      } finally {
        if (!abort.signal.aborted) setLoading(false);
      }
    }
    void load();
    return () => abort.abort();
  }, [tenantID, owner, revision, errorText]);

  async function change(member: MemberScriptPermissions, permission: ScriptPermission, granted: boolean) {
    if (pending) return;
    setPending(`${member.user_id}:${permission}`);
    setError("");
    setChanged(null);
    try {
      const next = await scripts.setPermission(tenantID, member, permission, granted);
      setMembers((rows) => rows.map((row) => (row.user_id === member.user_id ? { ...row, ...next, email: row.email, role: row.role } : row)));
      setChanged({ member, permission, granted });
    } catch (err) {
      setConflict(isConflict(err));
      if (!isConflict(err)) setError(errorText(err));
    } finally {
      setPending("");
    }
  }

  return {
    members,
    invites,
    loading,
    pending,
    error,
    conflict,
    changed,
    change,
    reload: () => {
      setChanged(null);
      setRevision((v) => v + 1);
    },
  };
}
