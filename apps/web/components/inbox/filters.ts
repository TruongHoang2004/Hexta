import type { Conversation, Queue } from "./types";

export function matchesQueue(item: Conversation, queue: Queue) {
  return (
    queue === "all" ||
    (queue === "unassigned" && !item.assignee && !item.closed) ||
    (queue === "attention" && item.needsHuman && !item.closed) ||
    (queue === "closed" && item.closed)
  );
}
