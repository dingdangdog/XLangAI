/**
 * 管理后台 CRUD 列表默认排序。
 * 主排序字段相同时必须用 id 等作稳定次序，否则更新行（如生成试听触发 updatedAt）后列表会乱跳。
 */
export type AdminListOrderBy = ReadonlyArray<Record<string, "asc" | "desc">>;

const idAsc = { id: "asc" } as const;

/** sortOrder → code → id（服务配置、语言、会员等级、提示词等） */
export function orderBySortOrderCode(): AdminListOrderBy {
  return [{ sortOrder: "asc" }, { code: "asc" }, idAsc];
}

/**
 * 语音角色（占位；实际列表走 voiceRoleList.ts 的 JOIN 排序）。
 * 顺序：语言 sort_order → 语言 code → 角色 sort_order → name → id
 */
export function orderByVoiceRole(): AdminListOrderBy {
  return [{ languageId: "asc" }, { sortOrder: "asc" }, { name: "asc" }, idAsc];
}

/** 系统变量：key → id */
export function orderByKeyAsc(): AdminListOrderBy {
  return [{ key: "asc" }, idAsc];
}

/** 用户 / 会话 / 消息：创建时间倒序 → id */
export function orderByCreatedDesc(): AdminListOrderBy {
  return [{ createdAt: "desc" }, idAsc];
}

/** 用量：日期倒序 → 用户 → id */
export function orderByUsageDateDesc(): AdminListOrderBy {
  return [{ date: "desc" }, { userId: "asc" }, idAsc];
}

/** 备份：注销时间倒序 → id */
export function orderByCancelledDesc(): AdminListOrderBy {
  return [{ cancelledAt: "desc" }, idAsc];
}
