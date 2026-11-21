"use client";

import * as React from "react";
import {
  MessageSquare,
  Sparkles,
  FileCheck,
  ArrowRightLeft,
  Clock,
  User,
  ShieldAlert,
  CheckCircle,
  HelpCircle,
} from "lucide-react";

type StepKey = "receive" | "prepare" | "confirm" | "handoff" | "track";

interface StepConfig {
  key: StepKey;
  stepNumber: string;
  name: string;
  shortDesc: string;
}

const STEPS: StepConfig[] = [
  {
    key: "receive",
    stepNumber: "01",
    name: "Tiếp nhận",
    shortDesc: "Hội thoại và nhu cầu",
  },
  {
    key: "prepare",
    stepNumber: "02",
    name: "Soạn phản hồi",
    shortDesc: "Đề xuất chờ duyệt",
  },
  {
    key: "confirm",
    stepNumber: "03",
    name: "Xác nhận đơn",
    shortDesc: "Kiểm tra đơn nháp",
  },
  {
    key: "handoff",
    stepNumber: "04",
    name: "Bàn giao",
    shortDesc: "Chọn người nhận việc",
  },
  {
    key: "track",
    stepNumber: "05",
    name: "Theo dõi",
    shortDesc: "Tiến độ và người phụ trách",
  },
];

