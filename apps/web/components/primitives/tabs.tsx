"use client";

import type { ReactNode } from "react";
import { Tabs as TabsRoot, TabsContent, TabsList, TabsTrigger } from "@hexta/ui";

export interface TabItem<T extends string> {
  value: T;
  label: string;
  badge?: ReactNode;
}

/** Tab bar plus the active panel (`children`), wired for ARIA and arrow keys by Radix. */
export function Tabs<T extends string>({
  label,
  items,
  value,
  onChange,
  children,
  panelClassName,
}: {
  label: string;
  items: TabItem<T>[];
  value: T;
  onChange: (value: T) => void;
  children: ReactNode;
  panelClassName?: string;
}) {
  return (
    <TabsRoot
      value={value}
      onValueChange={(next) => {
        const item = items.find((row) => row.value === next);
        if (item) onChange(item.value);
      }}
    >
      <TabsList aria-label={label}>
        {items.map((item) => (
          <TabsTrigger key={item.value} value={item.value}>
            {item.label}
            {item.badge}
          </TabsTrigger>
        ))}
      </TabsList>
      <TabsContent value={value} className={panelClassName}>
        {children}
      </TabsContent>
    </TabsRoot>
  );
}
