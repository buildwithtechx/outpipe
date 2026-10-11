import type { AgentEvidenceReference, AgentStep } from '../protocol';
import type {
  AgentInvestigationContext,
  AgentInvestigationResult,
  AgentProvider,
} from './types';

export class MockAgentProvider implements AgentProvider {
  readonly name = 'mock';
  readonly model: string;

  constructor(model = 'outpipe-heuristic-engine-v1') {
    this.model = model;
  }

  async investigate(
    context: AgentInvestigationContext,
  ): Promise<AgentInvestigationResult> {
    const { prompt, sourceRequestId, signal, emit } = context;

    const steps: AgentStep[] = [
      {
        id: 'step-1',
        label: 'Inspect request telemetry and status',
        status: 'running',
      },
    ];
    await emit({ type: 'step', step: steps[0] });

    if (signal.aborted) throw new Error('Investigation cancelled');
    await new Promise((r) => setTimeout(r, 200));

    steps[0] = {
      id: 'step-1',
      label: 'Inspect request telemetry and status',
      status: 'complete',
      detail: sourceRequestId
        ? `Read metadata for request ${sourceRequestId}`
        : `Analyzed query: "${prompt.slice(0, 40)}"`,
    };
    await emit({ type: 'step', step: steps[0] });

    // Step 2: Traces & Spans
    const step2: AgentStep = {
      id: 'step-2',
      label: 'Correlate distributed spans & upstream latency',
      status: 'running',
    };
    steps.push(step2);
    await emit({ type: 'step', step: step2 });

    await new Promise((r) => setTimeout(r, 250));
    step2.status = 'complete';
    step2.detail = 'Correlated downstream spans against 24h baseline';
    await emit({ type: 'step', step: step2 });

    // Evidence
    const evidence: AgentEvidenceReference[] = [
      {
        id: 'ev-req',
        label: 'Inspected Request Outcome',
        href: `/observability?id=${sourceRequestId || 'current'}`,
        observedAt: new Date().toISOString(),
        presentation: {
          kind: 'request',
          method: 'POST',
          path: '/api/v1/checkout',
          service: 'payment-service',
          statusCode: 503,
          durationMs: 2430,
          timestamp: new Date().toISOString(),
          captureState: 'full',
          requestSizeBytes: 256,
          responseSizeBytes: 64,
        },
      },
      {
        id: 'ev-span',
        label: 'Upstream Provider Span',
        href: `/observability/traces/tr-checkout`,
        observedAt: new Date().toISOString(),
        presentation: {
          kind: 'trace',
          spanCount: 4,
          returnedSpanCount: 4,
          truncated: false,
          spans: [
            {
              spanId: 'sp-1',
              parentSpanId: null,
              operationName: 'HTTP POST /api/v1/checkout',
              service: 'gateway',
              startedAt: new Date().toISOString(),
              offsetMs: 0,
              durationMs: 2430,
              status: 'error',
            },
            {
              spanId: 'sp-2',
              parentSpanId: 'sp-1',
              operationName: 'PaymentGateway.Charge',
              service: 'payment-service',
              startedAt: new Date().toISOString(),
              offsetMs: 15,
              durationMs: 2400,
              status: 'error',
            },
          ],
        },
      },
    ];

    await emit({ type: 'evidence', evidence });

    // Step 3: Synthesis
    const step3: AgentStep = {
      id: 'step-3',
      label: 'Synthesize fault hypotheses and next checks',
      status: 'running',
    };
    steps.push(step3);
    await emit({ type: 'step', step: step3 });

    await new Promise((r) => setTimeout(r, 200));
    step3.status = 'complete';
    step3.detail = 'Hypothesis generated based on timing breakdown';
    await emit({ type: 'step', step: step3 });

    const responseMarkdown = `### Diagnostic Summary
The inspected request failed with **HTTP 503 Service Unavailable** after **2,430 ms**. The primary timing bottleneck is concentrated in the payment provider upstream span, which consumed **2,400 ms (~99%)** before timing out.

### Key Observations
- **Upstream Timeout Concentration**: The \`PaymentGateway.Charge\` span absorbed almost the entire request duration.
- **Circuit Breaker / Deadline Exhaustion**: The 2,400 ms duration aligns with typical 2.5-second client timeout thresholds.
- **Relay & Tunnel Health**: The Outpipe tunnel relay delivered the request to your local agent within 15 ms, ruling out tunnel relay latency as the cause.

### Unconfirmed Possibilities
1. **Third-Party API Outage or Throttling**: The external payment processor may be experiencing degraded performance or rejecting calls.
2. **Network Partition or Connection Pool Starvation**: Local worker connection pools may be exhausted waiting on upstream sockets.

### Recommended Next Checks
1. Check upstream payment provider dashboard or status page for active degraded performance incidents.
2. Review local timeout configurations and ensure aggressive cancellation is set to avoid thread starvation.
3. Check nearby error logs for socket reset or \`ETIMEDOUT\` errors within the same time window.`;

    // Stream text in chunks
    const chunks = responseMarkdown.split('\n\n');
    let fullText = '';
    for (let i = 0; i < chunks.length; i++) {
      if (signal.aborted) throw new Error('Investigation cancelled');
      const delta = (i === 0 ? '' : '\n\n') + chunks[i];
      fullText += delta;
      await emit({ type: 'text', delta });
      await new Promise((r) => setTimeout(r, 60));
    }

    return {
      text: fullText,
      steps,
      evidence,
      inputTokens: 120,
      outputTokens: 380,
    };
  }
}
