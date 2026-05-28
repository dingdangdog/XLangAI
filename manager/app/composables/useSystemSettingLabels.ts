/** DB key 如 auth.password.enabled；i18n 为嵌套 JSON，用 pages.systemSettings.keys.auth.password.enabled 解析 */

export function useSystemSettingLabels() {
  const { t, te, locale } = useI18n();

  function labelForKey(k: string) {
    const path = `pages.systemSettings.keys.${k}`;
    return te(path) ? t(path) : k;
  }

  function hintForKey(k: string) {
    const path = `pages.systemSettings.hints.${k}`;
    return te(path) ? t(path) : "";
  }

  return { labelForKey, hintForKey, locale };
}
