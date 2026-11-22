export interface WorkflowStepInput {
  name: string;
  handoff_mode: "direct" | "acceptance";
}
export type HandoffMode = WorkflowStepInput["handoff_mode"];
export interface WorkflowVersion {
  id: string;
  definition_id: string;
  number: number;
  name: string;
  steps: (WorkflowStepInput & { id: string; ordinal: number })[];
}
export interface WorkflowDefinition {
  id: string;
  name: string;
  version: number;
  versions: WorkflowVersion[];
}
export interface WorkEvent {
  id: string;
  version: number;
  actor_id: string;
  action: string;
  from_responsible_id: string | null;
  responsible_id: string | null;
  recipient_id: string | null;
  from_step: number;
  current_step: number;
  state: string;
  created_at: string;
}
export interface WorkItem {
  id: string;
  order_id: string;
  workflow_version_id: string;
  current_step: number;
  responsible_id: string | null;
  recipient_id: string | null;
  state: "active" | "waiting" | "completed";
  version: number;
  workflow: WorkflowVersion;
  events?: WorkEvent[];
}
export interface WorkflowInput {
  name: string;
  expected_version: number;
  steps: WorkflowStepInput[];
}
export interface WorkInput {
  order_id: string;
  workflow_version_id: string;
  responsible_id: string;
}
export type WorkView = "all" | "mine" | "awaiting_me";
export type WorkCommand = "handoff" | "accept" | "assign" | "complete";
export interface WorkCommandFields {
  recipient_id?: string;
  responsible_id?: string | null;
}
