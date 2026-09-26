import { MongoClient } from "mongodb";

/** Thứ /readyz cần: kiểm tra kết nối tới MongoDB. */
export interface Pinger {
  ping(signal: AbortSignal): Promise<void>;
}

/**
 * Tạo MongoClient. Driver kết nối lười (khi có thao tác đầu tiên), nên BFF vẫn khởi động được
 * khi MongoDB chưa sẵn sàng; /readyz báo trạng thái thật.
 */
export function createMongo(uri: string): { client: MongoClient; pinger: Pinger } {
  const client = new MongoClient(uri, { serverSelectionTimeoutMS: 2000, connectTimeoutMS: 2000 });
  const pinger: Pinger = {
    async ping(signal) {
      await Promise.race([
        client.db("admin").command({ ping: 1 }),
        new Promise<never>((_, reject) => {
          signal.addEventListener("abort", () => reject(signal.reason), { once: true });
        }),
      ]);
    },
  };
  return { client, pinger };
}
