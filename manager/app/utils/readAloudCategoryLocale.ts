export type CategoryLocaleEntry = { name: string; description: string };

export type CategoryLocaleLabel = {
  languageId: string;
  languageCode?: string;
  languageName?: string;
  name: string;
  description?: string | null;
};

export function localeLabelsToMap(
  labels: CategoryLocaleLabel[] | undefined,
): Record<string, CategoryLocaleEntry> {
  const out: Record<string, CategoryLocaleEntry> = {};
  for (const row of labels ?? []) {
    const id = String(row.languageId ?? "").trim();
    if (!id) continue;
    out[id] = {
      name: String(row.name ?? "").trim(),
      description: String(row.description ?? "").trim(),
    };
  }
  return out;
}

export function mapToLocaleLabels(
  languageIds: string[],
  map: Record<string, CategoryLocaleEntry>,
): CategoryLocaleLabel[] {
  return languageIds
    .map((languageId) => {
      const entry = map[languageId];
      const name = entry?.name?.trim() ?? "";
      if (!name) return null;
      return {
        languageId,
        name,
        description: entry?.description?.trim() || null,
      };
    })
    .filter((x): x is CategoryLocaleLabel => x != null);
}

export function emptyLocaleMapByLanguageId(languageIds: string[]): Record<string, CategoryLocaleEntry> {
  const out: Record<string, CategoryLocaleEntry> = {};
  for (const id of languageIds) {
    out[id] = { name: "", description: "" };
  }
  return out;
}

export function resolveCategoryDisplayName(
  row: Record<string, unknown>,
  languageId: string,
): string {
  const labels = row.localeLabels as CategoryLocaleLabel[] | undefined;
  const hit = labels?.find((l) => l.languageId === languageId);
  if (hit?.name?.trim()) return hit.name.trim();
  return String(row.name ?? "");
}
