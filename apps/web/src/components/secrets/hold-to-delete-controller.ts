export interface HoldToDeleteControllerOptions {
  durationMs?: number;
  tickIntervalMs?: number;
  now?: () => number;
  schedule?: (callback: () => void, delayMs: number) => unknown;
  cancelScheduled?: (handle: unknown) => void;
  onProgress?: (progress: number) => void;
  onConfirm?: () => void;
}

export function createHoldToDeleteController(
  options: HoldToDeleteControllerOptions = {},
) {
  const durationMs = options.durationMs ?? 1200;
  const tickIntervalMs = options.tickIntervalMs ?? 16;
  const now = options.now ?? (() => performance.now());
  const schedule =
    options.schedule ?? ((cb, delay) => window.setTimeout(cb, delay));
  const cancelScheduled =
    options.cancelScheduled ??
    ((handle) => window.clearTimeout(handle as number));

  let startTime: number | null = null;
  let activeHandle: unknown = null;
  let currentProgress = 0;
  let activeSource: string | null = null;

  function updateProgress(value: number) {
    currentProgress = Math.min(1, Math.max(0, value));
    options.onProgress?.(currentProgress);
  }

  function tick() {
    if (startTime === null) return;
    const elapsed = now() - startTime;
    const progress = Math.min(1, elapsed / durationMs);
    updateProgress(progress);

    if (progress >= 1) {
      startTime = null;
      activeHandle = null;
      activeSource = null;
      options.onConfirm?.();
      return;
    }

    activeHandle = schedule(tick, tickIntervalMs);
  }

  function start(source = 'default') {
    if (startTime !== null) return;
    activeSource = source;
    startTime = now();
    updateProgress(0);
    activeHandle = schedule(tick, tickIntervalMs);
  }

  function cancel(source?: string) {
    if (
      source &&
      activeSource &&
      source !== activeSource &&
      !source.startsWith('pointer:')
    ) {
      return;
    }
    if (activeHandle !== null) {
      cancelScheduled(activeHandle);
      activeHandle = null;
    }
    startTime = null;
    activeSource = null;
    updateProgress(0);
  }

  return {
    start,
    cancel,
    getProgress: () => currentProgress,
    isHolding: () => startTime !== null,
  };
}
