# Hexta Landing Page (`hexta.org`)

Trang landing page chính thức cho nền tảng thương mại điện tử **Hexta**, xây dựng bằng **Next.js 16 (App Router)** kết hợp **Tailwind CSS v4**, thiết kế tối ưu hóa công cụ tìm kiếm (**SEO Max**), hỗ trợ đầy đủ các trang chính sách pháp lý tiêu chuẩn quốc tế và Việt Nam.

---

## 🌟 Điểm Nổi Bật & Tối Ưu SEO

1. **Chuẩn SEO Quốc Tế & Google Rich Snippets:**
   - **JSON-LD Schema Markup:** Định nghĩa cấu trúc chuẩn cho `Organization`, `WebSite`, `SoftwareApplication` và `FAQPage` (tự động tạo rich accordions trên Google Search).
   - **Tối ưu Thẻ Siêu Dữ Liệu (Metadata API):** Cấu hình `canonical URL`, `openGraph` (hình ảnh 1200x630, tiêu đề, mô tả), `twitter:card: summary_large_image`, `robots: index, follow`.
   - **Tự động sinh Sitemap & Robots:**
     - `/sitemap.xml`: Tự động cập nhật `changefreq`, `priority`, `lastmod` cho toàn bộ trang chủ và trang chính sách (`apps/landing/app/sitemap.ts`).
     - `/robots.txt`: Khai báo sitemap và cấu hình cho Googlebot / Bingbot (`apps/landing/app/robots.ts`).
   - **HTML Ngữ Nghĩa (Semantic HTML5):** Cấu trúc thẻ `header`, `nav`, `main`, `section`, `article`, `footer`, phân cấp `h1` $\to$ `h2` $\to$ `h3` chặt chẽ.

2. **Hệ Thống Trang Chính Sách Pháp Lý (Legal & Compliance):**
   - `/privacy`: **Chính sách bảo mật** tuân thủ GDPR, CCPA và Nghị định 13/2023/NĐ-CP (Bảo vệ dữ liệu cá nhân).
   - `/terms`: **Điều khoản dịch vụ** quy định bản quyền mã nguồn Apache 2.0, hạn mức API rate limiting và trách nhiệm pháp lý.
   - `/cookies`: **Chính sách Cookie** phân loại Cookie thiết yếu, Cookie tùy chọn giao diện và Cookie phân tích.
   - `/security`: **Bảo mật & Cam kết SLA 99.99%** chi tiết hóa cơ chế cô lập dữ liệu Multi-Tenant (PostgreSQL RLS), mã hóa AES-256 / TLS 1.3 và chính sách báo cáo lỗ hổng bảo mật.

3. **Giao Diện & Trải Nghiệm Người Dùng:**
   - Hỗ trợ chế độ Sáng / Tối (Dark / Light mode) mượt mà với `@hexta/ui` và `next-themes`.
   - Layout responsive hoàn hảo từ điện thoại di động đến màn hình 4K.
   - Mockup giao diện Terminal thời gian thực mô phỏng lưu lượng API Gateway (:8080), Kafka Event Stream và Qdrant Vector Search.

---

## 📁 Cấu Trúc Dự Án

```
apps/landing/
├── app/
│   ├── layout.tsx         # Root layout tích hợp ThemeProvider, Navbar, Footer & Meta SEO
│   ├── page.tsx           # Trang chủ hexta.org tổng hợp các section
│   ├── globals.css        # Cấu hình Tailwind CSS v4 & bảng màu design tokens
│   ├── sitemap.ts         # Sinh tự động /sitemap.xml
│   ├── robots.ts          # Sinh tự động /robots.txt
│   ├── privacy/page.tsx   # Trang Chính sách bảo mật
│   ├── terms/page.tsx     # Trang Điều khoản dịch vụ
│   ├── cookies/page.tsx   # Trang Chính sách Cookie
│   └── security/page.tsx  # Trang Bảo mật & SLA 99.99%
├── components/
│   ├── brand-logo.tsx     # Logo Hexta responsive chuyển đổi theo Dark/Light mode
│   ├── navbar.tsx         # Thanh điều hướng với Menu di động và nút CTA
│   ├── footer.tsx         # Chân trang với điều hướng hệ sinh thái và pháp lý
│   ├── legal-layout.tsx   # Bố cục 2 cột tiêu chuẩn cho các trang chính sách
│   ├── icons/             # Icon SVG chuẩn (GitHub, Twitter/X, LinkedIn)
│   ├── seo/
│   │   └── json-ld.tsx    # Thành phần nhúng Schema.org Structured Data
│   └── sections/
│       ├── hero-section.tsx         # Hero section & Interactive Terminal
│       ├── stats-section.tsx        # Chỉ số hiệu năng (Latency <15ms, 10k RPS)
│       ├── features-section.tsx     # 6 năng lực cốt lõi nền tảng
│       ├── architecture-section.tsx # Sơ đồ kiến trúc 3 tầng Monorepo
│       ├── solutions-section.tsx    # Bảng so sánh Hexta vs SaaS truyền thống
│       ├── pricing-section.tsx      # Bảng giá minh bạch 3 gói triển khai
│       ├── faq-section.tsx          # Câu hỏi thường gặp tương tác
│       └── cta-section.tsx          # Banner kêu gọi chuyển đổi hành động
└── public/
    └── brand/             # Logo và favicon chính thức của Hexta
```

---

## 🚀 Hướng Dẫn Chạy & Build

Tại thư mục gốc của monorepo:

```bash
# 1. Chạy môi trường phát triển (Port 3003)
pnpm --filter landing dev

# 2. Kiểm tra chất lượng mã nguồn (Linter)
pnpm --filter landing lint

# 3. Đóng gói bản dựng sản phẩm (Production Build)
pnpm --filter landing build

# 4. Khởi chạy bản dựng sản phẩm
pnpm --filter landing start
```

---

## 🌐 Cấu Hình Deploy Lên Tên Miền `hexta.org`

- **Vercel / Cloudflare Pages / Coolify:** Trỏ Root Directory vào `apps/landing`.
- **Cấu hình DNS:** Trỏ bản ghi `A` và `CNAME` của tên miền `hexta.org` về IP máy chủ hoặc CNAME của nền tảng triển khai.
- **Biến môi trường (nếu cần tùy biến):**
  - `NEXT_PUBLIC_APP_URL=https://hexta.org`
