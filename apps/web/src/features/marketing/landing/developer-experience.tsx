import { Activity, Code2, Terminal } from 'lucide-react';
import { useEffect, useState } from 'react';
import { MarketingContainer } from '#/components/layout';
import { PluginTabs } from './plugin-tabs';

const traffic = [
  ['200 OK', 'GET', '/api/users', '12ms'],
  ['201 Created', 'POST', '/api/webhooks', '45ms'],
  ['401 Unauth', 'GET', '/admin', '8ms'],
  ['200 OK', 'GET', '/api/settings', '24ms'],
  ['500 Error', 'POST', '/api/checkout', '120ms'],
  ['204 No Content', 'OPTIONS', '/api/tunnels', '4ms'],
  ['302 Found', 'GET', '/oauth/callback', '21ms'],
  ['200 OK', 'POST', '/api/deployments', '67ms'],
  ['201 Created', 'POST', '/api/keys', '31ms'],
];

const visibleTrafficItems = 8;

export function DeveloperExperience() {
  const [offset, setOffset] = useState(0);
  useEffect(() => {
    const timer = setInterval(() => setOffset((value) => value + 1), 1500);
    return () => clearInterval(timer);
  }, []);
  const liveTraffic = [...traffic, ...traffic].slice(
    offset % traffic.length,
    (offset % traffic.length) + visibleTrafficItems,
  );
  return (
    <section className="bg-black py-14 sm:py-16">
      <MarketingContainer>
        <h2 className="mx-auto max-w-3xl text-center text-4xl font-bold tracking-[-0.05em] sm:text-6xl">
          Your workflow,
          <br />
          already connected
        </h2>
        <div className="mt-10 grid items-start gap-6 lg:grid-cols-2">
          <div className="grid gap-6">
            <article className="group flex flex-col rounded-3xl border border-white/5 bg-white/[0.02] p-8 transition-colors hover:border-white/10">
              <div className="flex items-center gap-4">
                <span className="flex size-10 items-center justify-center rounded-full bg-indigo-300/10 transition-colors group-hover:bg-indigo-300/20">
                  <Terminal className="size-5 text-indigo-300" />
                </span>
                <h3 className="text-xl font-semibold">Online in one command</h3>
              </div>
              <p className="mt-6 text-white/45">
                One command and your local service has a public endpoint. No
                reverse proxy configuration required.
              </p>
              <div className="mt-8 rounded-2xl border border-white/10 bg-black/30 p-4 font-mono text-sm text-white/65">
                <span className="text-indigo-300">$</span> outpipe 3000
              </div>
            </article>
            <article className="group rounded-3xl border border-white/5 bg-white/[0.02] p-8 transition-colors hover:border-white/10">
              <div className="flex items-center gap-4">
                <span className="flex size-10 items-center justify-center rounded-full bg-indigo-300/10 transition-colors group-hover:bg-indigo-300/20">
                  <Code2 className="size-5 text-indigo-300" />
                </span>
                <h3 className="text-xl font-semibold">
                  Prefer your framework?
                </h3>
              </div>
              <p className="mt-6 text-white/45">
                Use a thin adapter in Vite, Next.js, NestJS, Express, or React
                while the same SDK handles the tunnel lifecycle.
              </p>
              <PluginTabs />
            </article>
          </div>
          <article className="group relative flex h-full flex-col overflow-hidden rounded-3xl border border-white/5 bg-white/[0.02] p-8 transition-colors hover:border-white/10">
            <div className="absolute -right-24 -top-24 size-64 rounded-full bg-indigo-400/10 blur-3xl" />
            <div className="relative">
              <div className="flex items-center gap-4">
                <span className="flex size-10 items-center justify-center rounded-full bg-indigo-300/10 transition-colors group-hover:bg-indigo-300/20">
                  <Activity className="size-5 text-indigo-300" />
                </span>
                <h3 className="text-xl font-semibold">Instant observability</h3>
              </div>
              <p className="mt-6 text-white/45">
                See live traffic as soon as the tunnel comes online, with
                status, path, duration, and response outcomes.
              </p>
            </div>
            <div className="relative mt-auto space-y-2.5 pt-8 font-mono text-xs">
              {liveTraffic.map(([status, method, path, time]) => (
                <div
                  key={`${status}-${path}-${offset}`}
                  className="grid grid-cols-[90px_48px_1fr_42px] gap-2 rounded-lg border border-white/5 bg-black/30 px-3 py-3 text-white/50"
                >
                  <span
                    className={
                      status.startsWith('5')
                        ? 'text-red-300'
                        : status.startsWith('4')
                          ? 'text-white/35'
                          : 'text-indigo-300'
                    }
                  >
                    {status}
                  </span>
                  <span>{method}</span>
                  <span>{path}</span>
                  <span className="text-right text-white/25">{time}</span>
                </div>
              ))}
            </div>
          </article>
        </div>
      </MarketingContainer>
    </section>
  );
}
