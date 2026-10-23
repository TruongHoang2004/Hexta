"use client";
import { BrandLogo } from "@/components/brand-logo";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { sdk } from "@/lib/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { onboarding, type InvitationResult } from "@/lib/services/onboarding";
import { errorText } from "@/lib/services/sales";
import {
  buttonClass,
  secondaryClass,
  Notice,
} from "@/components/sales/primitives";
export default function InvitationPage() {
  const router = useRouter();
  const [token, setToken] = useState("");
  const [preview, setPreview] = useState<InvitationResult | null>(null);
  const [needsLogin, setNeedsLogin] = useState(false);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
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
          throw new Error(
            "Open the private invitation link shared by the tenant owner.",
          );
        }
        setToken(value);
        if (!(await sdk.identity.getAccessToken())) {
          if (!ignore) setNeedsLogin(true);
          return;
        }
        const result = await onboarding.useInvitation(value, "preview");
        if (!ignore) setPreview(result);
      } catch (err) {
        if (!ignore) setError(errorText(err));
      } finally {
        if (!ignore) setLoading(false);
      }
    }
    void load();
    return () => {
      ignore = true;
    };
  }, []);
  async function switchAccount() {
    setBusy(true);
    setError("");
    try {
      await sdk.identity.logout();
      useAuthStore.getState().logout();
      sessionStorage.removeItem("hexta.workspace");
      router.push("/login");
    } catch {
      setError("Could not sign out. Please try again.");
    } finally {
      setBusy(false);
    }
  }
  async function act(action: "accept" | "decline") {
    setBusy(true);
    setError("");
    try {
      const result = await onboarding.useInvitation(token, action);
      setPreview(result);
      setDeclined(action === "decline");
      sessionStorage.removeItem("hexta.invitation");
    } catch (err) {
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
        <h1 className="text-2xl font-semibold">Tenant invitation</h1>
        <Notice error={error} />
        {loading ? (
          <p role="status">Checking invitation…</p>
        ) : needsLogin ? (
          <>
            <p>
              Sign in or create an account using the email address invited by
              the tenant owner.
            </p>
            <div className="flex gap-3">
              <Link href="/login" className={buttonClass}>
                Sign in
              </Link>
              <Link href="/register" className={secondaryClass}>
                Create account
              </Link>
            </div>
          </>
        ) : preview ? (
          <>
            <h2 className="text-xl font-semibold">{preview.tenant_name}</h2>
            <p className="text-sm text-muted">
              Invited email: {preview.invitation.email}
            </p>
            {declined ? (
              <p>Invitation declined. No tenant access was granted.</p>
            ) : preview.invitation.status === "accepted" ? (
              <>
                <p role="status">You have accepted this invitation.</p>
                <Link
                  className={`${buttonClass} inline-block`}
                  href={`/tenants/${preview.invitation.tenant_id}/orders`}
                >
                  Enter tenant
                </Link>
              </>
            ) : (
              <>
                <p>
                  Accept to join this tenant as a member. Until then, its
                  business resources remain unavailable.
                </p>
                <div className="flex gap-3">
                  <button
                    className={buttonClass}
                    disabled={busy}
                    onClick={() => void act("accept")}
                  >
                    Accept invitation
                  </button>
                  <button
                    className={secondaryClass}
                    disabled={busy}
                    onClick={() => void act("decline")}
                  >
                    Decline
                  </button>
                </div>
              </>
            )}
          </>
        ) : null}
        <Link href="/tenants" className="block text-sm text-primary underline">
          Back to your tenants
        </Link>
        {error && (
          <button
            type="button"
            disabled={busy}
            onClick={() => void switchAccount()}
            className="block text-sm text-primary underline"
          >
            Sign in with another account
          </button>
        )}
      </section>
    </main>
  );
}
