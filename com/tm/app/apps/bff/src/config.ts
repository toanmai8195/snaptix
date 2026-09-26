import { z } from "zod";

const LogLevel = z.enum(["debug", "info", "warn", "error"]);

const Env = z.object({
  BFF_HOST: z.string().min(1).default("0.0.0.0"),
  BFF_PORT: z.coerce.number().int().min(1).max(65535).default(3000),
  LOG_LEVEL: LogLevel.default("info"),
  MONGODB_URI: z.url({ protocol: /^mongodb(\+srv)?$/ }).default("mongodb://localhost:27017/snaptix"),
  CORE_BASE_URL: z.url({ protocol: /^https?$/ }).default("http://localhost:8080"),
});

export type Config = {
  host: string;
  port: number;
  logLevel: z.infer<typeof LogLevel>;
  mongoUri: string;
  coreBaseUrl: string;
};

/** Đọc cấu hình BFF từ biến môi trường; lỗi nêu đúng tên biến sai. */
export function loadConfig(env: Record<string, string | undefined>): Config {
  // Biến rỗng coi như không đặt để dùng giá trị mặc định.
  const cleaned = Object.fromEntries(Object.entries(env).filter(([, v]) => v !== undefined && v !== ""));
  const parsed = Env.safeParse(cleaned);
  if (!parsed.success) {
    const issues = parsed.error.issues.map((i) => `${i.path.join(".")}: ${i.message}`).join("; ");
    throw new Error(`cấu hình không hợp lệ — ${issues}`);
  }
  const e = parsed.data;
  return {
    host: e.BFF_HOST,
    port: e.BFF_PORT,
    logLevel: e.LOG_LEVEL,
    mongoUri: e.MONGODB_URI,
    coreBaseUrl: e.CORE_BASE_URL,
  };
}
