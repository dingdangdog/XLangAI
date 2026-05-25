import bcrypt from "bcryptjs";
import type { User } from "~~/prisma/generated/client";
import prisma from "../lib/prisma";
import { MANAGER_ADMIN_REMARK } from "../lib/managerAdminSeed";

export function parseManagerLoginUsername(raw: unknown): string {
  return typeof raw === "string" ? raw.trim() : "";
}

export function isEmailLogin(username: string): boolean {
  return username.includes("@");
}

export async function findManagerAdminByUsername(username: string): Promise<User | null> {
  if (!username) return null;

  const where = isEmailLogin(username)
    ? { email: username, deletedAt: null, remark: MANAGER_ADMIN_REMARK, status: "active" }
    : { phone: username, deletedAt: null, remark: MANAGER_ADMIN_REMARK, status: "active" };

  return prisma.user.findFirst({ where });
}

export async function verifyManagerPassword(user: User, plainPassword: string): Promise<boolean> {
  const hash = user.passwordHash;
  if (!hash || !plainPassword) return false;
  return bcrypt.compare(plainPassword, hash);
}

export function toManagerAuthUser(user: User, username: string) {
  return {
    id: user.id,
    username,
    nickname: user.nickname,
    phone: user.phone,
    email: user.email,
    isManagerAdmin: true as const,
  };
}
