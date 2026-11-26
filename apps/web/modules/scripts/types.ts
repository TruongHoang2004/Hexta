/** Mirrors the Go API script DTOs (services/api/internal/present/http/dto/script_dto.go). */
export type ScriptStatus = "inactive" | "active" | "archived";
export type ScriptStatusFilter = ScriptStatus | "all";

export type ScriptPermission = "script.create" | "script.edit" | "script.activate" | "script.archive";

export const scriptPermissions: readonly ScriptPermission[] = [
  "script.create",
  "script.edit",
  "script.activate",
  "script.archive",
];

/** API limits; the server trims, de-duplicates and enforces them again. */
export const scriptLimits = { name: 120, triggers: 20, trigger: 120, answer: 2000 } as const;

export interface ScriptVersion {
  id: string;
  number: number;
  triggers: string[];
  answer: string;
  draft: boolean;
  created_by: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
  published_at?: string;
  published_by?: string;
}

export interface ScriptSummary {
  id: string;
  name: string;
  status: ScriptStatus;
  active_version_id?: string;
  active_version_number?: number;
  draft_version_number?: number;
  version: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  archived_at?: string;
}

export interface Script {
  id: string;
  name: string;
  status: ScriptStatus;
  active_version_id?: string;
  version: number;
  created_by: string;
  created_at: string;
  updated_at: string;
  archived_at?: string;
  /** Newest first, at most 100. */
  versions: ScriptVersion[];
}

export interface ScriptPage {
  items: ScriptSummary[];
  total: number;
  page: number;
  size: number;
}

export interface ScriptDraftInput {
  name: string;
  triggers: string[];
  answer: string;
}

export interface MyScriptPermissions {
  owner: boolean;
  permissions: ScriptPermission[];
}

export interface MemberScriptPermissions {
  user_id: string;
  email?: string;
  role?: string;
  owner: boolean;
  permissions: ScriptPermission[];
  /** Optimistic version of this member's grants; 0 before any grant. */
  version: number;
  updated_at?: string;
}
