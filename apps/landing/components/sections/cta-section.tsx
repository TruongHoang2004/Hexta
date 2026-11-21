"use client";

import * as React from "react";
import { ArrowDown, Mail } from "lucide-react";

export function CtaSection() {
  const scrollTo = (id: string) => (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    const el = document.getElementById(id);
    if (el) {
      el.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <section className="py-20 md:py-24 relative overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="relative rounded-3xl bg-surface border border-border p-8 sm:p-12 md:p-16 shadow-xl text-center space-y-6 max-w-4xl mx-auto">
          {/* Status badge */}
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <span>Cùng phát triển Hexta</span>
          </div>

          {/* Heading */}
          <h2 className="text-3xl sm:text-4xl font-extrabold text-foreground tracking-tight">
            Xây dựng quy trình bán hàng rõ ràng cho doanh nghiệp
          </h2>

          {/* Description */}
          <p className="text-sm sm:text-base text-muted max-w-2xl mx-auto leading-relaxed">
            Hexta giúp chủ doanh nghiệp nhỏ hạn chế bỏ sót đơn hàng và luôn biết ai đang chịu trách nhiệm cho từng việc.
          </p>

          {/* Buttons: Smooth scroll, no fake signups */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
            <a
              href="#demo"
              onClick={scrollTo("demo")}
              className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3.5 text-sm font-semibold text-primary-foreground bg-primary hover:bg-primary-hover rounded-xl shadow-xs transition-all cursor-pointer"
            >
              <span>Xem quy trình mẫu</span>
              <ArrowDown className="w-4 h-4" />
            </a>

            <a
              href="mailto:contact@hexta.org"
              className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3.5 text-sm font-medium text-foreground bg-surface border border-border hover:bg-surface-hover rounded-xl transition-all cursor-pointer"
            >
              <span>Liên hệ qua email</span>
              <Mail className="w-4 h-4" />
            </a>
          </div>

          {/* Direct contact link */}
          <div className="pt-4 border-t border-border/60 flex flex-wrap items-center justify-center gap-2 text-xs text-muted">
            <Mail className="w-4 h-4 text-primary" />
            <span>Để trao đổi hoặc góp ý về nghiệp vụ, vui lòng gửi email tới</span>
            <a
              href="mailto:contact@hexta.org"
              className="font-semibold text-foreground underline hover:text-primary transition-colors"
            >
              contact@hexta.org
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
