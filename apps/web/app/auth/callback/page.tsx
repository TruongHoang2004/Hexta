"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Cookies from "js-cookie";
import { useAuthStore } from "@/store/useAuthStore";
import { Loader2 } from "lucide-react";
import { toast } from "@hexta/ui";

function AuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((state) => state.setAuth);

  useEffect(() => {
    // Read from search params (fallback/legacy) or from secure cookie set by backend
    const token = searchParams.get("token") || Cookies.get("auth_token");

    if (token) {
      // Ensure cookie is set if came from query param
      Cookies.set("auth_token", token, { expires: 7, path: "/" });
      // Update auth store
      setAuth();

      toast.success("Login successful!");
      router.push("/tenant");
    } else {
      toast.error("Authentication failed: token not found.");
      router.push("/login");
    }
  }, [router, searchParams, setAuth]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
      <Loader2 className="w-10 h-10 animate-spin text-primary mb-4" />
      <h2 className="text-xl font-medium tracking-tight">Completing authentication...</h2>
      <p className="text-muted text-sm mt-2">Please wait a moment</p>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
          <Loader2 className="w-10 h-10 animate-spin text-primary mb-4" />
          <h2 className="text-xl font-medium tracking-tight">Loading...</h2>
        </div>
      }
    >
      <AuthCallbackContent />
    </Suspense>
  );
}

