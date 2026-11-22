"use client";
import { BrandLogo } from "@/components/brand-logo";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { SDKApiError } from "@hexta/sdk";
import { sdk } from "@/lib/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { onboarding, type InvitationResult } from "@/lib/services/onboarding";
import { useErrorText } from "@/lib/errors";
import { useT } from "@/lib/i18n/provider";
import { buttonClass, secondaryClass, Notice, StatusChip } from "@/components/primitives";

type Problem = "" | "missing" | "wrongEmail" | "unavailable" | "other";

function problemOf(err: unknown): Problem {
  if (err instanceof SDKApiError) {
    if (err.status === 403) return "wrongEmail";
    if (err.status === 404 || err.status === 409) return "unavailable";
  }
  return "other";
}

export default function InvitationPage() {
  const router = useRouter();
  const t = useT();
  const errorText = useErrorText();
  const [token, setToken] = useState("");
  const [preview, setPreview] = useState<InvitationResult | null>(null);
  const [needsLogin, setNeedsLogin] = useState(false);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState<Problem>("");
  const [error, setError] = useState("");
  const [declined, setDeclined] = useState(false);
  useEffect(() => {
    let ignore = false;
    async function load() {
      try {
        const fragment = window.location.hash.slice(1);
        if (fragment) {
          sessionStorage.setItem("hexta.invitation", fragment);
          window.history.replaceState(null, "", window.location.pathname);
        }
        const value = sessionStorage.getItem("hexta.invitation") || "";
        if (!value) {
          if (!ignore) setProblem("missing");
          return;
        }
        setToken(value);
        if (!(await sdk.identity.getAccessToken())) {
          if (!ignore) setNeedsLogin(true);
          return;
        }
        const result = await onboarding.useInvitation(value, "preview");
        if (!ignore) setPreview(result);
      } catch (err) {
        if (!ignore) {
          setProblem(problemOf(err));
          setError(errorText(err));
        }
      } finally {
        if (!ignore) setLoading(false);
      }
    }
    void load();
    return () => {
      ignore = true;
    };
  }, [errorText]);
  async function switchAccount() {
    setBusy(true);
    try {
      await sdk.identity.logout();
      useAuthStore.getState().logout();
      sessionStorage.removeItem("hexta.workspace");
      router.push("/login");
    } catch {
      setProblem("other");
      setError(t("invitation.signOutFailed"));
    } finally {
      setBusy(false);
    }
  }
  async function act(action: "accept" | "decline") {
    setBusy(true);
    setProblem("");
    setError("");
    try {
      const result = await onboarding.useInvitation(token, action);
      setPreview(result);
      setDeclined(action === "decline");
      sessionStorage.removeItem("hexta.invitation");
    } catch (err) {
      setProblem(problemOf(err));
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  return (
    <main className="mx-auto max-w-xl space-y-5 px-5 py-14">
      <Link href="/tenants" className="inline-flex">
        <BrandLogo />
      </Link>
      <section className="space-y-5 rounded-xl border border-border bg-surface p-6">
        <h1 className="text-2xl font-semibold">{t("invitation.title")}</h1>
        {loading ? (
          <p role="status">{t("invitation.checking")}</p>
        ) : problem === "missing" ? (
          <p>{t("invitation.missingLink")}</p>
        ) : problem === "wrongEmail" ? (
          <div role="alert" className="space-y-3 rounded-lg border border-error bg-error/10 p-4">
            <h2 className="font-semibold">{t("invitation.wrongEmailTitle")}</h2>
            <p className="text-sm">{t("invitation.wrongEmail")}</p>
            <button type="button" className={buttonClass} disabled={busy} onClick={() => void switchAccount()}>
              {t("invitation.switchAccount")}
            </button>
          </div>
        ) : problem === "unavailable" ? (
          <div role="alert" className="space-y-2 rounded-lg border border-border bg-background p-4">
            <h2 className="font-semibold">{t("invitation.unavailableTitle")}</h2>
            <p className="text-sm">{t("invitation.unavailable")}</p>
          </div>
        ) : needsLogin ? (
          <>
            <p>{t("invitation.needsLogin")}</p>
            <div className="flex gap-3">
              <Link href="/login" className={buttonClass}>
                {t("invitation.signIn")}
              </Link>
              <Link href="/register" className={secondaryClass}>
                {t("invitation.createAccount")}
              </Link>
            </div>
          </>
        ) : preview ? (
          <>
            <h2 className="text-xl font-semibold">
              {t("invitation.invitedTo", { name: preview.tenant_name ?? "" })}
            </h2>
            <div className="flex flex-wrap items-center gap-2 text-sm text-muted">
              <span>{t("invitation.invitedEmail", { email: preview.invitation.email })}</span>
              <StatusChip>{t("invitation.role")}</StatusChip>
            </div>
            {declined ? (
              <p role="status">{t("invitation.declined")}</p>
            ) : preview.invitation.status === "accepted" ? (
              <>
                <p role="status">{t("invitation.accepted")}</p>
                <Link className={`${buttonClass} inline-block`} href={`/tenants/${preview.invitation.tenant_id}/inbox`}>
                  {t("invitation.enter")}
                </Link>
              </>
            ) : (
              <>
                <p>{t("invitation.acceptHelp")}</p>
                <div className="flex gap-3">
                  <button type="button" className={buttonClass} disabled={busy} onClick={() => void act("accept")}>
                    {t("invitation.accept")}
                  </button>
                  <button type="button" className={secondaryClass} disabled={busy} onClick={() => void act("decline")}>
                    {t("invitation.decline")}
                  </button>
                </div>
              </>
            )}
          </>
        ) : null}
        {problem === "other" && <Notice error={error} />}
        <Link href="/tenants" className="block text-sm text-primary underline">
          {t("invitation.back")}
        </Link>
      </section>
    </main>
  );
}
