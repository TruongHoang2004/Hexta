import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { ThemeProvider } from "@hexta/ui";
import { Navbar } from "@/components/navbar";
import { Footer } from "@/components/footer";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://hexta.org"),
  title: {
    default: "Hexta | Từ tư vấn đến hoàn tất đơn hàng. Luôn rõ việc, rõ người phụ trách.",
    template: "%s | Hexta",
  },
  description:
    "Hexta kết nối tư vấn khách hàng, tạo đơn và bàn giao công việc trên cùng một nền tảng. Mỗi bước đều có người phụ trách rõ ràng; AI hỗ trợ soạn tin và lập đơn nháp, nhân viên duyệt trước khi gửi.",
  keywords: [
    "hexta",
    "quản lý đơn hàng",
    "quy trình bán hàng",
    "tư vấn khách hàng",
    "bàn giao công việc",
    "vận hành doanh nghiệp nhỏ",
    "rõ người phụ trách",
  ],
  authors: [{ name: "Hexta Team", url: "https://hexta.org" }],
  creator: "Hexta Team",
  publisher: "Hexta",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  alternates: {
    canonical: "https://hexta.org",
  },
  openGraph: {
    type: "website",
    locale: "vi_VN",
    url: "https://hexta.org",
    siteName: "Hexta Platform",
    title: "Hexta | Từ tư vấn đến hoàn tất đơn hàng. Luôn rõ việc, rõ người phụ trách.",
    description:
      "Nền tảng kết nối tư vấn khách hàng, tạo đơn và bàn giao công việc dành cho doanh nghiệp nhỏ. AI đề xuất, nhân viên duyệt.",
    images: [
      {
        url: "/brand/hexta-dark.png",
        width: 1200,
        height: 630,
        alt: "Hexta - Từ tư vấn đến hoàn tất đơn hàng",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Hexta | Từ tư vấn đến hoàn tất đơn hàng. Luôn rõ việc, rõ người phụ trách.",
    description:
      "Nền tảng kết nối tư vấn khách hàng, tạo đơn và bàn giao công việc dành cho doanh nghiệp nhỏ.",
    images: ["/brand/hexta-dark.png"],
  },
  icons: {
    icon: "/brand/hexta-icon.png",
    shortcut: "/brand/hexta-icon.png",
    apple: "/brand/hexta-icon.png",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="vi"
      suppressHydrationWarning
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-background text-foreground">
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <Navbar />
          <main className="flex-1">{children}</main>
          <Footer />
        </ThemeProvider>
      </body>
    </html>
  );
}
