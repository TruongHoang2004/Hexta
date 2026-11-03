import { MessageSquareOff, AlertCircle, RefreshCw, FileQuestion } from "lucide-react";

export function ProblemsSection() {
  const problems = [
    {
      icon: MessageSquareOff,
      title: "Tin nhắn phân tán, dễ sót khách",
      description:
        "Tư vấn trên nhiều kênh, khách nhắn hỏi dồn dập vào giờ cao điểm hoặc ngoài giờ làm việc. Tin nhắn trôi nhanh, nhân viên trả lời chậm khiến khách hàng chuyển sang đối thủ.",
    },
    {
      icon: FileQuestion,
      title: "Chốt đơn trên chat, ghi chép thủ công",
      description:
        "Nhân viên phải sao chép địa chỉ, số điện thoại và sản phẩm từ ô chat ra sổ tay hoặc file Excel. Sai sót thông tin xảy ra thường xuyên và mất nhiều thời gian kiểm tra lại.",
    },
    {
      icon: AlertCircle,
      title: "Đơn chốt xong không rõ ai làm tiếp",
      description:
        "Sale tạo đơn xong nhắn vào nhóm chung nhưng không ai nhận cụ thể. Khi khách hỏi tiến độ hoặc giao trễ, các bộ phận đùn đẩy trách nhiệm vì thiếu người đứng tên phụ trách.",
    },
    {
      icon: RefreshCw,
      title: "Dùng nhiều phần mềm rời rạc",
      description:
        "Một công cụ để chat, một phần mềm để lên đơn, một ứng dụng quản lý việc. Dữ liệu đứt gãy, chủ doanh nghiệp muốn biết một đơn hàng đang ở đâu phải đi hỏi từng nhân sự.",
    },
  ];

  return (
    <section id="problems" className="py-20 md:py-24 bg-surface/40 border-y border-border relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center max-w-3xl mx-auto space-y-4 mb-14">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-amber-500/10 text-amber-600 dark:text-amber-400 text-xs font-semibold">
            <span>Thực Trạng Phổ Biến</span>
          </div>
          <h2 className="text-3xl sm:text-4xl font-extrabold text-foreground tracking-tight">
            Khi quy trình bán hàng bị đứt đoạn, <br className="hidden sm:inline" />
            <span className="text-amber-600 dark:text-amber-400">doanh nghiệp chịu thiệt hại gì?</span>
          </h2>
          <p className="text-base sm:text-lg text-muted">
            Doanh nghiệp nhỏ thường không thiếu khách hàng, mà thiếu một cách vận hành có đầu mối và quy trình liền mạch.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {problems.map((prob, idx) => {
            const Icon = prob.icon;
            return (
              <div
                key={idx}
                className="p-6 rounded-2xl border border-border bg-surface flex flex-col justify-between space-y-4 shadow-2xs hover:border-amber-500/40 transition-colors"
              >
                <div className="w-11 h-11 rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center shrink-0">
                  <Icon className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-lg font-bold text-foreground mb-2">
                    {prob.title}
                  </h3>
                  <p className="text-xs sm:text-sm text-muted leading-relaxed">
                    {prob.description}
                  </p>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
