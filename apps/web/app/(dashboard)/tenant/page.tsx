"use client";

import { useEffect, useState } from "react";
import { sdk, Tenant } from "@/lib/sdk";
import { useRouter } from "next/navigation";

export default function TenantDashboardPage() {
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    // Basic mock implementation for fetching tenant info
    // In a real app, you would fetch the tenant based on the authenticated user's token context.
    const loadTenant = async () => {
      try {
        // Mock hardcoded tenant ID for demo purposes.
        const data = await sdk.identity.getTenant("mock-tenant-id");
        setTenant(data);
      } catch (err) {
        console.error("Failed to load tenant", err);
      } finally {
        setLoading(false);
      }
    };
    
    // Check auth
    if (!localStorage.getItem("auth_token")) {
      router.push("/login");
      return;
    }

    loadTenant();
  }, [router]);

  if (loading) {
    return <div className="p-8 text-center">Loading workspace...</div>;
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 py-8">
      <h1 className="text-2xl font-semibold text-gray-900 dark:text-white">Workspace Overview</h1>
      
      <div className="mt-6 bg-white dark:bg-gray-800 shadow overflow-hidden sm:rounded-lg">
        <div className="px-4 py-5 sm:px-6">
          <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white">
            Tenant Information
          </h3>
          <p className="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
            Details and current subscription plan.
          </p>
        </div>
        <div className="border-t border-gray-200 dark:border-gray-700">
          <dl>
            <div className="bg-gray-50 dark:bg-gray-900 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Workspace Name</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
                {tenant?.name || "Acme Corporation"}
              </dd>
            </div>
            <div className="bg-white dark:bg-gray-800 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">URL Slug</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
                {tenant?.slug || "acme-corp"}
              </dd>
            </div>
            <div className="bg-gray-50 dark:bg-gray-900 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Plan</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
                <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800 dark:bg-green-800 dark:text-green-100">
                  {tenant?.plan || "ENTERPRISE"}
                </span>
              </dd>
            </div>
            <div className="bg-white dark:bg-gray-800 px-4 py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
              <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
                {tenant?.status || "ACTIVE"}
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}
