import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CircleStop, LoaderCircle, Play, ShieldOff } from 'lucide-react';
import { Button } from '#/components/ui/button';
import type { Tunnel } from '#/interfaces/tunnel';
import { revokeTunnel, setTunnelStatus } from '../services/tunnel-service';

type TunnelDetailActionsProps = {
  tunnel: Tunnel;
};

export function TunnelDetailActions({ tunnel }: TunnelDetailActionsProps) {
  const queryClient = useQueryClient();
  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ['tunnel', tunnel.id] });
  const statusMutation = useMutation({
    mutationFn: (status: Tunnel['status']) =>
      setTunnelStatus(tunnel.id, status),
    onSuccess: invalidate,
  });
  const revokeMutation = useMutation({
    mutationFn: () => revokeTunnel(tunnel.id),
    onSuccess: invalidate,
  });
  const isRevoked = tunnel.status === 'revoked';
  const isPending = statusMutation.isPending || revokeMutation.isPending;
  const error = statusMutation.isError || revokeMutation.isError;

  return (
    <div className="flex flex-wrap items-center gap-3">
      {tunnel.status === 'active' ? (
        <Button
          type="button"
          variant="outline"
          onClick={() => statusMutation.mutate('disconnected')}
          disabled={isPending || isRevoked}
          className="border-white/10 bg-white/[0.03] text-white hover:bg-white/10 hover:text-white"
        >
          {statusMutation.isPending ? (
            <LoaderCircle className="animate-spin" />
          ) : (
            <CircleStop />
          )}
          Disconnect
        </Button>
      ) : (
        <Button
          type="button"
          onClick={() => statusMutation.mutate('active')}
          disabled={isPending || isRevoked}
          className="bg-white text-black hover:bg-indigo-100"
        >
          {statusMutation.isPending ? (
            <LoaderCircle className="animate-spin" />
          ) : (
            <Play />
          )}
          Mark active
        </Button>
      )}
      <Button
        type="button"
        variant="outline"
        onClick={() => revokeMutation.mutate()}
        disabled={isPending || isRevoked}
        className="border-rose-300/20 bg-rose-400/[0.04] text-rose-100 hover:bg-rose-400/10 hover:text-rose-50"
      >
        {revokeMutation.isPending ? (
          <LoaderCircle className="animate-spin" />
        ) : (
          <ShieldOff />
        )}
        Revoke
      </Button>
      {error && (
        <p role="alert" className="basis-full text-sm text-rose-200">
          The tunnel action did not complete. Please try again.
        </p>
      )}
    </div>
  );
}
