import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  getAdminOrganizations,
  getAdminOverview,
  getAdminSubscriptions,
  getAdminTunnels,
  getAdminUsers,
  setAdminUserStatus,
} from '../services/admin-service';
export function useAdminOverview() {
  return useQuery({
    queryKey: ['admin', 'overview'],
    queryFn: getAdminOverview,
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
