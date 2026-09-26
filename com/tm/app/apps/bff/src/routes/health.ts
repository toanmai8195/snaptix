import type { FastifyInstance } from "fastify";
import type { Pinger } from "../plugins/mongo.js";

/** Giới hạn thời gian /readyz chờ MongoDB để load balancer nhận 503 nhanh. */
export const READY_TIMEOUT_MS = 2000;

export function healthRoutes(app: FastifyInstance, deps: { mongo: Pinger; readyTimeoutMs?: number }): void {
  const timeout = deps.readyTimeoutMs ?? READY_TIMEOUT_MS;

  app.get("/healthz", async () => ({ status: "ok" }));

  app.get("/readyz", async (req, reply) => {
    try {
      await deps.mongo.ping(AbortSignal.timeout(timeout));
      return { status: "ok" };
    } catch (err) {
      req.log.warn({ err }, "readyz: mongodb unavailable");
      return reply.code(503).send({ status: "unavailable", error: "mongodb" });
    }
  });
}
