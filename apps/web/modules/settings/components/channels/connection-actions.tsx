"use client";

import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConfirmedAction, ConnectionAction } from "../../types";

interface Props {
  /** Actions the API allows for this connection in its current state. */
  actions: ConnectionAction[];
  /** Any mutation is running. */
  disabled: boolean;
  /** This row is the one being changed. */
  saving: boolean;
  onAct: (action: ConfirmedAction) => void;
  /** Type-specific controls shown first (reconnect, rotate token). */
  children?: React.ReactNode;
}

/** Owner actions shared by every connection type, rendered from the API's allowed actions. */
export function ConnectionActions({ actions, disabled, saving, onAct, children }: Props) {
  const t = useT();
  return (
    <>
      {children}
      {actions.includes("pause") && (
        <button className={secondaryClass} disabled={disabled} onClick={() => onAct("pause")}>
          {t("channels.pause")}
        </button>
      )}
      {actions.includes("resume") && (
        <button className={secondaryClass} disabled={disabled} onClick={() => onAct("resume")}>
          {t("channels.resume")}
        </button>
      )}
      {actions.includes("disconnect") && (
        <button className={`${secondaryClass} text-error`} disabled={disabled} onClick={() => onAct("disconnect")}>
          {saving ? t("channels.saving") : t("channels.disconnect")}
        </button>
      )}
    </>
  );
}
