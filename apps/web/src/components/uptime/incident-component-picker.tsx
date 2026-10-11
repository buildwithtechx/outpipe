import { Check, ChevronLeft, ChevronRight, Search } from 'lucide-react';
import { memo, useMemo, useState } from 'react';
import { Button } from '#/components/ui/button';
import {
  INCIDENT_COMPONENT_PAGE_SIZE,
  incidentComponentRows,
  incidentComponentWindow,
  type UptimeComponent,
  type UptimeComponentGroup,
} from './incident-component-data';

export interface IncidentComponentPickerProps {
  components?: UptimeComponent[];
  groups?: UptimeComponentGroup[];
  selected: string[];
  onToggle: (id: string, checked: boolean) => void;
  error?: string;
}

export const IncidentComponentPicker = memo(function IncidentComponentPicker({
  components = [],
  groups = [],
  selected,
  onToggle,
  error,
}: IncidentComponentPickerProps) {
  const [search, setSearch] = useState('');
  const [pageIndex, setPageIndex] = useState(0);

  const rows = useMemo(
    () => incidentComponentRows(components, groups),
    [components, groups],
  );

  const windowData = useMemo(
    () => incidentComponentWindow(rows, search, pageIndex),
    [rows, search, pageIndex],
  );

  const selectedSet = useMemo(() => new Set(selected), [selected]);

  return (
    <fieldset className="space-y-2">
      <legend className="flex items-center justify-between w-full text-xs font-medium text-white/70">
        <span>Affected Components</span>
        {selected.length > 0 && (
          <span className="font-normal text-indigo-300">
            {selected.length} selected
          </span>
        )}
      </legend>

      <div className="overflow-hidden rounded-xl border border-white/10 bg-white/2.5">
        <div className="relative border-b border-white/10 p-2">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 size-3.5 text-white/40 pointer-events-none" />
          <input
            type="search"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPageIndex(0);
            }}
            placeholder="Find a component or service…"
            className="w-full rounded-md border border-white/5 bg-black/30 py-1.5 pl-8 pr-3 text-xs text-white placeholder-white/30 outline-none focus:border-indigo-500/50"
          />
        </div>

        <div className="max-h-48 overflow-y-auto p-1.5 divide-y divide-white/5">
          {windowData.groups.map((group) => (
            <div key={group.id} className="py-1 first:pt-0 last:pb-0">
              <p className="px-2 py-1 text-[10px] font-semibold text-white/40 uppercase tracking-wider">
                {group.name}
              </p>
              <div className="space-y-0.5">
                {group.components.map((component) => {
                  const isChecked = selectedSet.has(component.id);
                  return (
                    <label
                      key={component.id}
                      className="flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 text-xs text-white/80 hover:bg-white/5 transition-colors"
                    >
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={(e) =>
                          onToggle(component.id, e.target.checked)
                        }
                        className="peer sr-only"
                      />
                      <span
                        aria-hidden="true"
                        className={`flex size-4 shrink-0 items-center justify-center rounded border transition-colors ${
                          isChecked
                            ? 'border-indigo-500 bg-indigo-600 text-white'
                            : 'border-white/20 bg-white/5 text-transparent'
                        }`}
                      >
                        <Check className="size-3 stroke-[3]" />
                      </span>
                      <span className="flex-1 truncate">{component.name}</span>
                      {component.visible === false && (
                        <span className="text-[10px] text-white/40">
                          Private
                        </span>
                      )}
                    </label>
                  );
                })}
              </div>
            </div>
          ))}

          {windowData.groups.length === 0 && (
            <p className="py-4 text-center text-xs text-white/40">
              No components match &quot;{search}&quot;.
            </p>
          )}
        </div>

        {windowData.total > INCIDENT_COMPONENT_PAGE_SIZE && (
          <div className="flex items-center justify-between border-t border-white/10 px-3 py-1.5 text-[11px] text-white/50">
            <span>
              {windowData.start + 1}–{windowData.end} of {windowData.total}{' '}
              components
            </span>
            <div className="flex gap-1">
              <Button
                type="button"
                variant="ghost"
                size="xs"
                disabled={windowData.currentPage === 0}
                onClick={() => setPageIndex((p) => p - 1)}
                className="h-6 px-1.5 text-white/70"
              >
                <ChevronLeft className="size-3" />
                <span>Prev</span>
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="xs"
                disabled={windowData.end >= windowData.total}
                onClick={() => setPageIndex((p) => p + 1)}
                className="h-6 px-1.5 text-white/70"
              >
                <span>Next</span>
                <ChevronRight className="size-3" />
              </Button>
            </div>
          </div>
        )}
      </div>

      {error && (
        <p role="alert" className="text-xs text-rose-300">
          {error}
        </p>
      )}
    </fieldset>
  );
});
