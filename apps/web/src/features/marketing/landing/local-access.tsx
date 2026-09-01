import { LaptopMinimal, Network, ShieldCheck, Zap } from 'lucide-react';
import { MarketingContainer } from '#/components/layout';

const features = [
  {
    title: 'Private network reach',
    text: 'Connect services that live behind a firewall or on a trusted development network.',
    icon: Network,
    color: 'text-blue-300',
  },
  {
    title: 'Protected previews',
    text: 'Require a password before a public request reaches your local application.',
    icon: ShieldCheck,
    color: 'text-cyan-300',
  },
  {
    title: 'Zero ceremony',
    text: 'No port forwarding or hand-built reverse proxy for a temporary endpoint.',
    icon: Zap,
    color: 'text-emerald-300',
  },
];

export function LocalAccessSection() {
  return (
    <section className="py-16 sm:py-20">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-cyan-300/20 bg-cyan-300/10 px-3 py-1.5 text-sm text-cyan-200">
            <LaptopMinimal className="size-4" />
            Built around local development
          </div>
          <h2 className="text-4xl font-semibold tracking-[-0.05em] sm:text-5xl">
            Local apps, ready to share.
          </h2>
          <p className="mt-6 text-lg leading-8 text-white/50">
            Share a preview, receive a webhook, or test an OAuth provider
            without moving the service out of your workflow.
          </p>
        </div>
        <div className="mt-12 grid gap-10 lg:grid-cols-2">
          <div className="space-y-7">
            {features.map(({ title, text, icon: Icon, color }) => (
              <div key={title} className="flex items-start gap-4">
                <span className="flex size-10 shrink-0 items-center justify-center rounded-xl border border-white/10 bg-white/[0.03]">
                  <Icon className={`size-5 ${color}`} />
                </span>
                <div>
                  <h3 className="font-semibold">{title}</h3>
                  <p className="mt-1 text-sm leading-6 text-white/40">{text}</p>
                </div>
              </div>
            ))}
          </div>
          <div className="relative overflow-hidden rounded-2xl border border-white/10 bg-[#0c1019] shadow-2xl">
            <div className="flex items-center gap-2 border-b border-white/10 px-4 py-3">
              <span className="size-2.5 rounded-full bg-red-400/80" />
              <span className="size-2.5 rounded-full bg-amber-300/80" />
              <span className="size-2.5 rounded-full bg-emerald-400/80" />
            </div>
            <div className="space-y-4 p-6 font-mono text-sm">
              <p className="text-white/60">
                <span className="text-emerald-300">$</span> outpipe 3000
                --password
              </p>
              <p className="text-indigo-300">
                Tunnel: https://preview.outpipe.app
              </p>
              <p className="text-cyan-300">Protection: password required</p>
              <p className="text-white/35">
                No port forwarding. No application changes.
              </p>
            </div>
          </div>
        </div>
      </MarketingContainer>
    </section>
  );
}
