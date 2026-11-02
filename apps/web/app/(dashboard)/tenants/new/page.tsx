import { redirect } from "next/navigation";

export default function NewTenantRedirectPage() {
  redirect("/tenants/register");
}
