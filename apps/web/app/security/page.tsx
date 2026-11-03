import type { Metadata } from "next";
import { LegalLayout } from "@/components/legal-layout";
import { AlertTriangle } from "lucide-react";

export const metadata: Metadata = {
  title: "Nguyên Tắc An Toàn Dữ Liệu | Hexta",
  description:
    "Nguyên tắc bảo vệ dữ liệu, cô lập tài khoản và lưu ý về giai đoạn phát triển thử nghiệm của Hexta.",
  alternates: {
    canonical: "https://hexta.org/security",
  },
  openGraph: {
    title: "Nguyên Tắc An Toàn Dữ Liệu | Hexta",
    description: "Nguyên tắc bảo mật và quản lý dữ liệu trong giai đoạn phát triển.",
    url: "https://hexta.org/security",
    siteName: "Hexta",
  },
};

export default function SecurityPage() {
  return (
    <LegalLayout
      title="Nguyên Tắc An Toàn Dữ Liệu"
      subtitle="Bản dự thảo về các biện pháp kỹ thuật và nguyên tắc bảo vệ dữ liệu trong giai đoạn phát triển và thử nghiệm của Hexta."
      lastUpdated="Tháng 11/2026 (Bản dự thảo)"
      readTime="4 phút"
    >
      <div className="space-y-8 text-foreground/90">
        {/* Notice on Trial Stage */}
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs sm:text-sm text-foreground/90 space-y-1.5">
          <div className="flex items-center gap-2 font-semibold text-amber-700 dark:text-amber-400">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>Thông báo về tình trạng phát triển</span>
          </div>
          <p className="text-muted leading-relaxed">
            Hexta hiện đang trong giai đoạn phát triển và thử nghiệm quy trình nội bộ. Mọi cam kết thương mại chính thức, thỏa thuận mức độ dịch vụ (SLA) và chính sách bồi thường sẽ do chủ dự án ban hành kèm hợp đồng pháp lý chính thức khi phát hành phiên bản thương mại.
          </p>
        </div>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            1. Nguyên Tắc Cô Lập Dữ Liệu Theo Tổ Chức (Tenant Isolation)
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Kiến trúc của Hexta được thiết kế để dữ liệu của mỗi tổ chức hoặc doanh nghiệp (Tenant) được phân vùng logic độc lập. Người dùng thuộc tổ chức này không thể truy cập hoặc nhìn thấy dữ liệu của tổ chức khác trong phạm vi quyền hạn được cấp.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            2. Kiểm Soát Quyền Truy Cập (Role-Based Access)
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Hệ thống phân định quyền rõ ràng giữa chủ tài khoản (Owner), nhân viên tư vấn (Sales) và nhân viên vận hành. Mỗi hành động bàn giao công việc hay xác nhận đơn đều được gắn với định danh người thực hiện để đảm bảo tính minh bạch.
          </p>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            3. Ranh Giới An Toàn Cho Trợ Lý AI
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Để bảo vệ an toàn nghiệp vụ cho doanh nghiệp, các tính năng AI trong Hexta bị ràng buộc bởi các quy tắc nghiêm ngặt:
          </p>
          <ul className="list-disc pl-6 space-y-2 text-sm text-muted">
            <li>AI chỉ đề xuất câu trả lời và soạn đơn hàng nháp, con người là người kiểm duyệt cuối cùng.</li>
            <li>Khi nhân viên tiếp nhận hội thoại, cơ chế tự động gửi tin nhắn của hệ thống sẽ lập tức dừng lại.</li>
            <li>Hệ thống không tự ý áp mã giảm giá, không tự ý chốt đơn hay hủy/trả hàng nếu không có thao tác từ nhân viên hoặc chủ shop.</li>
          </ul>
        </section>

        <section className="space-y-3">
          <h2 className="text-xl sm:text-2xl font-bold text-foreground tracking-tight">
            4. Tiếp Nhận Báo Cáo Sự Cố &amp; Đóng Góp
          </h2>
          <p className="text-sm sm:text-base leading-relaxed text-muted">
            Trong quá trình thử nghiệm, nếu bạn phát hiện bất kỳ lỗi logic hoặc hành vi bất thường nào về an toàn dữ liệu, xin vui lòng gửi thông tin chi tiết về:
          </p>
          <p className="font-mono font-semibold text-primary text-sm pt-1">
            contact@hexta.org
          </p>
        </section>
      </div>
    </LegalLayout>
  );
}
