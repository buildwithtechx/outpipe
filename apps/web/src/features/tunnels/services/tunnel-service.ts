import type {
  CreateTunnelRequest,
  Tunnel,
  UpdateTunnelConfigurationRequest,
} from '#/interfaces/tunnel';
import { apiClient } from '#/lib/api-client';

export function getTunnels(organizationID: string) {
  return apiClient.get<Tunnel[]>(
    `/api/v1/organizations/${encodeURIComponent(organizationID)}/tunnels`,
  );
}

export function createTunnel(
  organizationID: string,
  request: CreateTunnelRequest,
) {
  return apiClient.post<Tunnel>(
    `/api/v1/organizations/${encodeURIComponent(organizationID)}/tunnels`,
    request,
  );
}

export function getTunnel(tunnelID: string) {
  return apiClient.get<Tunnel>(
    `/api/v1/tunnels/${encodeURIComponent(tunnelID)}`,
  );
}

export function setTunnelStatus(tunnelID: string, status: Tunnel['status']) {
  return apiClient.patch<void>(
    `/api/v1/tunnels/${encodeURIComponent(tunnelID)}/status`,
    { status },
  );
}

export function revokeTunnel(tunnelID: string) {
  return apiClient.delete<void>(
    `/api/v1/tunnels/${encodeURIComponent(tunnelID)}`,
  );
}

export function updateTunnelConfiguration(
  tunnelID: string,
  request: UpdateTunnelConfigurationRequest,
) {
  return apiClient.patch<Tunnel>(
    `/api/v1/tunnels/${encodeURIComponent(tunnelID)}/config`,
    request,
  );
}
