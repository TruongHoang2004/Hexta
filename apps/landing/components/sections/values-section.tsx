import { Check, ShieldCheck, UserCheck, MessageSquare, ArrowRight } from "lucide-react";

export function ValuesSection() {
  const pillars = [
    {
      step: "01",
      icon: MessageSquare,
      title: "Hội thoại gắn liền với hồ sơ khách hàng",
      description:
        "Tập trung tin nhắn tư vấn từ các kênh vào một nơi. Nhân viên tiếp nhận là người trực tiếp chịu trách nhiệm trả lời; hệ thống ghi nhận đầy đủ lịch sử tư vấn.",
    },
    {
      step: "02",
      icon: ShieldCheck,
      title: "Tạo đơn nháp ngay trong ngữ cảnh hội thoại",
      description:
        "Thay vì nhập liệu thủ công rời rạc, thông tin sản phẩm và địa chỉ được trích xuất để lập đơn nháp chờ kiểm tra. Người bán xác nhận trước khi đơn trở thành đơn chính thức.",
    },
    {
      step: "03",
      icon: UserCheck,
      title: "Công việc tuần tự, rõ người đứng tên",
      description:
        "Khi đơn hàng được xác nhận, công việc được chuyển sang các bước tiếp theo (chuẩn bị hàng, đóng gói, bàn giao). Người phụ trách mới phải bấm nhận việc để chuyển quyền.",
    },
  ];

  return (
    <section id="values" className="py-20 md:py-28 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <span>Giá Trị Cốt Lõi</span>
          </div>
          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Một dòng chảy liền mạch từ khi khách hỏi <br className="hidden sm:inline" />
            <span className="text-primary">đến lúc đơn hàng được giao xong</span>
          </h2>
          <p className="text-base sm:text-lg text-muted">
            Hexta không cố gắng trở thành một phần mềm phức tạp với hàng trăm biểu mẫu. Chúng tôi tập trung vào việc giúp mọi người trong đội ngũ luôn biết mình cần làm gì tiếp theo.
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

                <div className="pt-4 border-t border-border/60 flex items-center gap-2 text-xs font-semibold text-primary">
                  <span>Quy trình minh bạch</span>
                  <ArrowRight className="w-3.5 h-3.5" />
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
              Nguyên tắc vận hành cốt lõi:
            </h3>
            <p className="text-sm text-muted leading-relaxed">
              Mỗi công việc tại một thời điểm chỉ có <strong className="text-foreground">duy nhất một người chịu trách nhiệm</strong>. Khi bàn giao sang khâu tiếp theo, người mới phải xác nhận nhận việc — không còn tình trạng việc bị bỏ quên giữa chừng.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}
