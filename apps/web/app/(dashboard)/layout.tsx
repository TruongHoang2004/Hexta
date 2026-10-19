import { Workspace } from "@/components/sales/workspace";
export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <Workspace>{children}</Workspace>;
}
