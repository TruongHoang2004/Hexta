"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import Cookies from "js-cookie";
import { useAuthStore } from "@/store/useAuthStore";
import { Loader2 } from "lucide-react";
import { toast } from "@hexta/ui";

export default function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setAuth = useAuthStore((state) => state.setAuth);

  useEffect(() => {
    const token = searchParams.get("token");

    if (token) {
      // Set token to cookie
      Cookies.set("auth_token", token, { expires: 7, path: "/" });
      // Update auth store
      setAuth();
      
      toast.success("Đăng nhập thành công!");
      router.push("/tenant");
    } else {
      toast.error("Xác thực thất bại, không tìm thấy token.");
      router.push("/login");
    }
  }, [router, searchParams, setAuth]);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
      <Loader2 className="w-10 h-10 animate-spin text-primary mb-4" />
      <h2 className="text-xl font-medium tracking-tight">Đang hoàn tất quá trình xác thực...</h2>
      <p className="text-muted text-sm mt-2">Vui lòng chờ trong giây lát</p>
    </div>
  );
}
