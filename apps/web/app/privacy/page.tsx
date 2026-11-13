import type { Metadata } from "next";
import { LegalLayout } from "@/components/legal-layout";
import { AlertCircle } from "lucide-react";

export const metadata: Metadata = {
  title: "Chính Sách Bảo Mật (Dự thảo) | Hexta",
  description:
    "Bản dự thảo chính sách bảo vệ dữ liệu người dùng và thông tin thử nghiệm trên nền tảng Hexta (hexta.org).",
  alternates: {
    canonical: "https://hexta.org/privacy",
  },
  openGraph: {
    title: "Chính Sách Bảo Mật | Hexta",
    description: "Nguyên tắc bảo mật thông tin trong giai đoạn phát triển và thử nghiệm.",
    url: "https://hexta.org/privacy",
    siteName: "Hexta",
  },
};

export default function PrivacyPage() {
  return (
    <LegalLayout
      title="Chính Sách Bảo Mật Thông Tin"
      subtitle="Bản dự thảo giải thích cách Hexta xử lý dữ liệu người dùng trong giai đoạn phát triển, thử nghiệm tính năng và vận hành mô phỏng."
      lastUpdated="Tháng 11/2026 (Bản dự thảo)"
      readTime="5 phút"
    >
      <div className="space-y-8 text-foreground/90">
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs sm:text-sm text-foreground/90 space-y-1.5">
          <div className="flex items-center gap-2 font-semibold text-amber-700 dark:text-amber-400">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>Lưu ý về phiên bản tài liệu</span>
          </div>
          <p className="text-muted leading-relaxed">
            Đây là bản dự thảo nguyên tắc bảo vệ quyền riêng tư trong giai đoạn thử nghiệm. Bản quy chế pháp lý hoàn chỉnh kèm thông tin pháp nhân đăng ký chính thức sẽ do chủ dự án cung cấp trước khi phát hành thương mại.
          </p>
        </div>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            1. Phạm Vi Dữ Liệu Trong Giai Đoạn Thử Nghiệm
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Khi người dùng hoặc thành viên tham gia thử nghiệm nền tảng Hexta, chúng tôi có thể ghi nhận các thông tin cơ bản:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-sm sm:text-base text-muted">
            <li><strong className="text-foreground">Thông tin tài khoản thử nghiệm:</strong> Tên hiển thị, địa chỉ email dùng để xác thực và vai trò thành viên trong tổ chức.</li>
            <li><strong className="text-foreground">Dữ liệu mô phỏng nghiệp vụ:</strong> Nội dung hội thoại mẫu, danh mục sản phẩm thử nghiệm và lịch sử thao tác bàn giao công việc nhằm mục đích kiểm tra độ ổn định của hệ thống.</li>
            <li><strong className="text-foreground">Nhật ký kỹ thuật:</strong> Địa chỉ IP, thời gian phiên đăng nhập và lỗi phát sinh trong quá trình thử nghiệm.</li>
          </ul>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            2. Mục Đích Sử Dụng
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Dữ liệu thu thập chỉ được dùng nhằm:
          </p>
          <ol className="list-decimal pl-6 space-y-2 text-sm sm:text-base text-muted">
            <li>Cung cấp và hoàn thiện các chức năng quản trị đơn hàng, hội thoại và phân công công việc.</li>
            <li>Kiểm thử thuật toán đề xuất câu trả lời và trích xuất đơn hàng nháp có sự kiểm duyệt của con người.</li>
            <li>Khắc phục các lỗi phát sinh và cải thiện trải nghiệm giao diện người dùng.</li>
          </ol>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            3. Quyền Của Người Tham Gia Thử Nghiệm
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Người tham gia thử nghiệm có quyền yêu cầu xem lại, điều chỉnh hoặc xóa bỏ tài khoản thử nghiệm và dữ liệu liên quan khỏi hệ thống bất cứ lúc nào bằng cách gửi yêu cầu về email tiếp nhận.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            4. Liên Hệ &amp; Tiếp Nhận Ý Kiến
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Mọi thắc mắc hoặc yêu cầu liên quan đến việc xử lý dữ liệu trong giai đoạn phát triển xin vui lòng gửi về:
          </p>
          <p className="text-sm font-semibold text-primary pt-1">
            Email tiếp nhận: <a href="mailto:contact@hexta.org" className="underline">contact@hexta.org</a>
          </p>
        </section>
      </div>
    </LegalLayout>
  );
}
