import { Link, useNavigate } from "@tanstack/react-router";
import { Check, ChevronsUpDown, Plus } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { getOrganizations } from "#/features/organizations/services/organization-service";
import type { Organization } from "#/interfaces/organization";

interface WorkspaceSwitcherProps {
  currentOrgSlug: string;
}

export function WorkspaceSwitcher({ currentOrgSlug }: WorkspaceSwitcherProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const navigate = useNavigate();

  const { data: organizations = [], isLoading } = useQuery({
    queryKey: ["organizations"],
    queryFn: getOrganizations,
  });

  const currentOrg = organizations.find((org) => org.slug === currentOrgSlug) ||
    organizations[0] || { name: "Workspace", slug: currentOrgSlug };

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSelect = (org: Organization) => {
    setOpen(false);
    navigate({
      to: "/$orgSlug",
      params: { orgSlug: org.slug },
    });
  };

  return (
    <div className="relative" ref={containerRef}>
      <button
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        className="flex items-center gap-2 rounded-xl border border-white/10 bg-white/3 px-2.5 py-1.5 text-left text-xs font-medium text-white transition hover:border-white/20 hover:bg-white/6 focus:outline-none focus:ring-1 focus:ring-indigo-400"
        aria-expanded={open}
        aria-haspopup="listbox"
      >
        <div className="flex size-6 shrink-0 items-center justify-center rounded-lg bg-indigo-600/30 border border-indigo-400/40 text-[10px] font-bold text-indigo-300">
          {currentOrg.name.slice(0, 2).toUpperCase()}
        </div>
        <div className="hidden sm:flex flex-col text-left min-w-0 max-w-32.5">
          <span className="truncate font-semibold text-white/90 text-xs">
            {currentOrg.name}
          </span>
          <span className="truncate text-[10px] text-white/40 font-mono">
            {currentOrg.slug}
          </span>
        </div>
        <ChevronsUpDown className="size-3.5 shrink-0 text-white/40 ml-0.5" />
      </button>

      {open && (
        <div className="absolute left-0 top-full z-50 mt-2 w-64 rounded-2xl border border-white/15 bg-neutral-900/95 p-1.5 shadow-2xl backdrop-blur-xl animate-in fade-in zoom-in-95 duration-150">
          <div className="px-2.5 py-1.5 text-[10px] font-semibold uppercase tracking-wider text-white/40">
            Switch Workspace
          </div>
          <div className="max-h-56 overflow-y-auto space-y-0.5">
            {isLoading ? (
              <div className="px-3 py-2 text-xs text-white/40">
                Loading workspaces...
              </div>
            ) : (
              organizations.map((org) => {
                const isSelected = org.slug === currentOrgSlug;
                return (
                  <button
                    key={org.id}
                    type="button"
                    onClick={() => handleSelect(org)}
                    className={`flex w-full items-center justify-between gap-2 rounded-xl px-2.5 py-2 text-left text-xs transition ${
                      isSelected
                        ? "bg-indigo-600 text-white font-medium"
                        : "text-white/80 hover:bg-white/10 hover:text-white"
                    }`}
                  >
                    <div className="flex items-center gap-2.5 min-w-0">
                      <div
                        className={`flex size-6 shrink-0 items-center justify-center rounded-lg text-[10px] font-bold ${
                          isSelected
                            ? "bg-white/20 text-white"
                            : "bg-white/5 border border-white/10 text-white/60"
                        }`}
                      >
                        {org.name.slice(0, 2).toUpperCase()}
                      </div>
                      <div className="min-w-0">
                        <p className="truncate font-medium">{org.name}</p>
                        <p
                          className={`truncate text-[10px] font-mono ${
                            isSelected ? "text-white/70" : "text-white/40"
                          }`}
                        >
                          {org.slug}
                        </p>
                      </div>
                    </div>
                    {isSelected && (
                      <Check className="size-3.5 shrink-0 text-white" />
                    )}
                  </button>
                );
              })
            )}
          </div>

          <div className="mt-1 border-t border-white/10 pt-1">
            <Link
              to="/select"
              onClick={() => setOpen(false)}
              className="flex w-full items-center gap-2 rounded-xl px-2.5 py-2 text-xs font-medium text-indigo-300 transition hover:bg-white/5 hover:text-indigo-200"
            >
              <Plus className="size-3.5" />
              <span>Create or join workspace</span>
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
