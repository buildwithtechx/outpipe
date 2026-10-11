export interface UptimeComponent {
  id: string;
  name: string;
  groupId?: string;
  visible?: boolean;
}

export interface UptimeComponentGroup {
  id: string;
  name: string;
  components: UptimeComponent[];
}

export const INCIDENT_COMPONENT_PAGE_SIZE = 50;

export interface ComponentRow {
  component: UptimeComponent;
  groupId: string;
  groupName: string;
  searchName: string;
}

export function incidentComponentRows(
  components: UptimeComponent[] = [],
  groups: UptimeComponentGroup[] = [],
): ComponentRow[] {
  const groupedIds = new Set(
    groups.flatMap((g) => g.components.map((c) => c.id)),
  );
  const standalone = components.filter((c) => !groupedIds.has(c.id));

  const allGroups = [
    ...(standalone.length > 0
      ? [
          {
            id: 'standalone',
            name: 'Standalone components',
            components: standalone,
          },
        ]
      : []),
    ...groups,
  ];

  return allGroups.flatMap((group) =>
    group.components.map((component) => ({
      component,
      groupId: group.id,
      groupName: group.name,
      searchName: component.name.toLowerCase(),
    })),
  );
}

export function incidentComponentWindow(
  rows: ComponentRow[],
  search: string,
  pageIndex: number,
) {
  const query = search.trim().toLowerCase();
  const matches = query
    ? rows.filter((row) => row.searchName.includes(query))
    : rows;

  const totalPages = Math.max(
    1,
    Math.ceil(matches.length / INCIDENT_COMPONENT_PAGE_SIZE),
  );
  const currentPage = Math.max(0, Math.min(pageIndex, totalPages - 1));
  const start = currentPage * INCIDENT_COMPONENT_PAGE_SIZE;
  const visible = matches.slice(start, start + INCIDENT_COMPONENT_PAGE_SIZE);

  const groups: Array<{
    id: string;
    name: string;
    components: UptimeComponent[];
  }> = [];
  for (const row of visible) {
    const previous = groups[groups.length - 1];
    if (previous?.id === row.groupId) {
      previous.components.push(row.component);
    } else {
      groups.push({
        id: row.groupId,
        name: row.groupName,
        components: [row.component],
      });
    }
  }

  return {
    groups,
    total: matches.length,
    currentPage,
    totalPages,
    start,
    end: start + visible.length,
  };
}
