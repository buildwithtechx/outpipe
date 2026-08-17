import { Link } from '@tanstack/react-router';
import { ArrowRight, PackageCheck } from 'lucide-react';
import { MarketingContainer } from '#/components/layout';
import { pluginDefinitions } from './plugin-data';

export function PluginsPage() {
  return (
    <section className="pb-16 pt-28 sm:pt-32">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <div className="mx-auto inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-3 py-1.5 text-sm text-white/55">
            <PackageCheck className="size-4 text-cyan-300" />
            Official Outpipe integrations
          </div>
          <h1 className="mt-7 text-4xl font-semibold tracking-[-0.06em] sm:text-5xl">
            One lifecycle for every stack.
          </h1>
          <p className="mt-6 text-lg leading-8 text-white/50">
            Install the package that matches your runtime. Each integration
            handles its own framework details and shares the same secure Outpipe
            protocol underneath.
          </p>
        </div>
        <div className="mt-12 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
          {pluginDefinitions.map((plugin) => (
            <Link
              key={plugin.id}
              to="/plugins/$pluginId"
              params={{ pluginId: plugin.id }}
              className="group relative overflow-hidden rounded-3xl border border-white/10 bg-white/[0.025] p-7 transition-all hover:-translate-y-1 hover:border-indigo-300/35"
            >
              <div className="absolute inset-0 bg-gradient-to-br from-indigo-300/[0.08] to-transparent opacity-0 transition-opacity group-hover:opacity-100" />
              <div className="relative">
                <div className="flex items-center justify-between">
                  <span
                    className={`flex size-12 items-center justify-center rounded-2xl bg-white/[0.06] text-2xl ${plugin.colorClass}`}
                  >
                    <plugin.icon />
                  </span>
                  <ArrowRight className="size-4 text-white/25 transition-all group-hover:translate-x-1 group-hover:text-white" />
                </div>
                <h2 className="mt-8 text-xl font-semibold">{plugin.name}</h2>
                <p className="mt-1 font-mono text-xs text-white/35">
                  {plugin.packageName}
                </p>
                <p className="mt-5 min-h-12 text-sm leading-6 text-white/50">
                  {plugin.description}
                </p>
                <div className="mt-7 flex flex-wrap gap-2">
                  {plugin.features.slice(0, 2).map((feature) => (
                    <span
                      key={feature}
                      className="rounded-full border border-white/10 px-2.5 py-1 text-[11px] text-white/45"
                    >
                      {feature}
                    </span>
                  ))}
                </div>
              </div>
            </Link>
          ))}
        </div>
        <div className="mx-auto mt-14 max-w-3xl rounded-2xl border border-white/10 bg-[#0d1220] p-7 text-center sm:p-10">
          <p className="font-mono text-sm text-cyan-300">
            one protocol, every surface
          </p>
          <h2 className="mt-4 text-2xl font-semibold">
            No duplicate SDK install
          </h2>
          <p className="mt-3 leading-7 text-white/50">
            Framework adapters include the core SDK as a dependency. Install
            only the integration your project needs.
          </p>
        </div>
      </MarketingContainer>
    </section>
  );
}
