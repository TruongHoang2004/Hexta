import Link from "next/link";
import { BrandLogo } from "./brand-logo";
import { ShieldAlert } from "lucide-react";
import { GithubIcon } from "./icons/social-icons";

export function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="border-t border-border bg-surface/50 text-foreground">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 md:py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-8 lg:gap-12">
          {/* Brand Info & Mission */}
          <div className="lg:col-span-2 space-y-4">
            <BrandLogo className="w-32 md:w-36" />
            <p className="text-xs sm:text-sm text-muted leading-relaxed max-w-sm">
              Hexta là giải pháp kết nối vận hành bán hàng dành cho doanh nghiệp vừa và nhỏ: từ tư vấn khách hàng, tạo đơn đến phân công công việc tuần tự.
            </p>

            {/* Development Stage Indicator */}
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-amber-500/10 border border-amber-500/20 text-xs font-medium text-amber-700 dark:text-amber-400">
              <ShieldAlert className="w-3.5 h-3.5" />
              <span>Giai đoạn thử nghiệm &amp; mô phỏng quy trình</span>
            </div>

            {/* Source code / repo link */}
            <div className="pt-2">
              <Link
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-surface border border-border text-xs text-muted hover:text-foreground hover:bg-surface-hover transition-colors"
                aria-label="GitHub Repository"
              >
                <GithubIcon className="w-4 h-4" />
                <span>Kho mã nguồn dự án</span>
              </Link>
            </div>
          </div>

          {/* Navigation Links */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-light mb-4">
              Nội Dung Chính
            </h3>
            <ul className="space-y-2.5 text-xs sm:text-sm">
              <li>
                <Link href="/#problems" className="text-muted hover:text-foreground transition-colors">
                  Vấn Đề Vận Hành
                </Link>
              </li>
              <li>
                <Link href="/#values" className="text-muted hover:text-foreground transition-colors">
                  Giá Trị Cốt Lõi
                </Link>
              </li>
              <li>
                <Link href="/#demo" className="text-muted hover:text-foreground transition-colors font-medium text-primary">
                  Minh Họa 5 Bước
                </Link>
              </li>
              <li>
                <Link href="/#controlled-ai" className="text-muted hover:text-foreground transition-colors">
                  AI Có Kiểm Soát
                </Link>
              </li>
              <li>
                <Link href="/#roles" className="text-muted hover:text-foreground transition-colors">
                  Lợi Ích Theo Vai Trò
                </Link>
              </li>
            </ul>
          </div>

          {/* Resources & Information */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-light mb-4">
              Thông Tin Dự Án
            </h3>
            <ul className="space-y-2.5 text-xs sm:text-sm">
              <li>
                <Link href="/#faq" className="text-muted hover:text-foreground transition-colors">
                  Câu Hỏi Thường Gặp (FAQ)
                </Link>
              </li>
              <li>
                <a
                  href="mailto:contact@hexta.org"
                  className="text-muted hover:text-foreground transition-colors"
                >
                  Liên Hệ Góp Ý &amp; Trao Đổi
                </a>
              </li>
              <li>
                <span className="text-muted/60 cursor-default">
                  Tài Liệu Kỹ Thuật (Nội bộ)
                </span>
              </li>
              <li>
                <span className="text-muted/60 cursor-default">
                  Bản Thử Nghiệm MVP
                </span>
              </li>
            </ul>
          </div>

          {/* Legal & Policies */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-light mb-4">
              Chính Sách &amp; Pháp Lý
            </h3>
            <ul className="space-y-2.5 text-xs sm:text-sm">
              <li>
                <Link href="/privacy" className="text-muted hover:text-foreground transition-colors">
                  Chính Sách Bảo Mật (Dự thảo)
                </Link>
              </li>
              <li>
                <Link href="/terms" className="text-muted hover:text-foreground transition-colors">
                  Điều Khoản Dịch Vụ (Dự thảo)
                </Link>
              </li>
              <li>
                <Link href="/cookies" className="text-muted hover:text-foreground transition-colors">
                  Chính Sách Cookie
                </Link>
              </li>
              <li>
                <Link href="/security" className="text-muted hover:text-foreground transition-colors">
                  Nguyên Tắc An Toàn Dữ Liệu
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom bar */}
        <div className="mt-12 pt-8 border-t border-border flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-muted">
          <p>© {currentYear} Hexta (hexta.org). Dự án đang trong giai đoạn phát triển và thử nghiệm.</p>
          <div className="flex items-center gap-6">
            <Link href="/privacy" className="hover:text-foreground transition-colors">
              Bảo Mật
            </Link>
            <Link href="/terms" className="hover:text-foreground transition-colors">
              Điều Khoản
            </Link>
            <Link href="/security" className="hover:text-foreground transition-colors">
              An Toàn Dữ Liệu
            </Link>
            <Link href="/sitemap.xml" className="hover:text-foreground transition-colors">
              Sitemap
            </Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
