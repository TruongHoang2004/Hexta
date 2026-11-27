"use client";

import { useState } from "react";
import { Notice, buttonClass, secondaryClass } from "@/components/primitives";
import { useWorkspace } from "@/components/layout/workspace-context";
import { useT } from "@/lib/i18n/provider";
import { useScriptAccess } from "../../hooks/use-script-access";
import { useScriptList } from "../../hooks/use-script-list";
import { ScriptDetailPanel } from "./script-detail-panel";
import { ScriptDraftEditor } from "./script-draft-editor";
import { ScriptList } from "./script-list";

type Selection = { kind: "none" } | { kind: "new" } | { kind: "script"; id: string };

/** S-15 Scripts: list, version editor and history. Actions follow the caller's script permissions. */
export function ScriptsScreen() {
  const { tenant } = useWorkspace();
  const t = useT();
  const tenantID = tenant?.id ?? "";
  const [revision, setRevision] = useState(0);
  const [selection, setSelection] = useState<Selection>({ kind: "none" });
  const [success, setSuccess] = useState("");
  const access = useScriptAccess(tenantID, revision);
  const list = useScriptList(tenantID, revision);
  const refresh = () => setRevision((v) => v + 1);
  const select = (next: Selection) => {
    setSelection(next);
    setSuccess("");
  };

  return (
    <div className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">{t("scripts.title")}</h1>
          <p className="mt-1 text-sm text-muted">{t("scripts.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <button type="button" className={secondaryClass} onClick={refresh}>
            {t("scripts.refresh")}
          </button>
          {access.can("script.create") && (
            <button type="button" className={buttonClass} onClick={() => select({ kind: "new" })}>
              {t("scripts.create")}
            </button>
          )}
        </div>
      </div>
      <Notice error={list.error || access.error} success={success} />
      {access.readOnly && <p className="text-sm text-muted">{t("scripts.readOnly")}</p>}
      <div className="flex flex-wrap items-start gap-5">
        <ScriptList
          list={list}
          selectedID={selection.kind === "script" ? selection.id : undefined}
          onSelect={(id) => select({ kind: "script", id })}
        />
        <div className="min-w-0 flex-[999_1_480px]">
          {selection.kind === "new" && access.can("script.create") ? (
            <ScriptDraftEditor
              key={`new:${revision}`}
              script={null}
              canActivate={false}
              busy={false}
              onReload={refresh}
              onActivate={() => undefined}
              onSaved={(created) => {
                setSelection({ kind: "script", id: created.id });
                setSuccess(t("scripts.detail.created", { name: created.name }));
                refresh();
              }}
            />
          ) : selection.kind === "script" ? (
            <ScriptDetailPanel key={selection.id} scriptID={selection.id} can={access.can} onChanged={refresh} />
          ) : (
            <p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted">
              {t("scripts.select")}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
