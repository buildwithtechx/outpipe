import { Canvas } from '@react-three/fiber';
import { Link } from '@tanstack/react-router';
import { ArrowRight, Check, Copy } from 'lucide-react';
import { motion, useMotionValue, useSpring, useTransform } from 'motion/react';
import { useRef, useState } from 'react';
import { MarketingContainer } from '#/components/layout';
import { PluginBeam } from './plugin-beam';
import type { PluginDefinition } from './plugin-data';

const accents = {
  sdk: {
    text: 'text-cyan-300',
    button: 'bg-cyan-300 text-[#081116]',
  },
  react: {
    text: 'text-sky-300',
    button: 'bg-sky-300 text-[#071118]',
  },
  vite: {
    text: 'text-indigo-300',
    button: 'bg-indigo-300 text-[#0b0b18]',
  },
  next: {
    text: 'text-white',
    button: 'bg-white text-[#080b14]',
  },
  nest: {
    text: 'text-rose-300',
    button: 'bg-rose-400 text-[#19080d]',
  },
  express: {
    text: 'text-amber-300',
    button: 'bg-amber-300 text-[#171107]',
  },
} as const;

const particleAngles = [
  0,
  Math.PI / 4,
  Math.PI / 2,
  (Math.PI * 3) / 4,
  Math.PI,
  (Math.PI * 5) / 4,
  (Math.PI * 3) / 2,
  (Math.PI * 7) / 4,
];

export function PluginHero({ plugin }: { plugin: PluginDefinition }) {
  const [copied, setCopied] = useState(false);
  const accent = accents[plugin.id];
  const logoRef = useRef<HTMLDivElement>(null);
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);
  const rotateX = useSpring(useTransform(mouseY, [-0.5, 0.5], [15, -15]), {
    damping: 20,
    stiffness: 300,
  });
  const rotateY = useSpring(useTransform(mouseX, [-0.5, 0.5], [-15, 15]), {
    damping: 20,
    stiffness: 300,
  });

  function moveLogo(event: React.MouseEvent<HTMLDivElement>) {
    if (!logoRef.current) return;
    const bounds = logoRef.current.getBoundingClientRect();
    mouseX.set((event.clientX - bounds.left - bounds.width / 2) / bounds.width);
    mouseY.set(
      (event.clientY - bounds.top - bounds.height / 2) / bounds.height,
    );
  }

  async function copyInstall() {
    await navigator.clipboard.writeText(plugin.install);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  }

  return (
    <section className="relative flex min-h-[82vh] items-center overflow-hidden bg-black pb-16 pt-32 sm:pt-36">
      <div className="absolute inset-0 z-0">
        <Canvas camera={{ position: [0, 0, 15], fov: 45 }}>
          <color attach="background" args={['#000000']} />
          <PluginBeam />
        </Canvas>
      </div>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="relative z-10 w-full"
      >
        <MarketingContainer className="flex flex-col items-center text-center">
          <nav
            className="relative mb-8"
            onMouseMove={moveLogo}
            onMouseLeave={() => {
              mouseX.set(0);
              mouseY.set(0);
            }}
          >
            {particleAngles.map((angle) => (
              <motion.span
                key={angle}
                className={`absolute left-1/2 top-1/2 size-1 rounded-full ${accent.text.replace('text-', 'bg-')}`}
                animate={{
                  x: [Math.cos(angle) * 65, Math.cos(angle + Math.PI * 2) * 65],
                  y: [Math.sin(angle) * 65, Math.sin(angle + Math.PI * 2) * 65],
                  opacity: [0.25, 0.8, 0.25],
                }}
                transition={{
                  duration: 8,
                  repeat: Infinity,
                  ease: 'linear',
                  delay: angle * 0.04,
                }}
              />
            ))}
            <motion.div
              ref={logoRef}
              style={{ rotateX, rotateY, transformStyle: 'preserve-3d' }}
              className={`relative flex size-24 items-center justify-center rounded-3xl border border-white/15 bg-white/[0.06] text-5xl shadow-2xl ${accent.text}`}
            >
              <div
                className="absolute inset-0 rounded-3xl bg-white/[0.04]"
                style={{ transform: 'translateZ(-18px)' }}
              />
              <plugin.icon className="relative z-10" />
            </motion.div>
          </nav>
          <p className={`text-sm font-semibold ${accent.text}`}>
            {plugin.eyebrow}
          </p>
          <motion.h1
            initial={{ opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1 }}
            className="mx-auto mt-5 max-w-4xl text-5xl font-semibold tracking-[-0.07em] sm:text-7xl"
          >
            {plugin.headline.split('\n').map((line, index) => (
              <span
                key={line}
                className={index === 0 ? 'block' : 'block text-white/55'}
              >
                {line}
              </span>
            ))}
          </motion.h1>
          <p className="mx-auto mt-7 max-w-2xl text-lg leading-8 text-white/50">
            {plugin.description}
          </p>
          <div className="mt-9 flex flex-col items-center gap-3 sm:flex-row">
            <Link
              to="/signup"
              className={`group inline-flex items-center gap-2 rounded-full px-7 py-4 text-sm font-semibold transition-transform hover:-translate-y-0.5 ${accent.button}`}
            >
              Start building{' '}
              <ArrowRight className="size-4 transition-transform group-hover:translate-x-1" />
            </Link>
            <Link
              to="/docs/$"
              params={{ _splat: plugin.docsSlug }}
              className="rounded-full border border-white/15 px-7 py-4 text-sm text-white/70 hover:bg-white/5"
            >
              Documentation
            </Link>
          </div>
          <button
            type="button"
            onClick={copyInstall}
            className="group mt-8 inline-flex items-center gap-3 rounded-full border border-white/10 bg-black/35 px-5 py-3 font-mono text-xs text-white/55 hover:border-white/25"
          >
            <span className={accent.text}>$</span>
            {plugin.install}
            {copied ? (
              <Check className="size-4 text-emerald-300" />
            ) : (
              <Copy className="size-4 opacity-0 group-hover:opacity-100" />
            )}
          </button>
        </MarketingContainer>
      </motion.div>
    </section>
  );
}
