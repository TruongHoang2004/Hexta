"use client";

import { StatusChip } from "@/components/primitives";
import { useT } from "@/lib/i18n/provider";
import type { ScriptStatus } from "../../types";
import { statusTone } from "../../utils";

export function ScriptStatusChip({ status }: { status: ScriptStatus }) {
  const t = useT();
  return <StatusChip tone={statusTone(status)}>{t(`scripts.status.${status}`)}</StatusChip>;
}
