import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  createIncident,
  createMonitor,
  deleteMonitor,
  getIncidents,
  getMonitorChecks,
  getMonitors,
  getStatusPageConfig,
  testMonitor,
  updateIncident,
  updateStatusPageConfig,
} from '../services/uptime-service';

export function useMonitors(organizationId?: string) {
  return useQuery({
    queryKey: ['uptime-monitors', organizationId],
    queryFn: () => getMonitors(organizationId as string),
    enabled: Boolean(organizationId),
    refetchInterval: 15000,
  });
}

export function useMonitorChecks(organizationId?: string, monitorId?: string) {
  return useQuery({
    queryKey: ['uptime-monitor-checks', organizationId, monitorId],
    queryFn: () =>
      getMonitorChecks(organizationId as string, monitorId as string),
    enabled: Boolean(organizationId && monitorId),
    refetchInterval: 15000,
  });
}

export function useIncidents(organizationId?: string) {
  return useQuery({
    queryKey: ['uptime-incidents', organizationId],
    queryFn: () => getIncidents(organizationId as string),
    enabled: Boolean(organizationId),
  });
}

export function useStatusPageConfig(organizationId?: string) {
  return useQuery({
    queryKey: ['uptime-status-page', organizationId],
    queryFn: () => getStatusPageConfig(organizationId as string),
    enabled: Boolean(organizationId),
  });
}

export function useUptimeMutations(organizationId: string) {
  const queryClient = useQueryClient();

  const addMonitor = useMutation({
    mutationFn: (input: Parameters<typeof createMonitor>[1]) =>
      createMonitor(organizationId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['uptime-monitors', organizationId],
      });
    },
  });

  const removeMonitor = useMutation({
    mutationFn: (monitorId: string) => deleteMonitor(organizationId, monitorId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['uptime-monitors', organizationId],
      });
    },
  });

  const probeMonitor = useMutation({
    mutationFn: (monitorId: string) => testMonitor(organizationId, monitorId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['uptime-monitors', organizationId],
      });
      queryClient.invalidateQueries({
        queryKey: ['uptime-monitor-checks', organizationId],
      });
    },
  });

  const addIncident = useMutation({
    mutationFn: (input: {
      title: string;
      severity: 'minor' | 'major' | 'critical';
      message: string;
    }) => createIncident(organizationId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['uptime-incidents', organizationId],
      });
    },
  });

  const addIncidentUpdate = useMutation({
    mutationFn: ({
      incidentId,
      input,
    }: {
      incidentId: string;
      input: {
        status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
        message: string;
      };
    }) => updateIncident(organizationId, incidentId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['uptime-incidents', organizationId],
      });
    },
  });

  const saveStatusPage = useMutation({
    mutationFn: (input: {
      slug: string;
      title: string;
      description: string;
      is_public: boolean;
      customDomain?: string;
    }) => updateStatusPageConfig(organizationId, input),
    onSuccess: (page) => {
      queryClient.setQueryData(['uptime-status-page', organizationId], page);
      return queryClient.invalidateQueries({
        queryKey: ['uptime-status-page', organizationId],
      });
    },
  });

  return {
    addMonitor,
    removeMonitor,
    probeMonitor,
    addIncident,
    addIncidentUpdate,
    saveStatusPage,
  };
}
