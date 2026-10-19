"use client";
import { useWorkspace } from "@/components/sales/workspace";
export default function SubscriptionPage() {
  const { tenant } = useWorkspace();
  return (
    <section className="max-w-xl space-y-4 rounded-xl border border-border bg-surface p-6">
      <h1 className="text-2xl font-semibold">Tenant service plan</h1>
      <p className="text-muted">
        The service plan belongs to {tenant?.name}, not individual employee
        accounts.
      </p>
      <dl className="space-y-3">
        <div className="flex justify-between">
          <dt>Plan</dt>
          <dd className="font-semibold uppercase">{tenant?.plan}</dd>
        </div>
        <div className="flex justify-between">
          <dt>Access status</dt>
          <dd>{tenant?.status}</dd>
        </div>
      </dl>
      <p className="text-sm text-muted">
        Hexta manages the service plan and access status. Tenant approval does
        not mean a payment has been received. Online payments and invoices are
        not connected.
      </p>
    </section>
  );
}
