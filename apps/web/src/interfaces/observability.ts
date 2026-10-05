export interface TelemetrySpan {
  id: string;
  organizationId: string;
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  name: string;
  kind: string;
  startTime: string;
  endTime: string;
  durationMs: number;
  statusCode: string;
  statusMessage?: string;
  attributes?: string;
  events?: string;
  createdAt: string;
}

export interface TelemetryLog {
  id: string;
  organizationId: string;
  traceId?: string;
  spanId?: string;
  timestamp: string;
  severity: string;
  body: string;
  attributes?: string;
  createdAt: string;
}

export interface RequestCapture {
  id: string;
  organizationId: string;
  tunnelId: string;
  timestamp: string;
  method: string;
  path: string;
  statusCode: number;
  durationMs: number;
  requestHeaders?: string;
  requestBody?: string;
  requestBodySize: number;
  responseHeaders?: string;
  responseBody?: string;
  responseBodySize: number;
  createdAt: string;
}

export interface ObservabilityStats {
  totalRequests: number;
  requestsChange: number;
  totalBytes: number;
  dataTransferChange: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  chartData: Array<{
    time: string;
    requests: number;
    errors: number;
    bytes: number;
  }>;
}

export interface ReplayRequestInput {
  url: string;
  method: string;
  headers?: Record<string, string>;
  request_body?: string;
}

export interface ReplayResponseOutput {
  status_code: number;
  status_text: string;
  headers: Record<string, string>;
  body: string;
  duration_ms: number;
}
