import { Briefcase, MessageSquare, Package, Check } from "lucide-react";

export function RolesValueSection() {
  const roles = [
    {
      icon: Briefcase,
      role: "Dành Cho Chủ Doanh Nghiệp",
      subtitle: "Kiểm soát toàn cảnh, không đứt đoạn",
      benefits: [
        "Nhìn thấy ngay đơn hàng đang ở bước nào và ai đang giữ việc",
        "Loại bỏ văn hóa đùn đẩy trách nhiệm khi xảy ra sự cố giao chậm",
        "Dữ liệu đơn hàng và khách hàng được ghi nhận tập trung, không phân mảnh",
        "Dễ dàng bàn giao hoặc tái phân công công việc khi có nhân sự vắng mặt",
      ],
    },
    {
      icon: MessageSquare,
      role: "Dành Cho Nhân Viên Tư Vấn (Sale)",
      subtitle: "Phản hồi nhanh hơn, chốt đơn nhẹ nhàng hơn",
      benefits: [
        "Có sẵn câu trả lời gợi ý sát ngữ cảnh, giảm tải gõ lặp lại",
        "Lập đơn hàng nháp ngay trong khung chat, không cần copy paste qua Excel",
        "Đứng tên nhận hội thoại rõ ràng, không lo tranh chấp khách hàng",
        "Chủ động chuyển giao đơn đã chốt sang bộ phận vận hành có xác nhận",
      ],
    },
    {
      icon: Package,
      role: "Dành Cho Nhân Viên Vận Hành / Kho",
      subtitle: "Rõ việc cần làm, bàn giao có bằng chứng",
      benefits: [
        "Nhận việc qua danh sách phân công rõ ràng, không còn 'giao việc miệng'",
        "Bấm xác nhận nhận việc để chính thức bắt đầu, bảo vệ trách nhiệm cá nhân",
        "Nắm đầy đủ ghi chú đơn hàng của khách từ khâu tư vấn ban đầu",
        "Bàn giao sang khâu tiếp theo theo thứ tự tuần tự chuẩn xác",
      ],
    },
  ];

  return (
    <section id="roles" className="py-20 md:py-28 bg-surface/40 border-y border-border relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-16">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <span>Giá Trị Cho Từng Vị Trí</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Khi Mọi Người Đều Rõ Trách Nhiệm, <br className="hidden sm:inline" />
            <span className="text-primary">Vận Hành Tự Nhiên Trở Nên Nhẹ Nhàng</span>
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Hexta mang lại lợi ích thiết thực cho từng thành viên trong chuỗi vận hành bán hàng của doanh nghiệp.
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
