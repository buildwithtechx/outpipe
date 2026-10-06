import { useNavigate } from '@tanstack/react-router';
import { Search } from 'lucide-react';
import { type KeyboardEvent, useEffect, useRef, useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '#/components/ui/dialog';
import { Input } from '#/components/ui/input';
import {
  ADMIN_NAV_ITEMS,
  EXTERNAL_NAV_ITEMS,
  getWorkspaceNavItems,
} from './constants';

export function CommandPalette({
  open,
  onOpenChange,
  orgSlug,
  isPlatformAdmin = false,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgSlug: string;
  isPlatformAdmin?: boolean;
}) {
  const [search, setSearch] = useState('');
  const results = useRef<HTMLDivElement>(null);
  const input = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  useEffect(() => {
    function hotkey(event: globalThis.KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        onOpenChange(!open);
      }
    }
    window.addEventListener('keydown', hotkey);
    return () => window.removeEventListener('keydown', hotkey);
  }, [open, onOpenChange]);
  useEffect(() => {
    if (!open) setSearch('');
  }, [open]);
  const items = [
    ...getWorkspaceNavItems(orgSlug),
    ...(isPlatformAdmin ? ADMIN_NAV_ITEMS : []),
    ...EXTERNAL_NAV_ITEMS,
  ];
  const filtered = items.filter((item) =>
    `${item.label} ${item.desc ?? ''}`
      .toLowerCase()
      .includes(search.trim().toLowerCase()),
  );
  function move(event: KeyboardEvent) {
    if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return;
    if (event.target === input.current && ['Home', 'End'].includes(event.key))
      return;
    const buttons = Array.from(
      results.current?.querySelectorAll<HTMLButtonElement>('[data-command]') ??
        [],
    );
    if (!buttons.length) return;
    event.preventDefault();
    const index =
      document.activeElement instanceof HTMLButtonElement
        ? buttons.indexOf(document.activeElement)
        : -1;
    const next =
      event.key === 'Home'
        ? 0
        : event.key === 'End'
          ? buttons.length - 1
          : index < 0
            ? event.key === 'ArrowUp'
              ? buttons.length - 1
              : 0
            : (index + (event.key === 'ArrowUp' ? -1 : 1) + buttons.length) %
              buttons.length;
    buttons[next]?.focus();
  }
  function select(to: string) {
    onOpenChange(false);
    if (to.startsWith('/docs'))
      void navigate({ to: '/docs/$', params: { _splat: '' } });
    else void navigate({ to });
  }
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        onKeyDown={move}
        onOpenAutoFocus={(event) => {
          event.preventDefault();
          input.current?.focus();
        }}
        className="max-w-xl"
      >
        <DialogTitle className="sr-only">Jump to a page</DialogTitle>
        <DialogDescription className="sr-only">
          Search workspace pages. Use arrow keys to move through results and
          Enter to open a page.
        </DialogDescription>
        <div className="relative mr-6">
          <Search className="absolute left-3 top-3 size-4 text-muted-foreground" />
          <Input
            ref={input}
            aria-label="Search pages"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search pages…"
            className="pl-9"
            onKeyDown={(event) => {
              if (event.key === 'Enter' && filtered[0]) select(filtered[0].to);
            }}
          />
        </div>
        <div ref={results} className="max-h-80 space-y-1 overflow-y-auto">
          {filtered.map((item) => (
            <button
              key={item.to}
              data-command
              type="button"
              onClick={() => select(item.to)}
              className="flex min-h-11 w-full items-center gap-3 rounded-xl p-3 text-left text-sm hover:bg-accent focus:bg-accent"
            >
              <item.icon className="size-4 shrink-0 text-muted-foreground" />
              <span>
                <span className="block font-medium">{item.label}</span>
                <span className="text-xs text-muted-foreground">
                  {item.desc}
                </span>
              </span>
            </button>
          ))}
          {!filtered.length && (
            <p role="status" className="p-6 text-sm text-muted-foreground">
              No pages match “{search}”.
            </p>
          )}
        </div>
        <p className="text-xs text-muted-foreground">
          Ctrl/⌘ K to open · ↑ ↓ to navigate · Enter to select · Esc to close
        </p>
      </DialogContent>
    </Dialog>
  );
}
