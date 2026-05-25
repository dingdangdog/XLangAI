import { PrismaClient } from "~~/prisma/generated/client";
import { withAccelerate } from "@prisma/extension-accelerate";
import { PrismaPg } from "@prisma/adapter-pg";

const prismaClientSingleton = () => {
  return new PrismaClient({
    adapter: new PrismaPg({
      connectionString: process.env.DATABASE_URL,
    }),
  }).$extends(withAccelerate());
};

/** Nitro 注入的 `prisma` 实例类型（含 Accelerate 扩展），种子函数参数应使用本类型而非裸 `PrismaClient` */
export type AppPrismaClient = ReturnType<typeof prismaClientSingleton>;

declare const globalThis: {
  prismaGlobal: AppPrismaClient;
} & typeof global;

const prisma = globalThis.prismaGlobal ?? prismaClientSingleton();

export default prisma;

if (process.env.NODE_ENV !== "production") globalThis.prismaGlobal = prisma;
