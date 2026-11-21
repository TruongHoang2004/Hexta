import { Check, ShieldCheck, UserCheck, MessageSquare } from "lucide-react";

export function ValuesSection() {
  const pillars = [
    {
      step: "01",
      icon: MessageSquare,
      title: "Hội thoại gắn với hồ sơ khách hàng",
      description:
        "Tin nhắn từ các kênh được tập trung về một nơi. Nhân viên nhận hội thoại là người chịu trách nhiệm trả lời; toàn bộ lịch sử tư vấn được lưu lại.",
    },
    {
      step: "02",
      icon: ShieldCheck,
      title: "Tạo đơn nháp ngay từ hội thoại",
      description:
        "Thông tin sản phẩm và địa chỉ được trích xuất từ hội thoại thành đơn nháp. Nhân viên kiểm tra và xác nhận trước khi đơn trở thành đơn chính thức.",
    },
    {
      step: "03",
      icon: UserCheck,
      title: "Bàn giao từng bước, rõ người phụ trách",
      description:
        "Sau khi đơn được xác nhận, công việc lần lượt chuyển qua chuẩn bị hàng, đóng gói và bàn giao vận chuyển. Trách nhiệm chỉ được chuyển giao khi người nhận xác nhận nhận việc.",
    },
  ];

  return (
    <section id="values" className="py-20 md:py-28 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <span>Giải pháp</span>
          </div>
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Một quy trình liền mạch, <br className="hidden sm:inline" />
            <span className="text-primary">từ lúc khách hỏi đến khi đơn hàng được bàn giao</span>
          </h2>
          <p className="text-base sm:text-lg text-muted">
            Hexta không phải phần mềm phức tạp với hàng trăm biểu mẫu. Chúng tôi tập trung giúp mỗi thành viên biết rõ mình cần làm gì tiếp theo.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {pillars.map((item, idx) => {
            const Icon = item.icon;
            return (
              <div
                key={idx}
                className="p-8 rounded-2xl border border-border bg-surface flex flex-col justify-between space-y-6 shadow-xs hover:border-primary/40 transition-all"
              >
                <div>
                  <div className="flex items-center justify-between mb-4">
                    <div className="w-12 h-12 rounded-xl bg-primary/10 text-primary flex items-center justify-center">
                      <Icon className="w-6 h-6" />
                    </div>
                    <span className="text-2xl font-black font-mono text-muted/30">
                      {item.step}
                    </span>
                  </div>
                  <h3 className="text-xl font-bold text-foreground mb-3">
                    {item.title}
                  </h3>
                  <p className="text-sm text-muted leading-relaxed">
                    {item.description}
                  </p>
                </div>
              </div>
            );
          })}
        </div>

        {/* Core Philosophy Banner */}
        <div className="mt-14 p-6 sm:p-8 rounded-2xl border border-primary/20 bg-primary/5 max-w-4xl mx-auto flex flex-col sm:flex-row items-center gap-6 text-center sm:text-left">
          <div className="w-14 h-14 rounded-2xl bg-primary text-primary-foreground flex items-center justify-center shrink-0 shadow-md">
            <Check className="w-7 h-7" />
          </div>
          <div className="space-y-1.5">
            <h3 className="text-lg font-bold text-foreground">
              Nguyên tắc vận hành
            </h3>
            <p className="text-sm text-muted leading-relaxed">
              Tại mỗi thời điểm, mỗi công việc chỉ có <strong className="text-foreground">một người chịu trách nhiệm</strong>. Khi bàn giao, người nhận phải xác nhận, nhờ đó không công việc nào bị bỏ sót giữa chừng.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
