export interface AdminTabRouteOptions<T extends string> {
  tabKeys: readonly T[];
  defaultTab: T;
  queryKey?: string;
}

function parseTabQuery<T extends string>(
  value: unknown,
  tabKeys: readonly T[],
): T | null {
  const raw = Array.isArray(value) ? value[0] : value;
  if (typeof raw !== "string") return null;
  return tabKeys.includes(raw as T) ? (raw as T) : null;
}

export function useAdminTabRoute<T extends string>(options: AdminTabRouteOptions<T>) {
  const { tabKeys, defaultTab, queryKey = "tab" } = options;
  const route = useRoute();
  const router = useRouter();

  const activeTab = ref<T>(defaultTab);

  function replaceUrlTab(tab: T) {
    if (parseTabQuery(route.query[queryKey], tabKeys) === tab) return;
    router.replace({
      query: { ...route.query, [queryKey]: tab },
    });
  }

  function syncTabFromUrl() {
    const tabFromUrl = parseTabQuery(route.query[queryKey], tabKeys);
    if (tabFromUrl) {
      activeTab.value = tabFromUrl;
      return;
    }
    activeTab.value = defaultTab;
    replaceUrlTab(defaultTab);
  }

  function setTab(tab: T) {
    if (activeTab.value === tab) return;
    activeTab.value = tab;
    replaceUrlTab(tab);
  }

  onMounted(syncTabFromUrl);
  watch(() => route.query[queryKey], syncTabFromUrl);

  return { activeTab, setTab, syncTabFromUrl };
}
