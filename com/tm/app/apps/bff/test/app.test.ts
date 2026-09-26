import { PassThrough } from "node:stream";
import { describe, expect, it } from "vitest";
import { buildApp } from "../src/app.js";
import type { Pinger } from "../src/plugins/mongo.js";

const pingOK: Pinger = { ping: async () => {} };
const pingFail: Pinger = {
  ping: async () => {
    throw new Error("connection refused");
  },
};
// Treo tới khi signal abort — giống MongoDB không phản hồi.
const pingHang: Pinger = {
  ping: (signal) => new Promise((_, reject) => signal.addEventListener("abort", () => reject(signal.reason))),
};

function setup(mongo: Pinger, level: "info" | "debug" = "info") {
  const stream = new PassThrough();
  const lines: Record<string, unknown>[] = [];
  stream.on("data", (chunk: Buffer) => {
    for (const l of chunk.toString().split("\n").filter(Boolean)) lines.push(JSON.parse(l) as Record<string, unknown>);
  });
  const app = buildApp({ config: { logLevel: level }, mongo, logStream: stream, readyTimeoutMs: 50 });
  return { app, lines };
}

describe("health", () => {
  it.each([
    ["healthz không phụ thuộc Mongo", pingFail, "/healthz", 200, "ok"],
    ["readyz Mongo ok", pingOK, "/readyz", 200, "ok"],
    ["readyz Mongo lỗi", pingFail, "/readyz", 503, "unavailable"],
    ["readyz Mongo treo", pingHang, "/readyz", 503, "unavailable"],
  ] as const)("%s", async (_name, mongo, url, code, status) => {
    const { app } = setup(mongo);
    const start = Date.now();
    const res = await app.inject({ method: "GET", url });
    expect(res.statusCode).toBe(code);
    expect(res.json()).toMatchObject({ status });
    if (code === 503) expect(res.json()).toMatchObject({ error: "mongodb" });
    expect(Date.now() - start).toBeLessThan(1000);
  });

  it("route lạ → 404 JSON", async () => {
    const { app } = setup(pingOK);
    const res = await app.inject({ method: "GET", url: "/khong-ton-tai" });
    expect(res.statusCode).toBe(404);
    expect(res.json()).toMatchObject({ error: "not_found" });
  });
});

describe("request ID", () => {
  it("sinh 32 hex khi không gửi, mỗi request khác nhau", async () => {
    const { app } = setup(pingOK);
    const a = (await app.inject({ url: "/healthz" })).headers["x-request-id"];
    const b = (await app.inject({ url: "/healthz" })).headers["x-request-id"];
    expect(a).toMatch(/^[0-9a-f]{32}$/);
    expect(b).toMatch(/^[0-9a-f]{32}$/);
    expect(a).not.toBe(b);
  });

  it("dùng lại header hợp lệ, thay header không hợp lệ", async () => {
    const { app } = setup(pingOK);
    const ok = await app.inject({ url: "/healthz", headers: { "x-request-id": "abc-123.X_y" } });
    expect(ok.headers["x-request-id"]).toBe("abc-123.X_y");
    for (const bad of ["a".repeat(65), "có dấu", "<script>"]) {
      const res = await app.inject({ url: "/healthz", headers: { "x-request-id": bad } });
      expect(res.headers["x-request-id"]).toMatch(/^[0-9a-f]{32}$/);
    }
  });

  it("log request có reqId trùng header", async () => {
    const { app, lines } = setup(pingOK);
    await app.inject({ url: "/healthz", headers: { "x-request-id": "req-77" } });
    const completed = lines.find((l) => l.msg === "request completed");
    expect(completed).toMatchObject({ reqId: "req-77", level: "info", service: "bff" });
    expect(typeof completed?.time).toBe("string");
  });
});
