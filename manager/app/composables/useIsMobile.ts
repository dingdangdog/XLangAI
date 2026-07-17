/** 与 Tailwind `md` 断点一致：&lt;768px 视为手机端管理布局 */
const MOBILE_MQ = "(max-width: 767px)";
/** 与 Tailwind `lg` 断点一致：&lt;1024px 用于主从钻取（如跟读） */
const NARROW_MQ = "(max-width: 1023px)";

function useMatchMedia(query: string) {
  const matched = ref(false);

  function update() {
    if (!import.meta.client) return;
    matched.value = window.matchMedia(query).matches;
  }

  onMounted(() => {
    update();
    const mql = window.matchMedia(query);
    const handler = () => {
      matched.value = mql.matches;
    };
    mql.addEventListener("change", handler);
    onUnmounted(() => mql.removeEventListener("change", handler));
  });

  return matched;
}

export function useIsMobile() {
  return useMatchMedia(MOBILE_MQ);
}

export function useIsNarrow() {
  return useMatchMedia(NARROW_MQ);
}
