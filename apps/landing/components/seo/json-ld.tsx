interface FaqItem {
  question: string;
  answer: string;
}

export function JsonLd({ faqs }: { faqs?: FaqItem[] }) {
  const organizationSchema = {
    "@context": "https://schema.org",
    "@type": "Organization",
    name: "Hexta",
    url: "https://hexta.org",
    logo: "https://hexta.org/brand/hexta-light.png",
    description:
      "Nền tảng quản lý vận hành bán hàng từ tư vấn, xác nhận đơn đến phân công công việc dành cho doanh nghiệp vừa và nhỏ.",
  };

  const websiteSchema = {
    "@context": "https://schema.org",
    "@type": "WebSite",
    name: "Hexta - Quản trị từ tư vấn đến hoàn tất đơn hàng",
    url: "https://hexta.org",
  };

  const softwareSchema = {
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "Hexta Operations Platform (Đang phát triển & thử nghiệm)",
    applicationCategory: "BusinessApplication",
    operatingSystem: "Web Browser",
    softwareVersion: "Giai đoạn thử nghiệm MVP",
  };

  const faqSchema =
    faqs && faqs.length > 0
      ? {
          "@context": "https://schema.org",
          "@type": "FAQPage",
          mainEntity: faqs.map((faq) => ({
            "@type": "Question",
            name: faq.question,
            acceptedAnswer: {
              "@type": "Answer",
              text: faq.answer,
            },
          })),
        }
      : null;

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(organizationSchema) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(websiteSchema) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(softwareSchema) }}
      />
      {faqSchema && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(faqSchema) }}
        />
      )}
    </>
  );
}
