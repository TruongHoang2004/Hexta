"use client";

import { useState } from "react";
import type { Channel, Queue, Mode, Preview } from "./types";
import { conversations } from "./sample-data";
import { matchesQueue } from "./filters";

export function useInbox() {
  const [queue, setQueue] = useState<Queue>("all");
  const [channel, setChannel] = useState<Channel | "all">("all");
  const [query, setQuery] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(
    conversations[0].id,
  );
  const [mode, setMode] = useState<Mode>("reply");
  const [drafts, setDrafts] = useState<Record<string, string>>({});
  const [previews, setPreviews] = useState<Record<string, Preview[]>>({});
  const visible = conversations.filter(
    (item) =>
      matchesQueue(item, queue) &&
      (channel === "all" || item.channel === channel) &&
      `${item.name} ${item.message}`
        .toLowerCase()
        .includes(query.trim().toLowerCase()),
  );
  const selected = visible.find((item) => item.id === selectedId) ?? null;
  const draftKey = selected ? `${selected.id}:${mode}` : "";
  const draft = drafts[draftKey] ?? "";
  function addPreview() {
    if (!selected || !draft.trim() || (selected.closed && mode === "reply"))
      return;
    setPreviews((current) => ({
      ...current,
      [selected.id]: [
        ...(current[selected.id] ?? []),
        { text: draft.trim(), mode },
      ],
    }));
    setDrafts((current) => ({ ...current, [draftKey]: "" }));
  }
  return {
    queue,
    setQueue,
    channel,
    setChannel,
    query,
    setQuery,
    selected,
    setSelectedId,
    mode,
    setMode,
    visible,
    draft,
    previews,
    addPreview,
    setDraft: (value: string) =>
      setDrafts((current) => ({ ...current, [draftKey]: value })),
  };
}
