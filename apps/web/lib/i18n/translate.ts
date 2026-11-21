import type { Locale } from "./config";
import { vi } from "./messages/vi";
import { en } from "./messages/en";

type Shape<T> = { readonly [K in keyof T]: T[K] extends string ? string : Shape<T[K]> };

/** Every locale has exactly the keys of the Vietnamese source dictionary. */
export type Messages = Shape<typeof vi>;

type Paths<T, Prefix extends string = ""> = {
  [K in keyof T & string]: T[K] extends string
    ? `${Prefix}${K}`
    : Paths<T[K], `${Prefix}${K}.`>;
}[keyof T & string];

/** Dotted key of a message, e.g. "nav.inbox". */
export type MessageKey = Paths<Messages>;
export type MessageParams = Record<string, string | number>;
export type Translate = (key: MessageKey, params?: MessageParams) => string;

const dictionaries: Record<Locale, Messages> = { vi, en };

export function translate(locale: Locale, key: MessageKey, params?: MessageParams): string {
  let node: unknown = dictionaries[locale];
  for (const part of key.split(".")) {
    node = typeof node === "object" && node !== null ? (node as Record<string, unknown>)[part] : undefined;
  }
  const text = typeof node === "string" ? node : key;
  if (!params) return text;
  return text.replace(/\{(\w+)\}/g, (match, name: string) =>
    name in params ? String(params[name]) : match,
  );
}
