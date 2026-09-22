import { Link } from '@tanstack/react-router';
import { ArrowRight, Bug, History, Sparkles, Zap } from 'lucide-react';
import { motion } from 'motion/react';
import { MarketingContainer } from '#/components/layout';

const releases = [
  {
    date: 'July 30, 2026',
    type: 'Launch',
    title: 'One protocol across every client',
    description:
      'The CLI, desktop application, TypeScript SDK, and framework integrations now share one versioned tunnel lifecycle.',
    icon: Sparkles,
    tone: 'text-indigo-200 bg-indigo-300/10',
    highlights: [
      'Typed negotiation and tunnel state',
      'Reconnect-aware client lifecycle',
      'Vite, React, Next.js, NestJS, and Express adapters',
    ],
  },
  {
    date: 'June 18, 2026',
    type: 'Improvement',
    title: 'Traffic became easier to understand',
    description:
      'Usage and operational reporting now connect request outcomes, bandwidth, retention, and live relay health.',
    icon: Zap,
    tone: 'text-cyan-200 bg-cyan-300/10',
    highlights: [
      'Organization usage snapshots',
      'Request, error, and bandwidth measurements',
      'Relay health and operational metrics',
    ],
  },
  {
    date: 'May 4, 2026',
    type: 'Security',
    title: 'Stronger controls at the public edge',
    description:
      'Tunnel access gained password protection, scoped credentials, stricter target validation, and abuse controls.',
    icon: Bug,
    tone: 'text-amber-200 bg-amber-300/10',
    highlights: [
      'Password-protected HTTP endpoints',
      'Scoped API and tunnel credentials',
      'Rate limits and open-proxy protections',
    ],
  },
] as const;

export function ChangelogPage() {
  return (
    <section className="pb-20 pt-28 sm:pt-32">
      <MarketingContainer className="max-w-5xl">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          className="mx-auto max-w-3xl text-center"
        >
          <div className="inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/[0.04] px-4 py-2 text-sm text-white/55">
            <History className="size-4 text-indigo-200" /> Release notes
          </div>
          <h1 className="mt-6 text-4xl font-bold tracking-[-0.05em] sm:text-5xl lg:whitespace-nowrap">
            What changed, and why it matters.
          </h1>
          <p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-white/50">
            A running record of new capabilities, reliability work, and the
            details that improve everyday tunnel workflows.
          </p>
        </motion.div>

        <div className="relative mt-16">
          <div className="absolute bottom-0 left-[8.75rem] top-0 hidden w-px bg-gradient-to-b from-indigo-300/60 via-white/10 to-transparent md:block" />
          <div className="space-y-10">
            {releases.map((release, index) => (
              <motion.article
                key={release.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, margin: '-80px' }}
                transition={{ delay: index * 0.08 }}
                className="relative grid gap-4 md:grid-cols-[7rem_1fr] md:gap-12"
              >
                <time className="pt-1 text-sm text-white/35">
                  {release.date}
                </time>
                <span className="absolute left-[8.5rem] top-1.5 hidden size-2 rounded-full bg-indigo-200 shadow-[0_0_18px_rgba(199,210,254,0.8)] md:block" />
                <div className="rounded-3xl border border-white/10 bg-[#090909] p-6 sm:p-8">
                  <div
                    className={`inline-flex items-center gap-2 rounded-full px-3 py-1 text-xs font-medium ${release.tone}`}
                  >
                    <release.icon className="size-3.5" /> {release.type}
                  </div>
                  <h2 className="mt-5 text-2xl font-bold tracking-tight sm:text-3xl">
                    {release.title}
                  </h2>
                  <p className="mt-4 max-w-2xl leading-7 text-white/48">
                    {release.description}
                  </p>
                  <div className="mt-6 rounded-2xl border border-white/5 bg-white/[0.025] p-5">
                    <p className="text-xs font-semibold uppercase tracking-[0.18em] text-white/30">
                      Included in this release
                    </p>
                    <ul className="mt-4 grid gap-3 sm:grid-cols-2">
                      {release.highlights.map((highlight) => (
                        <li
                          key={highlight}
                          className="flex items-start gap-3 text-sm leading-6 text-white/60"
                        >
                          <span className="mt-2 size-1.5 shrink-0 rounded-full bg-indigo-300" />
                          {highlight}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </motion.article>
            ))}
          </div>
        </div>

        <div className="mt-14 flex flex-col items-center justify-between gap-5 rounded-3xl border border-white/10 bg-white/[0.025] p-6 text-center sm:flex-row sm:p-8 sm:text-left">
          <div>
            <h2 className="text-xl font-bold">Follow the implementation</h2>
            <p className="mt-1 text-sm text-white/40">
              Read the code, open an issue, or suggest the next improvement.
            </p>
          </div>
          <Link
            to="/docs/$"
            params={{ _splat: '' }}
            className="inline-flex shrink-0 items-center gap-2 rounded-full bg-white px-5 py-3 text-sm font-bold text-black"
          >
            Read the docs <ArrowRight className="size-4" />
          </Link>
        </div>
      </MarketingContainer>
    </section>
  );
}
