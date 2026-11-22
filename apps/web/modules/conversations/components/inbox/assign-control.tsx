"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";

const inputStyle = "w-full rounded-lg border border-border bg-background p-2 text-sm";

/** Owner-only control to assign the turn to a member or back to the queue. */
export function AssignControl({
  members,
  currentAssignee,
  disabled,
  onAssign,
}: {
  members: TenantMember[];
  currentAssignee: string;
  disabled: boolean;
  onAssign: (assignee: string | null) => void;
}) {
  const t = useT();
  const memberName = useMemberName(members);
  const [assignee, setAssignee] = useState(currentAssignee);
  return (
    <div className="space-y-2 border-t border-border pt-3">
      <label className="block text-sm">
        {t("desk.actions.assignTo")}
        <select className={inputStyle} disabled={disabled} value={assignee} onChange={(e) => setAssignee(e.target.value)}>
          <option value="">{t("desk.actions.assignQueue")}</option>
          {members.map((member) => (
            <option key={member.user_id} value={member.user_id}>
              {memberName(member.user_id)}
            </option>
          ))}
        </select>
      </label>
      <button type="button" className={secondaryClass} disabled={disabled} onClick={() => onAssign(assignee || null)}>
        {t("desk.actions.assign")}
      </button>
    </div>
  );
}
