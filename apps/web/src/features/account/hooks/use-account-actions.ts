import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  deleteAccount,
  transferOrganizationOwnership,
} from '../services/account-service';

export function useAccountActions(organizationId?: string) {
  const queryClient = useQueryClient();
  const remove = useMutation({
    mutationFn: deleteAccount,
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: ['auth', 'session'] });
      window.location.assign('/');
    },
  });
  const transfer = useMutation({
    mutationFn: (newOwnerId: string) =>
      transferOrganizationOwnership(organizationId ?? '', newOwnerId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['organizations'] }),
  });
  return { remove, transfer };
}
