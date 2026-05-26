export type ToastType = "success" | "error" | "warning" | "info";

export interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
}

const timers = new Map<number, ReturnType<typeof setTimeout>>();

export function useToast() {
  const toasts = useState<ToastItem[]>("manager-toasts", () => []);
  const nextId = useState("manager-toast-next-id", () => 1);

  function clearTimer(id: number) {
    const timer = timers.get(id);
    if (timer !== undefined) {
      clearTimeout(timer);
      timers.delete(id);
    }
  }

  function push(type: ToastType, message: string, durationMs = 3200) {
    if (!import.meta.client) return;

    const id = nextId.value++;
    toasts.value = [...toasts.value, { id, type, message }];

    clearTimer(id);
    timers.set(
      id,
      window.setTimeout(() => {
        timers.delete(id);
        toasts.value = toasts.value.filter((t) => t.id !== id);
      }, durationMs),
    );
  }

  function dismiss(id: number) {
    clearTimer(id);
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  return {
    toasts: readonly(toasts),
    success: (message: string) => push("success", message),
    error: (message: string) => push("error", message),
    warning: (message: string) => push("warning", message),
    info: (message: string) => push("info", message),
    dismiss,
  };
}
