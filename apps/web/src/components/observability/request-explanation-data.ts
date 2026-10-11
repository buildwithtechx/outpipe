import type { RequestCapture } from '#/interfaces';

export interface HttpRequestSummary {
  id: string;
  requestId?: string;
  timestamp: string;
  method: string;
  path: string;
  statusCode: number;
  duration: number;
  traceId?: string;
  service?: string;
  requestSize?: number;
  responseSize?: number;
  captureState?: 'metadata' | 'redacted' | 'full' | 'unknown';
}

export interface RequestExplanationPreview {
  category: 'server-error' | 'client-error' | 'slow' | 'healthy' | 'unknown';
  title: string;
  summary: string;
  evidence: {
    id: string;
    label: string;
    value: string;
    detail: string;
    target: 'response' | 'context';
  }[];
  hypothesis: { title: string; detail: string } | null;
  nextChecks: string[];
  limitations: string[];
  followUps: { id: string; label: string; answer: string }[];
}

const PREVIEW_SLOW_THRESHOLD_MS = 1_000;

function validStatus(value: unknown): value is number {
  return (
    typeof value === 'number' &&
    Number.isInteger(value) &&
    value >= 100 &&
    value <= 599
  );
}

function validDuration(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0;
}

function classify(
  request: HttpRequestSummary,
): RequestExplanationPreview['category'] {
  if (!validStatus(request.statusCode) || request.statusCode < 200)
    return 'unknown';
  if (request.statusCode >= 500) return 'server-error';
  if (request.statusCode >= 400) return 'client-error';
  if (!validDuration(request.duration)) return 'unknown';
  return request.duration >= PREVIEW_SLOW_THRESHOLD_MS ? 'slow' : 'healthy';
}

const categoryCopy: Record<
  RequestExplanationPreview['category'],
  Pick<RequestExplanationPreview, 'title' | 'hypothesis' | 'nextChecks'>
> = {
  'server-error': {
    title: 'A server-error response was recorded',
    hypothesis: {
      title: 'Possible upstream or server-side failure',
      detail:
        'Unconfirmed hypothesis: an upstream tunnel target, relay worker, or application server failed or timed out. Status and duration do not isolate code root causes.',
    },
    nextChecks: [
      'Inspect available response headers and payload body for diagnostic error descriptions.',
      'Check tunnel agent status, local listener health, and connection latency.',
      'Compare nearby requests to determine if the failure is isolated or systemic.',
    ],
  },
  'client-error': {
    title: 'A client-error response was recorded',
    hypothesis: {
      title: 'Possible request rejection',
      detail:
        'Unconfirmed hypothesis: authentication, route validation, query parameters, or permissions caused the 4xx rejection. Payload contents were not executed.',
    },
    nextChecks: [
      'Inspect response status and message to determine why the caller was rejected.',
      'Verify target path, required authorization headers, and HTTP method matching.',
      'Check client payload syntax against target API schemas.',
    ],
  },
  slow: {
    title: 'This request exceeded the latency threshold',
    hypothesis: {
      title: 'Possible latency bottleneck or connection wait',
      detail:
        'Unconfirmed hypothesis: processing spent significant time in network transit, upstream application logic, or database waits.',
    },
    nextChecks: [
      'Inspect trace spans to determine where latency concentrated.',
      'Compare similar requests to identify whether latency is an outlier or baseline.',
      'Review tunnel relay metrics for connection congestion or backlog.',
    ],
  },
  healthy: {
    title: 'This request completed within expected parameters',
    hypothesis: null,
    nextChecks: [
      'Inspect response headers and body if the client still reported unexpected behavior.',
      'Compare response sizes and status codes across subsequent calls.',
    ],
  },
  unknown: {
    title: 'Insufficient metadata to classify request',
    hypothesis: null,
    nextChecks: [
      'Verify that a completed HTTP status and duration were recorded.',
      'Ensure the tunnel agent is actively capturing stream telemetry.',
    ],
  },
};

