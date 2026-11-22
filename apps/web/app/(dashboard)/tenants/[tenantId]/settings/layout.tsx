import { SettingsTabs } from "@/modules/settings";

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return <SettingsTabs>{children}</SettingsTabs>;
}
