import prisma from "../lib/prisma";
import type { VoiceRole } from "~~/prisma/generated/client";

/**
 * 语音角色列表：语言 sort_order → 语言 code → 角色 sort_order → name → id。
 * 须 JOIN sys_languages，不能用仅 languageId 的 Prisma orderBy（无法按语言 code 排序）。
 */
export async function listVoiceRolesForAdmin(args: {
  skip: number;
  take: number;
}): Promise<VoiceRole[]> {
  const idRows = await prisma.$queryRaw<{ id: string }[]>`
    SELECT vr.id
    FROM sys_voice_roles vr
    LEFT JOIN sys_languages l ON l.id = vr.language_id
    ORDER BY
      l.sort_order ASC NULLS LAST,
      l.code ASC NULLS LAST,
      vr.sort_order ASC,
      vr.name ASC,
      vr.id ASC
    OFFSET ${args.skip}
    LIMIT ${args.take}
  `;
  const ids = idRows.map((r) => r.id);
  if (ids.length === 0) return [];

  const rows = await prisma.voiceRole.findMany({ where: { id: { in: ids } } });
  const byId = new Map(rows.map((r) => [r.id, r]));
  return ids.map((id) => byId.get(id)).filter((r): r is VoiceRole => r != null);
}
