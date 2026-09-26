import { randomBytes } from "node:crypto";
import Fastify, { type FastifyInstance, type FastifyServerOptions } from "fastify";
import type { Config } from "./config.js";
import type { Pinger } from "./plugins/mongo.js";
import { healthRoutes } from "./routes/health.js";

const REQUEST_ID_HEADER = "x-request-id";
// Chỉ nhận request ID ngắn, an toàn để ghi log và trả lại header (giống core).
const VALID_REQUEST_ID = /^[A-Za-z0-9._-]{1,64}$/;

export type AppDeps = {
  config: Pick<Config, "logLevel">;
  mongo: Pinger;
  /** Ghi log vào stream khác (test); mặc định stdout. */
  logStream?: NodeJS.WritableStream;
  readyTimeoutMs?: number;
};

/** Dựng Fastify app (chưa listen) — dùng chung cho server.ts và test (app.inject). */
export function buildApp(deps: AppDeps): FastifyInstance {
  const logger: FastifyServerOptions["logger"] = {
    level: deps.config.logLevel,
    base: { service: "bff" },
    // level dạng chữ ("info") thay vì số, thống nhất với log JSON của core
    formatters: { level: (label) => ({ level: label }) },
    timestamp: () => `,"time":"${new Date().toISOString()}"`,
    ...(deps.logStream ? { stream: deps.logStream } : {}),
  };

  const app = Fastify({
    logger,
    genReqId: (req) => {
      const incoming = req.headers[REQUEST_ID_HEADER];
      return typeof incoming === "string" && VALID_REQUEST_ID.test(incoming) ? incoming : randomBytes(16).toString("hex");
    },
  });

  app.addHook("onRequest", async (req, reply) => {
    reply.header(REQUEST_ID_HEADER, req.id);
  });

  app.setNotFoundHandler((req, reply) => {
    reply.code(404).send({ error: "not_found", path: req.url });
  });

  healthRoutes(app, { mongo: deps.mongo, ...(deps.readyTimeoutMs !== undefined ? { readyTimeoutMs: deps.readyTimeoutMs } : {}) });
  return app;
}
