import { sdk } from "@/lib/sdk";
import type {
  MemberScriptPermissions,
  MyScriptPermissions,
  Script,
  ScriptDraftInput,
  ScriptPage,
  ScriptPermission,
  ScriptStatusFilter,
} from "../types";

const base = (tenant: string) => `/api/v1/tenants/${tenant}`;

/** Script and script-permission endpoints; every call is authorized again by the API. */
export const scripts = {
  async list(tenant: string, status: ScriptStatusFilter, page: number, signal?: AbortSignal): Promise<ScriptPage> {
    const res = await sdk.client.get<{ data: ScriptPage }>(`${base(tenant)}/scripts`, {
      params: { page, size: 20, ...(status === "all" ? {} : { status }) },
      signal,
    });
    return res.data.data;
  },
  async get(tenant: string, id: string, signal?: AbortSignal): Promise<Script> {
    const res = await sdk.client.get<{ data: Script }>(`${base(tenant)}/scripts/${id}`, { signal });
    return res.data.data;
  },
  /** The idempotency key keeps a retried create from making a second script. */
  async create(tenant: string, input: ScriptDraftInput, key: string): Promise<Script> {
    const res = await sdk.client.request<{ data: Script }>({
      method: "POST",
      url: `${base(tenant)}/scripts`,
      data: input,
      headers: { "Idempotency-Key": key },
    });
    return res.data.data;
  },
  async saveDraft(tenant: string, script: Script, input: ScriptDraftInput): Promise<Script> {
    const res = await sdk.client.put<{ data: Script }>(`${base(tenant)}/scripts/${script.id}/draft`, {
      ...input,
      expected_version: script.version,
    });
    return res.data.data;
  },
  async activate(tenant: string, script: Script, versionID: string): Promise<Script> {
    const res = await sdk.client.post<{ data: Script }>(`${base(tenant)}/scripts/${script.id}/activate`, {
      version_id: versionID,
      expected_version: script.version,
    });
    return res.data.data;
  },
  async transition(tenant: string, script: Script, action: "deactivate" | "archive"): Promise<Script> {
    const res = await sdk.client.post<{ data: Script }>(`${base(tenant)}/scripts/${script.id}/${action}`, {
      expected_version: script.version,
    });
    return res.data.data;
  },
  async myPermissions(tenant: string, signal?: AbortSignal): Promise<MyScriptPermissions> {
    const res = await sdk.client.get<{ data: MyScriptPermissions }>(`${base(tenant)}/script-permissions/me`, { signal });
    return res.data.data;
  },
  async memberPermissions(tenant: string, signal?: AbortSignal): Promise<MemberScriptPermissions[]> {
    const res = await sdk.client.get<{ data: { members: MemberScriptPermissions[] } }>(
      `${base(tenant)}/script-permissions`,
      { signal },
    );
    return res.data.data.members;
  },
  async setPermission(
    tenant: string,
    member: MemberScriptPermissions,
    permission: ScriptPermission,
    granted: boolean,
  ): Promise<MemberScriptPermissions> {
    const res = await sdk.client.post<{ data: MemberScriptPermissions }>(
      `${base(tenant)}/script-permissions/${member.user_id}/${granted ? "grant" : "revoke"}`,
      { permission, expected_version: member.version },
    );
    return res.data.data;
  },
};
