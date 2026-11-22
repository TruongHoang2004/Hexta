"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { StatusChip } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import type { WorkCommand, WorkCommandFields, WorkItem } from "../../types";
import { stateTone } from "../../utils";
import { WorkActions } from "./work-actions";
import { WorkHandoffDialog } from "./work-handoff-dialog";
import { WorkHistory } from "./work-history";
import { WorkReassign } from "./work-reassign";
import { WorkStepTrack } from "./work-step-track";

interface Props {
  item: WorkItem;
  members: TenantMember[];
  busy: boolean;
  onCommand: (action: WorkCommand, fields?: WorkCommandFields) => Promise<boolean>;
  onRefresh: () => void;
}

/** One work item: progress, who holds it, available actions and its audit history. */
export function WorkDetail({ item, members, busy, onCommand, onRefresh }: Props) {
  const { isOwner } = useWorkspace();
  const t = useT();
  const person = useMemberName(members);
  const [handoffOpen, setHandoffOpen] = useState(false);
  return (
    <section className="space-y-4 rounded-xl border border-border bg-surface p-5">
      <div className="flex flex-wrap justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{item.workflow.name}</h2>
          <p className="text-xs text-muted">
            {t("work.workflowVersion", { number: item.workflow.number, version: item.version })}
          </p>
          <p className="break-all text-xs text-muted">{t("work.order", { id: item.order_id })}</p>
        </div>
        <StatusChip tone={stateTone[item.state]}>{t(`work.states.${item.state}`)}</StatusChip>
      </div>
      <WorkStepTrack item={item} />
      <p className="break-all text-sm">
        {t("work.responsible")}: <strong>{person(item.responsible_id)}</strong>
      </p>
      {item.state === "waiting" && (
        <p className="break-all rounded-lg border border-warning bg-warning/15 p-3 text-sm">
          {t("work.waitingFor", { name: person(item.recipient_id), responsible: person(item.responsible_id) })}
        </p>
      )}
      <WorkActions
        item={item}
        busy={busy}
        onHandoff={() => setHandoffOpen(true)}
        onCommand={onCommand}
        onRefresh={onRefresh}
      />
      {isOwner && item.state !== "completed" && (
        <WorkReassign item={item} members={members} busy={busy} onCommand={onCommand} />
      )}
      <WorkHistory events={item.events} person={person} />
      <WorkHandoffDialog
        open={handoffOpen}
        item={item}
        members={members}
        busy={busy}
        onCancel={() => setHandoffOpen(false)}
        onConfirm={async (recipient) => {
          if (await onCommand("handoff", { recipient_id: recipient })) setHandoffOpen(false);
        }}
      />
    </section>
  );
}
