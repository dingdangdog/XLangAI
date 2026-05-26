const LEGACY_AUTH_MESSAGES: Record<string, string> = {
  "用户名或密码不能为空": "AUTH_EMPTY_CREDENTIALS",
  "用户名或密码错误": "AUTH_INVALID_CREDENTIALS",
  "未配置 MANAGER_AUTH_SECRET，无法签发登录会话": "AUTH_SECRET_NOT_CONFIGURED",
};

export function useAuthApiError() {
  const { t } = useI18n();

  function localizeAuthError(payload: unknown, fallbackKey = "toast.loginFailed"): string {
    if (payload && typeof payload === "object") {
      const data = payload as {
        data?: { code?: string; message?: string; statusMessage?: string };
        message?: string;
        statusMessage?: string;
      };
      const code = data.data?.code;
      if (code) {
        const key = `errors.${code}`;
        const translated = t(key);
        if (translated !== key) return translated;
      }

      const message =
        data.data?.message ||
        data.data?.statusMessage ||
        data.message ||
        data.statusMessage;
      if (typeof message === "string" && message.trim()) {
        const legacyCode = LEGACY_AUTH_MESSAGES[message.trim()];
        if (legacyCode) {
          const key = `errors.${legacyCode}`;
          const translated = t(key);
          if (translated !== key) return translated;
        }
        return message;
      }
    }

    return t(fallbackKey);
  }

  return { localizeAuthError };
}
