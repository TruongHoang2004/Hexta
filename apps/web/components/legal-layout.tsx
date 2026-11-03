"use client";

import * as React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { ShieldCheck, FileText, Cookie, Lock, ArrowLeft, Calendar, Clock, Mail } from "lucide-react";

interface LegalLayoutProps {
  title: string;
  subtitle: string;
  lastUpdated: string;
  readTime: string;
  children: React.ReactNode;
}

const POLICY_NAV_ITEMS = [
  {
    title: "Chính Sách Bảo Mật",
    description: "Bản dự thảo bảo vệ dữ liệu người dùng thử nghiệm",
    href: "/privacy",
    icon: Lock,
  },
  {
    title: "Điều Khoản Dịch Vụ",
    description: "Quy định tham gia thử nghiệm & bản quyền mã nguồn",
    href: "/terms",
    icon: FileText,
  },
  {
    title: "Chính Sách Cookie",
    description: "Phân loại cookie phiên làm việc & tùy chọn giao diện",
    href: "/cookies",
    icon: Cookie,
  },
  {
    title: "Nguyên Tắc An Toàn Dữ Liệu",
    description: "Cô lập dữ liệu tổ chức & ranh giới an toàn",
    href: "/security",
    icon: ShieldCheck,
  },
];

export function LegalLayout({
  title,
  subtitle,
  lastUpdated,
  readTime,
  children,
}: LegalLayoutProps) {
  const pathname = usePathname();

  return (
    <div className="pt-24 pb-20 min-h-screen">
      {/* Background radial glow */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-7xl h-96 radial-glow pointer-events-none" />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        {/* Breadcrumb & back button */}
        <div className="flex items-center gap-2 text-xs text-muted mb-8">
          <Link href="/" className="inline-flex items-center gap-1 hover:text-foreground transition-colors">
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Trang chủ</span>
          </Link>
          <span>/</span>
          <span className="text-muted-light">Pháp lý & Chính sách</span>
          <span>/</span>
          <span className="text-foreground font-medium">{title}</span>
        </div>

        {/* Page Header */}
        <div className="border-b border-border pb-8 mb-10">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold mb-3">
            <ShieldCheck className="w-3.5 h-3.5" />
            <span>Tài Liệu Quy Chế &amp; Chính Sách • Hexta</span>
          </div>
          <h1 className="text-3xl sm:text-4xl md:text-5xl font-extrabold tracking-tight text-foreground">
            {title}
          </h1>
          <p className="mt-3 text-base sm:text-lg text-muted max-w-3xl">
            {subtitle}
          </p>

          <div className="mt-4 flex flex-wrap items-center gap-4 sm:gap-6 text-xs text-muted">
            <div className="flex items-center gap-1.5">
              <Calendar className="w-3.5 h-3.5 text-primary" />
              <span>{lastUpdated}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <Clock className="w-3.5 h-3.5 text-primary" />
              <span>Thời gian đọc: {readTime}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <ShieldCheck className="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" />
              <span>Bản dự thảo áp dụng trong giai đoạn thử nghiệm</span>
            </div>
          </div>
        </div>

        {/* Main 2-column layout */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-start">
          {/* Sidebar Navigation */}
          <aside className="lg:col-span-4 sticky top-28 space-y-6">
            <div className="glass-panel rounded-2xl p-5 shadow-xs">
              <h2 className="text-xs font-bold uppercase tracking-wider text-muted-light mb-4">
                Danh Mục Văn Bản
              </h2>
              <nav className="space-y-2">
                {POLICY_NAV_ITEMS.map((item) => {
                  const Icon = item.icon;
                  const isActive = pathname === item.href;
                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={`flex items-start gap-3 p-3 rounded-xl transition-all ${
                        isActive
                          ? "bg-primary text-primary-foreground shadow-xs font-medium"
                          : "hover:bg-muted/10 text-muted hover:text-foreground"
                      }`}
                    >
                      <Icon className={`w-5 h-5 shrink-0 mt-0.5 ${isActive ? "text-primary-foreground" : "text-primary"}`} />
                      <div>
                        <div className="text-sm font-semibold">{item.title}</div>
                        <div
                          className={`text-xs mt-0.5 leading-snug line-clamp-1 ${
                            isActive ? "text-primary-foreground/80" : "text-muted"
                          }`}
                        >
                          {item.description}
                        </div>
                      </div>
                    </Link>
                  );
                })}
              </nav>
            </div>

            {/* Quick Contact Box */}
            <div className="rounded-2xl border border-border bg-surface p-5 text-sm space-y-3">
              <div className="flex items-center gap-2 font-semibold text-foreground">
                <Mail className="w-4 h-4 text-primary" />
                <span>Liên Hệ Đóng Góp &amp; Hỏi Đáp</span>
              </div>
              <p className="text-xs text-muted leading-relaxed">
                Nếu bạn có câu hỏi về việc xử lý dữ liệu hoặc góp ý cho quy định vận hành của Hexta:
              </p>
              <div className="pt-1">
                <a
                  href="mailto:contact@hexta.org"
                  className="inline-block text-xs font-semibold text-primary hover:underline"
                >
                  contact@hexta.org
                </a>
              </div>
            </div>
          </aside>

          {/* Policy Article Content */}
          <article className="lg:col-span-8 bg-surface border border-border rounded-2xl p-6 sm:p-10 shadow-xs prose prose-slate dark:prose-invert max-w-none">
            {children}
          </article>
        </div>
      </div>
    </div>
  );
}
