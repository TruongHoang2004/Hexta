import { test } from "node:test";
import assert from "node:assert/strict";
import { readSseFrames } from "./sse-frames.ts";

function stream(text, width = 1) {
  const bytes = new TextEncoder().encode(text);
  return new ReadableStream({
    start(controller) {
      for (let i = 0; i < bytes.length; i += width) controller.enqueue(bytes.slice(i, i + width));
      controller.close();
    },
  });
}

test("frames survive CRLF and byte-by-byte delivery, heartbeats are skipped", async () => {
  const frames = [];
  await readSseFrames(
    stream(
      'event: ready\r\ndata: {}\r\n\r\n: keep-alive\r\n\r\nevent: changed\r\ndata: {"thread_id":"Tư 🛒"}\r\n\r\n',
    ),
    (frame) => frames.push(frame),
  );
  assert.deepEqual(frames, [
    { event: "ready", data: "{}" },
    { event: "changed", data: '{"thread_id":"Tư 🛒"}' },
  ]);
});

test("an unterminated trailing frame is not delivered", async () => {
  const frames = [];
  await readSseFrames(stream('event: changed\ndata: {"thread_id":"x"}'), (frame) => frames.push(frame));
  assert.equal(frames.length, 0);
});