export function DemoMockupSection() {
  const [activeStep, setActiveStep] = React.useState<StepKey>("receive");

  return (
    <section id="demo" className="py-20 md:py-28 bg-surface/50 border-y border-border relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-12">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <Sparkles className="w-3.5 h-3.5" />
            <span>Quy trình mẫu</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Một đơn hàng đi qua Hexta như thế nào
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Theo dõi một đơn hàng từ tin nhắn tư vấn đầu tiên đến khi được bàn giao cho bộ phận vận hành.
          </p>

          {/* Prominent Demo Notice Badge */}
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-amber-500/10 border border-amber-500/30 text-xs text-amber-800 dark:text-amber-300 font-medium">
            <ShieldAlert className="w-4 h-4 shrink-0 text-amber-600 dark:text-amber-400" />
            <span>
              <strong>Dữ liệu minh họa:</strong> Nội dung và số liệu dưới đây chỉ nhằm minh họa giao diện đang thử nghiệm, không phải dữ liệu khách hàng thật.
            </span>
          </div>
        </div>

        {/* 5 Steps Interactive Selector */}
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 sm:gap-3 max-w-4xl mx-auto mb-8">
          {STEPS.map((s) => {
            const isActive = activeStep === s.key;
            return (
              <button
                key={s.key}
                type="button"
                onClick={() => setActiveStep(s.key)}
                className={`p-3.5 rounded-xl border text-left transition-all cursor-pointer focus:outline-hidden focus:ring-2 focus:ring-primary ${
                  isActive
                    ? "bg-primary text-primary-foreground border-primary shadow-md scale-[1.02]"
                    : "bg-surface border-border text-foreground hover:bg-surface-hover"
                }`}
                aria-pressed={isActive}
              >
                <div
                  className={`text-xs font-mono font-bold mb-1 ${
                    isActive ? "text-primary-foreground/80" : "text-muted"
                  }`}
                >
                  BƯỚC {s.stepNumber}
                </div>
                <div className="text-xs sm:text-sm font-bold leading-snug">
                  {s.name}
                </div>
                <div
                  className={`text-[11px] mt-1 line-clamp-1 ${
                    isActive ? "text-primary-foreground/80" : "text-muted"
                  }`}
                >
                  {s.shortDesc}
                </div>
              </button>
            );
          })}
        </div>

        {/* Active Mockup Card Container */}
        <div className="max-w-4xl mx-auto rounded-3xl border border-border bg-surface p-6 sm:p-10 shadow-xl space-y-6">
          {/* Top meta bar */}
          <div className="flex flex-wrap items-center justify-between gap-3 pb-5 border-b border-border/70 text-xs">
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse" />
              <span className="font-semibold text-foreground">
                Bước {STEPS.find((s) => s.key === activeStep)?.stepNumber}: {STEPS.find((s) => s.key === activeStep)?.name}
              </span>
            </div>
          </div>

          {/* STEP 1: RECEIVE */}
          {activeStep === "receive" && (
            <div className="space-y-6 animate-in fade-in duration-300">
              <div className="p-4 rounded-2xl bg-background border border-border space-y-3">
                <div className="flex items-center justify-between text-xs text-muted">
                  <div className="flex items-center gap-2">
                    <User className="w-4 h-4 text-primary" />
                    <span className="font-semibold text-foreground">Khách hàng: Trần Thu Trang (Khách mới)</span>
                  </div>
                  <span>Nguồn: Kênh nhắn tin đã kết nối</span>
                </div>
                <div className="p-3.5 rounded-xl bg-surface border border-border/60 text-xs sm:text-sm text-foreground leading-relaxed">
                  &ldquo;Chào shop, mình muốn đặt 2 hũ Granola yến mạch hạnh nhân 500g loại ít ngọt. Ship về Liễu Giai, Ba Đình trước thứ Sáu được không ạ?&rdquo;
                </div>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                <div className="p-4 rounded-xl border border-border bg-surface-hover/30 space-y-1.5">
                  <span className="font-semibold text-foreground flex items-center gap-1.5">
                    <MessageSquare className="w-4 h-4 text-primary" />
                    Tiếp nhận hội thoại
                  </span>
                  <p className="text-muted leading-relaxed">
                    Tin nhắn vào danh sách chờ tiếp nhận. Nhân viên tư vấn bấm nhận hội thoại để trở thành người phụ trách.
                  </p>
                </div>

                <div className="p-4 rounded-xl border border-border bg-surface-hover/30 space-y-1.5">
                  <span className="font-semibold text-foreground flex items-center gap-1.5">
                    <HelpCircle className="w-4 h-4 text-primary" />
                    Nhu cầu của khách
                  </span>
                  <p className="text-muted leading-relaxed">
                    2 hũ Granola 500g, vị ít ngọt, nhận hàng trước thứ Sáu, khu vực Ba Đình, Hà Nội.
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* STEP 2: PREPARE REPLY */}
          {activeStep === "prepare" && (
            <div className="space-y-6 animate-in fade-in duration-300">
              <div className="p-4 rounded-2xl bg-background border border-border space-y-3">
                <div className="flex items-center justify-between text-xs">
                  <div className="flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-primary" />
                    <span className="font-semibold text-foreground">Nhân viên tư vấn: Mai Trần (đã nhận hội thoại)</span>
                  </div>
                  <span className="text-[11px] text-primary bg-primary/10 px-2 py-0.5 rounded font-medium">
                    Chế độ: Gợi ý phản hồi
                  </span>
                </div>

                <div className="p-4 rounded-xl bg-primary/5 border border-primary/20 space-y-2">
                  <div className="flex items-center justify-between text-xs font-semibold text-primary">
                    <span className="flex items-center gap-1.5">
                      <Sparkles className="w-3.5 h-3.5" />
                      Phản hồi đề xuất, chờ nhân viên duyệt:
                    </span>
                    <span className="text-[11px] text-muted">Chưa gửi</span>
                  </div>
                  <p className="text-xs sm:text-sm text-foreground leading-relaxed">
                    &ldquo;Dạ, chào chị Trang. Bên em hiện còn Granola yến mạch hạnh nhân 500g loại ít ngọt. Với địa chỉ tại Ba Đình, bên em có thể giao hỏa tốc trong thứ Năm, kịp thời gian chị cần. Chị vui lòng cho em xin địa chỉ chi tiết và số điện thoại để em lên đơn ạ.&rdquo;
                  </p>
                </div>
              </div>

              <div className="p-4 rounded-xl border border-amber-500/20 bg-amber-500/5 text-xs text-foreground/90 space-y-1.5">
                <div className="font-semibold text-amber-700 dark:text-amber-400">
                  Khi nhân viên đã nhận hội thoại:
                </div>
                <p className="text-muted leading-relaxed">
                  Hệ thống <strong>ngừng tự động gửi tin</strong>. Nhân viên có thể xem, chỉnh sửa hoặc viết lại nội dung; chế độ tự động chỉ bật lại khi nhân viên chủ động chọn.
                </p>
              </div>
            </div>
          )}

          {/* STEP 3: CONFIRM ORDER */}
          {activeStep === "confirm" && (
            <div className="space-y-6 animate-in fade-in duration-300">
              <div className="p-4 rounded-2xl bg-background border border-border space-y-4">
                <div className="flex items-center justify-between text-xs">
                  <span className="font-semibold text-foreground flex items-center gap-1.5">
                    <FileCheck className="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                    Đơn nháp chờ kiểm tra
                  </span>
                  <span className="text-[11px] bg-amber-500/15 text-amber-700 dark:text-amber-400 px-2.5 py-0.5 rounded font-bold">
                    Trạng thái: Đơn nháp
                  </span>
                </div>

                <div className="rounded-xl border border-border bg-surface p-4 space-y-3 text-xs sm:text-sm">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pb-3 border-b border-border/60 text-xs">
                    <div>
                      <span className="text-muted">Người nhận:</span>{" "}
                      <span className="font-semibold text-foreground">Trần Thu Trang (0987.xxx.xxx)</span>
                    </div>
                    <div>
                      <span className="text-muted">Địa chỉ:</span>{" "}
                      <span className="font-semibold text-foreground">Số 12 Liễu Giai, Ba Đình, Hà Nội</span>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex justify-between text-xs text-muted">
                      <span>Sản phẩm</span>
                      <span>Thành tiền (Minh họa)</span>
                    </div>
                    <div className="flex justify-between font-medium text-foreground text-xs sm:text-sm">
                      <span>02x Granola yến mạch hạnh nhân 500g (Ít ngọt)</span>
                      <span>320.000đ</span>
                    </div>
                    <div className="flex justify-between text-xs text-muted">
                      <span>Phí vận chuyển dự kiến</span>
                      <span>25.000đ</span>
                    </div>
                  </div>

                  <div className="pt-3 border-t border-border/60 flex justify-between items-center text-sm font-bold text-foreground">
                    <span>Tổng cộng:</span>
                    <span className="text-primary text-base">345.000đ</span>
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-surface-hover/30 text-xs text-muted space-y-1.5">
                <span className="font-semibold text-foreground">Phạm vi kiểm soát:</span>
                <p>
                  Hệ thống không tự áp mã giảm giá hay tự xác nhận đơn. Đơn nháp chỉ trở thành đơn chính thức sau khi nhân viên kiểm tra đủ thông tin giao hàng.
                </p>
              </div>
            </div>
          )}

          {/* STEP 4: HANDOFF */}
          {activeStep === "handoff" && (
            <div className="space-y-6 animate-in fade-in duration-300">
              <div className="p-5 rounded-2xl bg-background border border-border space-y-4">
                <div className="flex items-center justify-between text-xs">
                  <span className="font-semibold text-foreground flex items-center gap-1.5">
                    <ArrowRightLeft className="w-4 h-4 text-primary" />
                    Phiếu bàn giao: Chuẩn bị và đóng gói
                  </span>
                  <span className="text-[11px] bg-primary/10 text-primary px-2 py-0.5 rounded font-medium">
                    Đơn #HX-1048
                  </span>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
                  <div className="p-3.5 rounded-xl border border-border bg-surface space-y-1">
                    <span className="text-muted">Người bàn giao:</span>
                    <div className="font-semibold text-foreground flex items-center gap-1.5 pt-1">
                      <User className="w-3.5 h-3.5 text-primary" />
                      <span>Mai Trần (Tư vấn)</span>
                    </div>
                    <span className="text-[11px] text-muted">Đã xác nhận đơn hàng</span>
                  </div>

                  <div className="p-3.5 rounded-xl border border-amber-500/30 bg-amber-500/5 space-y-1">
                    <span className="text-muted">Người nhận việc:</span>
                    <div className="font-semibold text-foreground flex items-center gap-1.5 pt-1">
                      <User className="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" />
                      <span>Lê Tuấn (Kho)</span>
                    </div>
                    <span className="text-[11px] text-amber-700 dark:text-amber-300 font-medium">
                      Trạng thái: Chờ Lê Tuấn nhận việc
                    </span>
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-surface-hover/30 text-xs text-muted space-y-1.5">
                <span className="font-semibold text-foreground">Nguyên tắc bàn giao:</span>
                <p>
                  Trách nhiệm chỉ chuyển sang người nhận khi họ bấm <strong>Xác nhận nhận việc</strong>. Trước đó, người bàn giao vẫn chịu trách nhiệm, nên không có khoảng trống giữa hai bước.
                </p>
              </div>
            </div>
          )}

          {/* STEP 5: TRACK */}
          {activeStep === "track" && (
            <div className="space-y-6 animate-in fade-in duration-300">
              <div className="p-5 rounded-2xl bg-background border border-border space-y-4">
                <div className="flex items-center justify-between text-xs">
                  <span className="font-semibold text-foreground flex items-center gap-1.5">
                    <Clock className="w-4 h-4 text-primary" />
                    Nhật ký tiến độ
                  </span>
                  <span className="text-[11px] font-mono text-muted">Đơn #HX-1048</span>
                </div>

                <div className="space-y-3">
                  <div className="flex items-start gap-3 p-3 rounded-xl bg-surface border border-border text-xs">
                    <CheckCircle className="w-4 h-4 text-emerald-500 shrink-0 mt-0.5" />
                    <div className="flex-1 space-y-0.5">
                      <div className="flex justify-between font-semibold text-foreground">
                        <span>Tiếp nhận và xác nhận đơn hàng</span>
                        <span className="text-muted font-normal text-[11px]">10:15</span>
                      </div>
                      <p className="text-muted">Hoàn thành bởi Mai Trần • Đơn đã được xác nhận</p>
                    </div>
                  </div>

                  <div className="flex items-start gap-3 p-3 rounded-xl bg-surface border border-primary/30 text-xs">
                    <span className="w-4 h-4 rounded-full border-2 border-primary border-t-transparent animate-spin shrink-0 mt-0.5" />
                    <div className="flex-1 space-y-0.5">
                      <div className="flex justify-between font-semibold text-primary">
                        <span>Đóng gói và dán vận đơn</span>
                        <span className="text-primary font-normal text-[11px]">Đang thực hiện</span>
                      </div>
                      <p className="text-muted">Lê Tuấn nhận bàn giao lúc 10:22 • Đang chuẩn bị hàng</p>
                    </div>
                  </div>

                  <div className="flex items-start gap-3 p-3 rounded-xl bg-surface/50 border border-border/50 text-xs opacity-60">
                    <span className="w-4 h-4 rounded-full border border-muted shrink-0 mt-0.5" />
                    <div className="flex-1 space-y-0.5">
                      <div className="flex justify-between font-medium text-muted">
                        <span>Bàn giao đơn vị vận chuyển</span>
                        <span className="text-[11px]">Chờ bước trước</span>
                      </div>
                      <p className="text-muted">Bắt đầu sau khi đóng gói xong</p>
                    </div>
                  </div>
                </div>
              </div>

              <div className="p-4 rounded-xl border border-border bg-surface-hover/30 text-xs text-muted space-y-1.5">
                <span className="font-semibold text-foreground">Theo dõi tập trung:</span>
                <p>
                  Chỉ cần xem bảng theo dõi, chủ doanh nghiệp biết đơn hàng đang ở bước nào và ai đang phụ trách, không cần hỏi từng người.
                </p>
              </div>
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
