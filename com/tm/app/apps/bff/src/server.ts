import { buildApp } from "./app.js";
import { loadConfig } from "./config.js";
import { createMongo } from "./plugins/mongo.js";

async function main(): Promise<void> {
  const config = loadConfig(process.env);
  const mongo = createMongo(config.mongoUri);
  const app = buildApp({ config, mongo: mongo.pinger });

  app.addHook("onClose", async () => {
    await mongo.client.close();
  });

  await app.listen({ host: config.host, port: config.port });
  app.log.info({ addr: `${config.host}:${config.port}` }, "bff started");
}

main().catch((err: unknown) => {
  console.error("bff:", err instanceof Error ? err.message : err);
  process.exit(1);
});
