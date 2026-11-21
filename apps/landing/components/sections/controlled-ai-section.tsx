import { ShieldCheck, Sparkles, UserCheck, Lock, AlertTriangle } from "lucide-react";

export function ControlledAiSection() {
  return (
    <section id="controlled-ai" className="py-20 md:py-28 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
          {/* Status Badge */}
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-blue-500/10 text-primary text-xs font-semibold">
            <Sparkles className="w-3.5 h-3.5" />
            <span>AI có kiểm soát</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            AI hỗ trợ, <br className="hidden sm:inline" />
            <span className="text-primary">con người quyết định</span>
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Hexta dùng AI để giảm các thao tác soạn tin và nhập liệu lặp lại, nhưng không giao cho AI quyền ra quyết định kinh doanh.
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
                <span>Có nhân viên duyệt</span>
              </div>

              <h3 className="text-xl font-bold text-foreground">
                Gợi ý phản hồi
              </h3>

              <p className="text-sm text-muted leading-relaxed">
                Hệ thống phân tích câu hỏi của khách và soạn sẵn câu trả lời phù hợp. Nhân viên đọc, chỉnh sửa nếu cần rồi bấm gửi.
              </p>
            </div>

            <div className="p-4 rounded-xl bg-background border border-border/80 text-xs text-muted space-y-1.5">
              <span className="font-semibold text-foreground flex items-center gap-1.5">
                <UserCheck className="w-4 h-4 text-primary" />
                Kiểm soát:
              </span>
              <p>Nhân viên luôn là người gửi tin cuối cùng.</p>
            </div>
          </div>

          {/* Mode 2 */}
          <div className="p-8 rounded-2xl border border-border bg-surface flex flex-col justify-between space-y-6 shadow-xs">
            <div className="space-y-4">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-md bg-amber-500/10 text-amber-700 dark:text-amber-400 text-xs font-bold">
                <span>Chế độ 2</span>
                <span>•</span>
                <span>Thử nghiệm</span>
              </div>

              <h3 className="text-xl font-bold text-foreground">
                Tự động trả lời có giới hạn
              </h3>

              <p className="text-sm text-muted leading-relaxed">
                Trả lời nhanh các câu hỏi thường gặp ngoài giờ làm việc như bảng size, giờ mở cửa hay thông tin sản phẩm, dựa trên nội dung soạn sẵn.
              </p>
            </div>

            <div className="p-4 rounded-xl bg-background border border-border/80 text-xs text-muted space-y-1.5">
              <span className="font-semibold text-foreground flex items-center gap-1.5">
                <Lock className="w-4 h-4 text-amber-600 dark:text-amber-400" />
                Nguyên tắc dừng:
              </span>
              <p>Khi nhân viên nhận hội thoại, chế độ tự động gửi dừng ngay.</p>
            </div>
          </div>
        </div>

        {/* Safety Boundaries */}
        <div className="max-w-5xl mx-auto rounded-2xl border border-border bg-surface-hover/30 p-6 sm:p-8 space-y-6">
          <div className="flex items-center gap-2 text-foreground font-bold text-base">
            <ShieldCheck className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
            <span>Những việc AI không làm</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 text-xs sm:text-sm">
            <div className="space-y-2">
              <div className="font-semibold text-foreground flex items-center gap-1.5">
                <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
                <span>Không tự xác nhận đơn</span>
              </div>
              <p className="text-muted leading-relaxed text-xs">
                AI chỉ trích xuất thông tin để lập đơn nháp. Đơn hàng phải được nhân viên hoặc chủ doanh nghiệp xác nhận.
              </p>
            </div>

            <div className="space-y-2">
              <div className="font-semibold text-foreground flex items-center gap-1.5">
                <AlertTriangle className="w-4 h-4 text-amber-500 shrink-0" />
                <span>Không tự giảm giá hay hoàn tiền</span>
              </div>
              <p className="text-muted leading-relaxed text-xs">
                AI không tự áp mã giảm giá, thương lượng giá hay hoàn tiền khi chưa có chính sách và sự cho phép của chủ doanh nghiệp.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
