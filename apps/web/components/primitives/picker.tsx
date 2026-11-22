"use client";

import { Combobox, type ComboboxProps } from "@hexta/ui";
import { useT } from "@/lib/i18n/provider";

export type PickerOption = ComboboxProps["options"][number];

type Props = Omit<ComboboxProps, "placeholder" | "searchPlaceholder" | "emptyText"> & {
  placeholder?: string;
};

/** The app's dropdown: a searchable @hexta/ui Combobox with localized default texts. */
export function Picker({ placeholder, ...props }: Props) {
  const t = useT();
  return (
    <Combobox
      placeholder={placeholder ?? t("common.choose")}
      searchPlaceholder={t("common.search")}
      emptyText={t("common.noResults")}
      {...props}
    />
  );
}
