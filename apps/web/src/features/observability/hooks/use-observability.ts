import { useMutation, useQuery } from '@tanstack/react-query';
import {
  executeReplay,
  getCapture,
  getCaptures,
  getLogs,
  getObservabilityStats,
  getTraces,
  getTraceWaterfall,
  type ReplayRequestInput,
} from '../services/observability-service';

export function useObservabilityStats(organizationId?: string, range = '24h') {
  return useQuery({
    queryKey: ['observability-stats', organizationId, range],
    queryFn: () => getObservabilityStats(organizationId as string, range),
    enabled: Boolean(organizationId),
    refetchInterval: 30000,
  });
}

export function useTraces(organizationId?: string, limit = 50) {
  return useQuery({
    queryKey: ['observability-traces', organizationId, limit],
    queryFn: () => getTraces(organizationId as string, limit),
    enabled: Boolean(organizationId),
    refetchInterval: 15000,
  });
}

export function useTraceWaterfall(organizationId?: string, traceId?: string) {
  return useQuery({
    queryKey: ['observability-waterfall', organizationId, traceId],
    queryFn: () =>
      getTraceWaterfall(organizationId as string, traceId as string),
    enabled: Boolean(organizationId && traceId),
  });
}

export function useLogs(
  organizationId?: string,
  search?: string,
  severity?: string,
  limit = 100,
) {
  return useQuery({
    queryKey: ['observability-logs', organizationId, search, severity, limit],
    queryFn: () => getLogs(organizationId as string, search, severity, limit),
    enabled: Boolean(organizationId),
    refetchInterval: 10000,
  });
}

export function useCaptures(
  organizationId?: string,
  tunnelId?: string,
  limit = 50,
) {
  return useQuery({
    queryKey: ['observability-captures', organizationId, tunnelId, limit],
    queryFn: () => getCaptures(organizationId as string, tunnelId, limit),
    enabled: Boolean(organizationId),
    refetchInterval: 10000,
  });
}

export function useCapture(organizationId?: string, captureId?: string) {
  return useQuery({
    queryKey: ['observability-capture', organizationId, captureId],
    queryFn: () => getCapture(organizationId as string, captureId as string),
    enabled: Boolean(organizationId && captureId),
  });
}

export function useReplayMutation(organizationId: string) {
  return useMutation({
    mutationFn: (input: ReplayRequestInput) =>
      executeReplay(organizationId, input),
  });
}
