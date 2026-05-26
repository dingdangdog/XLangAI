const DEFAULT_AUTH_REDIRECT = "/";

function normalizeCallbackUrl(value: unknown): string | null {
  const callbackUrl = Array.isArray(value) ? value[0] : value;
  if (typeof callbackUrl !== "string") return null;

  const trimmed = callbackUrl.trim();
  if (!trimmed) return null;

  if (trimmed.startsWith("/") && !trimmed.startsWith("//")) {
    return trimmed;
  }

  if (import.meta.client && typeof window !== "undefined") {
    try {
      const url = new URL(trimmed, window.location.origin);
      if (url.origin === window.location.origin) {
        return `${url.pathname}${url.search}${url.hash}`;
      }
    } catch {
      return null;
    }
  }

  return null;
}

function isLoginPath(path: string, loginPath: string): boolean {
  return path === loginPath || path.startsWith(`${loginPath}?`);
}

export function resolveAuthRedirect(
  value: unknown,
  loginPath: string,
  fallback = DEFAULT_AUTH_REDIRECT,
): string {
  const callbackUrl = normalizeCallbackUrl(value);
  if (!callbackUrl) return fallback;
  if (isLoginPath(callbackUrl, loginPath)) return fallback;
  return callbackUrl;
}

export function buildLoginRedirect(loginPath: string, callbackUrl?: unknown) {
  const safeCallbackUrl = normalizeCallbackUrl(callbackUrl);
  if (!safeCallbackUrl || isLoginPath(safeCallbackUrl, loginPath)) {
    return loginPath;
  }

  return {
    path: loginPath,
    query: { callbackUrl: safeCallbackUrl },
  };
}

export { DEFAULT_AUTH_REDIRECT };
