import { WorkItems } from "@/components/sales/work-items";

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  const value = (await searchParams).work_id;
  const id =
    typeof value === "string" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(
      value,
    )
      ? value
      : undefined;
  return (
    <WorkItems
      key={id || "list"}
      initialID={id}
      invalidID={value !== undefined && !id}
    />
  );
}
