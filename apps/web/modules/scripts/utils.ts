import type { ChipTone } from "@/components/primitives";
import type { MessageKey } from "@/lib/i18n/translate";
import type { Script, ScriptPermission, ScriptStatus, ScriptVersion } from "./types";

/** Splits the comma or newline separated "when customers ask about" field into phrases. */
export function parseTriggers(text: string): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const raw of text.split(/[,\n]/)) {
    const phrase = raw.trim().replace(/\s+/g, " ");
    if (phrase && !seen.has(phrase.toLowerCase())) {
      seen.add(phrase.toLowerCase());
      result.push(phrase);
    }
  }
  return result;
}

export function formatTriggers(triggers: string[]): string {
  return triggers.join(", ");
}

export function statusTone(status: ScriptStatus): ChipTone {
  if (status === "active") return "success";
  if (status === "archived") return "neutral";
  return "warning";
}

/** The open draft, if any. */
export function openDraft(script: Script): ScriptVersion | undefined {
  return script.versions.find((v) => v.draft);
}

export function activeVersion(script: Script): ScriptVersion | undefined {
  return script.versions.find((v) => v.id === script.active_version_id);
}

/** Number the next saved draft will have: the open draft's, or one past the newest version. */
export function draftNumber(script: Script): number {
  return openDraft(script)?.number ?? (script.versions[0]?.number ?? 0) + 1;
}

/** Content the editor starts from: the open draft, else the active version, else the newest version. */
export function editorSource(script: Script): ScriptVersion | undefined {
  return openDraft(script) ?? activeVersion(script) ?? script.versions[0];
}

/** Short label key of a permission, e.g. "script.edit" → "Sửa" / "Edit". */
export function permissionLabelKey(permission: ScriptPermission): MessageKey {
  const keys: Record<ScriptPermission, MessageKey> = {
    "script.create": "scriptPermissions.permissions.create",
    "script.edit": "scriptPermissions.permissions.edit",
    "script.activate": "scriptPermissions.permissions.activate",
    "script.archive": "scriptPermissions.permissions.archive",
  };
  return keys[permission];
}
