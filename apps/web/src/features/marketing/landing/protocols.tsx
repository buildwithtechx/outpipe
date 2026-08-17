import { ArrowUpRight, Database, Gamepad2, Globe2 } from 'lucide-react';
import { MarketingContainer } from '#/components/layout';

const protocols = [
  {
    title: 'HTTP and HTTPS',
    detail: 'Webhooks, previews, callbacks, and APIs.',
    command: 'outpipe http 3000',
    icon: Globe2,
    color: 'text-sky-300',
  },
  {
    title: 'TCP',
    detail: 'Databases, SSH, and private services.',
    command: 'outpipe tcp 5432',
    icon: Database,
    color: 'text-violet-300',
  },
  {
    title: 'UDP',
    detail: 'Real-time services and game servers.',
    command: 'outpipe udp 25565',
    icon: Gamepad2,
    color: 'text-amber-300',
  },
];

export function ProtocolsSection() {
  return (
    <section className="py-16 sm:py-20">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <h2 className="text-4xl font-semibold tracking-tighter sm:text-5xl">
            Every protocol. One workflow.
          </h2>
          <p className="mt-5 text-lg leading-8 text-white/50">
            Use one CLI, one account, and one tunnel lifecycle for HTTP, HTTPS,
            TCP, and UDP services.
          </p>
        </div>
        <div className="mx-auto mt-10 max-w-4xl overflow-hidden rounded-3xl border border-white/10 bg-white/2.5 p-3 sm:p-4">
          <div className="mb-2 flex items-center justify-between px-3 py-2 font-mono text-xs text-white/35">
            <span>protocols</span>
            <span className="text-emerald-300">ready</span>
          </div>
          <div className="space-y-2">
            {protocols.map(({ title, detail, command, icon: Icon, color }) => (
              <article
                key={title}
                className="group grid gap-4 rounded-2xl border border-transparent bg-black/25 p-4 transition-colors hover:border-white/10 hover:bg-white/[0.035] sm:grid-cols-[auto_1fr_auto] sm:items-center"
              >
                <span
                  className={`flex size-10 items-center justify-center rounded-xl bg-white/5 ${color}`}
                >
                  <Icon className="size-5" />
                </span>
                <div>
                  <h3 className="font-medium text-white">{title}</h3>
                  <p className="mt-1 text-sm text-white/40">{detail}</p>
                </div>
                <div className="flex items-center justify-between gap-4 font-mono text-xs text-white/50 sm:justify-end">
                  <code>{command}</code>
                  <ArrowUpRight className={`size-4 ${color}`} />
                </div>
              </article>
            ))}
          </div>
        </div>
      </MarketingContainer>
    </section>
  );
}
