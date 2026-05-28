export type ServiceProbeResult = {
  ok: boolean;
  latencyMs: number;
  message: string;
  detail?: string;
};

export function useServiceConfigProbe(resourceBase: string) {
  const requestFetch = useRequestFetch();
  const probing = ref(false);
  const lastResult = ref<ServiceProbeResult | null>(null);

  async function probe(id: string, body?: Record<string, unknown>) {
    probing.value = true;
    lastResult.value = null;
    try {
      const result = await requestFetch<ServiceProbeResult>(`${resourceBase}/${id}/probe`, {
        method: "POST",
        body: body ?? {},
      });
      lastResult.value = result;
      return result;
    } finally {
      probing.value = false;
    }
  }

  function clearProbe() {
    lastResult.value = null;
  }

  return { probe, probing, lastResult, clearProbe };
}
