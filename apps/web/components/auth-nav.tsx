"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@hexta/ui";
import { useAuthStore } from "@/store/useAuthStore";
import { sdk } from "@/lib/sdk";
import { useRouter } from "next/navigation";
import { User, LogOut, LayoutDashboard } from "lucide-react";

export function AuthNav() {
  const { isAuthenticated, user, initialize, logout } = useAuthStore();
  const [isClient, setIsClient] = useState(false);
  const router = useRouter();

  useEffect(() => {
    setIsClient(true);
    initialize();
  }, [initialize]);

  const handleLogout = async () => {
    await sdk.identity.logout();
    logout();
    router.push("/");
  };

  if (!isClient) {
    return (
      <div className="flex items-center gap-4">
        <div className="w-20 h-9 animate-pulse bg-muted/20 rounded-md" />
        <div className="w-28 h-9 animate-pulse bg-muted/20 rounded-full" />
      </div>
    );
  }

  if (isAuthenticated) {
    return (
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-sm text-foreground bg-muted/10 px-3 py-1.5 rounded-full border border-border">
          <User className="w-4 h-4 text-primary" />
          <span className="font-medium truncate max-w-[150px]">
            {user?.email || "User"}
          </span>
        </div>
        
        <Button asChild variant="ghost" size="sm" className="hover:text-primary transition-colors text-muted hidden sm:flex">
          <Link href="/tenant">
            <LayoutDashboard className="w-4 h-4 mr-2" />
            Dashboard
          </Link>
        </Button>

        <Button onClick={handleLogout} variant="outline" size="sm" className="rounded-full shadow-sm hover:bg-destructive/10 hover:text-destructive hover:border-destructive/30 transition-all">
          <LogOut className="w-4 h-4 sm:mr-2" />
          <span className="hidden sm:inline">Đăng xuất</span>
        </Button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-4">
      <Button asChild variant="ghost" className="hover:text-primary transition-colors text-muted">
        <Link href="/login">Đăng nhập</Link>
      </Button>
      <Button asChild className="rounded-full shadow-lg shadow-primary/25 transition-all">
        <Link href="/register">Bắt đầu ngay</Link>
      </Button>
    </div>
  );
}
