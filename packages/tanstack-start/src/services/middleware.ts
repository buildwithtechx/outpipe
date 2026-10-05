import type { TanStackStartTunnel } from '../interfaces/options';

const UUID_SEGMENT =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const INTEGER_SEGMENT = /^\d+$/;
const LONG_HEX_SEGMENT = /^[0-9a-f]{16,}$/i;
const ULID_SEGMENT = /^[0-9A-HJKMNP-TV-Z]{26}$/i;

export function normalizeTanStackRoute(pathname: string): string {
  const normalized = pathname
    .split('/')
    .map((segment) =>
      UUID_SEGMENT.test(segment) ||
      INTEGER_SEGMENT.test(segment) ||
      LONG_HEX_SEGMENT.test(segment) ||
      ULID_SEGMENT.test(segment)
        ? ':id'
        : segment,
    )
    .join('/');

  const route = normalized.startsWith('/') ? normalized : `/${normalized}`;
  return (route || '/').slice(0, 256);
}

export function tunnelStatus(tunnel: TanStackStartTunnel): Response {
  const state = tunnel.state();
  return Response.json(
    { ...state, error: state.error?.message },
    { headers: { 'content-type': 'application/json' } },
  );
}

export async function tunnelLifecycle(
  tunnel: TanStackStartTunnel,
  next: () => Promise<Response> | Response,
): Promise<Response> {
  if (tunnel.autoStart !== false) await tunnel.start();
  return next();
}
