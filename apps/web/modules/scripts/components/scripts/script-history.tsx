"use client";

import { StatusChip, secondaryClass } from "@/components/primitives";
import { useLocale, useT } from "@/lib/i18n/provider";
import type { Script } from "../../types";
import { formatTriggers } from "../../utils";

interface Props {
  script: Script;
  /** Names a member ID for "edited by". */
  nameOf: (id: string) => string;
  /** Present only when the caller may activate and the script is not archived. */
  onActivate?: (versionID: string) => void;
  busy: boolean;
}

/** Version history, newest first. Published versions are read-only; an earlier one can be re-activated. */
export function ScriptHistory({ script, nameOf, onActivate, busy }: Props) {
  const t = useT();
  const locale = useLocale();
  const date = (value: string | undefined) => (value ? new Date(value).toLocaleString(locale) : "");
  return (
    <section className="space-y-3 rounded-xl border border-border bg-surface p-5">
      <h2 className="font-semibold">{t("scripts.history.title")}</h2>
      <ol className="divide-y divide-border">
        {script.versions.map((version) => {
          const active = version.id === script.active_version_id;
          return (
            <li key={version.id} className="space-y-1 py-3 text-sm">
              <div className="flex flex-wrap items-center gap-2">
                <span className="font-medium">{t("scripts.version", { number: version.number })}</span>
                {version.draft && <StatusChip tone="warning">{t("scripts.history.draft")}</StatusChip>}
                {active && <StatusChip tone="success">{t("scripts.history.active")}</StatusChip>}
                <span className="text-muted">
                  {version.draft
                    ? t("scripts.history.editedBy", { name: nameOf(version.updated_by), date: date(version.updated_at) })
                    : t("scripts.history.publishedBy", { name: nameOf(version.published_by ?? ""), date: date(version.published_at) })}
                </span>
                {onActivate && !version.draft && !active && (
                  <button type="button" className={`${secondaryClass} ml-auto`} disabled={busy} onClick={() => onActivate(version.id)}>
                    {t("scripts.history.activate", { number: version.number })}
                  </button>
                )}
              </div>
              <p className="text-muted">{formatTriggers(version.triggers)}</p>
              <p className="whitespace-pre-wrap break-words">{version.answer}</p>
            </li>
          );
        })}
      </ol>
      <p className="text-xs text-muted">{t("scripts.history.note")}</p>
    </section>
  );
}
