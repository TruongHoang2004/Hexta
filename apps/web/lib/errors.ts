"use client";

import { useCallback } from "react";
import { SDKApiError } from "@hexta/sdk";
import { useT } from "@/lib/i18n/provider";
import type { MessageKey } from "@/lib/i18n/translate";

export interface DescribedError {
  /** Localized message key; null means the server detail is shown as-is. */
  key: MessageKey | null;
  detail: string;
  status?: number;
  traceId?: string;
}

function traceIdOf(error: SDKApiError): string | undefined {
  const body = error.response?.data;
  if (typeof body !== "object" || body === null) return undefined;
  const value = (body as Record<string, unknown>).trace_id;
  return typeof value === "string" && value ? value : undefined;
}

/** Maps an API/client failure to a safe, localizable description. */
export function describeError(error: unknown): DescribedError {
  if (error instanceof SDKApiError) {
    const traceId = traceIdOf(error);
    const base = { detail: error.message, status: error.status, traceId };
    if (error.status === 409) return { ...base, key: "errors.conflict" };
    if (error.status === 401 || error.status === 403) return { ...base, key: "errors.forbidden" };
    if (error.status === 404) return { ...base, key: "errors.notFound" };
    // Validation and business-rule rejections carry a specific server detail.
    if (error.status >= 400 && error.status < 500 && error.message) return { ...base, key: null };
    return { ...base, key: "errors.generic" };
  }
  if (error instanceof Error && /network/i.test(error.message)) {
    return { key: "errors.network", detail: error.message };
  }
  return { key: "errors.generic", detail: error instanceof Error ? error.message : "" };
}

export function isConflict(error: unknown): boolean {
  return error instanceof SDKApiError && error.status === 409;
}

/** Returns a formatter producing localized error text with the support code when available. */
export function useErrorText(): (error: unknown) => string {
  const t = useT();
  return useCallback(
    (error: unknown) => {
      const described = describeError(error);
      const text = described.key ? t(described.key) : described.detail;
      return described.traceId
        ? `${text} ${t("errors.supportCode", { code: described.traceId })}`
        : text;
    },
    [t],
  );
}
