import prisma from "../lib/prisma";

export type CategoryLocaleInput = {
  languageId: string;
  name: string;
  description?: string | null;
};

export function extractLocaleLabels(body: Record<string, unknown>): CategoryLocaleInput[] {
  const raw = body.localeLabels;
  if (!Array.isArray(raw)) return [];
  const out: CategoryLocaleInput[] = [];
  for (const item of raw) {
    if (!item || typeof item !== "object") continue;
    const row = item as Record<string, unknown>;
    const languageId = String(row.languageId ?? "").trim();
    const name = String(row.name ?? "").trim();
    if (!languageId || !name) continue;
    const description = row.description != null ? String(row.description).trim() : "";
    out.push({
      languageId,
      name,
      description: description || null,
    });
  }
  return out;
}

/** 从写入 body 中剥离虚拟字段，避免 Prisma 报错 */
export function stripReadAloudCategoryVirtualFields(data: Record<string, unknown>) {
  const next = { ...data };
  delete next.localeLabels;
  delete next.localizedNames;
  return next;
}

export async function saveReadAloudCategoryLocales(
  categoryId: string,
  labels: CategoryLocaleInput[],
) {
  await prisma.$transaction([
    prisma.readAloudCategoryLocale.deleteMany({ where: { categoryId } }),
    ...labels.map((l) =>
      prisma.readAloudCategoryLocale.create({
        data: {
          categoryId,
          languageId: l.languageId,
          name: l.name,
          description: l.description ?? null,
        },
      }),
    ),
  ]);
}

export async function deleteReadAloudCategoryLocales(categoryId: string) {
  await prisma.readAloudCategoryLocale.deleteMany({ where: { categoryId } });
}

/** 兼容旧客户端：同步 en 语言行到 name_en / description_en */
export async function legacyEnFieldsFromLocales(
  labels: CategoryLocaleInput[],
): Promise<{ nameEn: string | null; descriptionEn: string | null } | null> {
  if (!labels.length) return null;
  const langs = await prisma.language.findMany({
    where: { id: { in: labels.map((l) => l.languageId) } },
  });
  const enLang = langs.find((l) => String(l.code).toLowerCase() === "en");
  if (!enLang) return null;
  const row = labels.find((l) => l.languageId === enLang.id);
  if (!row) return null;
  return {
    nameEn: row.name,
    descriptionEn: row.description ?? null,
  };
}

export async function attachReadAloudCategoryLocales(
  rows: Record<string, unknown>[],
): Promise<Record<string, unknown>[]> {
  if (!rows.length) return rows;
  const categoryIds = rows
    .map((r) => r.id)
    .filter((id): id is string => typeof id === "string" && id.length > 0);
  if (!categoryIds.length) return rows;

  const locales = await prisma.readAloudCategoryLocale.findMany({
    where: { categoryId: { in: categoryIds } },
  });
  const langIds = [...new Set(locales.map((l) => l.languageId))];
  const languages = langIds.length
    ? await prisma.language.findMany({ where: { id: { in: langIds } } })
    : [];
  const langById = new Map(languages.map((l) => [l.id, l]));

  const byCategory = new Map<string, typeof locales>();
  for (const loc of locales) {
    const list = byCategory.get(loc.categoryId) ?? [];
    list.push(loc);
    byCategory.set(loc.categoryId, list);
  }

  return rows.map((row) => {
    const id = String(row.id ?? "");
    const list = byCategory.get(id) ?? [];
    const localeLabels = list.map((loc) => {
      const lang = langById.get(loc.languageId);
      return {
        languageId: loc.languageId,
        languageCode: lang?.code ?? "",
        languageName: lang?.name ?? "",
        name: loc.name,
        description: loc.description,
      };
    });
    return { ...row, localeLabels };
  });
}
