import { Link } from '@tanstack/react-router';
import { ArrowRight, Code2 } from 'lucide-react';
import { MarketingContainer } from '#/components/layout';
import { sdkDefinitions } from './sdk-data';

export function SdkPage() {
  return (
    <main className="pb-24 pt-32 sm:pt-40">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <div className="mx-auto inline-flex items-center gap-2 rounded-full border border-indigo-300/20 bg-indigo-300/8 px-3 py-1.5 text-sm text-indigo-200">
            <Code2 className="size-4" />
            One protocol, your language
          </div>
          <h1 className="mt-7 text-4xl font-semibold tracking-[-0.06em] text-white sm:text-6xl">
            Build around the runtime
            <span className="block text-white/45">you already trust.</span>
          </h1>
          <p className="mx-auto mt-6 max-w-2xl text-lg leading-8 text-white/50">
            Outpipe keeps authentication, tunnel state, and relay behavior
            consistent across every official SDK. Pick a package, keep your
            stack, and connect local services without rewriting your workflow.
          </p>
        </div>

        <div className="mt-16 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {sdkDefinitions.map((sdk) => {
            const Icon = sdk.icon;
            return (
              <Link
                key={sdk.name}
                to="/docs/$"
                params={{ _splat: `integrations/${sdk.docsSlug}` }}
                className="group relative overflow-hidden rounded-3xl border border-white/10 bg-white/[0.025] p-6 transition-all hover:-translate-y-1 hover:border-indigo-300/35 hover:bg-indigo-300/[0.04]"
              >
                <div className="flex items-start justify-between">
                  <span
                    className={`flex size-12 items-center justify-center rounded-2xl bg-white/[0.06] text-2xl ${sdk.color}`}
                  >
                    <Icon />
                  </span>
                  <ArrowRight className="size-4 text-white/25 transition-transform group-hover:translate-x-1 group-hover:text-white" />
                </div>
                <h2 className="mt-8 text-xl font-semibold text-white">
                  {sdk.name}
                </h2>
                <p className="mt-1 font-mono text-xs text-white/35">
                  {sdk.packageName}
                </p>
                <p className="mt-5 min-h-12 text-sm leading-6 text-white/50">
                  {sdk.description}
                </p>
                <div className="mt-7 rounded-xl border border-white/10 bg-black/35 px-3 py-2.5 font-mono text-xs text-white/55">
                  <span className="mr-2 text-indigo-300">$</span>
                  {sdk.install}
                </div>
              </Link>
            );
          })}
        </div>

        <div className="mx-auto mt-16 max-w-3xl rounded-3xl border border-white/10 bg-[#0b0d14] p-8 text-center sm:p-10">
          <p className="font-mono text-sm text-cyan-300">
            same contract underneath
          </p>
          <h2 className="mt-4 text-2xl font-semibold text-white">
            Documentation that stays in sync.
          </h2>
          <p className="mt-3 leading-7 text-white/50">
            Each guide follows the versioned API, authentication model, and
            relay lifecycle used by the Outpipe CLI.
          </p>
          <Link
            to="/docs/$"
            params={{ _splat: 'integrations/overview' }}
            className="mt-7 inline-flex items-center gap-2 text-sm font-medium text-indigo-300 transition-colors hover:text-indigo-200"
          >
            Read the integration overview <ArrowRight className="size-4" />
          </Link>
        </div>
      </MarketingContainer>
    </main>
  );
}
