"use client";

import { useEffect, useState, useSyncExternalStore } from "react";
import Link from "next/link";
import { Button, ThemeToggle } from "@hexta/ui";
import { useAuthStore } from "@/store/useAuthStore";
import { sdk } from "@/lib/sdk";
import { useRouter } from "next/navigation";
import { User, LogOut, LayoutDashboard } from "lucide-react";


const emptySubscribe = () => () => {};

function useIsClient() {
  return useSyncExternalStore(
    emptySubscribe,
    () => true,
    () => false
  );
}

export function AuthNav() {
  const { isAuthenticated, user, initialize, logout } = useAuthStore();
  const isClient = useIsClient();
  const router = useRouter();
  const [signingOut, setSigningOut] = useState(false);
  const [logoutError, setLogoutError] = useState("");

  useEffect(() => {
    initialize();
  }, [initialize]);

  const handleLogout = async () => {
    setSigningOut(true);
    setLogoutError("");
    try {
      await sdk.identity.logout();
      logout();
      sessionStorage.removeItem("hexta.workspace");
      router.push("/");
    } catch {
      setLogoutError("Could not sign out. Please try again.");
    } finally {
      setSigningOut(false);
    }
  };

  if (!isClient) {
    return (
      <div className="flex items-center gap-4">
        <div className="w-9 h-9 animate-pulse bg-muted/20 rounded-full" />
        <div className="w-20 h-9 animate-pulse bg-muted/20 rounded-md" />
        <div className="w-28 h-9 animate-pulse bg-muted/20 rounded-full" />
      </div>
    );
  }

  if (isAuthenticated) {
    return (
      <div className="flex items-center gap-4">
        <ThemeToggle />
        <div className="flex items-center gap-2 text-sm text-foreground bg-muted/10 px-3 py-1.5 rounded-full border border-border">
          <User className="w-4 h-4 text-primary" />
          <span className="font-medium truncate max-w-[150px]">
            {user?.email || "User"}
          </span>
        </div>
        
        <Button asChild variant="ghost" size="sm" className="hover:text-primary transition-colors text-muted hidden sm:flex">
          <Link href="/tenants">
            <LayoutDashboard className="w-4 h-4 mr-2" />
            My tenants
          </Link>
        </Button>

        <Button onClick={handleLogout} disabled={signingOut} variant="outline" size="sm" className="rounded-full shadow-sm hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-all">
          <LogOut className="w-4 h-4 sm:mr-2" />
          <span className="hidden sm:inline">{signingOut ? "Signing out…" : "Sign out"}</span>
        </Button>
        {logoutError && <p role="alert" className="max-w-48 text-sm text-destructive">{logoutError}</p>}
      </div>
    );
  }

  return (
    <div className="flex items-center gap-4">
      <ThemeToggle />
      <Button asChild variant="ghost" className="hover:text-primary transition-colors text-muted">
        <Link href="/login">Sign in</Link>
      </Button>
      <Button asChild className="rounded-full shadow-lg shadow-primary/25 transition-all">
        <Link href="/register">Get started</Link>
      </Button>
    </div>
  );
}

