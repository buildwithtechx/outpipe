import { ArrowRight, Globe2, Laptop, Server } from 'lucide-react';
import { motion } from 'motion/react';
import { MarketingContainer } from '#/components/layout';

const connectionSteps = [
  {
    title: 'Your local service',
    detail: 'localhost:3000',
    icon: Laptop,
    color: 'text-amber-300',
  },
  {
    title: 'Outpipe relay',
    detail: 'encrypted session',
    icon: Server,
    color: 'text-indigo-300',
  },
  {
    title: 'Public endpoint',
    detail: 'preview.outpipe.app',
    icon: Globe2,
    color: 'text-cyan-300',
  },
];

export function NetworkDiagram() {
  return (
    <section className="py-16 sm:py-20">
      <MarketingContainer>
        <div className="mx-auto max-w-3xl text-center">
          <p className="text-sm font-medium text-indigo-300">Connection path</p>
          <h2 className="mt-4 text-4xl font-semibold tracking-[-0.05em] sm:text-5xl">
            One path to a public URL.
          </h2>
          <p className="mt-5 text-lg leading-8 text-white/50">
            Your client opens an encrypted outbound session. The relay keeps it
            connected, then routes each public request back to your service.
          </p>
        </div>
        <div className="mx-auto mt-10 max-w-5xl rounded-3xl border border-white/10 bg-white/[0.025] p-4 sm:p-6">
          <div className="flex items-center justify-between border-b border-white/10 pb-4 font-mono text-xs text-white/35">
            <span>outpipe 3000</span>
            <span className="flex items-center gap-2 text-emerald-300">
              <span className="size-2 rounded-full bg-emerald-300" /> session
              active
            </span>
          </div>
          <div className="mt-5 flex flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:gap-2">
            {connectionSteps.map(
              ({ title, detail, icon: Icon, color }, index) => (
                <div key={title} className="contents">
                  <article className="flex flex-1 items-center gap-3 rounded-2xl bg-black/30 p-4">
                    <span
                      className={`flex size-10 shrink-0 items-center justify-center rounded-xl bg-white/[0.05] ${color}`}
                    >
                      <Icon className="size-5" />
                    </span>
                    <div className="min-w-0">
                      <h3 className="text-sm font-medium text-white">
                        {title}
                      </h3>
                      <p className="mt-1 truncate font-mono text-xs text-white/40">
                        {detail}
                      </p>
                    </div>
                  </article>
                  {index < connectionSteps.length - 1 && (
                    <span className="flex shrink-0 justify-center text-indigo-300">
                      <motion.span
                        animate={{ x: [0, 4, 0] }}
                        transition={{
                          duration: 1.4,
                          repeat: Infinity,
                          ease: 'easeInOut',
                        }}
                      >
                        <ArrowRight className="size-5 rotate-90 sm:rotate-0" />
                      </motion.span>
                    </span>
                  )}
                </div>
              ),
            )}
          </div>
          <p className="mt-5 text-center text-sm text-white/40">
            No inbound port opening. No reverse-proxy setup. Just an outbound
            connection your client controls.
          </p>
        </div>
      </MarketingContainer>
    </section>
  );
}
