import { Briefcase, MessageSquare, Package, Check } from "lucide-react";

export function RolesValueSection() {
  const roles = [
    {
      icon: Briefcase,
      role: "Chủ doanh nghiệp",
      subtitle: "Nắm toàn bộ tiến độ, không bị đứt đoạn",
      benefits: [
        "Biết ngay đơn hàng đang ở bước nào và ai đang phụ trách",
        "Xác định rõ người chịu trách nhiệm khi giao hàng chậm hoặc có sự cố",
        "Dữ liệu khách hàng và đơn hàng được lưu tập trung",
        "Dễ dàng giao lại công việc khi có nhân sự vắng mặt",
      ],
    },
    {
      icon: MessageSquare,
      role: "Nhân viên tư vấn",
      subtitle: "Phản hồi nhanh hơn, lên đơn dễ dàng hơn",
      benefits: [
        "Có sẵn câu trả lời gợi ý theo ngữ cảnh, giảm thao tác gõ lặp lại",
        "Lập đơn nháp ngay trong khung chat, không cần sao chép sang Excel",
        "Nhận hội thoại rõ ràng, tránh trùng lặp khi chăm sóc khách",
        "Bàn giao đơn đã xác nhận cho bộ phận vận hành, có xác nhận từ người nhận",
      ],
    },
    {
      icon: Package,
      role: "Nhân viên vận hành / kho",
      subtitle: "Rõ việc cần làm, bàn giao có ghi nhận",
      benefits: [
        "Nhận việc qua danh sách phân công, không còn giao việc bằng lời",
        "Xác nhận nhận việc trước khi bắt đầu, phân định rõ trách nhiệm",
        "Xem đầy đủ ghi chú của khách từ khâu tư vấn",
        "Bàn giao sang bước tiếp theo đúng thứ tự",
      ],
    },
  ];

  return (
    <section id="roles" className="py-20 md:py-28 bg-surface/40 border-y border-border relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <span>Giá trị theo vai trò</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Khi trách nhiệm rõ ràng, <br className="hidden sm:inline" />
            <span className="text-primary">vận hành trở nên nhẹ nhàng hơn</span>
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Hexta mang lại lợi ích cụ thể cho từng vị trí trong quy trình bán hàng.
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {roles.map((item, idx) => {
            const Icon = item.icon;
            return (
              <div
                key={idx}
                className="p-8 rounded-2xl border border-border bg-surface flex flex-col justify-between space-y-6 shadow-xs hover:border-primary/40 transition-all"
              >
                <div>
                  <div className="w-12 h-12 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-6">
                    <Icon className="w-6 h-6" />
                  </div>

                  <h3 className="text-xl font-bold text-foreground mb-1">
                    {item.role}
                  </h3>

                  <p className="text-xs text-primary font-medium mb-6">
                    {item.subtitle}
                  </p>

                  <ul className="space-y-3.5 text-xs sm:text-sm text-muted">
                    {item.benefits.map((benefit, i) => (
                      <li key={i} className="flex items-start gap-2.5">
                        <Check className="w-4 h-4 text-primary shrink-0 mt-0.5" />
                        <span className="text-foreground/85 leading-relaxed">{benefit}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
