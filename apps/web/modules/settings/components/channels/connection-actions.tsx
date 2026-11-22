"use client";

import { secondaryClass } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ConnectionAction, ConnectionStatus } from "../../types";

interface Props {
  status: ConnectionStatus;
  /** Any mutation is running. */
  disabled: boolean;
  /** This row is the one being changed. */
  saving: boolean;
  onAct: (action: ConnectionAction) => void;
  /** Platform-specific actions shown first (reconnect, rotate token). */
  children?: React.ReactNode;
}

/** Owner actions shared by every channel: pause, resume, disconnect. */
export function ConnectionActions({ status, disabled, saving, onAct, children }: Props) {
  const t = useT();
  return (
    <>
      {children}
      {(status === "active" || status === "auth_expired") && (
        <button className={secondaryClass} disabled={disabled} onClick={() => onAct("pause")}>
          {t("channels.pause")}
        </button>
      )}
      {status === "paused" && (
        <button className={secondaryClass} disabled={disabled} onClick={() => onAct("resume")}>
          {t("channels.resume")}
        </button>
      )}
      <button className={`${secondaryClass} text-error`} disabled={disabled} onClick={() => onAct("disconnect")}>
        {saving ? t("channels.saving") : t("channels.disconnect")}
      </button>
    </>
  );
}
