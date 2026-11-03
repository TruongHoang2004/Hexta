"use client";

import * as React from "react";
import Link from "next/link";
import { BrandLogo } from "./brand-logo";
import { ThemeToggle } from "@hexta/ui";
import { Menu, X, ChevronRight, ArrowDown } from "lucide-react";

const NAV_LINKS = [
  { name: "Vấn Đề", href: "/#problems" },
  { name: "Giải Pháp", href: "/#values" },
  { name: "Minh Họa Quy Trình", href: "/#demo" },
  { name: "AI Có Kiểm Soát", href: "/#controlled-ai" },
  { name: "Theo Vai Trò", href: "/#roles" },
  { name: "FAQ", href: "/#faq" },
];

export function Navbar() {
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const [scrolled, setScrolled] = React.useState(false);

  React.useEffect(() => {
    const handleScroll = () => {
      setScrolled(window.scrollY > 20);
    };
    window.addEventListener("scroll", handleScroll);
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const scrollToDemo = (e: React.MouseEvent<HTMLAnchorElement>) => {
    const el = document.getElementById("demo");
    if (el) {
      e.preventDefault();
      el.scrollIntoView({ behavior: "smooth" });
      setMobileMenuOpen(false);
    }
  };

  return (
    <header
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrolled
          ? "bg-background/90 backdrop-blur-md border-b border-border shadow-xs"
          : "bg-transparent border-b border-transparent"
      }`}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 md:h-20">
          {/* Brand Logo & Stage Badge */}
          <div className="flex items-center gap-3 md:gap-4">
            <BrandLogo className="w-28 md:w-32" />
            <span className="hidden lg:inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-muted/15 text-muted border border-border">
              Bản thử nghiệm
            </span>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center gap-5 lg:gap-7" aria-label="Main Navigation">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.name}
                href={link.href}
                className="text-xs lg:text-sm font-medium text-muted hover:text-foreground transition-colors"
              >
                {link.name}
              </Link>
            ))}
          </nav>

          {/* Right actions: ThemeToggle + CTA Khám phá */}
          <div className="hidden sm:flex items-center gap-3">
            <ThemeToggle />

            <a
              href="#demo"
              onClick={scrollToDemo}
              className="inline-flex items-center justify-center gap-1.5 px-4 py-2 text-xs md:text-sm font-semibold text-primary-foreground bg-primary hover:bg-primary-hover rounded-lg transition-all shadow-xs cursor-pointer"
            >
              <span>Khám phá Hexta</span>
              <ArrowDown className="w-3.5 h-3.5" />
            </a>
          </div>

          {/* Mobile hamburger menu button */}
          <div className="flex items-center gap-2 md:hidden">
            <ThemeToggle />
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="p-2 text-muted hover:text-foreground rounded-lg focus:outline-hidden focus:ring-2 focus:ring-primary"
              aria-label="Mở bảng điều hướng"
              aria-expanded={mobileMenuOpen}
            >
              {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile menu dropdown */}
      {mobileMenuOpen && (
        <div className="md:hidden border-b border-border bg-background/95 backdrop-blur-xl px-4 pt-3 pb-6 space-y-3">
          <div className="px-3 py-1 text-xs text-muted flex items-center justify-between">
            <span>Trạng thái sản phẩm:</span>
            <span className="px-2 py-0.5 rounded text-[11px] font-medium bg-muted/15 text-muted">
              Đang thử nghiệm
            </span>
          </div>
          <nav className="flex flex-col space-y-1">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.name}
                href={link.href}
                onClick={() => setMobileMenuOpen(false)}
                className="flex items-center justify-between px-3 py-2 text-sm font-medium text-muted hover:text-foreground rounded-md hover:bg-muted/10 transition-colors"
              >
                <span>{link.name}</span>
                <ChevronRight className="w-4 h-4 text-muted/50" />
              </Link>
            ))}
          </nav>

          <div className="pt-3 border-t border-border">
            <a
              href="#demo"
              onClick={scrollToDemo}
              className="w-full flex items-center justify-center gap-2 px-4 py-2.5 text-sm font-semibold text-primary-foreground bg-primary rounded-lg shadow-xs"
            >
              <span>Khám phá mô phỏng quy trình</span>
              <ArrowDown className="w-4 h-4" />
            </a>
          </div>
        </div>
      )}
    </header>
  );
}
