import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ConfirmAction } from '#/components/feedback/confirm-action';
import type { Tunnel } from '#/interfaces/tunnel';
import { revokeTunnel } from '../services/tunnel-service';

export function TunnelDetailActions({ tunnel }: { tunnel: Tunnel }) {
  const queryClient = useQueryClient();
  const invalidate = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ['tunnel', tunnel.id] }),
      queryClient.invalidateQueries({
        queryKey: ['tunnels', tunnel.organizationId],
      }),
    ]);
  };
  const revoke = useMutation({
    mutationFn: () => revokeTunnel(tunnel.id),
    onSuccess: invalidate,
  });
  const pending = revoke.isPending;
  if (tunnel.status === 'revoked' || tunnel.status === 'expired') return null;
  return (
    <div className="flex flex-wrap items-center gap-3">
      <ConfirmAction
        title="Revoke this endpoint?"
        description="This permanently revokes the endpoint and its access. Your agent will no longer be able to connect to it. Create a new tunnel if you need another endpoint."
        label="Revoke tunnel"
        pending={pending}
        onConfirm={() => revoke.mutateAsync()}
      />
    </div>
  );
}
