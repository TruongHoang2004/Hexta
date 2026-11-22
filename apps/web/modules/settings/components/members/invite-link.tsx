"use client";

import { useState } from "react";
import { inputClass, secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";

/** Read-only invitation link with copy; key it by link so "copied" resets for a new one. */
export function InviteLink({ link }: { link: string }) {
  const t = useT();
  const [copied, setCopied] = useState(false);
  async function copy() {
    try {
      await navigator.clipboard.writeText(link);
      setCopied(true);
    } catch {
      setCopied(false);
    }
  }
  return (
    <div className="space-y-2 rounded-lg border border-primary bg-background p-4">
      <label className="block text-sm font-medium">
        {t("members.linkLabel")}
        <span className="mt-2 flex gap-2">
          <input className={inputClass} readOnly value={link} onFocus={(e) => e.target.select()} />
          <button type="button" className={secondaryClass} onClick={() => void copy()}>
            {copied ? t("members.copied") : t("members.copy")}
          </button>
        </span>
      </label>
      <p className="text-sm text-muted">{t("members.linkNote")}</p>
    </div>
  );
}
