import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createApiKey, revokeApiKey } from '../services/api-key-service';

export function useApiKeyMutations(organizationId: string | undefined) {
  const queryClient = useQueryClient();

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['api-keys', organizationId] });

  const create = useMutation({
    mutationFn: (input: {
      name: string;
      scopes: string[];
      expiresAt?: string;
    }) => createApiKey(organizationId ?? '', input),
    onSuccess: invalidate,
  });

  const revoke = useMutation({
    mutationFn: (apiKeyId: string) =>
      revokeApiKey(organizationId ?? '', apiKeyId),
    onSuccess: invalidate,
  });

  return { create, revoke };
}
