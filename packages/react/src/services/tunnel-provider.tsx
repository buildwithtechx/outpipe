import {
  type OpenTunnel,
  type OpenTunnelAck,
  RelayConnection,
  type RelayConnectionOptions,
} from '@outpipe/sdk';
import {
  createContext,
  type PropsWithChildren,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import type { TunnelContextValue, TunnelStatus } from '../interfaces/tunnel';

export const TunnelContext = createContext<TunnelContextValue | undefined>(
  undefined,
);

export type TunnelProviderOptions = PropsWithChildren<{
  options: RelayConnectionOptions;
}>;

export function TunnelProvider({ options, children }: TunnelProviderOptions) {
  const connection = useMemo(() => new RelayConnection(options), [options]);
  const [status, setStatus] = useState<TunnelStatus>('idle');
  const [tunnel, setTunnel] = useState<OpenTunnelAck>();
  const [error, setError] = useState<Error>();

  useEffect(() => {
    const removeConnected = connection.on('connected', () => {
      setStatus('connected');
      setError(undefined);
    });
    const removeDisconnected = connection.on('disconnected', () => {
      setStatus('closed');
      setTunnel(undefined);
    });
    const removeError = connection.on('error', (nextError) => {
      setStatus('error');
      setError(nextError);
    });
    const removeTunnelOpened = connection.on('tunnel_opened', setTunnel);
    const removeTunnelClosed = connection.on('tunnel_closed', (closed) => {
      setTunnel((current) =>
        current?.tunnel_id === closed.tunnel_id ? undefined : current,
      );
    });

    return () => {
      removeConnected();
      removeDisconnected();
      removeError();
      removeTunnelOpened();
      removeTunnelClosed();
      connection.close();
    };
  }, [connection]);

  const connect = useCallback(async () => {
    setStatus('connecting');
    setError(undefined);
    try {
      await connection.connect();
      setStatus('connected');
    } catch (nextError) {
      const normalized = normalizeError(nextError);
      setStatus('error');
      setError(normalized);
      throw normalized;
    }
  }, [connection]);

  const disconnect = useCallback(() => {
    connection.close();
    setStatus('closed');
  }, [connection]);

  const openTunnel = useCallback(
    async (request: Omit<OpenTunnel, 'token'>) => {
      const opened = await connection.openTunnel(request);
      setTunnel(opened);
      return opened;
    },
    [connection],
  );

  const closeTunnel = useCallback(
    async (tunnelId = tunnel?.tunnel_id, reason?: string) => {
      if (!tunnelId) {
        return;
      }
      await connection.closeTunnel(tunnelId, reason);
      setTunnel((current) =>
        current?.tunnel_id === tunnelId ? undefined : current,
      );
    },
    [connection, tunnel?.tunnel_id],
  );

  const value = useMemo<TunnelContextValue>(
    () => ({
      connection,
      status,
      tunnel,
      error,
      connect,
      disconnect,
      openTunnel,
      closeTunnel,
    }),
    [
      connection,
      status,
      tunnel,
      error,
      connect,
      disconnect,
      openTunnel,
      closeTunnel,
    ],
  );

  return (
    <TunnelContext.Provider value={value}>{children}</TunnelContext.Provider>
  );
}

function normalizeError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value));
}
