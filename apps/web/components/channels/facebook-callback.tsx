"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { errorText } from "@/lib/services/sales";
import {
  channels,
  type OAuthPages,
  type SelectablePage,
} from "@/lib/services/channels";
import { buttonClass, Notice } from "@/components/sales/primitives";

const availabilityText: Record<SelectablePage["availability"], string> = {
  available: "Available",
  already_connected_this_tenant: "Connected here — select to reconnect",
  owned_by_other_tenant: "Connected to another workspace",
};

export function FacebookCallback() {
  const params = useSearchParams();
  const router = useRouter();
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
    channels
      .completeFacebook({
        code: params.get("code") ?? "",
        state,
        error: params.get("error") ?? "",
      })
      .then((pages) => {
        setResult(pages);
        setSelected(
          pages.pages
            .filter((p) => p.availability === "available")
            .map((p) => p.page_id),
        );
      })
      .catch((err) => setError(errorText(err)));
  }, [params, state]);
  async function connect(event: React.FormEvent) {
    event.preventDefault();
    if (!result || busy || !selected.length) return;
    setBusy(true);
    setError("");
    try {
      await channels.connect(result.tenant_id, result.oauth_session_id, selected);
      router.push(`/tenants/${result.tenant_id}/settings/channels`);
    } catch (err) {
      setError(errorText(err));
      setBusy(false);
    }
  }
  if (!state) {
    return (
      <Notice error="This page is opened by Facebook after you start a connection from Channel settings." />
    );
  }
  return (
    <div className="space-y-5">
      <h1 className="text-3xl font-semibold">Choose Facebook Pages</h1>
      <Notice error={error} />
      {!result && !error && <p role="status">Checking Facebook authorization…</p>}
      {error && (
        <Link href="/tenants" className="text-sm text-primary underline">
          Back to workspaces
        </Link>
      )}
      {result && (
        <form
          onSubmit={connect}
          className="space-y-4 rounded-xl border border-border bg-surface p-5"
        >
          <p className="text-sm text-muted">
            Select the Pages this workspace should receive and answer messages
            for. This selection expires at{" "}
            {new Date(result.expires_at).toLocaleTimeString()}.
          </p>
          {!result.pages.length ? (
            <p className="text-sm">
              Facebook returned no Pages you manage. Check the Pages granted in
              the Facebook dialog and try again.
            </p>
          ) : (
            <fieldset className="space-y-2">
              <legend className="sr-only">Facebook Pages</legend>
              {result.pages.map((page) => {
                const taken = page.availability === "owned_by_other_tenant";
                return (
                  <label
                    key={page.page_id}
                    className={`flex items-start gap-3 rounded-lg border border-border p-3 ${taken ? "opacity-60" : ""}`}
                  >
                    <input
                      type="checkbox"
                      className="mt-1"
                      disabled={taken || busy}
                      checked={selected.includes(page.page_id)}
                      onChange={(e) =>
                        setSelected((current) =>
                          e.target.checked
                            ? [...current, page.page_id]
                            : current.filter((id) => id !== page.page_id),
                        )
                      }
                    />
                    <span>
                      <span className="block font-medium">{page.page_name}</span>
                      <span className="block text-xs text-muted">
                        {availabilityText[page.availability]}
                      </span>
                    </span>
                  </label>
                );
              })}
            </fieldset>
          )}
          <div className="flex flex-wrap gap-2">
            <button className={buttonClass} disabled={busy || !selected.length}>
              {busy ? "Connecting…" : `Connect ${selected.length || ""} Page${selected.length === 1 ? "" : "s"}`}
            </button>
            <Link
              href={`/tenants/${result.tenant_id}/settings/channels`}
              className="rounded-lg border border-border px-3 py-2 text-sm hover:bg-surface"
            >
              Cancel
            </Link>
          </div>
        </form>
      )}
    </div>
  );
}
