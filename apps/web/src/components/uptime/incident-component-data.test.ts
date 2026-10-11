import { describe, expect, it } from 'vitest';
import {
  incidentComponentRows,
  incidentComponentWindow,
  type UptimeComponent,
  type UptimeComponentGroup,
} from './incident-component-data';

describe('incident-component-data', () => {
  const components: UptimeComponent[] = [
    { id: 'c1', name: 'HTTP Gateway' },
    { id: 'c2', name: 'Relay US-East' },
    { id: 'c3', name: 'Web Dashboard' },
  ];

  const groups: UptimeComponentGroup[] = [
    {
      id: 'core',
      name: 'Core Services',
      components: [
        { id: 'g1', name: 'Relay Auth' },
        { id: 'g2', name: 'Control Plane' },
      ],
    },
  ];

  it('flattens components and groups into rows', () => {
    const rows = incidentComponentRows(components, groups);
    expect(rows).toHaveLength(5);
    expect(rows[0]?.groupId).toBe('standalone');
    expect(rows[0]?.searchName).toBe('http gateway');
    expect(rows[3]?.groupId).toBe('core');
  });

  it('filters rows by search term case-insensitively', () => {
    const rows = incidentComponentRows(components, groups);
    const windowed = incidentComponentWindow(rows, 'relay', 0);
    expect(windowed.total).toBe(2); // 'Relay US-East' and 'Relay Auth'
    expect(windowed.groups.flatMap((g) => g.components)).toHaveLength(2);
  });

  it('paginates correctly when exceeding page size', () => {
    const manyComponents: UptimeComponent[] = Array.from(
      { length: 65 },
      (_, i) => ({
        id: `comp-${i}`,
        name: `Component ${i}`,
      }),
    );
    const rows = incidentComponentRows(manyComponents, []);
    const page0 = incidentComponentWindow(rows, '', 0);
    expect(page0.total).toBe(65);
    expect(page0.totalPages).toBe(2);
    expect(page0.start).toBe(0);
    expect(page0.end).toBe(50);

    const page1 = incidentComponentWindow(rows, '', 1);
    expect(page1.start).toBe(50);
    expect(page1.end).toBe(65);
  });
});
