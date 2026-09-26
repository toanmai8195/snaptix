import { describe, expect, it } from "vitest";
import { loadConfig } from "../src/config.js";

describe("loadConfig", () => {
  it("dùng giá trị mặc định khi không đặt biến", () => {
    expect(loadConfig({})).toEqual({
      host: "0.0.0.0",
      port: 3000,
      logLevel: "info",
      mongoUri: "mongodb://localhost:27017/snaptix",
      coreBaseUrl: "http://localhost:8080",
    });
  });

  it("biến rỗng coi như không đặt", () => {
    expect(loadConfig({ BFF_PORT: "", LOG_LEVEL: "" }).port).toBe(3000);
  });

  it("đọc giá trị override", () => {
    const c = loadConfig({ BFF_PORT: "4000", LOG_LEVEL: "debug", MONGODB_URI: "mongodb://m:27017/x", CORE_BASE_URL: "http://core:8080" });
    expect(c).toMatchObject({ port: 4000, logLevel: "debug", mongoUri: "mongodb://m:27017/x", coreBaseUrl: "http://core:8080" });
  });

  it.each([
    ["LOG_LEVEL", "verbose"],
    ["BFF_PORT", "abc"],
    ["BFF_PORT", "70000"],
    ["MONGODB_URI", "http://not-mongo"],
  ])("%s=%s → lỗi nêu tên biến", (key, value) => {
    expect(() => loadConfig({ [key]: value })).toThrow(key);
  });
});
