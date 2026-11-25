import { sdk } from "@/lib/sdk";

export type ReplyMode = "suggest" | "auto_send";

/** The tenant's bot reply mode. `configured` is false until the owner first chooses one. */
export interface ReplySettings {
  reply_mode: ReplyMode;
  configured: boolean;
  version: number;
  updated_at?: string;
}

const base = (tenant: string) => `/api/v1/tenants/${tenant}/settings/reply-mode`;

/** Reply mode endpoints; the API allows every member to read and only the owner to change. */
export const replyMode = {
  async get(tenant: string, signal?: AbortSignal): Promise<ReplySettings> {
    const res = await sdk.client.get<{ data: ReplySettings }>(base(tenant), { signal });
    return res.data.data;
  },
  /** expected_version is the version last read; a stale version is rejected with 409. */
  async update(tenant: string, mode: ReplyMode, expectedVersion: number): Promise<ReplySettings> {
    const res = await sdk.client.put<{ data: ReplySettings }>(base(tenant), {
      reply_mode: mode,
      expected_version: expectedVersion,
    });
    return res.data.data;
  },
};
