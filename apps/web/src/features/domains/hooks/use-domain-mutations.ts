import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createDomain, verifyDomain } from '../services/domain-service';

export function useDomainMutations(organizationId: string | undefined) {
  const queryClient = useQueryClient();

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['domains', organizationId] });

  const create = useMutation({
    mutationFn: (input: { hostname: string; verificationMethod: string }) =>
      createDomain(organizationId ?? '', input),
    onSuccess: invalidate,
  });

  const verify = useMutation({
    mutationFn: (input: { domainId: string; token: string }) =>
      verifyDomain(input.domainId, input.token),
    onSuccess: invalidate,
  });

  return { create, verify };
}
