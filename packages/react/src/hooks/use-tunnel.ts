import { useContext } from 'react';
import { TunnelContext } from '../services/tunnel-provider';

export function useTunnel() {
  const context = useContext(TunnelContext);

  if (!context) {
    throw new Error('useTunnel must be used inside TunnelProvider');
  }

  return context;
}