export function buildRequestExplanationPreview(
  request: HttpRequestSummary,
): RequestExplanationPreview {
  const category = classify(request);
  const copy = categoryCopy[category];
  const status = validStatus(request.statusCode) ? request.statusCode : null;
  const duration = validDuration(request.duration) ? request.duration : null;
  const tracePresent =
    typeof request.traceId === 'string' && request.traceId.trim().length > 0;

  const limitations = [
    'This is a deterministic frontend preview, not an external AI diagnosis.',
    'No downstream databases, internal application logs, or external dependencies were inspected.',
    `The ${PREVIEW_SLOW_THRESHOLD_MS}ms latency threshold is a heuristic cutoff, not an enforced SLA.`,
  ];

  if (status === null) {
    limitations.push('A valid HTTP status was not recorded.');
  } else if (status < 200) {
    limitations.push('Interim status recorded; final outcome is unknown.');
  }

  if (duration === null) {
    limitations.push('A valid request duration was not recorded.');
  }

  if (!tracePresent) {
    limitations.push(
      'No trace identifier is attached; distributed waterfall is unavailable.',
    );
  }

  const evidence: RequestExplanationPreview['evidence'] = [
    {
      id: 'http-status',
      label: 'Response Status',
      value: status === null ? 'Not recorded' : `HTTP ${status}`,
      detail:
        status === null
          ? 'No status available.'
          : status >= 500
            ? 'A 5xx server-error was returned by the target or relay.'
            : status >= 400
              ? 'A 4xx client-error was returned to the caller.'
              : 'A successful 2xx/3xx response was completed.',
      target: 'response',
    },
    {
      id: 'duration',
      label: 'Latency',
      value: duration === null ? 'Not recorded' : `${duration} ms`,
      detail:
        duration === null
          ? 'Duration not recorded.'
          : duration >= PREVIEW_SLOW_THRESHOLD_MS
            ? `Crossed the ${PREVIEW_SLOW_THRESHOLD_MS}ms preview threshold.`
            : `Within expected sub-second latency window (${duration}ms).`,
      target: 'context',
    },
    {
      id: 'trace',
      label: 'Trace Context',
      value: tracePresent ? 'Trace attached' : 'Not attached',
      detail: tracePresent
        ? 'Distributed trace ID correlates downstream spans.'
        : 'Request has no associated trace header.',
      target: 'context',
    },
    {
      id: 'payload-size',
      label: 'Payload Exchange',
      value: `${request.requestSize ?? 0} B in / ${request.responseSize ?? 0} B out`,
      detail: `Request body size: ${request.requestSize ?? 0} bytes. Response body size: ${request.responseSize ?? 0} bytes.`,
      target: 'response',
    },
  ];

  const summary = `${status === null ? 'HTTP status is unavailable' : `HTTP ${status} recorded`} with a duration of ${duration ?? 0}ms. ${
    category === 'server-error' || category === 'client-error'
      ? 'This identifies the response category, not root cause.'
      : category === 'slow'
        ? 'Exceeds standard 1,000ms latency threshold.'
        : 'Completed successfully.'
  }`;

  const nextChecks = [...copy.nextChecks];

  return {
    category,
    title: copy.title,
    summary,
    evidence,
    hypothesis: copy.hypothesis ? { ...copy.hypothesis } : null,
    nextChecks,
    limitations,
    followUps: [
      {
        id: 'root-cause',
        label: 'What caused this?',
        answer:
          copy.hypothesis?.detail ??
          'Available telemetry indicates the outcome but does not establish internal application cause without inspecting server logs.',
      },
      {
        id: 'confidence',
        label: 'What is confirmed?',
        answer: `Confirmed facts: HTTP ${status ?? 'unknown'}, duration ${duration ?? 0}ms, payload size ${request.requestSize ?? 0}/${request.responseSize ?? 0}B. Hypotheses remain unconfirmed until verified.`,
      },
      {
        id: 'next-checks',
        label: 'What should I check next?',
        answer: nextChecks.join(' '),
      },
    ],
  };
}

export function captureToSummary(capture: RequestCapture): HttpRequestSummary {
  return {
    id: capture.id,
    requestId: capture.id,
    timestamp: capture.timestamp,
    method: capture.method,
    path: capture.path,
    statusCode: capture.statusCode,
    duration: capture.durationMs,
    requestSize: capture.requestBodySize,
    responseSize: capture.responseBodySize,
    captureState:
      capture.requestBody || capture.responseBody ? 'full' : 'metadata',
  };
}

export const requestExplanationDemoScenarios: {
  value: 'server-error' | 'slow' | 'healthy';
  label: string;
  request: HttpRequestSummary;
}[] = [
  {
    value: 'server-error',
    label: 'Failed request (503)',
    request: {
      id: 'demo-failed',
      requestId: 'demo-req-503',
      timestamp: new Date().toISOString(),
      method: 'POST',
      path: '/api/v1/checkout',
      statusCode: 503,
      duration: 2430,
      traceId: 'tr_503_abc123',
      service: 'payment-svc',
      requestSize: 240,
      responseSize: 64,
      captureState: 'full',
    },
  },
  {
    value: 'slow',
    label: 'Slow request (1840ms)',
    request: {
      id: 'demo-slow',
      requestId: 'demo-req-slow',
      timestamp: new Date().toISOString(),
      method: 'GET',
      path: '/api/v1/orders',
      statusCode: 200,
      duration: 1840,
      traceId: 'tr_slow_xyz789',
      service: 'orders-svc',
      requestSize: 0,
      responseSize: 1024,
      captureState: 'full',
    },
  },
  {
    value: 'healthy',
    label: 'Healthy request (200)',
    request: {
      id: 'demo-healthy',
      requestId: 'demo-req-ok',
      timestamp: new Date().toISOString(),
      method: 'GET',
      path: '/healthz',
      statusCode: 200,
      duration: 24,
      service: 'gateway',
      requestSize: 0,
      responseSize: 16,
      captureState: 'metadata',
    },
  },
];
