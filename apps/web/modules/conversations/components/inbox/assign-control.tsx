"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { Picker, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";

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
      <label className="block space-y-1.5 text-sm">
        {t("desk.actions.assignTo")}
        <Picker
          disabled={disabled}
          value={assignee}
          onChange={setAssignee}
          options={[
            { value: "", label: t("desk.actions.assignQueue") },
            ...members.map((member) => ({ value: member.user_id, label: memberName(member.user_id) })),
          ]}
        />
      </label>
      <button type="button" className={secondaryClass} disabled={disabled} onClick={() => onAssign(assignee || null)}>
        {t("desk.actions.assign")}
      </button>
    </div>
  );
}
