export type Channel = "Facebook" | "Telegram" | "Web chat";
export type Queue = "all" | "unassigned" | "attention" | "closed";
export type Mode = "reply" | "note";
export interface Conversation {
  id: string;
  name: string;
  initials: string;
  channel: Channel;
  message: string;
  time: string;
  unread: number;
  assignee: string | null;
  needsHuman: boolean;
  closed: boolean;
}
export interface Preview {
  text: string;
  mode: Mode;
}
