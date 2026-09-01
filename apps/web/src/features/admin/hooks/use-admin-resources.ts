import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getAdminOrganization,
  getAdminOrganizations,
  getAdminOverview,
  getAdminSubscriptions,
  getAdminTunnels,
  getAdminUsage,
  getAdminUser,
  getAdminUsers,
  setAdminUserStatus,
} from '../services/admin-service';
export function useAdminOverview() {
  return useQuery({
    queryKey: ['admin', 'overview'],
    queryFn: getAdminOverview,
  });
}
export function useAdminUsage() {
  return useQuery({ queryKey: ['admin', 'usage'], queryFn: getAdminUsage });
}
export function useAdminUser(userId: string) {
  return useQuery({
    queryKey: ['admin', 'user', userId],
    queryFn: () => getAdminUser(userId),
    enabled: Boolean(userId),
  });
}
export function useAdminOrganization(organizationId: string) {
  return useQuery({
    queryKey: ['admin', 'organization', organizationId],
    queryFn: () => getAdminOrganization(organizationId),
    enabled: Boolean(organizationId),
  });
}
export function useAdminUsers() {
  return useQuery({ queryKey: ['admin', 'users'], queryFn: getAdminUsers });
}
export function useAdminOrganizations() {
  return useQuery({
    queryKey: ['admin', 'organizations'],
    queryFn: getAdminOrganizations,
  });
}
export function useAdminTunnels() {
  return useQuery({ queryKey: ['admin', 'tunnels'], queryFn: getAdminTunnels });
}
export function useAdminSubscriptions() {
  return useQuery({
    queryKey: ['admin', 'subscriptions'],
    queryFn: getAdminSubscriptions,
  });
}
export function useAdminUserStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      status,
    }: {
      userId: string;
      status: 'active' | 'disabled';
    }) => setAdminUserStatus(userId, status),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] }),
  });
}
