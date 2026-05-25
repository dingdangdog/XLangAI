export type ToastType = "success" | "error" | "warning" | "info";

export interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
}

const toasts = ref<ToastItem[]>([]);
let nextId = 1;

export function useToast() {
  function push(type: ToastType, message: string, durationMs = 3200) {
    const id = nextId++;
    toasts.value = [...toasts.value, { id, type, message }];
    if (import.meta.client) {
      window.setTimeout(() => {
        toasts.value = toasts.value.filter((t) => t.id !== id);
      }, durationMs);
    }
  }

  return {
    toasts: readonly(toasts),
    success: (message: string) => push("success", message),
    error: (message: string) => push("error", message),
    warning: (message: string) => push("warning", message),
    info: (message: string) => push("info", message),
    dismiss: (id: number) => {
      toasts.value = toasts.value.filter((t) => t.id !== id);
    },
  };
}
