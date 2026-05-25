import { createError } from "h3";
import prisma from "../../lib/prisma";
import { getManagerAuthPayload } from "../../utils/jwt";
import { MANAGER_ADMIN_REMARK } from "../../lib/managerAdminSeed";
import { toManagerAuthUser } from "../../utils/managerAuth";

export default defineEventHandler(async (event) => {
  const payload = getManagerAuthPayload(event);
  if (!payload?.id) {
    throw createError({ statusCode: 401, message: "未登录或登录失效" });
  }

  const user = await prisma.user.findFirst({
    where: {
      id: payload.id,
      deletedAt: null,
      remark: MANAGER_ADMIN_REMARK,
      status: "active",
    },
  });

  if (!user) {
    throw createError({ statusCode: 401, message: "账号不可用或已失效" });
  }

  const username =
    payload.username ||
    user.phone ||
    user.email ||
    "";

  return toManagerAuthUser(user, username);
});
