/** 后台新建记录时自动生成唯一 code，运营无需手填 */
export function autoAdminCode(seed?: string): string {
  const base =
    (seed ?? "")
      .trim()
      .toLowerCase()
      .replace(/\s+/g, "_")
      .replace(/[^a-z0-9_-]/g, "")
      .slice(0, 28) || "item";
  return `${base}-${Date.now()}`;
}
