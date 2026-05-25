import jwt from "jsonwebtoken";
import type { H3Event } from "h3";

export type ManagerAuthPayload = {
  id: string;
  username: string;
  nickname?: string | null;
  isManagerAdmin: boolean;
};

function readToken(event: H3Event): string | null {
  let token = getHeader(event, "Authorization") || getCookie(event, "Authorization");
  if (token?.startsWith("Bearer ")) token = token.slice(7);
  return token?.trim() || null;
}

export function getManagerAuthPayload(event: H3Event): ManagerAuthPayload | null {
  const token = readToken(event);
  const secret = useRuntimeConfig().manager.authSecret?.trim();
  if (!token || !secret) return null;

  try {
    const decoded = jwt.verify(token, secret) as ManagerAuthPayload & {
      iat?: number;
      exp?: number;
    };
    if (!decoded?.id || !decoded.isManagerAdmin) return null;
    return {
      id: String(decoded.id),
      username: String(decoded.username ?? ""),
      nickname: decoded.nickname ?? null,
      isManagerAdmin: true,
    };
  } catch {
    return null;
  }
}

export function signManagerAuthToken(
  payload: ManagerAuthPayload,
  secret: string,
  expiresInSeconds: number,
): string {
  return jwt.sign(payload, secret, { expiresIn: expiresInSeconds });
}
