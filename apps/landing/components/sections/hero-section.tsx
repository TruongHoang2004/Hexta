"use client";

import * as React from "react";
import { ArrowDown, CheckCircle2, MessageSquare, ShoppingBag, GitFork, UserCheck } from "lucide-react";

export function HeroSection() {
  const scrollTo = (id: string) => (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <section className="relative pt-28 pb-16 md:pt-36 md:pb-24 overflow-hidden">
      {/* Subtle ambient light */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[300px] md:w-[700px] md:h-[400px] bg-primary/10 blur-[100px] rounded-full pointer-events-none -z-10" />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-12 items-center">
          {/* Left Column: Headline, Copy & CTAs */}
          <div className="lg:col-span-7 space-y-6 text-left">
            {/* Status indicator */}
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-border bg-surface text-xs font-medium text-muted shadow-2xs">
              <span className="w-2 h-2 rounded-full bg-amber-500" />
              <span>Giai đoạn phát triển &amp; thử nghiệm mô phỏng</span>
            </div>

            {/* Main Headline */}
            <h1 className="text-3xl sm:text-4xl md:text-5xl lg:text-[54px] font-extrabold tracking-tight text-foreground leading-[1.18]">
              Từ tư vấn đến hoàn tất đơn hàng. <br />
              <span className="text-primary">
                Luôn rõ việc, rõ người phụ trách.
              </span>
            </h1>

            {/* Subhead */}
            <p className="text-base sm:text-lg text-muted leading-relaxed max-w-2xl">
              Hexta kết nối liền mạch giữa khâu tư vấn khách hàng, tạo đơn và phân công xử lý sau bán.
              Mỗi bước đều có người đứng tên nhận việc, AI hỗ trợ soạn tin và lập đơn nháp trong phạm vi kiểm soát an toàn.
            </p>

            {/* CTAs */}
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 pt-2">
              <a
                href="#demo"
                onClick={scrollTo("demo")}
                className="inline-flex items-center justify-center gap-2 px-6 py-3.5 text-sm font-semibold text-primary-foreground bg-primary hover:bg-primary-hover rounded-xl shadow-xs transition-all cursor-pointer"
              >
                <span>Xem trải nghiệm minh họa</span>
                <ArrowDown className="w-4 h-4" />
              </a>

              <a
                href="#values"
                onClick={scrollTo("values")}
                className="inline-flex items-center justify-center gap-2 px-5 py-3.5 text-sm font-medium text-foreground bg-surface border border-border hover:bg-surface-hover rounded-xl transition-all cursor-pointer"
              >
                <span>Tìm hiểu cách Hexta giải quyết</span>
              </a>
            </div>

            {/* Key Assurance Points */}
            <div className="pt-4 grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs text-muted border-t border-border/70">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-primary shrink-0" />
                <span>Rõ trách nhiệm từng khâu</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-primary shrink-0" />
                <span>AI có con người duyệt duyệt</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="w-4 h-4 text-primary shrink-0" />
                <span>Không phân tán qua nhiều app</span>
              </div>
            </div>
          </div>

          {/* Right Column: Visual Product Workflow Mockup (Desktop) */}
          <div className="lg:col-span-5 w-full">
            <div className="rounded-2xl border border-border bg-surface p-5 sm:p-6 shadow-xl space-y-4">
              {/* Mockup Header */}
              <div className="flex items-center justify-between border-b border-border/80 pb-3">
                <div className="flex items-center gap-2">
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-500" />
                  <span className="text-xs font-semibold text-foreground">
                    Luồng xử lý đơn hàng #HX-1048
                  </span>
                </div>
                <span className="text-[11px] font-mono text-muted bg-muted/10 px-2 py-0.5 rounded">
                  Minh họa quy trình
                </span>
              </div>

              {/* Step 1: Customer Consultation */}
              <div className="p-3.5 rounded-xl border border-border/80 bg-background/60 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2 text-foreground font-semibold">
                    <MessageSquare className="w-4 h-4 text-primary" />
                    <span>1. Hội thoại tư vấn</span>
                  </div>
                  <span className="text-[11px] text-primary font-medium bg-primary/10 px-2 py-0.5 rounded">
                    Sale phụ trách: Mai Trần
                  </span>
                </div>
                <p className="text-xs text-muted italic line-clamp-1 bg-surface p-2 rounded border border-border/40">
                  &ldquo;Shop còn set áo linen trắng size M không, mình cần giao gấp mai nhé?&rdquo;
                </p>
              </div>

              {/* Step 2: Order Created */}
              <div className="p-3.5 rounded-xl border border-border/80 bg-background/60 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2 text-foreground font-semibold">
                    <ShoppingBag className="w-4 h-4 text-amber-500" />
                    <span>2. Xác nhận đơn hàng</span>
                  </div>
                  <span className="text-[11px] text-amber-600 dark:text-amber-400 font-medium bg-amber-500/10 px-2 py-0.5 rounded">
                    Đã kiểm tra thông tin
                  </span>
                </div>
                <div className="text-xs text-muted flex items-center justify-between bg-surface p-2 rounded border border-border/40">
                  <span>01x Set Linen Trắng (M) • COD 480.000đ</span>
                  <span className="font-semibold text-foreground">Đã chốt</span>
                </div>
              </div>

              {/* Step 3: Sequential Work Assigned */}
              <div className="p-3.5 rounded-xl border border-primary/40 bg-primary/5 space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2 text-primary font-semibold">
                    <GitFork className="w-4 h-4" />
                    <span>3. Bàn giao công việc</span>
                  </div>
                  <span className="text-[11px] font-medium text-primary bg-primary/15 px-2 py-0.5 rounded flex items-center gap-1">
                    <UserCheck className="w-3 h-3" />
                    Người nhận: Tuấn Kho
                  </span>
                </div>
                <div className="text-xs text-muted bg-surface p-2 rounded border border-border/40 space-y-1">
                  <div className="flex justify-between font-medium text-foreground">
                    <span>Đóng gói &amp; dán vận đơn</span>
                    <span className="text-primary">Đang thực hiện</span>
                  </div>
                  <p className="text-[11px] text-muted">
                    Tuấn Kho đã bấm nhận bàn giao từ Mai Trần • Có lịch sử ghi nhận
                  </p>
                </div>
              </div>

              {/* Mockup note */}
              <div className="text-[11px] text-center text-muted-light pt-1">
                Minh họa chuỗi phân công nhiệm vụ tuần tự trong Hexta
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
