/** 与 jimily 一致：统一注入 Authorization 头，所有 API 请求带 credentials */
export function getAuthHeaders(): Record<string, string> {
  return {
    Authorization: useCookie("Authorization").value || "",
  };
}

export function withAuthFetchOptions<T extends Record<string, unknown>>(options: T = {} as T) {
  return {
    ...options,
    headers: {
      ...(options.headers as Record<string, string> | undefined),
      ...getAuthHeaders(),
    },
    credentials: "include" as const,
  };
}
