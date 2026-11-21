"use client";

import * as React from "react";
import { ChevronDown, HelpCircle } from "lucide-react";

export const FAQS = [
  {
    question: "Hexta hiện đang ở giai đoạn phát triển nào?",
    answer:
      "Hexta đang trong giai đoạn phát triển và thử nghiệm. Chúng tôi đang kiểm chứng các quy trình cốt lõi (tiếp nhận hội thoại, đề xuất phản hồi và bàn giao công việc) trước khi phát hành rộng rãi.",
  },
  {
    question: "Hexta khác gì so với các phần mềm chat đa kênh hoặc CRM thông thường?",
    answer:
      "Phần lớn công cụ hiện nay dừng ở việc gộp tin nhắn hoặc lưu danh sách đơn hàng. Hexta đi tiếp một bước: tạo đơn nháp ngay từ hội thoại và chuyển đơn thành chuỗi công việc, mỗi bước có một người phụ trách. Nhờ vậy, đơn đã xác nhận không bị bỏ sót giữa các bộ phận.",
  },
  {
    question: "AI của Hexta có tự gửi tin sai hoặc tự xác nhận đơn không?",
    answer:
      "Không. AI chỉ soạn câu trả lời đề xuất và trích xuất thông tin để lập đơn nháp; nhân viên kiểm tra và bấm gửi. Khi nhân viên đã nhận hội thoại, hệ thống ngừng hoàn toàn việc tự động gửi tin.",
  },
  {
    question: "Hexta kết nối với những kênh nhắn tin nào?",
    answer:
      "Hexta đang thử nghiệm kết nối với các kênh nhắn tin phổ biến để đưa hội thoại về một giao diện chung. Danh sách kênh hỗ trợ sẽ được công bố khi sản phẩm phát hành chính thức.",
  },
  {
    question: "Hexta có quản lý kho, kế toán hay tính lương không?",
    answer:
      "Không. Hexta tập trung vào một bài toán: từ tư vấn đến hoàn tất đơn hàng, luôn rõ việc và rõ người phụ trách. Kế toán chuyên sâu, chấm công, tính lương hay thanh toán phức tạp nằm ngoài phạm vi, để sản phẩm gọn nhẹ và dễ dùng với doanh nghiệp nhỏ.",
  },
  {
    question: "Làm thế nào để liên hệ với Hexta?",
    answer:
      "Quý doanh nghiệp có thể xem phần Quy trình mẫu trên trang này hoặc gửi email tới contact@hexta.org để trao đổi về nhu cầu vận hành thực tế.",
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
            <span>Hỏi đáp</span>
          </div>

          <h2 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-foreground tracking-tight">
            Câu hỏi thường gặp
          </h2>

          <p className="text-base sm:text-lg text-muted">
            Thông tin về phạm vi sản phẩm, giới hạn của AI và định hướng phát triển của Hexta.
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
