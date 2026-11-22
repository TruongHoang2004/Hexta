"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useErrorText } from "@/lib/errors";
import { facebook } from "../services/channels";
import type { OAuthPages } from "../types";

/** Exchanges the OAuth callback once, then connects the Pages the owner selects. */
export function useFacebookOAuth() {
  const params = useSearchParams();
  const router = useRouter();
  const errorText = useErrorText();
  const started = useRef("");
  const [result, setResult] = useState<OAuthPages | null>(null);
  const [selected, setSelected] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const state = params.get("state") ?? "";

  useEffect(() => {
    // The state is single-use; never exchange it twice (e.g. on a development double effect).
    if (!state || started.current === state) return;
    started.current = state;
    facebook
      .complete({ code: params.get("code") ?? "", state, error: params.get("error") ?? "" })
      .then((pages) => {
        setResult(pages);
        setSelected(pages.pages.filter((p) => p.availability === "available").map((p) => p.page_id));
      })
      .catch((err) => setError(errorText(err)));
  }, [params, state, errorText]);

  function toggle(pageID: string, checked: boolean) {
    setSelected((current) => (checked ? [...current, pageID] : current.filter((id) => id !== pageID)));
  }

  async function connect() {
    if (!result || busy || !selected.length) return;
    setBusy(true);
    setError("");
    try {
      await facebook.connectPages(result.tenant_id, result.oauth_session_id, selected);
      router.push(`/tenants/${result.tenant_id}/settings/channels`);
    } catch (err) {
      setError(errorText(err));
      setBusy(false);
    }
  }

  return { started: !!state, result, selected, toggle, error, busy, connect };
}
