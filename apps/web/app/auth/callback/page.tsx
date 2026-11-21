"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Cookies from "js-cookie";
import { DefaultBrowserStorage } from "@hexta/sdk";
import { useAuthStore } from "@/store/useAuthStore";
import { Loader2 } from "lucide-react";
import { toast } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";

const storage = new DefaultBrowserStorage();

function AuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((state) => state.setAuth);
  const t = useT();

  useEffect(() => {
    // Read from search params (fallback/legacy) or from secure cookie/storage
    const token = searchParams.get("token") || storage.get("auth_token") || Cookies.get("auth_token");

    if (token) {
      storage.set("auth_token", token);
      Cookies.set("auth_token", token, { expires: 7, path: "/" });
      // Update auth store
      setAuth();

      toast.success(t("authPages.callbackSuccess"));
      sessionStorage.removeItem("hexta.workspace");
      router.push(sessionStorage.getItem("hexta.invitation") ? "/invitations" : "/tenants");
    } else {
      toast.error(t("authPages.callbackFailed"));
      router.push("/login");
    }
  }, [router, searchParams, setAuth, t]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
      <Loader2 className="w-10 h-10 animate-spin text-primary mb-4" />
      <h2 className="text-xl font-medium tracking-tight">{t("authPages.callbackTitle")}</h2>
      <p className="text-muted text-sm mt-2">{t("authPages.callbackWait")}</p>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
          <Loader2 aria-hidden="true" className="w-10 h-10 animate-spin text-primary mb-4" />
        </div>
      }
    >
      <AuthCallbackContent />
    </Suspense>
  );
}
