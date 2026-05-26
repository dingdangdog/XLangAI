import {
  buildLoginRedirect as buildLoginRedirectPath,
  resolveAuthRedirect as resolveAuthRedirectPath,
} from "~/utils/authRedirect";

export function useAuthRedirect() {
  const localePath = useLocalePath();
  const loginPath = computed(() => localePath("/login"));

  function resolveAuthRedirect(value: unknown, fallback = "/") {
    return resolveAuthRedirectPath(value, loginPath.value, fallback);
  }

  function buildLoginRedirect(callbackUrl?: unknown) {
    return buildLoginRedirectPath(loginPath.value, callbackUrl);
  }

  return {
    loginPath,
    resolveAuthRedirect,
    buildLoginRedirect,
  };
}
