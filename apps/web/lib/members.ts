"use client";

import { useCallback } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { useT } from "@/lib/i18n/provider";

/** Display name for a member ID: "You", the member's email, or a short ID for former members. */
export function useMemberName(members: TenantMember[]): (id: string | null | undefined) => string {
  const t = useT();
  const me = useAuthStore((state) => state.user?.id);
  return useCallback(
    (id) => {
      if (!id) return t("common.unassigned");
      if (id === me) return t("common.you");
      const member = members.find((row) => row.user_id === id);
      if (member?.email) return member.email;
      return member ? `#${id.slice(-6)}` : t("common.formerMember");
    },
    [members, me, t],
  );
}
