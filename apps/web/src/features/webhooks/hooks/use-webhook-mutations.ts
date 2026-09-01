import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createWebhook, deleteWebhook } from '../services/webhook-service';

export function useWebhookMutations(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['webhooks', organizationId] });
  const create = useMutation({
    mutationFn: (input: { name: string; url: string; events: string[] }) =>
      createWebhook(organizationId ?? '', input),
    onSuccess: invalidate,
  });
  const remove = useMutation({
    mutationFn: (webhookId: string) =>
      deleteWebhook(organizationId ?? '', webhookId),
    onSuccess: invalidate,
  });
  return { create, remove };
}
