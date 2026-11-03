"use client";

import * as React from "react";
import { ChevronDown, HelpCircle } from "lucide-react";

export const FAQS = [
  {
    question: "Hexta hiện đang ở giai đoạn phát triển nào?",
    answer:
      "Hexta hiện đang trong giai đoạn phát triển, thử nghiệm nội bộ và mô phỏng các quy trình nghiệp vụ cốt lõi. Chúng tôi đang kiểm thử thực tế các kịch bản tiếp nhận hội thoại, đề xuất phản hồi và bàn giao công việc tuần tự trước khi mở rộng phiên bản công khai.",
  },
  {
    question: "Hexta khác gì so với các phần mềm chat đa kênh hoặc CRM thông thường?",
    answer:
      "Phần lớn phần mềm hiện tại dừng lại ở việc gộp tin nhắn hoặc lưu trữ danh sách đơn hàng. Hexta giải quyết nút thắt tiếp theo: kết nối trực tiếp hội thoại với khâu tạo đơn nháp, và chuyển đơn thành chuỗi công việc có người đứng tên phụ trách cụ thể. Không còn tình trạng đơn chốt xong bị bỏ quên giữa các bộ phận.",
  },
  {
    question: "AI của Hexta có tự ý gửi tin nhắn sai hoặc tự ý chốt đơn của khách không?",
    answer:
      "Tuyệt đối không. Hexta xây dựng AI theo nguyên tắc có con người kiểm duyệt (Human-in-the-loop). AI chỉ hỗ trợ soạn câu trả lời đề xuất và bóc tách thông tin lập đơn nháp. Nhân viên tư vấn là người kiểm tra và bấm gửi cuối cùng. Khi nhân viên đã tiếp nhận hội thoại, cơ chế tự động gửi của hệ thống sẽ dừng lại hoàn toàn.",
  },
  {
    question: "Hexta kết nối với Facebook, Zalo hay Pancake như thế nào?",
    answer:
      "Trong giai đoạn MVP, Pancake được nghiên cứu như phương án cầu nối kỹ thuật để lấy hội thoại từ các kênh mạng xã hội về một giao diện tập trung. Đây là hướng tiếp cận phục vụ thử nghiệm, Hexta chưa tuyên bố tính năng tích hợp thực tế đã vận hành đầy đủ trên quy mô lớn.",
  },
  {
    question: "Hexta có bao gồm quản lý kho đầy đủ, kế toán thuế hay tính lương không?",
    answer:
      "Không. Hexta tập trung tối đa vào việc giải quyết bài toán: 'Từ tư vấn đến hoàn tất đơn hàng, luôn rõ việc và rõ người phụ trách'. Chúng tôi chủ động không đưa vào các tính năng như kế toán chuyên sâu, chấm công, tính lương hay thanh toán phức tạp để giữ sản phẩm gọn nhẹ và dễ tiếp cận cho doanh nghiệp nhỏ.",
  },
  {
    question: "Làm thế nào để tôi có thể trao đổi hoặc tìm hiểu thêm về Hexta?",
    answer:
      "Bạn có thể cuộn lên xem phần Trải nghiệm minh họa quy trình trên trang này, hoặc liên hệ trực tiếp với đội ngũ phát triển qua email contact@hexta.org để trao đổi về nhu cầu vận hành thực tế của doanh nghiệp bạn.",
  },
];

export function FaqSection() {
  const [openIndex, setOpenIndex] = React.useState<number | null>(0);

  const toggle = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  const handleKeyDown = (e: React.KeyboardEvent, index: number) => {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      toggle(index);
    }
  };

  return (
    <section id="faq" className="py-20 md:py-28 relative">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="text-center space-y-4 mb-14">
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-primary/10 text-primary text-xs font-semibold">
            <HelpCircle className="w-3.5 h-3.5" />
            <span>Giải Đáp Thắc Mắc</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Câu Hỏi Thường Gặp
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Minh bạch về phạm vi sản phẩm, giới hạn công nghệ và định hướng phát triển của Hexta.
          </p>
        </div>

        <div className="space-y-4">
          {FAQS.map((faq, index) => {
            const isOpen = openIndex === index;
            const contentId = `faq-content-${index}`;
            const buttonId = `faq-button-${index}`;

            return (
              <div
                key={index}
                className="rounded-2xl border border-border bg-surface transition-all overflow-hidden shadow-2xs"
              >
                <button
                  id={buttonId}
                  type="button"
                  onClick={() => toggle(index)}
                  onKeyDown={(e) => handleKeyDown(e, index)}
                  className="w-full flex items-center justify-between p-5 sm:p-6 text-left font-semibold text-foreground hover:text-primary transition-colors focus:outline-hidden focus:ring-2 focus:ring-primary cursor-pointer"
                  aria-expanded={isOpen}
                  aria-controls={contentId}
                >
                  <span className="text-sm sm:text-base pr-4 leading-snug">{faq.question}</span>
                  <ChevronDown
                    className={`w-5 h-5 shrink-0 text-muted transition-transform duration-200 ${
                      isOpen ? "rotate-180 text-primary" : ""
                    }`}
                  />
                </button>

                {isOpen && (
                  <div
                    id={contentId}
                    role="region"
                    aria-labelledby={buttonId}
                    className="px-5 sm:px-6 pb-6 text-xs sm:text-sm text-muted leading-relaxed border-t border-border/50 pt-4 animate-in fade-in duration-200"
                  >
                    {faq.answer}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
