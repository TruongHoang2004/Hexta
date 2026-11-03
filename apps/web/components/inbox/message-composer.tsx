import { Send } from "lucide-react";
import type { Mode } from "./types";
interface MessageComposerProps {
  closed: boolean;
  mode: Mode;
  draft: string;
  onModeChange: (mode: Mode) => void;
  onDraftChange: (draft: string) => void;
  onAddPreview: () => void;
}
export function MessageComposer({
  closed,
  mode,
  draft,
  onModeChange,
  onDraftChange,
  onAddPreview,
}: MessageComposerProps) {
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        onAddPreview();
      }}
      className="m-3 rounded-xl border border-border p-3"
    >
      <div className="flex gap-4 border-b border-border pb-2">
        {(["reply", "note"] as const).map((value) => (
          <button
            type="button"
            key={value}
            aria-pressed={mode === value}
            onClick={() => onModeChange(value)}
            className={`text-xs ${mode === value ? "font-semibold text-violet-600" : "text-muted"}`}
          >
            {value === "reply" ? "Reply" : "Internal note"}
          </button>
        ))}
      </div>
      <textarea
        aria-label={mode === "reply" ? "Reply draft" : "Internal note draft"}
        value={draft}
        onChange={(event) => onDraftChange(event.target.value)}
        placeholder={
          mode === "reply" ? "Preview a reply…" : "Write a team note…"
        }
        className="min-h-24 w-full resize-y bg-transparent py-3 text-sm outline-violet-500"
      />
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-[10px] text-muted">
          {closed && mode === "reply"
            ? "Closed conversation · Notes available"
            : "Preview only · Not sent to customer"}
        </span>
        <button
          disabled={!draft.trim() || (closed && mode === "reply")}
          className="flex items-center gap-2 rounded-lg bg-violet-600 px-3 py-2 text-xs font-medium text-white disabled:cursor-not-allowed disabled:opacity-40"
        >
          <Send className="size-3" />
          Add preview
        </button>
      </div>
    </form>
  );
}
