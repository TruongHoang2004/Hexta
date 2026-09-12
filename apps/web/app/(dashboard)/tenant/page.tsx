"use client";

import { useEffect, useState } from "react";
import { sdk } from "@/lib/sdk";
import { Tenant, TenantMember } from "@ubi/sdk";
import { useRouter } from "next/navigation";
import Cookies from "js-cookie";
import { Building2, Users, Plus, Shield, CheckCircle2 } from "lucide-react";

export default function TenantDashboardPage() {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [activeTenant, setActiveTenant] = useState<Tenant | null>(null);
  const [members, setMembers] = useState<TenantMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newTenantName, setNewTenantName] = useState("");
  const [newTenantSlug, setNewTenantSlug] = useState("");
  const [newTenantPlan, setNewTenantPlan] = useState("free");
  const [inviteUserId, setInviteUserId] = useState("");
  const [inviteRole, setInviteRole] = useState("member");
  const [inviting, setInviting] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const router = useRouter();

  const loadTenants = async () => {
    try {
      setLoading(true);
      setErrorMessage("");
      const userTenants = await sdk.identity.getTenants();
      setTenants(userTenants || []);

      if (userTenants && userTenants.length > 0) {
        const current = userTenants[0];
        setActiveTenant(current);
        await loadMembers(current.id);
      } else {
        setActiveTenant(null);
        setMembers([]);
      }
    } catch (err: any) {
      console.error("Failed to load workspaces", err);
      setErrorMessage(err.message || "Failed to load workspaces");
    } finally {
      setLoading(false);
    }
  };

  const loadMembers = async (tenantId: string) => {
    try {
      const tenantMembers = await sdk.identity.getUsers(tenantId);
      setMembers(tenantMembers || []);
    } catch (err: any) {
      console.error("Failed to load members", err);
    }
  };

  useEffect(() => {
    const token = Cookies.get("auth_token") || (typeof window !== "undefined" ? localStorage.getItem("auth_token") : null);
    if (!token) {
      router.push("/login");
      return;
    }

    loadTenants();
  }, [router]);

  const handleSelectTenant = async (tenant: Tenant) => {
    setActiveTenant(tenant);
    await loadMembers(tenant.id);
  };

  const handleCreateTenant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTenantName || !newTenantSlug) return;
    try {
      setCreating(true);
      setErrorMessage("");
      const created = await sdk.identity.createTenant({
        name: newTenantName,
        slug: newTenantSlug,
        plan: newTenantPlan,
      });
      setNewTenantName("");
      setNewTenantSlug("");
      await loadTenants();
      setActiveTenant(created);
      await loadMembers(created.id);
    } catch (err: any) {
      setErrorMessage(err.message || "Failed to create workspace");
    } finally {
      setCreating(false);
    }
  };

  const handleInviteMember = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeTenant || !inviteUserId) return;
    try {
      setInviting(true);
      setErrorMessage("");
      await sdk.identity.inviteMember(activeTenant.id, {
        user_id: inviteUserId,
        role: inviteRole,
      });
      setInviteUserId("");
      await loadMembers(activeTenant.id);
    } catch (err: any) {
      setErrorMessage(err.message || "Failed to invite member");
    } finally {
      setInviting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-[400px] items-center justify-center">
        <div className="text-center">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent mx-auto mb-4" />
          <p className="text-sm text-gray-500 dark:text-gray-400">Loading workspaces...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 py-8">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-gray-900 dark:text-white">Workspace Overview</h1>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Manage your organizations, multi-tenant workspaces, and team memberships.
          </p>
        </div>
        {tenants.length > 0 && (
          <div className="flex items-center gap-2">
            <select
              className="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2 text-sm text-gray-900 dark:text-white shadow-sm focus:border-indigo-500 focus:outline-none"
              value={activeTenant?.id || ""}
              onChange={(e) => {
                const found = tenants.find((t) => t.id === e.target.value);
                if (found) handleSelectTenant(found);
              }}
            >
              {tenants.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name} ({t.slug})
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      {errorMessage && (
        <div className="mt-4 p-4 rounded-md bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-400 text-sm">
          {errorMessage}
        </div>
      )}

      {tenants.length === 0 ? (
        <div className="mt-8 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-700 p-8 text-center">
          <Building2 className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-base font-semibold text-gray-900 dark:text-white">No workspaces found</h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Get started by creating your first organization workspace.
          </p>
          <form onSubmit={handleCreateTenant} className="mt-6 max-w-md mx-auto space-y-4 text-left">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Workspace Name</label>
              <input
                type="text"
                required
                placeholder="e.g. Acme Corp"
                value={newTenantName}
                onChange={(e) => setNewTenantName(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Slug</label>
              <input
                type="text"
                required
                placeholder="e.g. acme-corp"
                value={newTenantSlug}
                onChange={(e) => setNewTenantSlug(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Plan</label>
              <select
                value={newTenantPlan}
                onChange={(e) => setNewTenantPlan(e.target.value)}
                className="mt-1 block w-full rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-3 py-2 text-sm text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              >
                <option value="free">Free</option>
                <option value="pro">Pro</option>
                <option value="enterprise">Enterprise</option>
              </select>
            </div>
            <button
              type="submit"
              disabled={creating}
              className="w-full inline-flex justify-center items-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 disabled:opacity-50"
            >
              <Plus className="h-4 w-4" />
              {creating ? "Creating..." : "Create Workspace"}
            </button>
          </form>
        </div>
      ) : (
        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
          {/* Workspace Info Card */}
          <div className="lg:col-span-2 bg-white dark:bg-gray-800 shadow rounded-lg overflow-hidden">
            <div className="px-4 py-5 sm:px-6 flex items-center justify-between border-b border-gray-200 dark:border-gray-700">
              <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white flex items-center gap-2">
                  <Building2 className="h-5 w-5 text-indigo-500" />
                  Tenant Details
                </h3>
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  Detailed subscription and identity information.
                </p>
              </div>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300">
                <CheckCircle2 className="h-3.5 w-3.5" />
                {activeTenant?.status?.toUpperCase() || "ACTIVE"}
              </span>
            </div>
            <dl className="divide-y divide-gray-200 dark:divide-gray-700">
              <div className="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Workspace Name</dt>
                <dd className="mt-1 text-sm font-semibold text-gray-900 dark:text-white sm:mt-0 sm:col-span-2">
                  {activeTenant?.name}
                </dd>
              </div>
              <div className="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">URL Slug</dt>
                <dd className="mt-1 text-sm text-gray-900 dark:text-white sm:mt-0 sm:col-span-2 font-mono">
                  {activeTenant?.slug}
                </dd>
              </div>
              <div className="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Plan</dt>
                <dd className="mt-1 text-sm sm:mt-0 sm:col-span-2">
                  <span className="px-2 py-0.5 inline-flex text-xs font-semibold rounded-md bg-indigo-50 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800 uppercase">
                    {activeTenant?.plan || "FREE"}
                  </span>
                </dd>
              </div>
              <div className="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Owner ID</dt>
                <dd className="mt-1 text-xs text-gray-600 dark:text-gray-300 font-mono sm:mt-0 sm:col-span-2">
                  {activeTenant?.owner_id}
                </dd>
              </div>
              <div className="px-4 py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="text-sm font-medium text-gray-500 dark:text-gray-400">Workspace ID</dt>
                <dd className="mt-1 text-xs text-gray-600 dark:text-gray-300 font-mono sm:mt-0 sm:col-span-2">
                  {activeTenant?.id}
                </dd>
              </div>
            </dl>
          </div>

          {/* Members & Invite Card */}
          <div className="bg-white dark:bg-gray-800 shadow rounded-lg overflow-hidden flex flex-col">
            <div className="px-4 py-5 sm:px-6 border-b border-gray-200 dark:border-gray-700">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white flex items-center gap-2">
                <Users className="h-5 w-5 text-indigo-500" />
                Team Members ({members.length})
              </h3>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                Workspace collaborators and their roles.
              </p>
            </div>
            <div className="p-4 flex-1 overflow-y-auto max-h-[260px] divide-y divide-gray-100 dark:divide-gray-700">
              {members.map((member) => (
                <div key={member.id} className="py-2.5 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Shield className="h-4 w-4 text-gray-400" />
                    <div>
                      <p className="text-xs font-mono text-gray-900 dark:text-white truncate max-w-[160px]">
                        {member.user_id}
                      </p>
                    </div>
                  </div>
                  <span className="px-2 py-0.5 rounded text-xs font-medium uppercase bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300">
                    {member.role}
                  </span>
                </div>
              ))}
            </div>

            {/* Invite Form */}
            <form onSubmit={handleInviteMember} className="p-4 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 space-y-2">
              <label className="block text-xs font-medium text-gray-700 dark:text-gray-300">Invite User</label>
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="Target User ID"
                  required
                  value={inviteUserId}
                  onChange={(e) => setInviteUserId(e.target.value)}
                  className="flex-1 rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-2.5 py-1.5 text-xs text-gray-900 dark:text-white focus:outline-none"
                />
                <select
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value)}
                  className="rounded-md border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-2 py-1.5 text-xs text-gray-900 dark:text-white focus:outline-none"
                >
                  <option value="member">Member</option>
                  <option value="admin">Admin</option>
                </select>
                <button
                  type="submit"
                  disabled={inviting}
                  className="rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white shadow-sm hover:bg-indigo-700 disabled:opacity-50"
                >
                  {inviting ? "..." : "Invite"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
