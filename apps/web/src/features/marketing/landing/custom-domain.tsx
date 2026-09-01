import { ArrowRight, Check, Globe2, Link2, Route } from 'lucide-react';
import { MarketingContainer } from '#/components/layout';

const endpoints = [
  {
    label: 'Reserved endpoint',
    value: 'checkout.outpipe.app',
    detail: 'Keep a recognizable URL for repeatable previews.',
    color: 'text-indigo-300',
  },
  {
    label: 'Custom domain',
    value: 'api.yourcompany.com',
    detail: 'Route your own hostname through the same tunnel.',
    color: 'text-cyan-300',
  },
];

export function CustomDomainSection() {
  return (
    <section className="py-14 sm:py-16">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <p className="text-sm font-medium text-indigo-300">
            Endpoint identity
          </p>
          <h2 className="mt-4 text-4xl font-semibold tracking-[-0.05em] sm:text-5xl">
            A URL people can remember.
          </h2>
          <p className="mt-4 text-lg leading-8 text-white/50">
            Give a preview a stable name or bring the domain your team already
            knows. Both lead to the same connected service.
          </p>
        </div>
        <div className="relative mx-auto mt-8 max-w-5xl overflow-hidden rounded-3xl border border-white/10 bg-white/[0.025] p-4 sm:p-5">
          <div className="absolute inset-x-20 top-0 h-px bg-linear-to-r from-transparent via-indigo-300/50 to-transparent" />
          <div className="flex items-center justify-between border-b border-white/10 pb-4 font-mono text-xs text-white/35">
            <span>endpoint routing</span>
            <span className="flex items-center gap-2 text-emerald-300">
              <span className="size-2 rounded-full bg-emerald-300" /> connected
            </span>
          </div>
          <div className="relative mt-4 flex flex-col gap-3 sm:flex-row sm:items-stretch">
            <article className="flex-1 rounded-2xl bg-black/30 p-4 transition-colors hover:bg-white/[0.04]">
              <div className="flex items-center gap-2.5">
                <span
                  className={`flex size-9 items-center justify-center rounded-xl bg-white/[0.05] ${endpoints[0].color}`}
                >
                  <Route className="size-5" />
                </span>
                <span className="text-sm font-medium text-white/70">
                  {endpoints[0].label}
                </span>
              </div>
              <p className="mt-4 break-all font-mono text-sm text-white sm:text-base">
                https://{endpoints[0].value}
              </p>
              <p className="mt-2 text-sm leading-5 text-white/40">
                {endpoints[0].detail}
              </p>
            </article>
            <div className="relative z-10 flex shrink-0 items-center justify-center py-1 sm:px-1">
              <span className="flex size-10 items-center justify-center rounded-full border border-indigo-300/25 bg-[#101421] text-indigo-300">
                <ArrowRight className="size-5 rotate-90 sm:rotate-0" />
              </span>
            </div>
            <article className="flex-1 rounded-2xl bg-black/30 p-4 transition-colors hover:bg-white/[0.04]">
              <div className="flex items-center gap-2.5">
                <span
                  className={`flex size-9 items-center justify-center rounded-xl bg-white/[0.05] ${endpoints[1].color}`}
                >
                  <Globe2 className="size-5" />
                </span>
                <span className="text-sm font-medium text-white/70">
                  {endpoints[1].label}
                </span>
              </div>
              <p className="mt-4 break-all font-mono text-sm text-white sm:text-base">
                https://{endpoints[1].value}
              </p>
              <p className="mt-2 text-sm leading-5 text-white/40">
                {endpoints[1].detail}
              </p>
            </article>
          </div>
          <div className="mt-4 grid gap-2 text-sm text-white/50 sm:grid-cols-3 sm:gap-3">
            <span className="inline-flex items-center justify-center gap-2 rounded-lg bg-white/[0.025] px-3 py-2">
              <Check className="size-4 text-emerald-300" /> Reserve a hostname
            </span>
            <span className="inline-flex items-center justify-center gap-2 rounded-lg bg-white/[0.025] px-3 py-2">
              <Link2 className="size-4 text-cyan-300" /> Point a CNAME at the
              relay
            </span>
            <span className="inline-flex items-center justify-center gap-2 rounded-lg bg-white/[0.025] px-3 py-2">
              <Check className="size-4 text-emerald-300" /> Keep application
              routing unchanged
            </span>
          </div>
        </div>
      </MarketingContainer>
    </section>
  );
}
