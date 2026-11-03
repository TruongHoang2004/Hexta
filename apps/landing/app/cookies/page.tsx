import type { Metadata } from "next";
import { LegalLayout } from "@/components/legal-layout";

export const metadata: Metadata = {
  title: "Chính Sách Cookie (Cookie Policy) | Hexta",
  description:
    "Tìm hiểu cách Hexta sử dụng cookie và công nghệ lưu trữ cục bộ để nâng cao trải nghiệm người dùng, duy trì phiên làm việc và bảo vệ an toàn trên hexta.org.",
  alternates: {
    canonical: "https://hexta.org/cookies",
  },
  openGraph: {
    title: "Chính Sách Cookie | Hexta Platform",
    description: "Giải thích chi tiết về việc sử dụng cookie và quyền kiểm soát của bạn trên hexta.org.",
    url: "https://hexta.org/cookies",
    siteName: "Hexta",
  },
};

export default function CookiesPage() {
  return (
    <LegalLayout
      title="Chính Sách Sử Dụng Cookie"
      subtitle="Minh bạch về cách chúng tôi sử dụng cookie và các công nghệ lưu trữ cục bộ để tối ưu hóa trải nghiệm của bạn trên hexta.org."
      lastUpdated="Ngày 03 tháng 11 năm 2026"
      readTime="5 phút"
    >
      <div className="space-y-8 text-foreground/90">
        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            1. Cookie Là Gì?
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Cookie là một tệp văn bản nhỏ được lưu trữ trên trình duyệt hoặc thiết bị của bạn khi bạn truy cập một trang web. Cookie cho phép trang web nhận diện thiết bị của bạn, ghi nhớ các tùy chọn cá nhân (ví dụ: chế độ Dark/Light mode, ngôn ngữ hiển thị) và bảo vệ an toàn cho các phiên đăng nhập.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            2. Các Loại Cookie Được Sử Dụng Trên Hexta
          </h2>
          <div className="space-y-4">
            <div className="p-4 rounded-xl border border-border bg-surface-hover/50 space-y-1.5">
              <h3 className="font-semibold text-foreground text-base">
                A. Cookie Cần Thiết & Bảo Mật (Strictly Necessary Cookies)
              </h3>
              <p className="text-xs sm:text-sm text-muted leading-relaxed">
                Các cookie này bắt buộc phải có để trang web hoạt động bình thường. Chúng quản lý phiên đăng nhập an toàn, bảo vệ chống giả mạo yêu cầu (CSRF token), và ghi nhớ mã định danh Tenant trong môi trường Multi-tenant. Bạn không thể tắt loại cookie này trong hệ thống của chúng tôi.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-surface-hover/50 space-y-1.5">
              <h3 className="font-semibold text-foreground text-base">
                B. Cookie Tùy Chọn & Giao Diện (Preferences Cookies)
              </h3>
              <p className="text-xs sm:text-sm text-muted leading-relaxed">
                Cho phép ghi nhớ cấu hình giao diện bạn đã chọn, chẳng hạn như chuyển đổi giữa giao diện Sáng và Tối (thông qua thư viện <code className="text-xs text-primary font-mono">next-themes</code>) hoặc ngôn ngữ yêu thích, giúp bạn không phải thiết lập lại ở mỗi lần truy cập.
              </p>
            </div>

            <div className="p-4 rounded-xl border border-border bg-surface-hover/50 space-y-1.5">
              <h3 className="font-semibold text-foreground text-base">
                C. Cookie Phân Tích & Hiệu Năng (Analytics Cookies)
              </h3>
              <p className="text-xs sm:text-sm text-muted leading-relaxed">
                Giúp chúng tôi đo lường lưu lượng truy cập và tốc độ tải trang để cải thiện trải nghiệm người dùng. Mọi dữ liệu thu thập từ các cookie phân tích đều được tổng hợp ẩn danh và không chứa thông tin định danh cá nhân trực tiếp.
              </p>
            </div>
          </div>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            3. Cách Bạn Kiểm Soát & Tắt Cookie
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Hầu hết các trình duyệt web hiện đại đều cho phép bạn kiểm tra, chặn hoặc xóa cookie thông qua phần Cài đặt (Settings):
          </p>
          <ul className="list-disc pl-6 space-y-1.5 text-xs sm:text-sm text-muted">
            <li>Google Chrome: Cài đặt &gt; Quyền riêng tư và bảo mật &gt; Cookie của bên thứ ba.</li>
            <li>Apple Safari: Cài đặt &gt; Quyền riêng tư &gt; Chặn tất cả cookie.</li>
            <li>Mozilla Firefox: Cài đặt &gt; Quyền riêng tư &amp; Bảo mật &gt; Cookie và Dữ liệu trang web.</li>
            <li>Microsoft Edge: Cài đặt &gt; Quyền đối với cookie và trang web.</li>
          </ul>
          <p className="text-xs text-muted-light italic">
            * Lưu ý: Nếu bạn vô hiệu hóa toàn bộ cookie thiết yếu, một số tính năng như đăng nhập tài khoản quản trị hay lưu trữ giỏ hàng có thể không hoạt động chính xác.
          </p>
        </section>
      </div>
    </LegalLayout>
  );
}
