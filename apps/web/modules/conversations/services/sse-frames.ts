export interface SseFrame {
  event: string;
  data: string;
}

const MAX_BUFFER = 1024 * 1024;

/** Reads server-sent event frames until the stream ends. Comment lines (heartbeats) are skipped. */
export async function readSseFrames(
  stream: ReadableStream<Uint8Array>,
  onFrame: (frame: SseFrame) => void,
) {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  const consume = () => {
    let boundary: number;
    while ((boundary = buffer.indexOf("\n\n")) >= 0) {
      const lines = buffer.slice(0, boundary).split("\n");
      buffer = buffer.slice(boundary + 2);
      const event = lines.find((line) => line.startsWith("event:"))?.slice(6).trim() ?? "message";
      const data = lines
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.slice(5).trimStart())
        .join("\n");
      if (lines.every((line) => line.startsWith(":") || line === "")) continue;
      onFrame({ event, data });
    }
  };
  try {
    while (true) {
      const chunk = await reader.read();
      if (chunk.done) break;
      buffer = (buffer + decoder.decode(chunk.value, { stream: true })).replace(/\r\n/g, "\n");
      if (buffer.length > MAX_BUFFER) throw new Error("Event stream frame too large");
      consume();
    }
    buffer += decoder.decode();
    consume();
  } finally {
    await reader.cancel().catch(() => undefined);
    reader.releaseLock();
  }
}
