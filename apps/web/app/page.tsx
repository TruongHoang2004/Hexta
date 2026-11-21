import { redirect } from "next/navigation";

// The marketing site is the separate landing app; the web app starts at sign-in
// (the proxy sends signed-in users on to their workspaces).
export default function Home() {
  redirect("/login");
}
