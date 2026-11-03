import { test } from "node:test";
import assert from "node:assert/strict";
import { readConsultEvents } from "./consultant-sse.ts";

function stream(text, width = 1) {
  const bytes = new TextEncoder().encode(text);
  return new ReadableStream({
    start(controller) {
      for (let i = 0; i < bytes.length; i += width)
        controller.enqueue(bytes.slice(i, i + width));
      controller.close();
    },
  });
}
const message =
  'event: message\r\ndata: {"type":"message","text":"Tư vấn 🛒"}\r\n\r\n';
const done = 'event: done\ndata: {"type":"done"}\n\n';
test("SSE preserves UTF-8 and CRLF across every byte boundary", async () => {
  const events = [];
  await readConsultEvents(
    stream(": heartbeat\r\n\r\n" + message + done),
    (event) => events.push(event),
  );
  assert.equal(events.length, 2);
  assert.equal(events[0].text, "Tư vấn 🛒");
});
test("SSE detects truncated streams and provider error events", async () => {
  await assert.rejects(
    readConsultEvents(stream(message), () => {}),
    /interrupted/,
  );
  await assert.rejects(
    readConsultEvents(
      stream('event: error\ndata: {"type":"error","error":"Unavailable"}\n\n'),
      () => {},
    ),
    /Unavailable/,
  );
});
test("SSE rejects malformed JSON and conflicting event types", async () => {
  await assert.rejects(
    readConsultEvents(stream("data: invalid\n\n"), () => {}),
  );
  await assert.rejects(
    readConsultEvents(
      stream('event: products\ndata: {"type":"message"}\n\n'),
      () => {},
    ),
    /Invalid consultation event/,
  );
});
test("SSE propagates consumer failures and releases the reader", async () => {
  const source = stream(message + done);
  await assert.rejects(
    readConsultEvents(source, () => {
      throw new Error("consumer failed");
    }),
    /consumer failed/,
  );
  assert.equal(source.locked, false);
});
