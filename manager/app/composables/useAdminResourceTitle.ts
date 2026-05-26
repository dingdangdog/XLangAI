export function useAdminResourceTitle() {
  const { t } = useI18n();

  function adminResourceTitle(slug: string): string {
    const key = `resources.${slug.replace(/-/g, "_")}`;
    const translated = t(key);
    return translated === key ? slug : translated;
  }

  return { adminResourceTitle };
}
