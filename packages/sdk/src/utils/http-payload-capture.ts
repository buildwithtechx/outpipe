import type {
  IncomingHttpHeaders,
  IncomingMessage,
  ServerResponse,
} from 'node:http';

const REDACTED = '[REDACTED]';
export const DEFAULT_MAX_BODY_BYTES = 16 * 1024;
export const HARD_MAX_BODY_BYTES = 64 * 1024;
export const DEFAULT_MAX_HEADER_BYTES = 8 * 1024;
export const HARD_MAX_HEADER_BYTES = 32 * 1024;

export const DEFAULT_SENSITIVE_HEADERS = [
  'authorization',
  'proxy-authorization',
  'cookie',
  'set-cookie',
  'api-key',
  'x-api-key',
  'auth-token',
  'x-auth-token',
  'access-token',
  'x-access-token',
  'refresh-token',
  'x-refresh-token',
  'client-secret',
  'x-client-secret',
  'private-key',
  'x-private-key',
  'session',
  'session-id',
  'x-csrf-token',
  'x-xsrf-token',
] as const;

export const DEFAULT_SENSITIVE_FIELDS = [
  'authorization',
  'cookie',
  'password',
  'passwd',
  'pwd',
  'secret',
  'token',
  'access-token',
  'refresh-token',
  'auth-token',
  'api-key',
  'private-key',
  'client-secret',
  'session',
  'session-id',
  'card-number',
  'cvv',
  'cvc',
  'ssn',
  'pin',
  'otp',
] as const;

export const OUTPIPE_HTTP_CAPTURE_ATTRIBUTES = {
  requestHeaders: 'outpipe.http.request.headers',
  requestBody: 'outpipe.http.request.body',
  requestBodyTruncated: 'outpipe.http.request.body.truncated',
  responseHeaders: 'outpipe.http.response.headers',
  responseBody: 'outpipe.http.response.body',
  responseBodyTruncated: 'outpipe.http.response.body.truncated',
} as const;

export interface HttpPayloadCaptureOptions {
  requestHeaders?: boolean;
  requestBody?: boolean;
  responseHeaders?: boolean;
  responseBody?: boolean;
  maxBodyBytes?: number;
  maxHeaderBytes?: number;
  redactedHeaders?: readonly string[];
  redactedFields?: readonly string[];
}

export type HttpPayloadCaptureSetting = boolean | HttpPayloadCaptureOptions;

export function redactHeaders(
  headers: IncomingHttpHeaders | Record<string, unknown>,
  additionalSensitive?: readonly string[],
): Record<string, string> {
  const sensitive = new Set<string>([
    ...DEFAULT_SENSITIVE_HEADERS,
    ...(additionalSensitive || []).map((h) => h.toLowerCase()),
  ]);

  const result: Record<string, string> = {};
  for (const [key, value] of Object.entries(headers)) {
    const lowerKey = key.toLowerCase();
    if (sensitive.has(lowerKey)) {
      result[key] = REDACTED;
    } else if (Array.isArray(value)) {
      result[key] = value.join(', ');
    } else if (typeof value === 'string') {
      result[key] = value;
    }
  }
  return result;
}

export function redactJsonValue(
  value: unknown,
  additionalSensitive?: readonly string[],
  depth = 0,
): unknown {
  if (depth > 8 || value === null || typeof value !== 'object') {
    return value;
  }

  const sensitive = new Set<string>([
    ...DEFAULT_SENSITIVE_FIELDS,
    ...(additionalSensitive || []).map((f) => f.toLowerCase()),
  ]);

  if (Array.isArray(value)) {
    return value.map((item) =>
      redactJsonValue(item, additionalSensitive, depth + 1),
    );
  }

  const record = value as Record<string, unknown>;
  const output: Record<string, unknown> = {};

  for (const [k, v] of Object.entries(record)) {
    if (sensitive.has(k.toLowerCase())) {
      output[k] = REDACTED;
    } else {
      output[k] = redactJsonValue(v, additionalSensitive, depth + 1);
    }
  }
  return output;
}

export interface CapturedPayloads {
  requestHeaders?: Record<string, string>;
  requestBody?: string;
  responseHeaders?: Record<string, string>;
  responseBody?: string;
}

export type NodeHttpPayloadCaptureMiddleware = (
  request: IncomingMessage & {
    body?: unknown;
    outpipePayloads?: CapturedPayloads;
  },
  response: ServerResponse & { outpipePayloads?: CapturedPayloads },
  next: (error?: unknown) => void,
) => void;

export function createNodeHttpPayloadCaptureMiddleware(
  setting: HttpPayloadCaptureSetting = true,
): NodeHttpPayloadCaptureMiddleware {
  if (setting === false) {
    return (_req, _res, next) => next();
  }

  const options: HttpPayloadCaptureOptions =
    typeof setting === 'object' ? setting : {};
  const maxBodyBytes = Math.min(
    Math.max(options.maxBodyBytes ?? DEFAULT_MAX_BODY_BYTES, 1024),
    HARD_MAX_BODY_BYTES,
  );

  return (req, res, next) => {
    const payloads: CapturedPayloads = {};
    req.outpipePayloads = payloads;
    res.outpipePayloads = payloads;

    if (options.requestHeaders !== false) {
      payloads.requestHeaders = redactHeaders(
        req.headers,
        options.redactedHeaders,
      );
    }

    if (options.requestBody !== false && req.body) {
      try {
        const redacted = redactJsonValue(req.body, options.redactedFields);
        const serialized = JSON.stringify(redacted);
        payloads.requestBody =
          serialized.length > maxBodyBytes
            ? `${serialized.slice(0, maxBodyBytes)}...[truncated]`
            : serialized;
      } catch {
        // Ignored
      }
    }

    if (options.responseBody !== false) {
      const originalWrite = res.write;
      const originalEnd = res.end;
      const chunks: Buffer[] = [];
      let totalLength = 0;

      res.write = (chunk: unknown, ...args: unknown[]) => {
        if (chunk && totalLength < maxBodyBytes) {
          const buf = Buffer.isBuffer(chunk)
            ? chunk
            : Buffer.from(String(chunk));
          chunks.push(buf);
          totalLength += buf.length;
        }
        return (originalWrite as (...a: unknown[]) => boolean).apply(res, [
          chunk,
          ...args,
        ]);
      };

      res.end = (chunk?: unknown, ...args: unknown[]) => {
        if (chunk && totalLength < maxBodyBytes) {
          const buf = Buffer.isBuffer(chunk)
            ? chunk
            : Buffer.from(String(chunk));
          chunks.push(buf);
          totalLength += buf.length;
        }

        if (chunks.length > 0) {
          try {
            const raw = Buffer.concat(chunks).toString('utf-8');
            try {
              const parsed = JSON.parse(raw);
              const redacted = redactJsonValue(parsed, options.redactedFields);
              payloads.responseBody = JSON.stringify(redacted);
            } catch {
              payloads.responseBody = raw.slice(0, maxBodyBytes);
            }
          } catch {
            // Ignored
          }
        }

        if (options.responseHeaders !== false) {
          payloads.responseHeaders = redactHeaders(
            res.getHeaders(),
            options.redactedHeaders,
          );
        }

        return (originalEnd as (...a: unknown[]) => ServerResponse).apply(res, [
          chunk,
          ...args,
        ]);
      };
    }

    next();
  };
}
