import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '#/components/ui/button';
import type { Tunnel } from '#/interfaces/tunnel';
import { setTunnelCapture } from '../services/tunnel-service';

export function TunnelCaptureControl({ tunnel }: { tunnel: Tunnel }) {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: () => setTunnelCapture(tunnel.id, !tunnel.captureEnabled),
    onSuccess: (updated) => {
      queryClient.setQueryData(['tunnel', tunnel.id], updated);
      queryClient.invalidateQueries({
        queryKey: ['tunnels', tunnel.organizationId],
      });
    },
  });

  return (
    <div className="mt-6 rounded-xl border border-white/10 p-4">
      <h2 className="text-sm font-medium">Request capture</h2>
      <p className="mt-1 text-sm text-white/60">
        Capture bounded JSON payloads and headers for inspection. Credential
        fields are redacted.
      </p>
      <Button
        className="mt-3"
        variant="outline"
        disabled={mutation.isPending || tunnel.status === 'revoked'}
        onClick={() => mutation.mutate()}
        aria-pressed={Boolean(tunnel.captureEnabled)}
      >
        {mutation.isPending
          ? 'Updating…'
          : tunnel.captureEnabled
            ? 'Disable capture'
            : 'Enable capture'}
      </Button>
      {mutation.isError && (
        <p role="alert" className="mt-2 text-sm text-rose-300">
          Capture settings could not be saved.
        </p>
      )}
    </div>
  );
}
