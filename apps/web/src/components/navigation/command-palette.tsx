import { useNavigate } from '@tanstack/react-router';
import { ArrowRight, Search, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import {
  ADMIN_NAV_ITEMS,
  EXTERNAL_NAV_ITEMS,
  getWorkspaceNavItems,
} from './constants';

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgSlug: string;
  isPlatformAdmin?: boolean;
}

export function CommandPalette({
  open,
  onOpenChange,
  orgSlug,
  isPlatformAdmin = false,
}: CommandPaletteProps) {
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        onOpenChange(!open);
      }
      if (e.key === 'Escape' && open) {
        onOpenChange(false);
      }
    }
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [open, onOpenChange]);

  if (!open) return null;

  const workspaceNav = getWorkspaceNavItems(orgSlug);
  const adminNav = isPlatformAdmin ? ADMIN_NAV_ITEMS : [];
  const allItems = [...workspaceNav, ...adminNav, ...EXTERNAL_NAV_ITEMS];

  const filtered = search.trim()
    ? allItems.filter(
        (item) =>
          item.label.toLowerCase().includes(search.toLowerCase()) ||
          item.desc?.toLowerCase().includes(search.toLowerCase()),
      )
    : allItems;

  const handleSelect = (to: string) => {
    onOpenChange(false);
    setSearch('');
    if (to.startsWith('/docs')) {
      navigate({ to: '/docs/$', params: { _splat: '' } });
    } else {
      navigate({ to });
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center p-4 pt-[12vh] animate-in fade-in duration-150"
      role="dialog"
      aria-modal="true"
      aria-label="Command palette navigation"
    >
      <button
        type="button"
        aria-label="Close dialog backdrop"
        onClick={() => onOpenChange(false)}
        className="fixed inset-0 bg-black/70 backdrop-blur-md cursor-default"
      />
      <div className="relative z-10 w-full max-w-xl rounded-2xl border border-white/15 bg-neutral-900/95 p-3 shadow-2xl backdrop-blur-2xl animate-in zoom-in-95 duration-150 text-white">
        <div className="flex items-center gap-3 border-b border-white/10 px-3 pb-3">
          <Search className="size-4 shrink-0 text-white/40" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Type a command or search sections..."
            className="w-full bg-transparent text-sm text-white placeholder-white/35 focus:outline-none"
          />
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className="rounded-lg p-1 text-white/40 hover:bg-white/10 hover:text-white transition"
          >
            <X className="size-4" />
          </button>
        </div>

        <div className="max-h-80 overflow-y-auto p-1 space-y-1">
          {filtered.length === 0 ? (
            <div className="p-8 text-center text-xs text-white/40">
              No matching sections found for "{search}"
            </div>
          ) : (
            filtered.map((item) => {
              const Icon = item.icon;
              return (
                <button
                  key={item.to + item.label}
                  type="button"
                  onClick={() => handleSelect(item.to)}
                  className="flex w-full items-center justify-between gap-3 rounded-xl px-3 py-2.5 text-left text-xs transition hover:bg-white/10 focus:bg-white/10 focus:outline-none group"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-white/5 border border-white/10 text-white/70 group-hover:text-indigo-300 group-hover:border-indigo-400/30 transition">
                      <Icon className="size-3.5" />
                    </div>
                    <div className="min-w-0">
                      <p className="font-semibold text-white/90 group-hover:text-white">
                        {item.label}
                      </p>
                      {item.desc && (
                        <p className="text-[11px] text-white/45 truncate">
                          {item.desc}
                        </p>
                      )}
                    </div>
                  </div>
                  <ArrowRight className="size-3 text-white/20 group-hover:text-indigo-300 group-hover:translate-x-0.5 transition" />
                </button>
              );
            })
          )}
        </div>

        <div className="flex items-center justify-between border-t border-white/10 px-3 pt-2 text-[10px] text-white/40">
          <span>Navigate with ⌘K</span>
          <span>Esc to close</span>
        </div>
      </div>
    </div>
  );
}
