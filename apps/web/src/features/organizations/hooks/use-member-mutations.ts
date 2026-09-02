import { useMutation, useQueryClient } from '@tanstack/react-query';
import { inviteMember, removeMember } from '../services/member-service';

export function useMemberMutations(organizationId: string | undefined) {
  const queryClient = useQueryClient();

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['members', organizationId] });

  const invite = useMutation({
    mutationFn: (input: {
      email: string;
      role: 'admin' | 'member' | 'viewer';
    }) => inviteMember(organizationId ?? '', input),
    onSuccess: invalidate,
  });

  const remove = useMutation({
    mutationFn: (memberId: string) =>
      removeMember(organizationId ?? '', memberId),
    onSuccess: invalidate,
  });

  return { invite, remove };
}
