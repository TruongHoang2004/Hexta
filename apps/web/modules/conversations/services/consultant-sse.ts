import type { ConsultEvent } from "./consultant";

export async function readConsultEvents(
  stream: ReadableStream<Uint8Array>,
  onEvent: (event: ConsultEvent) => void,
) {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buffer = "",
    done = false;
  const consume = () => {
    let boundary: number;
    while ((boundary = buffer.indexOf("\n\n")) >= 0) {
      const frame = buffer.slice(0, boundary);
      buffer = buffer.slice(boundary + 2);
      const lines = frame.split("\n");
      const name = lines
        .find((line) => line.startsWith("event:"))
        ?.slice(6)
        .trim();
      const data = lines
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.slice(5).trimStart())
        .join("\n");
      if (!data) continue;
      const event = JSON.parse(data) as ConsultEvent;
      if (name && name !== event.type)
        throw new Error("Invalid consultation event");
      if (event.type === "error")
        throw new Error(event.error || "Consultation failed");
      if (event.type === "done") done = true;
      onEvent(event);
    }
  };
  try {
    while (true) {
      const chunk = await reader.read();
      if (chunk.done) break;
      buffer += decoder.decode(chunk.value, { stream: true });
      buffer = buffer.replace(/\r\n/g, "\n");
      if (buffer.length > 2 * 1024 * 1024)
        throw new Error("Consultation response too large");
      consume();
    }
    buffer += decoder.decode();
    consume();
    if (!done) throw new Error("Consultation interrupted. Please try again.");
  } finally {
    await reader.cancel();
    reader.releaseLock();
  }
}
