import type { Metadata } from "next";
import { LegalLayout } from "@/components/legal-layout";
import { AlertCircle } from "lucide-react";

export const metadata: Metadata = {
  title: "Điều Khoản Dịch Vụ (Dự thảo) | Hexta",
  description:
    "Bản dự thảo điều khoản sử dụng và quy định tham gia thử nghiệm nền tảng Hexta (hexta.org).",
  alternates: {
    canonical: "https://hexta.org/terms",
  },
  openGraph: {
    title: "Điều Khoản Dịch Vụ | Hexta",
    description: "Quy định sử dụng trong giai đoạn phát triển và thử nghiệm.",
    url: "https://hexta.org/terms",
    siteName: "Hexta",
  },
};

export default function TermsPage() {
  return (
    <LegalLayout
      title="Điều Khoản Sử Dụng"
      subtitle="Bản dự thảo thỏa thuận giữa người dùng thử nghiệm và đội ngũ phát triển Hexta khi truy cập hoặc sử dụng hệ thống."
      lastUpdated="Tháng 11/2026 (Bản dự thảo)"
      readTime="5 phút"
    >
      <div className="space-y-8 text-foreground/90">
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs sm:text-sm text-foreground/90 space-y-1.5">
          <div className="flex items-center gap-2 font-semibold text-amber-700 dark:text-amber-400">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>Thông báo về điều khoản thử nghiệm</span>
          </div>
          <p className="text-muted leading-relaxed">
            Hexta hiện đang trong giai đoạn phát triển và thử nghiệm tính năng. Các điều khoản thương mại chính thức, biểu phí dịch vụ và cam kết thanh toán sẽ do chủ dự án ban hành khi mở bán chính thức.
          </p>
        </div>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            1. Phạm Vi Áp Dụng
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Khi truy cập website hexta.org hoặc sử dụng các phiên bản thử nghiệm nội bộ, bạn đồng ý với các nguyên tắc sử dụng được mô tả tại đây. Mục đích chính của hệ thống trong giai đoạn này là khảo sát nhu cầu, thu thập phản hồi và cải thiện quy trình vận hành.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            2. Trách Nhiệm Sử Dụng Tài Khoản Thử Nghiệm
          </h2>
          <ul className="list-disc pl-6 space-y-2 text-sm sm:text-base text-muted">
            <li>Bạn có trách nhiệm bảo mật mật khẩu và thông tin đăng nhập tài khoản thử nghiệm của mình.</li>
            <li>Không sử dụng hệ thống để thực hiện các hành vi quét lỗ hổng trái phép hoặc cố ý gây gián đoạn hạ tầng máy chủ thử nghiệm.</li>
            <li>Không đưa dữ liệu nhạy cảm hoặc bí mật kinh doanh chưa được phép vào môi trường thử nghiệm.</li>
          </ul>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            3. Quyền Sở Hữu Trí Tuệ
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Mã nguồn cốt lõi của dự án Hexta được quản lý theo giấy phép mã nguồn mở Apache 2.0. Thương hiệu &ldquo;Hexta&rdquo; và bộ nhận diện thuộc quyền sở hữu của đội ngũ sáng lập Hexta. Dữ liệu thử nghiệm do bạn nhập vào hệ thống thuộc quyền sở hữu của bạn.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            4. Giới Hạn Trách Nhiệm Trong Giai Đoạn Thử Nghiệm
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Trong thời gian thử nghiệm và mô phỏng, phần mềm được cung cấp trên nguyên tắc &ldquo;nguyên trạng&rdquo; (as-is). Đội ngũ phát triển không chịu trách nhiệm đối với các gián đoạn dịch vụ kỹ thuật hoặc mất mát dữ liệu thử nghiệm phát sinh ngoài ý muốn.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            5. Góp Ý &amp; Phản Hồi
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Mọi ý kiến đóng góp cho quy trình và sản phẩm vui lòng gửi về:
          </p>
          <p className="text-sm font-semibold text-primary pt-1">
            Email: <a href="mailto:contact@hexta.org" className="underline">contact@hexta.org</a>
          </p>
        </section>
      </div>
    </LegalLayout>
  );
}
