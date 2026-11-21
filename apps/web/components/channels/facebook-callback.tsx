"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import {
  channels,
  type OAuthPages,
} from "@/lib/services/channels";
import { buttonClass, Notice } from "@/components/sales/primitives";

export function FacebookCallback() {
  const params = useSearchParams();
  const router = useRouter();
  const t = useT();
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
  }, [params, state, errorText]);
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
      <Notice error={t("oauth.notStarted")} />
    );
  }
  return (
    <div className="space-y-5">
      <h1 className="text-2xl font-semibold">{t("oauth.title")}</h1>
      <Notice error={error} />
      {!result && !error && <p role="status">{t("oauth.checking")}</p>}
      {error && (
        <Link href="/tenants" className="text-sm text-primary underline">
          {t("oauth.back")}
        </Link>
      )}
      {result && (
        <form
          onSubmit={connect}
          className="space-y-4 rounded-xl border border-border bg-surface p-5"
        >
          <p className="text-sm text-muted">
            {t("oauth.intro", { time: new Date(result.expires_at).toLocaleTimeString() })}
          </p>
          {!result.pages.length ? (
            <p className="text-sm">
              {t("oauth.noPages")}
            </p>
          ) : (
            <fieldset className="space-y-2">
              <legend className="sr-only">{t("oauth.pagesLegend")}</legend>
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
                        {t(`oauth.availability.${page.availability}`)}
                      </span>
                    </span>
                  </label>
                );
              })}
            </fieldset>
          )}
          <div className="flex flex-wrap gap-2">
            <button className={buttonClass} disabled={busy || !selected.length}>
              {busy ? t("oauth.connecting") : t("oauth.connect", { count: selected.length })}
            </button>
            <Link
              href={`/tenants/${result.tenant_id}/settings/channels`}
              className="rounded-lg border border-border px-3 py-2 text-sm hover:bg-surface"
            >
              {t("oauth.cancel")}
            </Link>
          </div>
        </form>
      )}
    </div>
  );
}
