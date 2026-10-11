import { readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const project = resolve(__dirname, '../../../tinybird');

const organizationEndpoints = [
  'tunnel_http_stats',
  'tunnel_http_chart',
  'tunnel_requests',
  'tunnel_protocol_stats',
  'tunnel_protocol_chart',
  'tunnel_protocol_recent',
  'tunnel_capture',
  'tunnel_overview_stats',
  'tunnel_overview_chart',
];

const adminEndpoints = [
  'tunnel_admin_active_series',
  'tunnel_admin_http_chart',
  'tunnel_admin_active_chart',
  'tunnel_admin_usage',
];

const dataSources = [
  'tunnel_events',
  'tunnel_protocol_events',
  'tunnel_request_captures',
  'tunnel_active_snapshots',
];

const dateTimeParameters: Record<string, string[]> = Object.fromEntries(
  [...organizationEndpoints, ...adminEndpoints].map((name) => [
    name,
    [
      'start',
      'end',
      ...(name === 'tunnel_overview_stats' ? ['previous_start'] : []),
      ...(name === 'tunnel_capture' ? ['timestamp'] : []),
    ],
  ]),
);

async function endpoint(name: string): Promise<string> {
  return readFile(resolve(project, `endpoints/${name}.pipe`), 'utf8');
}

describe('Tunnel Tinybird & ClickHouse Analytics Contract', () => {
  for (const [name, parameters] of Object.entries(dateTimeParameters)) {
    it(`${name} declares required millisecond UTC time parameters for deployment validation`, async () => {
      const source = await endpoint(name);
      for (const parameter of parameters) {
        const interpolations = [...source.matchAll(/\{\{\s*([^{}]+?)\s*\}\}/g)]
          .map((m) => m[1].trim())
          .filter((exp) =>
            new RegExp(`\\(\\s*${parameter}\\s*(?:,|\\))`).test(exp),
          );
        expect(interpolations.length).toBeGreaterThan(0);
        for (const exp of interpolations) {
          expect(exp).toMatch(
            new RegExp(`^DateTime64\\(\\s*${parameter}\\s*\\)$`),
          );
        }
        const utcConversions = [
          ...source.matchAll(
            new RegExp(
              `toDateTime64\\(\\s*\\{\\{\\s*DateTime64\\(\\s*${parameter}\\s*\\)\\s*\\}\\}\\s*,\\s*3\\s*,\\s*'UTC'\\s*\\)`,
              'g',
            ),
          ),
        ];
        expect(utcConversions.length).toBe(interpolations.length);
      }
      expect(source).not.toMatch(/parseDateTime\w*\(\s*\{\{\s*String\(/);
    });
  }

  for (const name of organizationEndpoints) {
    it(`${name} keeps tenant and retention checks on every source read`, async () => {
      const source = await endpoint(name);
      expect(source).toMatch(/TOKEN OUTPIPE_QUERY_TOKEN READ/);
      expect(source).toMatch(/TYPE endpoint/);
      const sourceNodes = source
        .split(/\nNODE /)
        .filter((node) => /FROM tunnel_/.test(node));
      expect(sourceNodes.length).toBeGreaterThan(0);
      for (const node of sourceNodes) {
        expect(node).toMatch(/FROM tunnel_\w+ FINAL/);
        expect(node).toMatch(
          /organization_id = \{\{String\(organization_id\)\}\}/,
        );
        expect(node).toMatch(
          /timestamp >= now64\(3, 'UTC'\) - toIntervalDay\(least\(greatest\(retention_days, 1\), 90\)\)/,
        );
      }
    });
  }

  for (const name of adminEndpoints) {
    it(`${name} deduplicates retry batches and rejects logically expired evidence`, async () => {
      const source = await endpoint(name);
      const sourceNodes = source
        .split(/\nNODE /)
        .filter((node) => /FROM tunnel_/.test(node));
      for (const node of sourceNodes) {
        expect(node).toMatch(/FROM tunnel_\w+ FINAL/);
        expect(node).toMatch(
          /(?:timestamp|ts) >= now64\(3, 'UTC'\) - (?:toIntervalDay\(least\(greatest\(retention_days, 1\), 90\)\)|INTERVAL 90 DAY)/,
        );
      }
    });
  }

  for (const name of dataSources) {
    it(`${name} supports millisecond timestamps, retry deduplication and bounded TTL`, async () => {
      const source = await readFile(
        resolve(project, `datasources/${name}.datasource`),
        'utf8',
      );
      expect(source).toMatch(/TOKEN OUTPIPE_INGEST_TOKEN APPEND/);
      expect(source).toMatch(/DateTime64\(3, 'UTC'\)/);
      expect(source).toMatch(/ENGINE "ReplacingMergeTree"/);
      expect(source).toMatch(/ENGINE_VER ingested_at/);
      expect(source).toMatch(/ENGINE_TTL .*90/);
      if (name === 'tunnel_events' || name === 'tunnel_protocol_events') {
        expect(source).toMatch(
          /ENGINE_SORTING_KEY "organization_id, tunnel_id, timestamp, event_id"/,
        );
      }
    });
  }

  it('capture bodies preserve null and headers are serialized JSON', async () => {
    const source = await readFile(
      resolve(project, 'datasources/tunnel_request_captures.datasource'),
      'utf8',
    );
    expect(source).toMatch(/`request_headers` String .*DEFAULT '\{\}'/);
    expect(source).toMatch(/`response_headers` String .*DEFAULT '\{\}'/);
    expect(source).toMatch(/`request_body` Nullable\(String\)/);
    expect(source).toMatch(/`response_body` Nullable\(String\)/);
  });

  it('capture exact IDs cannot bypass organization, tunnel or retention checks', async () => {
    const source = await endpoint('tunnel_capture');
    const globalScope = source.slice(
      0,
      source.indexOf('{% if defined(request_id) %}'),
    );
    expect(globalScope).toMatch(
      /organization_id = \{\{String\(organization_id\)\}\}/,
    );
    expect(globalScope).toMatch(
      /has\(JSONExtract\(\{\{String\(tunnel_ids\)\}\}, 'Array\(String\)'\), tunnel_id\)/,
    );
    expect(globalScope).toMatch(
      /toIntervalDay\(least\(greatest\(retention_days, 1\), 90\)\)/,
    );
    expect(source).toMatch(/ORDER BY abs\(dateDiff\('millisecond'/);
    expect(source).toMatch(/LIMIT 1/);
  });

  it('request search is literal substring matching with stable time/ID ordering', async () => {
    const source = await endpoint('tunnel_requests');
    expect(source).toMatch(
      /positionCaseInsensitive\(path, \{\{String\(search\)\}\}\)/,
    );
    expect(source).not.toMatch(/\bI?LIKE\b/i);
    expect(source).toMatch(/ORDER BY timestamp DESC, event_id DESC/);
    expect(source).toMatch(
      /LIMIT least\(greatest\(\{\{UInt32\(limit, 50\)\}\}, 1\), 501\)/,
    );
  });

  it('comparison windows share one captured start/end boundary', async () => {
    const source = await endpoint('tunnel_overview_stats');
    expect(
      (
        source.match(
          /timestamp >= toDateTime64\(\{\{DateTime64\(previous_start\)\}\}/g,
        ) || []
      ).length,
    ).toBe(2);
    expect(
      (
        source.match(/timestamp < toDateTime64\(\{\{DateTime64\(end\)\}\}/g) ||
        []
      ).length,
    ).toBe(2);
    expect(source).toMatch(/http_errors/);
    expect(source).toMatch(/previous_http_errors/);
    expect(source).not.toMatch(/event_type .*http_errors/);
  });

  it('administrator request chart counts HTTP requests rather than active tunnels', async () => {
    const source = await endpoint('tunnel_admin_http_chart');
    expect(source).toMatch(/count\(\) AS requests/);
    expect(source).toMatch(/FROM tunnel_events FINAL/);
    expect(source).not.toMatch(/active_tunnels/);
  });
});
