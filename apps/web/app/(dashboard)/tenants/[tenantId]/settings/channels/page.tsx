import { FacebookChannelSettings } from "@/components/channels/facebook-settings";
import { TelegramChannelSettings } from "@/components/channels/telegram-settings";

export default function ChannelsPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-semibold">Channels</h1>
        <p className="mt-2 text-sm text-muted">
          Connect Facebook Pages and Telegram bots so customer messages arrive
          in the Inbox and assigned sales can reply. Each Page or bot can
          belong to one workspace.
        </p>
      </div>
      <FacebookChannelSettings />
      <TelegramChannelSettings />
    </div>
  );
}
