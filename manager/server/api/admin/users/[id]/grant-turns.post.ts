import { createError } from "h3";
import prisma from "~~/server/lib/prisma";

const MAX_GRANT_TURNS = 1_000_000;
const MAX_REASON_LENGTH = 200;

export default defineEventHandler(async (event) => {
	const userId = getRouterParam(event, "id")?.trim();
	if (!userId) {
		throw createError({ statusCode: 400, message: "缺少用户 ID" });
	}

	const body = (await readBody(event)) as Record<string, unknown> | null;
	const amount = Math.floor(Number(body?.amount));
	const reason = String(body?.reason ?? "").trim();
	if (!Number.isFinite(amount) || amount <= 0 || amount > MAX_GRANT_TURNS) {
		throw createError({
			statusCode: 400,
			message: `发放次数须为 1–${MAX_GRANT_TURNS} 的整数`,
		});
	}
	if (!reason || reason.length > MAX_REASON_LENGTH) {
		throw createError({
			statusCode: 400,
			message: `发放原因必填，且不能超过 ${MAX_REASON_LENGTH} 个字符`,
		});
	}

	const result = await prisma.$transaction(async (tx) => {
		const existing = await tx.user.findFirst({
			where: { id: userId, deletedAt: null },
			select: { id: true, turnBalance: true },
		});
		if (!existing) {
			throw createError({ statusCode: 404, message: "用户不存在或已删除" });
		}
		const updated = await tx.user.update({
			where: { id: userId },
			data: { turnBalance: { increment: amount } },
			select: { id: true, turnBalance: true },
		});
		return {
			userId: updated.id,
			amount,
			before: existing.turnBalance,
			after: updated.turnBalance,
		};
	});

	const manager = event.context.managerAuth as
		| { id?: string; username?: string }
		| undefined;
	console.info(
		"[quota-grant]",
		JSON.stringify({
			...result,
			reason,
			managerId: manager?.id ?? "unknown",
			managerUsername: manager?.username ?? "unknown",
			createdAt: new Date().toISOString(),
		}),
	);

	return result;
});
