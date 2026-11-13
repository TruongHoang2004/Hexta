import { ShieldCheck, Sparkles, UserCheck, Lock, AlertTriangle, Layers } from "lucide-react";

export function ControlledAiSection() {
  return (
    <section id="controlled-ai" className="py-20 md:py-28 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
          {/* Status Badge */}
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blue-500/10 text-primary text-xs font-semibold">
            <Sparkles className="w-3.5 h-3.5" />
            <span>Năng Lực Đang Phát Triển &amp; Thử Nghiệm</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            AI Hỗ Trợ Nhưng Luôn Trong <br className="hidden sm:inline" />
            <span className="text-primary">Phạm Vi Kiểm Soát An Toàn</span>
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Hexta ứng dụng AI để giảm tải các thao tác soạn tin và nhập liệu lặp lại cho nhân viên, nhưng không giao toàn quyền ra quyết định kinh doanh cho AI.
          </p>
        </div>

        {/* 2 Response Modes Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-5xl mx-auto mb-12">
          {/* Mode 1 */}
          <div className="p-8 rounded-2xl border border-border bg-surface flex flex-col justify-between space-y-6 shadow-xs">
            <div className="space-y-4">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-md bg-primary/10 text-primary text-xs font-bold">
                <span>Chế độ 1</span>
                <span>•</span>
                <span>Hỗ Trợ Đề Xuất (Đã chốt thiết kế)</span>
              </div>

              <h3 className="text-xl font-bold text-foreground">
                Gợi ý phản hồi chờ nhân viên duyệt
              </h3>

              <p className="text-sm text-muted leading-relaxed">
                Hệ thống phân tích ngữ cảnh câu hỏi của khách hàng và chuẩn bị sẵn một câu trả lời mẫu sát với nhu cầu. Nhân viên tư vấn chỉ cần đọc lướt, chỉnh sửa nếu cần và bấm gửi.
              </p>
            </div>

            <div className="p-4 rounded-xl bg-background border border-border/80 text-xs text-muted space-y-1.5">
              <span className="font-semibold text-foreground flex items-center gap-1.5">
                <UserCheck className="w-4 h-4 text-primary" />
                Kiểm soát:
              </span>
              <p>Con người là người gửi tin cuối cùng, đảm bảo tính cá nhân hóa và sự chính xác trước khách hàng.</p>
            </div>
          </div>

          {/* Mode 2 */}
          <div className="p-8 rounded-2xl border border-border bg-surface flex flex-col justify-between space-y-6 shadow-xs">
            <div className="space-y-4">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-md bg-amber-500/10 text-amber-700 dark:text-amber-400 text-xs font-bold">
                <span>Chế độ 2</span>
                <span>•</span>
                <span>Tự Động Có Giới Hạn (Đang thử nghiệm)</span>
              </div>

              <h3 className="text-xl font-bold text-foreground">
                Tự động trả lời thông tin cơ bản
              </h3>

              <p className="text-sm text-muted leading-relaxed">
                Hỗ trợ phản hồi nhanh các thắc mắc phổ biến ngoài giờ làm việc (bảng size, giờ mở cửa, thông tin sản phẩm cơ bản có kịch bản sẵn) để giữ chân khách hàng.
              </p>
            </div>

            <div className="p-4 rounded-xl bg-background border border-border/80 text-xs text-muted space-y-1.5">
              <span className="font-semibold text-foreground flex items-center gap-1.5">
                <Lock className="w-4 h-4 text-amber-600 dark:text-amber-400" />
                Nguyên tắc dừng gửi:
              </span>
              <p>Ngay khi nhân viên Sale bấm nhận hội thoại, cơ chế tự động gửi dừng lại ngay lập tức.</p>
            </div>
          </div>
        </div>

        {/* Safety Boundaries & Pancake Bridge */}
        <div className="max-w-5xl mx-auto rounded-2xl border border-border bg-surface-hover/30 p-6 sm:p-8 space-y-6">
          <div className="flex items-center gap-2 text-foreground font-bold text-base">
            <ShieldCheck className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
            <span>Ranh Giới Bảo Vệ Nghiệp Vụ Nghiêm Ngặt</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 text-xs sm:text-sm">
            <div className="space-y-2">
              <div className="font-semibold text-foreground flex items-center gap-1.5">
                <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
                <span>Không tự xác nhận đơn</span>
              </div>
              <p className="text-muted leading-relaxed text-xs">
                AI chỉ bóc tách thông tin để lập đơn hàng nháp. Việc xác nhận đơn bắt buộc phải do nhân viên hoặc chủ doanh nghiệp phê duyệt.
              </p>
            </div>

            <div className="space-y-2">
              <div className="font-semibold text-foreground flex items-center gap-1.5">
                <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
                <span>Không tự giảm giá hay hoàn tiền</span>
              </div>
              <p className="text-muted leading-relaxed text-xs">
                Hệ thống không tự ý áp dụng mã giảm giá, đàm phán giá hoặc xử lý hoàn tiền nếu không có chính sách và sự cho phép rõ ràng từ chủ shop.
              </p>
            </div>

            <div className="space-y-2">
              <div className="font-semibold text-foreground flex items-center gap-1.5">
                <Layers className="w-4 h-4 text-blue-500 shrink-0" />
                <span>Cầu nối Pancake (Kế hoạch MVP)</span>
              </div>
              <p className="text-muted leading-relaxed text-xs">
                Pancake được nghiên cứu như phương án cầu nối cho phiên bản MVP để lấy hội thoại đa kênh; Hexta chưa tuyên bố tính năng tích hợp thực tế đã hoạt động đại trà.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
