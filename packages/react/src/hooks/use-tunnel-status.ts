import { useTunnel } from './use-tunnel';

export function useTunnelStatus() {
  const { status, error, tunnel } = useTunnel();

  return { status, error, tunnel };
}
