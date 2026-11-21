import Link from "next/link";
import { BrandLogo } from "./brand-logo";
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
              Hexta là nền tảng kết nối tư vấn khách hàng, tạo đơn và bàn giao công việc dành cho doanh nghiệp nhỏ.
            </p>

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
              Nội dung chính
            </h3>
            <ul className="space-y-2.5 text-xs sm:text-sm">
              <li>
                <Link href="/#problems" className="text-muted hover:text-foreground transition-colors">
                  Vấn đề
                </Link>
              </li>
              <li>
                <Link href="/#values" className="text-muted hover:text-foreground transition-colors">
                  Giải pháp
                </Link>
              </li>
              <li>
                <Link href="/#demo" className="text-muted hover:text-foreground transition-colors font-medium text-primary">
                  Quy trình mẫu
                </Link>
              </li>
              <li>
                <Link href="/#controlled-ai" className="text-muted hover:text-foreground transition-colors">
                  AI có kiểm soát
                </Link>
              </li>
              <li>
                <Link href="/#roles" className="text-muted hover:text-foreground transition-colors">
                  Theo vai trò
                </Link>
              </li>
            </ul>
          </div>

          {/* Resources & Information */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-light mb-4">
              Thông tin
            </h3>
            <ul className="space-y-2.5 text-xs sm:text-sm">
              <li>
                <Link href="/#faq" className="text-muted hover:text-foreground transition-colors">
                  Hỏi đáp
                </Link>
              </li>
              <li>
                <a
                  href="mailto:contact@hexta.org"
                  className="text-muted hover:text-foreground transition-colors"
                >
                  Liên hệ
                </a>
              </li>
            </ul>
          </div>

          {/* Legal & Policies */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-light mb-4">
              Chính sách và pháp lý
            </h3>
            <ul className="space-y-2.5 text-xs sm:text-sm">
              <li>
                <Link href="/privacy" className="text-muted hover:text-foreground transition-colors">
                  Chính sách bảo mật (Dự thảo)
                </Link>
              </li>
              <li>
                <Link href="/terms" className="text-muted hover:text-foreground transition-colors">
                  Điều khoản dịch vụ (Dự thảo)
                </Link>
              </li>
              <li>
                <Link href="/cookies" className="text-muted hover:text-foreground transition-colors">
                  Chính sách cookie
                </Link>
              </li>
              <li>
                <Link href="/security" className="text-muted hover:text-foreground transition-colors">
                  Nguyên tắc an toàn dữ liệu
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom bar */}
        <div className="mt-12 pt-8 border-t border-border flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-muted">
          <p>© {currentYear} Hexta (hexta.org).</p>
          <div className="flex items-center gap-6">
            <Link href="/privacy" className="hover:text-foreground transition-colors">
              Bảo mật
            </Link>
            <Link href="/terms" className="hover:text-foreground transition-colors">
              Điều khoản
            </Link>
            <Link href="/security" className="hover:text-foreground transition-colors">
              An toàn dữ liệu
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
