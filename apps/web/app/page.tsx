import Link from "next/link";
import { ArrowRight, Shield, Zap, LayoutDashboard, Smartphone } from "lucide-react";
import { Button } from "@hexta/ui";
import { AuthNav } from "@/components/auth-nav";

export default function LandingPage() {
  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans selection:bg-primary/20 selection:text-primary">
      {/* Navigation */}
      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur-md">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex gap-2 items-center">
            <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
              <span className="text-white font-bold text-xl leading-none">H</span>
            </div>
            <span className="font-bold text-xl tracking-tight">Hexta</span>
          </div>
          <nav className="hidden md:flex items-center gap-8 text-sm font-medium text-muted">
            <Link href="#features" className="hover:text-foreground transition-colors">Tính năng</Link>
            <Link href="#solutions" className="hover:text-foreground transition-colors">Giải pháp</Link>
            <Link href="#pricing" className="hover:text-foreground transition-colors">Bảng giá</Link>
          </nav>
          <AuthNav />
        </div>
      </header>

      <main className="flex-1 flex flex-col">
        {/* Hero Section */}
        <section className="relative px-6 py-24 md:py-32 overflow-hidden flex items-center justify-center text-center">
          {/* Background decorations */}
          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-primary/10 rounded-full blur-[100px] -z-10 pointer-events-none" />
          
          <div className="max-w-4xl mx-auto space-y-8 relative z-10">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-surface border border-border text-sm font-medium text-primary shadow-sm">
              <span className="flex h-2 w-2 rounded-full bg-primary animate-pulse" />
              Ra mắt phiên bản Hexta 2.0
            </div>
            
            <h1 className="text-5xl md:text-7xl font-bold tracking-tighter text-balance">
              Nền tảng quản lý <br className="hidden md:block" />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary to-blue-400">
                toàn diện cho doanh nghiệp
              </span>
            </h1>
            
            <p className="text-lg md:text-xl text-muted max-w-2xl mx-auto text-balance">
              Hexta cung cấp hệ sinh thái công cụ số mạnh mẽ giúp tối ưu hóa quy trình, tăng cường hiệu suất và thúc đẩy sự phát triển không giới hạn.
            </p>
            
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
              <Button asChild size="lg" className="w-full sm:w-auto h-14 px-8 text-base rounded-full hover:scale-105 active:scale-95 transition-all shadow-xl shadow-primary/20">
                <Link href="/register">Trải nghiệm miễn phí <ArrowRight className="w-4 h-4 ml-2" /></Link>
              </Button>
              <Button asChild variant="outline" size="lg" className="w-full sm:w-auto h-14 px-8 text-base rounded-full bg-surface hover:bg-muted/10">
                <Link href="#demo">Xem Demo</Link>
              </Button>
            </div>
          </div>
        </section>

        {/* Features / Bento Grid */}
        <section id="features" className="px-6 py-24 bg-surface/50 border-t border-border/50">
          <div className="max-w-6xl mx-auto space-y-16">
            <div className="text-center space-y-4 max-w-2xl mx-auto">
              <h2 className="text-3xl md:text-4xl font-bold tracking-tight">Mọi công cụ bạn cần ở cùng một nơi</h2>
              <p className="text-muted">Được thiết kế tỉ mỉ để mang lại trải nghiệm tối ưu nhất, giúp bạn quản lý mọi thứ từ dữ liệu đến quy trình một cách trơn tru.</p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Feature 1 */}
              <div className="col-span-1 md:col-span-2 bg-background border border-border rounded-3xl p-8 shadow-sm hover:shadow-md transition-shadow group">
                <div className="w-12 h-12 bg-primary/10 rounded-2xl flex items-center justify-center text-primary mb-6 group-hover:scale-110 transition-transform">
                  <LayoutDashboard className="w-6 h-6" />
                </div>
                <h3 className="text-2xl font-bold mb-3">Dashboard Trực Quan</h3>
                <p className="text-muted max-w-md">Theo dõi mọi chỉ số quan trọng của doanh nghiệp theo thời gian thực với giao diện thân thiện, dễ nhìn và có thể tùy biến hoàn toàn.</p>
              </div>

              {/* Feature 2 */}
              <div className="col-span-1 bg-background border border-border rounded-3xl p-8 shadow-sm hover:shadow-md transition-shadow group">
                <div className="w-12 h-12 bg-warning/10 rounded-2xl flex items-center justify-center text-warning mb-6 group-hover:scale-110 transition-transform">
                  <Zap className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold mb-3">Tốc Độ Tối Đa</h3>
                <p className="text-muted">Được tối ưu hóa bằng các công nghệ lõi tiên tiến nhất, mang đến độ trễ gần như bằng không.</p>
              </div>

              {/* Feature 3 */}
              <div className="col-span-1 bg-background border border-border rounded-3xl p-8 shadow-sm hover:shadow-md transition-shadow group">
                <div className="w-12 h-12 bg-success/10 rounded-2xl flex items-center justify-center text-success mb-6 group-hover:scale-110 transition-transform">
                  <Shield className="w-6 h-6" />
                </div>
                <h3 className="text-xl font-bold mb-3">Bảo Mật Tuyệt Đối</h3>
                <p className="text-muted">Dữ liệu được mã hóa hai chiều chuẩn quân đội, đảm bảo an toàn tuyệt đối trước mọi rủi ro.</p>
              </div>

              {/* Feature 4 */}
              <div className="col-span-1 md:col-span-2 bg-background border border-border rounded-3xl p-8 shadow-sm hover:shadow-md transition-shadow group">
                <div className="w-12 h-12 bg-primary/10 rounded-2xl flex items-center justify-center text-primary mb-6 group-hover:scale-110 transition-transform">
                  <Smartphone className="w-6 h-6" />
                </div>
                <h3 className="text-2xl font-bold mb-3">Đa Nền Tảng</h3>
                <p className="text-muted max-w-md">Làm việc hiệu quả trên mọi thiết bị từ Desktop, Tablet cho đến Mobile mà không gặp bất kỳ giới hạn nào về tính năng.</p>
              </div>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-border bg-background py-12 px-6">
        <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
          <div className="flex gap-2 items-center">
             <div className="w-6 h-6 rounded bg-foreground flex items-center justify-center">
              <span className="text-background font-bold text-sm leading-none">H</span>
            </div>
            <span className="font-bold text-lg tracking-tight">Hexta</span>
          </div>
          <p className="text-sm text-muted">© {new Date().getFullYear()} Hexta Inc. Mọi quyền được bảo lưu.</p>
          <div className="flex items-center gap-6 text-sm text-muted">
            <Link href="#" className="hover:text-foreground transition-colors">Điều khoản</Link>
            <Link href="#" className="hover:text-foreground transition-colors">Bảo mật</Link>
            <Link href="#" className="hover:text-foreground transition-colors">Hỗ trợ</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
