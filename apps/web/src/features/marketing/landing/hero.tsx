import { Canvas } from '@react-three/fiber';
import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { AnimatePresence, motion, useReducedMotion } from 'motion/react';
import { useEffect, useState } from 'react';
import { MarketingContainer } from '#/components/layout';
import { CliInstall } from '#/features/tunnels/components/cli-install';
import { docsUrl } from '#/lib/docs';
import { BeamGroup } from './beam-group';
import { TerminalWindow } from './terminal-window';

const logs = [
  ['GET', '/api/health', '200'],
  ['POST', '/webhooks/stripe', '201'],
  ['GET', '/oauth/callback', '302'],
  ['GET', '/api/projects', '200'],
];

export function Hero() {
  const reducedMotion = useReducedMotion();
  const [hovered, setHovered] = useState(false);
  const [visibleLogs, setVisibleLogs] = useState(
    logs.slice(0, 3).map((log, id) => ({ log, id })),
  );

  useEffect(() => {
    if (!hovered || reducedMotion) return;
    let index = 3;
    const timer = setInterval(() => {
      setVisibleLogs((current) => [
        ...current.slice(1),
        { log: logs[index++ % logs.length], id: index },
      ]);
    }, 800);
    return () => clearInterval(timer);
  }, [hovered, reducedMotion]);

  return (
    <section className="relative min-h-screen overflow-hidden bg-black pb-12 pt-28 sm:pb-16 sm:pt-20">
      <div className="pointer-events-none absolute inset-0 z-0 md:translate-x-[-10%]">
        {!reducedMotion && (
          <Canvas camera={{ position: [0, 0, 15], fov: 45 }}>
            <color attach="background" args={['#000000']} />
            <BeamGroup />
          </Canvas>
        )}
      </div>
      <MarketingContainer className="relative z-10 flex flex-col items-center">
        <div className="inline-flex max-w-full items-center gap-2 rounded-full border border-white/10 bg-white/4 px-3 py-1.5 text-center text-[10px] text-white/55 backdrop-blur-xs sm:mt-20 sm:text-xs">
          <span className="size-1.5 shrink-0 rounded-full bg-cyan-300" />
          Secure public endpoints for your local apps
        </div>
        <h1 className="mt-6 w-full text-center text-[clamp(1.75rem,6vw,4.5rem)] font-bold leading-[1.1] tracking-[-0.055em] sm:mt-8 sm:leading-[1.02]">
          <span className="block sm:whitespace-nowrap">
            <motion.button
              type="button"
              className="relative inline-block cursor-default appearance-none border-0 bg-transparent p-0 text-inherit"
              onMouseEnter={() => setHovered(true)}
              onMouseLeave={() => setHovered(false)}
            >
              <motion.span
                animate={{
                  rotate: hovered && !reducedMotion ? -5 : 0,
                  y: hovered && !reducedMotion ? -4 : 0,
                }}
                transition={{ type: 'spring', stiffness: 300, damping: 20 }}
                className="relative z-10 inline-block rounded-2xl border border-indigo-300/35 bg-indigo-300/15 px-2 py-1 sm:px-4"
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
            <span className="inline-block whitespace-nowrap">
              your local app
            </span>
          </span>
          <span className="block">with the world</span>
        </h1>
        <p className="mt-6 max-w-2xl text-center text-base leading-7 text-white/55 sm:mt-8 sm:text-xl sm:leading-8">
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
          <a
            href={docsUrl('installation')}
            className="inline-flex min-h-11 items-center rounded-full border border-white/10 px-6 text-sm text-white/70"
          >
            Read the quickstart
          </a>
        </div>
        <div className="mt-8 w-full max-w-xl rounded-2xl border border-border bg-card p-5 text-left">
          <CliInstall />
        </div>
        <TerminalWindow />
      </MarketingContainer>
    </section>
  );
}
