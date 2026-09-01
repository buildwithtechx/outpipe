import { Canvas } from '@react-three/fiber';
import { Link } from '@tanstack/react-router';
import { ArrowRight, Check, Copy } from 'lucide-react';
import { AnimatePresence, motion } from 'motion/react';
import { useEffect, useState } from 'react';
import { MarketingContainer } from '#/components/layout';
import { BeamGroup } from './beam-group';

const logs = [
  ['GET', '/api/health', '200'],
  ['POST', '/webhooks/stripe', '201'],
  ['GET', '/oauth/callback', '302'],
  ['GET', '/api/projects', '200'],
];

const terminalRequestSequence = [
  ['GET', '/api/health', '200', '12ms'],
  ['POST', '/webhooks', '201', '45ms'],
  ['GET', '/oauth/callback', '302', '8ms'],
  ['GET', '/favicon.ico', '304', '2ms'],
  ['GET', '/api/projects', '200', '24ms'],
  ['POST', '/api/uploads', '201', '230ms'],
  ['DELETE', '/api/sessions/123', '204', '35ms'],
  ['PATCH', '/api/profile', '200', '67ms'],
  ['POST', '/api/checkout', '500', '120ms'],
] as const;

function statusColor(status: string) {
  const code = Number(status);

  if (code >= 500) return 'text-red-300';
  if (code >= 400) return 'text-orange-300';
  if (code >= 300) return 'text-amber-300';
  return 'text-emerald-300';
}

export function Hero() {
  const [copied, setCopied] = useState(false);
  const [hovered, setHovered] = useState(false);
  const [visibleLogs, setVisibleLogs] = useState(
    logs.slice(0, 3).map((log, id) => ({ log, id })),
  );

  useEffect(() => {
    if (!hovered) return;
    let index = 3;
    const timer = setInterval(() => {
      setVisibleLogs((current) => [
        ...current.slice(1),
        { log: logs[index++ % logs.length], id: index },
      ]);
    }, 800);
    return () => clearInterval(timer);
  }, [hovered]);

  async function copyCommand() {
    await navigator.clipboard.writeText(
      'curl -fsSL https://cli.outpipe.dev | bash',
    );
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  }

  return (
    <section className="relative min-h-screen overflow-hidden bg-black pb-16 pt-20">
      <div className="pointer-events-none absolute inset-0 z-0 md:translate-x-[-10%]">
        <Canvas camera={{ position: [0, 0, 15], fov: 45 }}>
          <color attach="background" args={['#000000']} />
          <BeamGroup />
        </Canvas>
      </div>
      <MarketingContainer className="relative z-10 flex flex-col items-center">
        <div className="mt-20 inline-flex items-center gap-2 rounded-full border border-white/10 bg-white/4 px-3 py-1.5 text-xs text-white/55 backdrop-blur-sm">
          <span className="size-1.5 rounded-full bg-cyan-300" />
          Open-source tunnel infrastructure for developers
        </div>
        <h1 className="mt-8 w-full text-center text-[clamp(1.8rem,6vw,4.5rem)] font-bold leading-[1.02] tracking-[-0.055em]">
          <span className="block whitespace-nowrap">
            <motion.button
              type="button"
              className="relative inline-block cursor-default appearance-none border-0 bg-transparent p-0 text-inherit"
              onMouseEnter={() => setHovered(true)}
              onMouseLeave={() => setHovered(false)}
            >
              <motion.span
                animate={{ rotate: hovered ? -5 : 0, y: hovered ? -4 : 0 }}
                transition={{ type: 'spring', stiffness: 300, damping: 20 }}
                className="relative z-10 inline-block rounded-2xl border border-indigo-300/35 bg-indigo-300/15 px-4 py-1"
              >
                Share
              </motion.span>
              <span className="pointer-events-none absolute inset-0 flex overflow-hidden rounded-2xl border border-indigo-300/30 bg-black px-3 py-1 font-mono text-[0.16em] leading-tight">
                <span className="flex w-full flex-col justify-center">
                  <AnimatePresence mode="popLayout" initial={false}>
                    {visibleLogs.map(({ log, id }) => (
                      <motion.span
                        key={id}
                        initial={{ opacity: 0, x: -10 }}
                        animate={{ opacity: hovered ? 1 : 0, x: 0 }}
                        exit={{ opacity: 0 }}
                        className="flex items-center gap-2 whitespace-nowrap"
                      >
                        <span className="w-10 text-left text-indigo-300">
                          {log[0]}
                        </span>
                        <span className="flex-1 text-left text-white/50">
                          {log[1]}
                        </span>
                        <span className="text-emerald-300">{log[2]}</span>
                      </motion.span>
                    ))}
                  </AnimatePresence>
                </span>
              </span>
            </motion.button>{' '}
            your local app
          </span>
          <span className="block">with the world</span>
        </h1>
        <p className="mt-8 max-w-2xl text-center text-lg leading-8 text-white/55 sm:text-xl">
          Outpipe gives your local apps a secure, observable public endpoint for
          previews, webhooks, OAuth callbacks, and CI workflows.
        </p>
        <div className="mt-9 flex w-full flex-col items-center justify-center gap-3 sm:flex-row">
          <Link
            to="/signup"
            className="group inline-flex w-full items-center justify-center gap-2 rounded-full bg-white px-7 py-4 text-base font-semibold text-[#080b14] transition-transform hover:-translate-y-0.5 sm:w-auto"
          >
            Get started free{' '}
            <ArrowRight className="size-5 transition-transform group-hover:translate-x-1" />
          </Link>
          <button
            type="button"
            onClick={copyCommand}
            className="group inline-flex w-full items-center justify-center gap-3 rounded-full border border-white/10 bg-white/4 px-6 py-4 font-mono text-xs text-white/60 transition-colors hover:border-indigo-300/40 hover:bg-white/8 sm:w-auto"
          >
            <span className="text-indigo-300">$</span> curl .../install.sh{' '}
            {copied ? (
              <Check className="size-4 text-emerald-300" />
            ) : (
              <Copy className="size-4 opacity-0 transition-opacity group-hover:opacity-100" />
            )}
          </button>
        </div>
        <TerminalWindow />
      </MarketingContainer>
    </section>
  );
}

