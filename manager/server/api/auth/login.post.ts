import { createError } from "h3";
import {
  findManagerAdminByUsername,
  parseManagerLoginUsername,
  toManagerAuthUser,
  verifyManagerPassword,
} from "../../utils/managerAuth";
import { signManagerAuthToken } from "../../utils/jwt";

const SESSION_DAYS = 30;

function authError(statusCode: number, code: string, message: string) {
  throw createError({ statusCode, message, data: { code } });
}

export default defineEventHandler(async (event) => {
  const body = await readBody(event);
  const username = parseManagerLoginUsername(body?.username);
  const password = typeof body?.password === "string" ? body.password : "";

  if (!username || !password) {
    authError(400, "AUTH_EMPTY_CREDENTIALS", "用户名或密码不能为空");
  }

  const secret = useRuntimeConfig().manager.authSecret?.trim();
  if (!secret) {
    authError(500, "AUTH_SECRET_NOT_CONFIGURED", "未配置 MANAGER_AUTH_SECRET，无法签发登录会话");
  }

  const user = await findManagerAdminByUsername(username);
  if (!user || !(await verifyManagerPassword(user, password))) {
    authError(401, "AUTH_INVALID_CREDENTIALS", "用户名或密码错误");
  }

  const expiresInSeconds = SESSION_DAYS * 24 * 60 * 60;
  const authUser = toManagerAuthUser(user, username);
  const token = signManagerAuthToken(
    {
      id: authUser.id,
      username: authUser.username,
      nickname: authUser.nickname,
      isManagerAdmin: true,
    },
    secret,
    expiresInSeconds,
  );

  const isProduction = process.env.NODE_ENV === "production";
  setCookie(event, "Authorization", token, {
    maxAge: expiresInSeconds,
    path: "/",
    httpOnly: true,
    sameSite: "lax",
    secure: isProduction,
  });

  return authUser;
});
