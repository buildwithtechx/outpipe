import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { UpdateTunnelConfigurationRequest } from '#/interfaces/tunnel';
import { updateTunnelConfiguration } from '../services/tunnel-service';

export function useTunnelConfigurationMutation(tunnelID: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: UpdateTunnelConfigurationRequest) =>
      updateTunnelConfiguration(tunnelID, request),
    onSuccess: (tunnel) => {
      queryClient.setQueryData(['tunnel', tunnelID], tunnel);
    },
  });
}