function TerminalWindow() {
  const [visibleRequests, setVisibleRequests] = useState(() =>
    terminalRequestSequence.slice(0, 8).map((request, id) => ({ request, id })),
  );

  useEffect(() => {
    let nextRequest = 8;
    const timer = setInterval(() => {
      setVisibleRequests((current) => [
        ...current.slice(1),
        {
          request:
            terminalRequestSequence[
              nextRequest % terminalRequestSequence.length
            ],
          id: nextRequest++,
        },
      ]);
    }, 1400);

    return () => clearInterval(timer);
  }, []);

  return (
    <div className="mt-14 w-full max-w-5xl overflow-hidden rounded-[1.25rem] border border-white/15 bg-[#090a0c] text-left font-mono text-sm shadow-2xl shadow-indigo-950/50">
      <div className="relative flex items-center gap-3 border-b border-white/10 bg-white/6 px-6 py-4">
        <span className="size-3 rounded-full bg-red-400" />
        <span className="size-3 rounded-full bg-amber-300" />
        <span className="size-3 rounded-full bg-emerald-400" />
        <span className="absolute inset-x-0 text-center text-sm text-white/35">
          user@outpipe-cli
        </span>
      </div>
      <div className="grid gap-2 px-6 py-7 text-xs leading-6 sm:px-8 sm:py-8 sm:text-sm">
        <p className="text-white/85">
          <span className="text-emerald-300">➜</span>{' '}
          <span className="text-cyan-300">~</span> outpipe 3000
        </p>
        <p className="text-cyan-300">Connecting to Outpipe...</p>
        <p className="text-emerald-300">Linked to local port 3000</p>
        <p className="text-fuchsia-300">
          Tunnel ready: https://quiet-moon.outpipe.app
        </p>
        <p className="text-amber-300">
          Keep this process running to keep the tunnel active.
        </p>
        <div className="mt-3 grid gap-2 text-white/45">
          <AnimatePresence initial={false} mode="popLayout">
            {visibleRequests.map(({ request, id }) => (
              <motion.div
                key={id}
                initial={{ opacity: 0, y: 8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                className="grid grid-cols-[4.5rem_minmax(0,1fr)_3.5rem_3.5rem] items-center gap-3"
              >
                <span>{request[0]}</span>
                <span className="truncate text-white/65">{request[1]}</span>
                <span className={`text-right ${statusColor(request[2])}`}>
                  {request[2]}
                </span>
                <span className="text-right text-white/30">{request[3]}</span>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
