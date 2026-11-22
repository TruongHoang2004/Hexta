"use client";

import { useEffect, useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useErrorText } from "@/lib/errors";
import { sdk } from "@/lib/sdk";
import { onboarding, type Invitation } from "@/lib/services/onboarding";

/** Members, and for owners the invitations, with invite and revoke. */
export function useMembers(tenantID: string, owner: boolean) {
  const errorText = useErrorText();
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [invites, setInvites] = useState<Invitation[]>([]);
  const [link, setLink] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);

  useEffect(() => {
    if (!tenantID) return;
    let ignore = false;
    async function load() {
      setLoading(true);
      try {
        const [m, i] = await Promise.all([
          sdk.identity.getUsers(tenantID),
          owner ? onboarding.invitations(tenantID) : Promise.resolve([]),
        ]);
        if (!ignore) {
          setMembers(m);
          setInvites(i);
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
  }, [tenantID, owner, revision, errorText]);

  /** Creates an invitation and exposes its one-time link; resolves to whether it succeeded. */
  async function invite(email: string): Promise<boolean> {
    if (!tenantID) return false;
    setBusy(true);
    setError("");
    setLink("");
    try {
      const result = await onboarding.invite(tenantID, email);
      setLink(`${window.location.origin}/invitations#${result.token}`);
      setRevision((v) => v + 1);
      return true;
    } catch (err) {
      setError(errorText(err));
      return false;
    } finally {
      setBusy(false);
    }
  }

  async function revoke(id: string) {
    if (!tenantID) return;
    setBusy(true);
    setError("");
    try {
      await onboarding.revoke(tenantID, id);
      setRevision((v) => v + 1);
      setLink("");
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }

  function refresh() {
    setError("");
    setRevision((v) => v + 1);
  }

  return { members, invites, link, error, busy, loading, invite, revoke, refresh };
}
