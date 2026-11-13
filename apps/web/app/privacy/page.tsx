import type { Metadata } from "next";
import Link from "next/link";
import { BrandLogo } from "@/components/brand-logo";
import { AuthNav } from "@/components/auth-nav";
import { Button } from "@hexta/ui";
import {
  ShieldCheck,
  Lock,
  Cpu,
  Database,
  EyeOff,
  UserCheck,
  FileText,
  Mail,
  ArrowLeft,
  ChevronRight,
  Server,
  Layers,
} from "lucide-react";

export const metadata: Metadata = {
  title: "Privacy Policy — Hexta",
  description:
    "Learn how Hexta collects, isolates, and protects customer, tenant, and business data across our multi-tenant sales and AI runtime platform.",
};

export default function PrivacyPolicyPage() {
  const lastUpdated = "November 3, 2026";
  const effectiveDate = "November 3, 2026";

  const sections = [
    { id: "overview", label: "1. Overview & Scope" },
    { id: "information-collected", label: "2. Information We Collect" },
    { id: "usage", label: "3. How We Use Your Information" },
    { id: "tenant-isolation", label: "4. Multi-Tenant Isolation & Boundaries" },
    { id: "ai-policy", label: "5. AI & Large Language Model Policy" },
    { id: "sharing-subprocessors", label: "6. Data Sharing & Subprocessors" },
    { id: "retention-deletion", label: "7. Retention & Deletion" },
    { id: "privacy-rights", label: "8. Your Privacy Rights (GDPR / CCPA)" },
    { id: "security-audits", label: "9. Security & Audit Logging" },
    { id: "cookies-storage", label: "10. Cookies & Session Storage" },
    { id: "international-transfers", label: "11. International Transfers" },
    { id: "changes", label: "12. Changes to this Policy" },
    { id: "contact", label: "13. Contact & Inquiries" },
  ];

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans selection:bg-primary/20 selection:text-primary">
      {/* Top Navigation */}
      <header className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur-md">
        <div className="container mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link href="/" aria-label="Hexta home">
              <BrandLogo />
            </Link>
            <div className="hidden sm:flex items-center text-xs font-medium text-muted">
              <ChevronRight className="w-3.5 h-3.5 mx-1" />
              <span className="text-foreground">Privacy Policy</span>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Button asChild variant="ghost" size="sm" className="hidden sm:inline-flex">
              <Link href="/">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Home
              </Link>
            </Button>
            <AuthNav />
          </div>
        </div>
      </header>

      {/* Hero Banner */}
      <section className="relative px-6 py-16 md:py-20 border-b border-border/40 overflow-hidden bg-surface/30">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[500px] bg-primary/10 rounded-full blur-[120px] -z-10 pointer-events-none" />
        
        <div className="max-w-4xl mx-auto space-y-6 text-center">
          <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-surface border border-border text-xs font-semibold text-primary shadow-sm">
            <ShieldCheck className="w-4 h-4" />
            <span>Data Protection & Privacy Standard</span>
          </div>

          <h1 className="text-4xl md:text-5xl font-bold tracking-tight text-balance">
            Hexta Privacy Policy
          </h1>

          <p className="text-base md:text-lg text-muted max-w-2xl mx-auto text-balance">
            We are committed to safeguarding your business data through strict multi-tenant boundary isolation, modern cryptographic security, and transparent AI governance.
          </p>

          <div className="flex flex-wrap items-center justify-center gap-4 text-xs text-muted pt-2">
            <span className="bg-surface px-3 py-1 rounded-md border border-border">
              <strong>Effective Date:</strong> {effectiveDate}
            </span>
            <span className="bg-surface px-3 py-1 rounded-md border border-border">
              <strong>Last Updated:</strong> {lastUpdated}
            </span>
            <span className="bg-surface px-3 py-1 rounded-md border border-border">
              <strong>Version:</strong> 2.0
            </span>
          </div>
        </div>
      </section>

      {/* High-level Pillars */}
      <section className="px-6 py-10 border-b border-border/40 bg-surface/20">
        <div className="max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="p-5 rounded-2xl bg-surface border border-border shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-3">
              <Layers className="w-5 h-5" />
            </div>
            <h3 className="font-semibold text-sm mb-1">Strict Tenant Isolation</h3>
            <p className="text-xs text-muted leading-relaxed">
              Every tenant operates within a cryptographically and logically isolated workspace boundary.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-surface border border-border shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-3">
              <EyeOff className="w-5 h-5" />
            </div>
            <h3 className="font-semibold text-sm mb-1">Zero AI Training</h3>
            <p className="text-xs text-muted leading-relaxed">
              Your private business data, conversations, and records are never used to train foundational AI models.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-surface border border-border shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-3">
              <Lock className="w-5 h-5" />
            </div>
            <h3 className="font-semibold text-sm mb-1">Enterprise Encryption</h3>
            <p className="text-xs text-muted leading-relaxed">
              All communications are encrypted using TLS 1.3 in-transit and AES-256 for data at-rest.
            </p>
          </div>

          <div className="p-5 rounded-2xl bg-surface border border-border shadow-xs">
            <div className="w-10 h-10 rounded-xl bg-primary/10 text-primary flex items-center justify-center mb-3">
              <UserCheck className="w-5 h-5" />
            </div>
            <h3 className="font-semibold text-sm mb-1">Role-Based Access</h3>
            <p className="text-xs text-muted leading-relaxed">
              Granular tenant RBAC ensures team members and operators only access what they are authorized to view.
            </p>
          </div>
        </div>
      </section>

      {/* Main Content Area with Sticky Navigation */}
      <main className="flex-1 px-6 py-12">
        <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-12 gap-10">
          {/* Table of Contents - Sidebar */}
          <aside className="lg:col-span-4 hidden lg:block">
            <div className="sticky top-24 space-y-4 p-5 rounded-2xl bg-surface/50 border border-border">
              <div className="text-xs font-bold uppercase tracking-wider text-muted">
                Table of Contents
              </div>
              <nav className="space-y-1">
                {sections.map((sec) => (
                  <a
                    key={sec.id}
                    href={`#${sec.id}`}
                    className="block px-3 py-1.5 text-xs text-muted hover:text-foreground hover:bg-surface rounded-lg transition-colors"
                  >
                    {sec.label}
                  </a>
                ))}
              </nav>

              <div className="pt-4 border-t border-border/60">
                <p className="text-xs text-muted mb-2">Have privacy questions?</p>
                <a
                  href="mailto:privacy@hexta.io"
                  className="inline-flex items-center text-xs font-semibold text-primary hover:underline"
                >
                  <Mail className="w-3.5 h-3.5 mr-1.5" />
                  privacy@hexta.io
                </a>
              </div>
            </div>
          </aside>

          {/* Legal and Technical Articles */}
          <article className="lg:col-span-8 space-y-12 text-sm leading-relaxed text-foreground/90">
            {/* 1. Overview */}
            <section id="overview" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <FileText className="w-6 h-6 text-primary" />
                1. Overview & Scope
              </h2>
              <p>
                This Privacy Policy applies to all services, web applications, SDKs, APIs, and AI runtimes provided by <strong>Hexta Inc.</strong> (&ldquo;Hexta&rdquo;, &ldquo;we&rdquo;, &ldquo;us&rdquo;, or &ldquo;our&rdquo;), including our enterprise multi-tenant management portal located at <code>hexta.io</code> and its associated subdomains.
              </p>
              <p>
                Hexta operates as a multi-tenant business-to-business (B2B) platform offering sales workflow orchestration, customer relationship management, catalog and order fulfillment, automated AI agents, and messaging hubs. In this context:
              </p>
              <ul className="list-disc pl-5 space-y-1.5 text-muted">
                <li>
                  <strong className="text-foreground">Data Controller:</strong> For account registration credentials (email, authentication identifiers, billing owner identity), Hexta acts as the Data Controller.
                </li>
                <li>
                  <strong className="text-foreground">Data Processor:</strong> For all data ingested, stored, processed, or generated inside a tenant workspace (such as customer profiles, product catalogs, order manifests, and messaging transcripts), our customer (&ldquo;Tenant&rdquo;) is the Data Controller, and Hexta acts as a Data Processor.
                </li>
              </ul>
            </section>

            {/* 2. Information We Collect */}
            <section id="information-collected" className="scroll-mt-24 space-y-4">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Database className="w-6 h-6 text-primary" />
                2. Information We Collect
              </h2>
              <p>
                We collect personal and business information strictly necessary to provide our services and maintain platform integrity. The categories of information include:
              </p>

              <div className="space-y-3">
                <div className="p-4 rounded-xl bg-surface border border-border">
                  <h3 className="font-semibold text-foreground mb-1">A. User & Identity Information</h3>
                  <p className="text-xs text-muted mb-2">
                    When you register an account or sign in via Google OAuth:
                  </p>
                  <ul className="list-disc pl-5 space-y-1 text-xs text-muted">
                    <li>Email address, full name, and avatar URL (if provided via OAuth).</li>
                    <li>Cryptographically salted password hashes (we never store plain-text passwords).</li>
                    <li>Assigned role and workspace membership permissions (e.g. Owner, Operator, Member).</li>
                  </ul>
                </div>

                <div className="p-4 rounded-xl bg-surface border border-border">
                  <h3 className="font-semibold text-foreground mb-1">B. Tenant & Organization Profile Data</h3>
                  <p className="text-xs text-muted mb-2">
                    When a tenant is registered and onboarded onto Hexta:
                  </p>
                  <ul className="list-disc pl-5 space-y-1 text-xs text-muted">
                    <li>Business name, legal entity identifier, unique workspace slug, and domain.</li>
                    <li>Subscription tier, operational status, review records, and approval history.</li>
                  </ul>
                </div>

                <div className="p-4 rounded-xl bg-surface border border-border">
                  <h3 className="font-semibold text-foreground mb-1">C. Business Workspace Content</h3>
                  <p className="text-xs text-muted mb-2">
                    Data created or managed within your authorized workspace:
                  </p>
                  <ul className="list-disc pl-5 space-y-1 text-xs text-muted">
                    <li>Customer contacts, company names, transaction records, and order line items.</li>
                    <li>Product catalogs, SKU data, inventory counts, pricing tiers, and metadata.</li>
                    <li>Messaging threads, AI chat interactions, execution traces, and workflow run outputs.</li>
                  </ul>
                </div>

                <div className="p-4 rounded-xl bg-surface border border-border">
                  <h3 className="font-semibold text-foreground mb-1">D. Technical Telemetry & Operational Logs</h3>
                  <p className="text-xs text-muted mb-2">
                    Collected automatically to maintain system security and availability:
                  </p>
                  <ul className="list-disc pl-5 space-y-1 text-xs text-muted">
                    <li>IP addresses, user agent strings, browser version, and operating system.</li>
                    <li>API request timestamps, response latency, HTTP status codes, and error diagnostics.</li>
                    <li>Security audit events, including login attempts, privilege escalations, and API token generations.</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* 3. Usage */}
            <section id="usage" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <UserCheck className="w-6 h-6 text-primary" />
                3. How We Use Your Information
              </h2>
              <p>We process your data exclusively under legitimate legal bases, including contractual necessity, legal compliance, and legitimate interests:</p>
              
              <div className="overflow-x-auto">
                <table className="w-full text-left border border-border rounded-xl overflow-hidden text-xs">
                  <thead className="bg-surface/80 border-b border-border text-foreground font-semibold">
                    <tr>
                      <th className="p-3">Purpose of Processing</th>
                      <th className="p-3">Data Categories</th>
                      <th className="p-3">Legal Basis</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border text-muted">
                    <tr>
                      <td className="p-3 font-medium text-foreground">Service delivery & workspace provisioning</td>
                      <td className="p-3">Account identity, tenant profile, workspace records</td>
                      <td className="p-3">Contractual necessity</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium text-foreground">AI agent execution & sales automation</td>
                      <td className="p-3">Prompts, message threads, business context</td>
                      <td className="p-3">Contractual necessity</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium text-foreground">Platform security, fraud detection, tenant isolation</td>
                      <td className="p-3">IP addresses, audit logs, authentication tokens</td>
                      <td className="p-3">Legitimate interest / Legal duty</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-medium text-foreground">Operational notices, incident alerting & customer support</td>
                      <td className="p-3">User email, operator communication history</td>
                      <td className="p-3">Contractual necessity</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

            {/* 4. Multi-Tenant Isolation */}
            <section id="tenant-isolation" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Layers className="w-6 h-6 text-primary" />
                4. Multi-Tenant Isolation & Boundaries
              </h2>
              <p>
                Multi-tenancy security is the core foundational principle of Hexta. Unlike legacy shared platforms, Hexta implements rigorous defense-in-depth boundaries:
              </p>
              <div className="p-5 rounded-2xl bg-surface border border-border space-y-3">
                <div className="flex items-start gap-3">
                  <div className="w-8 h-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0 mt-0.5">
                    <Server className="w-4 h-4" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-foreground text-sm">Tenant Scoping at Storage & API Layers</h4>
                    <p className="text-xs text-muted mt-1">
                      Every database query, gRPC interaction, and HTTP endpoint enforces tenant authorization. All relational tables carry explicit <code>tenant_id</code> foreign keys. Cross-tenant queries are blocked at both application middleware and database boundaries.
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3">
                  <div className="w-8 h-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0 mt-0.5">
                    <Lock className="w-4 h-4" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-foreground text-sm">Role-Based Access Control (RBAC)</h4>
                    <p className="text-xs text-muted mt-1">
                      Access to resources within a tenant is partitioned by role (Owner, Operator, Member). Users cannot inspect data from another tenant, even if they share an identical user account across multiple organizations.
                    </p>
                  </div>
                </div>
              </div>
            </section>

            {/* 5. AI Policy */}
            <section id="ai-policy" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Cpu className="w-6 h-6 text-primary" />
                5. AI & Large Language Model Policy
              </h2>
              <div className="p-4 rounded-xl bg-primary/5 border border-primary/20 text-foreground">
                <div className="flex items-center gap-2 font-semibold text-primary mb-1">
                  <ShieldCheck className="w-5 h-5" />
                  <span>Our Zero-Training Commitment</span>
                </div>
                <p className="text-xs text-muted">
                  Hexta strictly enforces enterprise-tier privacy with all AI model providers. <strong>Your private business data, customer messages, order details, and sales transcripts are NEVER used to train, retrain, or improve foundational AI models.</strong>
                </p>
              </div>

              <p>
                When you activate AI capabilities (such as automated customer responses, inventory recommendations, or sales insights), the following protocols apply:
              </p>
              <ul className="list-disc pl-5 space-y-1.5 text-muted text-xs">
                <li>
                  <strong className="text-foreground">Stateless Inference:</strong> Prompts sent to model providers are processed in ephemeral memory for the duration of the API call and are discarded immediately upon completion.
                </li>
                <li>
                  <strong className="text-foreground">Provider Selection & Verification:</strong> All underlying AI providers (such as Google Gemini Enterprise and Anthropic) operate under signed Business Associate / Data Processing Agreements prohibiting model fine-tuning on customer input.
                </li>
                <li>
                  <strong className="text-foreground">Context Scoping:</strong> Only the minimal necessary context relevant to a specific execution run is injected into AI prompts.
                </li>
              </ul>
            </section>

            {/* 6. Data Sharing */}
            <section id="sharing-subprocessors" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <EyeOff className="w-6 h-6 text-primary" />
                6. Data Sharing & Subprocessors
              </h2>
              <p>
                <strong>We do not sell, rent, or trade your personal or tenant data to third parties.</strong> Data is shared only with verified technical subprocessors required to host, execute, and deliver our platform:
              </p>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <span className="font-semibold text-foreground">Cloud Compute & Database</span>
                  <p className="text-muted mt-1">Hosting infrastructure, managed PostgreSQL, Redis caching, and cold storage backups.</p>
                </div>
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <span className="font-semibold text-foreground">AI Intelligence Providers</span>
                  <p className="text-muted mt-1">Google Cloud (Gemini) with enterprise zero-retention and zero-training policies.</p>
                </div>
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <span className="font-semibold text-foreground">Authentication Providers</span>
                  <p className="text-muted mt-1">Google OAuth for optional single-sign-on (SSO) authentication.</p>
                </div>
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <span className="font-semibold text-foreground">Legal & Regulatory Authorities</span>
                  <p className="text-muted mt-1">Only where strictly mandated by valid legal court orders, warrants, or applicable law.</p>
                </div>
              </div>
            </section>

            {/* 7. Retention & Deletion */}
            <section id="retention-deletion" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Database className="w-6 h-6 text-primary" />
                7. Data Retention & Deletion
              </h2>
              <p>
                We retain tenant data for as long as your workspace account remains active and in good standing.
              </p>
              <ul className="list-disc pl-5 space-y-1.5 text-muted text-xs">
                <li>
                  <strong className="text-foreground">Workspace Termination:</strong> Upon cancellation or deletion of a workspace by its Owner, all tenant-scoped operational data (customers, products, orders, chat threads) is scheduled for soft-deletion for 30 days to protect against accidental loss, after which it is permanently purged.
                </li>
                <li>
                  <strong className="text-foreground">Right to be Forgotten:</strong> Individual users can request deletion of their personal identity profile at any time by contacting our support team.
                </li>
                <li>
                  <strong className="text-foreground">Backup Rotation:</strong> Encrypted backup snapshots are retained in immutable archival storage for up to 90 days before automated rotation.
                </li>
              </ul>
            </section>

            {/* 8. Privacy Rights */}
            <section id="privacy-rights" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <UserCheck className="w-6 h-6 text-primary" />
                8. Your Privacy Rights (GDPR & CCPA)
              </h2>
              <p>
                Depending on your location (such as the European Economic Area, United Kingdom, or State of California), you hold statutory privacy rights:
              </p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <strong className="text-foreground">Right of Access & Portability</strong>
                  <p className="text-muted mt-0.5">Request a complete copy of your workspace data in structured, machine-readable format (JSON/CSV).</p>
                </div>
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <strong className="text-foreground">Right to Rectification</strong>
                  <p className="text-muted mt-0.5">Correct inaccurate, incomplete, or outdated business or profile records.</p>
                </div>
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <strong className="text-foreground">Right to Erasure (Deletion)</strong>
                  <p className="text-muted mt-0.5">Request permanent erasure of personal data held by Hexta.</p>
                </div>
                <div className="p-3 rounded-xl bg-surface border border-border">
                  <strong className="text-foreground">Right to Restriction & Objection</strong>
                  <p className="text-muted mt-0.5">Restrict or object to processing carried out under legitimate interest grounds.</p>
                </div>
              </div>
              <p className="text-xs text-muted">
                To exercise any of these rights, email us at <a href="mailto:privacy@hexta.io" className="text-primary hover:underline font-medium">privacy@hexta.io</a>. We respond to all verified requests within 30 calendar days without discrimination.
              </p>
            </section>

            {/* 9. Security & Audits */}
            <section id="security-audits" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Lock className="w-6 h-6 text-primary" />
                9. Security Practices & Audit Logging
              </h2>
              <p>
                Hexta employs robust organizational, physical, and technical controls to safeguard information against unauthorized disclosure, accidental destruction, and misuse:
              </p>
              <ul className="list-disc pl-5 space-y-1 text-xs text-muted">
                <li><strong>Transport Security:</strong> Strict HTTPS with TLS 1.3 encryption across all public web and API endpoints.</li>
                <li><strong>At-Rest Protection:</strong> Enterprise AES-256 encryption across all primary databases and cold backups.</li>
                <li><strong>Credential Protection:</strong> Passwords hashed with state-of-the-art key derivation algorithms (Argon2 / bcrypt).</li>
                <li><strong>Audit Trails:</strong> Real-time logging of authentication attempts, tenant membership changes, and administrative actions.</li>
              </ul>
            </section>

            {/* 10. Cookies */}
            <section id="cookies-storage" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Server className="w-6 h-6 text-primary" />
                10. Cookies & Session Storage
              </h2>
              <p>
                Hexta does not deploy intrusive third-party advertising trackers or commercial remarketing cookies. We use only strictly necessary first-party cookies and browser storage:
              </p>
              <div className="overflow-x-auto">
                <table className="w-full text-left border border-border rounded-xl overflow-hidden text-xs">
                  <thead className="bg-surface/80 border-b border-border text-foreground font-semibold">
                    <tr>
                      <th className="p-3">Identifier</th>
                      <th className="p-3">Type</th>
                      <th className="p-3">Purpose</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border text-muted">
                    <tr>
                      <td className="p-3 font-mono text-foreground">access_token / session</td>
                      <td className="p-3">HttpOnly Cookie</td>
                      <td className="p-3">Authenticates your active session securely against cross-site scripting (XSS).</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-foreground">hexta.workspace</td>
                      <td className="p-3">SessionStorage</td>
                      <td className="p-3">Maintains currently selected tenant workspace for navigation.</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-foreground">theme</td>
                      <td className="p-3">LocalStorage</td>
                      <td className="p-3">Stores your dark/light UI preference.</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

            {/* 11. International Transfers */}
            <section id="international-transfers" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <ShieldCheck className="w-6 h-6 text-primary" />
                11. International Data Transfers
              </h2>
              <p>
                Our infrastructure and servers may be situated globally. Where data is transferred across international boundaries, Hexta ensures standard contractual clauses (SCCs) approved by the European Commission and relevant data protection authorities are executed with all subprocessors to ensure equivalent data protection.
              </p>
            </section>

            {/* 12. Changes */}
            <section id="changes" className="scroll-mt-24 space-y-3">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <FileText className="w-6 h-6 text-primary" />
                12. Changes to this Policy
              </h2>
              <p>
                We may periodically update this Privacy Policy to reflect platform enhancements, new architectural capabilities, or legal requirements. In the event of material modifications, we will notify registered tenant administrators via email or prominent in-app notification prior to the change taking effect.
              </p>
            </section>

            {/* 13. Contact */}
            <section id="contact" className="scroll-mt-24 space-y-4">
              <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-2">
                <Mail className="w-6 h-6 text-primary" />
                13. Contact & Inquiries
              </h2>
              <p>
                If you have questions, feedback, or requests regarding this Privacy Policy, our data protection officer and security team are available to assist:
              </p>

              <div className="p-6 rounded-2xl bg-surface border border-border shadow-xs flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                  <h3 className="font-semibold text-base text-foreground">Hexta Data Protection Office</h3>
                  <p className="text-xs text-muted mt-1">
                    Email: <a href="mailto:privacy@hexta.io" className="text-primary hover:underline">privacy@hexta.io</a>
                  </p>
                  <p className="text-xs text-muted">
                    Support: <a href="mailto:support@hexta.io" className="text-primary hover:underline">support@hexta.io</a>
                  </p>
                  <p className="text-xs text-muted mt-2">
                    Hexta Inc. — Enterprise Trust & Safety
                  </p>
                </div>
                <Button asChild>
                  <a href="mailto:privacy@hexta.io">
                    <Mail className="w-4 h-4 mr-2" />
                    Contact Privacy Team
                  </a>
                </Button>
              </div>
            </section>
          </article>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-border bg-background py-12 px-6 mt-16">
        <div className="max-w-6xl mx-auto flex flex-col md:flex-row items-center justify-between gap-6">
          <Link href="/" aria-label="Hexta home">
            <BrandLogo />
          </Link>
          <p className="text-sm text-muted">
            © {new Date().getFullYear()} Hexta Inc. All rights reserved.
          </p>
          <div className="flex items-center gap-6 text-sm text-muted">
            <Link href="/" className="hover:text-foreground transition-colors">
              Home
            </Link>
            <Link href="/privacy" className="text-primary font-medium transition-colors">
              Privacy
            </Link>
            <Link href="/tenants" className="hover:text-foreground transition-colors">
              Workspaces
            </Link>
          </div>
        </div>
      </footer>
    </div>
  );
}
