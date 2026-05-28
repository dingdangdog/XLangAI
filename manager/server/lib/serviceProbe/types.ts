export type ServiceProbeResult = {
  ok: boolean;
  latencyMs: number;
  message: string;
  detail?: string;
};
