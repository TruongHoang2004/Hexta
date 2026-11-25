"use client";

import { useProposalActions } from "../../hooks/use-proposal-actions";
import type { MessagingTurn, ReplyProposal } from "../../services/messaging";
import type { ReplyMode } from "../../services/reply-mode";
import { proposalAccess } from "../../utils";
import { ProposalCard } from "./proposal-card";

/** The live bot proposal of the selected conversation, wired to its send/discard/regenerate actions. */
export function ThreadProposal({
  tenantID,
  turn,
  proposal,
  mode,
  userID,
  owner,
  assigneeName,
  onChanged,
  onPoll,
}: {
  tenantID: string;
  turn: MessagingTurn;
  proposal: ReplyProposal;
  mode: ReplyMode | null;
  userID: string | undefined;
  owner: boolean;
  assigneeName: string;
  onChanged: () => void;
  onPoll: () => void;
}) {
  const actions = useProposalActions({ tenantID, turnID: turn.id, proposal, onChanged, onPoll });
  return (
    <ProposalCard
      proposal={proposal}
      mode={mode}
      access={proposalAccess(turn, userID, owner)}
      assigneeName={assigneeName}
      busy={actions.busy}
      error={actions.error}
      uncertain={actions.uncertain}
      onSend={(text) => void actions.send(text)}
      onDiscard={() => void actions.discard()}
      onRegenerate={() => void actions.regenerate()}
    />
  );
}
