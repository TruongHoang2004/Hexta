import { HeroSection } from "@/components/sections/hero-section";
import { ProblemsSection } from "@/components/sections/problems-section";
import { ValuesSection } from "@/components/sections/values-section";
import { DemoMockupSection } from "@/components/sections/demo-mockup-section";
import { ControlledAiSection } from "@/components/sections/controlled-ai-section";
import { RolesValueSection } from "@/components/sections/roles-value-section";
import { FaqSection, FAQS } from "@/components/sections/faq-section";
import { CtaSection } from "@/components/sections/cta-section";
import { JsonLd } from "@/components/seo/json-ld";

export default function HomePage() {
  return (
    <>
      {/* Search Engine Structured Data without fake ratings */}
      <JsonLd faqs={FAQS} />

      {/* Main Landing Sections */}
      <div className="relative overflow-hidden">
        <HeroSection />
        <ProblemsSection />
        <ValuesSection />
        <DemoMockupSection />
        <ControlledAiSection />
        <RolesValueSection />
        <FaqSection />
        <CtaSection />
      </div>
    </>
  );
}
