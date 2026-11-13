import { UserRound } from "lucide-react";
import type { Conversation } from "./types";

export function CustomerDetails({
  conversation: item,
}: {
  conversation: Conversation;
}) {
  return (
    <div className="space-y-5 pt-5">
      <div className="flex items-center gap-3">
        <UserRound className="size-8 text-violet-400" />
        <div>
          <p className="text-sm font-semibold">{item.name}</p>
          <p className="mt-1 text-xs text-muted">{item.channel} customer</p>
        </div>
      </div>
      <dl className="space-y-3 border-y border-border py-5 text-xs">
        {[
          ["Lifecycle", item.closed ? "Closed" : "Open"],
          ["Assigned to", item.assignee ?? "Unassigned"],
          ["Control", item.assignee ? "Sales" : "System"],
          ["Escalation", item.needsHuman ? "Needs human" : "None"],
          ["Email", "Not provided"],
          ["Source ID", item.id],
        ].map(([label, value]) => (
          <div key={label} className="flex justify-between gap-3">
            <dt className="text-muted">{label}</dt>
            <dd className="text-right">{value}</dd>
          </div>
        ))}
      </dl>
      <div>
        <h3 className="text-xs font-semibold">TEAM CONTEXT</h3>
        <p className="mt-3 text-xs leading-relaxed text-muted">
          {item.needsHuman
            ? "A team member needs to review this conversation."
            : "Confirm customer context and product details before replying."}
        </p>
      </div>
    </div>
  );
}
