import { sdk } from "@/lib/sdk";
export interface Application {
  id: string;
  name: string;
  slug: string;
  applicant_id: string;
  status: "pending" | "approved" | "rejected";
  review_reason: string;
  tenant_id: string | null;
  created_at: string;
}
export interface Invitation {
  id: string;
  tenant_id: string;
  email: string;
  status: string;
  expires_at: string;
  accepted_by: string;
}
export interface InvitationResult {
  invitation: Invitation;
  token?: string;
  tenant_name?: string;
  delivery?: string;
}
export const onboarding = {
  async applications(platform = false): Promise<Application[]> {
    const r = await sdk.client.get<{ data: { items: Application[] } }>(
      `/api/v1/${platform ? "platform/" : ""}tenant-requests`,
    );
    return r.data.data.items;
  },
  async apply(name: string, slug: string): Promise<Application> {
    const r = await sdk.client.post<{ data: Application }>(
      "/api/v1/tenant-requests",
      { name, slug },
    );
    return r.data.data;
  },
  async operator(): Promise<boolean> {
    const r = await sdk.client.get<{ data: { operator: boolean } }>(
      "/api/v1/platform/access",
    );
    return r.data.data.operator;
  },
  async review(
    id: string,
    decision: string,
    reason: string,
  ): Promise<Application> {
    const r = await sdk.client.post<{ data: Application }>(
      `/api/v1/platform/tenant-requests/${id}/review`,
      { decision, reason },
    );
    return r.data.data;
  },
  async invitations(tenant: string): Promise<Invitation[]> {
    const r = await sdk.client.get<{ data: { items: Invitation[] } }>(
      `/api/v1/tenants/${tenant}/invitations`,
    );
    return r.data.data.items;
  },
  async invite(tenant: string, email: string): Promise<InvitationResult> {
    const r = await sdk.client.post<{ data: InvitationResult }>(
      `/api/v1/tenants/${tenant}/invitations`,
      { email },
    );
    return r.data.data;
  },
  async revoke(tenant: string, id: string): Promise<void> {
    await sdk.client.post(`/api/v1/tenants/${tenant}/invitations/${id}/revoke`);
  },
  async useInvitation(
    token: string,
    action: "preview" | "accept" | "decline",
  ): Promise<InvitationResult> {
    const r = await sdk.client.post<{ data: InvitationResult }>(
      `/api/v1/invitations/${action}`,
      { token },
    );
    return r.data.data;
  },
};
