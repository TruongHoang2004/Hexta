"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { Field, Picker, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import type { WorkCommand, WorkCommandFields, WorkItem } from "../../types";

interface Props {
  item: WorkItem;
  members: TenantMember[];
  busy: boolean;
  onCommand: (action: WorkCommand, fields?: WorkCommandFields) => Promise<boolean>;
}

/** Owner control: move responsibility to another member or back to the queue. */
export function WorkReassign({ item, members, busy, onCommand }: Props) {
  const t = useT();
  const person = useMemberName(members);
  const [responsible, setResponsible] = useState(item.responsible_id || "");
  return (
    <fieldset disabled={busy} className="space-y-3 border-t border-border pt-4">
      <legend className="sr-only">{t("work.owner.title")}</legend>
      <Field title={t("work.owner.reassignLabel")}>
        <Picker
          value={responsible}
          onChange={setResponsible}
          options={[
            { value: "", label: t("work.owner.queue") },
            ...(item.responsible_id && !members.some((row) => row.user_id === item.responsible_id)
              ? [{ value: item.responsible_id, label: t("work.formerMember") }]
              : []),
            ...members.map((row) => ({ value: row.user_id, label: person(row.user_id) })),
          ]}
        />
      </Field>
      <p className="text-xs text-muted">{t("work.owner.reassignNote")}</p>
      <button
        type="button"
        className={secondaryClass}
        onClick={() => void onCommand("assign", { responsible_id: responsible || null })}
      >
        {t("work.actions.reassign")}
      </button>
    </fieldset>
  );
}
