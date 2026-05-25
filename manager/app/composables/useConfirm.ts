export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
}

const visible = ref(false);
const loading = ref(false);
const options = ref<ConfirmOptions>({ message: "" });
let resolver: ((value: boolean) => void) | null = null;

export function useConfirm() {
  async function confirm(opts: ConfirmOptions): Promise<boolean> {
    if (resolver) {
      resolver(false);
    }
    options.value = {
      title: opts.title ?? "确认",
      message: opts.message,
      confirmLabel: opts.confirmLabel ?? "确定",
      cancelLabel: opts.cancelLabel ?? "取消",
      danger: opts.danger ?? false,
    };
    visible.value = true;
    return new Promise<boolean>((resolve) => {
      resolver = resolve;
    });
  }

  function resolveConfirm(ok: boolean) {
    visible.value = false;
    const r = resolver;
    resolver = null;
    r?.(ok);
  }

  return {
    visible: readonly(visible),
    loading: readonly(loading),
    options: readonly(options),
    confirm,
    resolveConfirm,
    setLoading: (v: boolean) => {
      loading.value = v;
    },
  };
}
