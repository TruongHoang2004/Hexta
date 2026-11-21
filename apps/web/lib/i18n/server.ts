import { cookies } from "next/headers";
import { defaultLocale, isLocale, localeCookie, type Locale } from "./config";

/** Locale chosen by the viewer, read from the request cookie; Vietnamese by default. */
export async function getLocale(): Promise<Locale> {
  const value = (await cookies()).get(localeCookie)?.value;
  return isLocale(value) ? value : defaultLocale;
}
