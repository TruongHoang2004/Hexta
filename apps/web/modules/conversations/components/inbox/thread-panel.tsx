"use client";

import { useState } from "react";
import type { TenantMember } from "@hexta/sdk";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import { platformLabel, type ChannelConnection } from "@/modules/settings";
import { useAuthStore } from "@/store/useAuthStore";
import { useReplySender } from "../../hooks/use-reply-sender";
import { useThread } from "../../hooks/use-thread";
import type { HandoffReason, ThreadCommand } from "../../types";
import { threadAccess, threadPlatform } from "../../utils";
import { AssignControl } from "./assign-control";
import { CustomerPanel } from "./customer-panel";
import { HandoffDialog } from "./handoff-dialog";
import { MessageHistory } from "./message-history";
import { ReplyComposer } from "./reply-composer";
import { ThreadHeader } from "./thread-header";

/** The selected conversation: header, history, reply box and the customer panel. */
export function ThreadPanel({
  tenantID,
  threadID,
  owner,
  members,
  connections,
  externalRefresh,
  live,
  onChanged,
}: {
  tenantID: string;
  threadID: string;
  owner: boolean;
  members: TenantMember[];
  // null while unknown; the API still rejects sends on inactive connections.
  connections: ChannelConnection[] | null;
  externalRefresh: number;
  live: number;
  onChanged: () => void;
}) {
  const t = useT();
  const memberName = useMemberName(members);
  const userID = useAuthStore((state) => state.user?.id);
  const thread = useThread({ tenantID, threadID, externalRefresh, live, onChanged });
  const { detail } = thread;
  const platform = threadPlatform(detail);
  const sender = useReplySender({
    tenantID,
    realChannel: !!platform,
    onSent: () => {
      thread.reload();
      onChanged();
    },
  });
  const [handoffOpen, setHandoffOpen] = useState(false);
  const turn = detail?.conversation.turn;
  const access = threadAccess({
    t,
    detail,
    connections,
    userID,
    owner,
    channelName: platform ? platformLabel[platform] : "",
    assigneeName: turn?.assignee_id ? memberName(turn.assignee_id) : "",
  });
  const busy = thread.busy || sender.busy;
  const commandsDisabled = busy || sender.uncertain;
  const error = thread.error || sender.error;

  async function run(action: ThreadCommand, assignee: string | null = null, reason: HandoffReason = "staff_handoff") {
    const done = await thread.command(action, assignee, reason);
    if (done) setHandoffOpen(false);
  }

  return (
    <>
      <section
        className="flex min-w-0 flex-[999_1_420px] flex-col rounded-xl border border-border bg-surface"
        aria-label={t("desk.detailLabel")}
      >
        <ThreadHeader
          detail={detail}
          access={access}
          disabled={commandsDisabled}
          loading={thread.loading || busy}
          onClaim={() => void run("claim")}
          onHandoff={() => setHandoffOpen(true)}
          onReturnToBot={() => void run("return-to-bot")}
          onRefresh={thread.reload}
        />
        <div className="flex flex-1 flex-col gap-3 bg-background/60 p-4">
          {error && (
            <p role="alert" className="text-sm text-error">
              {error} {t("desk.composer.errorSuffix")}
            </p>
          )}
          {sender.uncertain && (
            <p role="status" className="rounded-lg border border-warning bg-warning/15 p-2 text-sm">
              {t("desk.composer.uncertain")}
            </p>
          )}
          {thread.loading ? (
            <p role="status">{t("desk.loadingMessages")}</p>
          ) : (
            detail && (
              <MessageHistory
                detail={detail}
                platform={platform}
                busy={busy}
                onLoadOlder={() => void thread.loadOlder()}
              />
            )
          )}
        </div>
        {detail && (
          <ReplyComposer
            platform={platform}
            body={sender.body}
            onBodyChange={sender.setBody}
            canReply={access.canReply}
            readOnlyReason={access.readOnlyReason}
            busy={busy}
            uncertain={sender.uncertain}
            onSubmit={() => {
              if (turn && access.canReply) void sender.send(turn);
            }}
          />
        )}
      </section>
      <CustomerPanel tenantID={tenantID} detail={detail}>
        {owner && turn && (
          <AssignControl
            key={`${turn.id}:${turn.version}`}
            members={members}
            currentAssignee={turn.assignee_id ?? ""}
            disabled={commandsDisabled}
            onAssign={(assignee) => void run("assign", assignee)}
          />
        )}
      </CustomerPanel>
      <HandoffDialog
        open={handoffOpen}
        busy={busy}
        onCancel={() => setHandoffOpen(false)}
        onConfirm={(reason) => void run("handoff", null, reason)}
      />
    </>
  );
}
