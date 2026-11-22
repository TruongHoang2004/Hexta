"use client";

import { Textarea } from "@hexta/ui";
import { buttonClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import { platformLabel, type ChannelPlatform } from "@/modules/settings";

/** Reply box. Shows why it is locked when the viewer cannot reply. */
export function ReplyComposer({
  platform,
  body,
  onBodyChange,
  canReply,
  readOnlyReason,
  busy,
  uncertain,
  onSubmit,
}: {
  platform: ChannelPlatform | null;
  body: string;
  onBodyChange: (value: string) => void;
  canReply: boolean;
  readOnlyReason: string;
  busy: boolean;
  uncertain: boolean;
  onSubmit: () => void;
}) {
  const t = useT();
  const channel = platform ? platformLabel[platform] : "";
  const placeholder = !canReply
    ? t("desk.composer.placeholderLocked")
    : platform
      ? t("desk.composer.placeholderChannel", { channel })
      : t("desk.composer.placeholderSimulated");
  const submitLabel = uncertain
    ? platform
      ? t("desk.composer.retrySend")
      : t("desk.composer.retrySave")
    : platform
      ? t("desk.composer.send", { channel })
      : t("desk.composer.saveSimulated");
  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit();
      }}
      className="space-y-2 border-t border-border p-4"
    >
      {readOnlyReason && (
        <p role="status" className="rounded-lg border border-border bg-background p-2 text-sm">
          {readOnlyReason}
        </p>
      )}
      <label className="block text-sm">
        <span className="sr-only">{t("desk.composer.label")}</span>
        <Textarea
          rows={3}
          maxLength={platform ? 2000 : 4000}
          value={body}
          disabled={busy || uncertain || !canReply}
          onChange={(e) => onBodyChange(e.target.value)}
          placeholder={placeholder}
        />
      </label>
      <div className="flex justify-end">
        <button className={buttonClass} disabled={busy || !canReply || !body.trim()}>
          {submitLabel}
        </button>
      </div>
    </form>
  );
}
