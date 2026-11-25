"use client";

import { useState } from "react";
import { ConflictBanner, Notice } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useMemberName } from "@/lib/members";
import { useTenantMembers } from "@/modules/conversations";
import { useScriptDetail } from "../../hooks/use-script-detail";
import type { Script, ScriptPermission } from "../../types";
import { ScriptDetailHeader } from "./script-detail-header";
import { ScriptDraftEditor } from "./script-draft-editor";
import { ScriptHistory } from "./script-history";

interface Props {
  scriptID: string;
  can: (permission: ScriptPermission) => boolean;
  /** Called after every command attempt so the list and the caller's permissions refresh. */
  onChanged: (script: Script) => void;
}

/** Selected script: status and actions, draft editor (with script.edit) and version history. */
export function ScriptDetailPanel({ scriptID, can, onChanged }: Props) {
  const t = useT();
  const { tenant } = useWorkspace();
  const detail = useScriptDetail(tenant?.id ?? "", scriptID, onChanged);
  const { members } = useTenantMembers(tenant?.id ?? "");
  const nameOf = useMemberName(members);
  const [success, setSuccess] = useState("");
  const script = detail.script;

  if (!script) {
    return detail.loading ? <p role="status">{t("scripts.detail.loading")}</p> : <Notice error={detail.error} />;
  }
  const archived = script.status === "archived";
  const after = (message: string) => (ok: boolean) => {
    setSuccess(ok ? message : "");
    return ok;
  };
  const activate = (versionID: string) => {
    const number = script.versions.find((v) => v.id === versionID)?.number ?? 0;
    void detail.activate(versionID).then(after(t("scripts.detail.activated", { name: script.name, number })));
  };
  return (
    <div className="space-y-4">
      <ScriptDetailHeader
        script={script}
        canActivate={can("script.activate")}
        canArchive={can("script.archive")}
        busy={detail.busy}
        onDeactivate={() => void detail.deactivate().then(after(t("scripts.detail.deactivated", { name: script.name })))}
        onArchive={() => detail.archive().then(after(t("scripts.detail.archived", { name: script.name })))}
      />
      {detail.conflict ? (
        <ConflictBanner message={t("scripts.detail.conflict")} actionLabel={t("scripts.detail.reload")} onReload={detail.reload} />
      ) : (
        <Notice error={detail.error} success={success} />
      )}
      {can("script.edit") && !archived && (
        <ScriptDraftEditor
          key={`${script.id}:${script.version}`}
          script={script}
          canActivate={can("script.activate")}
          busy={detail.busy}
          onReload={detail.reload}
          onActivate={activate}
          onSaved={(next) => {
            detail.accept(next);
            setSuccess(t("scripts.detail.saved", { number: next.versions.find((v) => v.draft)?.number ?? 0 }));
          }}
        />
      )}
      <ScriptHistory
        script={script}
        nameOf={nameOf}
        busy={detail.busy}
        onActivate={can("script.activate") && !archived ? activate : undefined}
      />
    </div>
  );
}
