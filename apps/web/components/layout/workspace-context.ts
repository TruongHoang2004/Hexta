"use client";

import { createContext, useContext } from "react";
import type { Tenant } from "@hexta/sdk";

/** Tenant role as seen by the UI; the API still authorizes every action. */
export type WorkspaceRole = "owner" | "member";

export interface WorkspaceState {
  /** The workspace in the URL, or null while resolving or outside a workspace. */
  tenant: Tenant | null;
  /** Every workspace the user belongs to. */
  tenants: Tenant[];
  role: WorkspaceRole;
  isOwner: boolean;
  reload: () => Promise<void>;
}

export const WorkspaceContext = createContext<WorkspaceState | null>(null);

/** Current workspace and the user's workspaces; only valid inside the dashboard layout. */
export function useWorkspace() {
  const state = useContext(WorkspaceContext);
  if (!state) throw new Error("Tenant context is required");
  return state;
}
