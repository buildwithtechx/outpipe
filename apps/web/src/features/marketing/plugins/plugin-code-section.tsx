import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { motion } from 'motion/react';
import { MarketingContainer } from '#/components/layout';
import type { PluginDefinition } from './plugin-data';

export function PluginCodeSection({ plugin }: { plugin: PluginDefinition }) {
  return (
    <section className="py-16 sm:py-20">
      <MarketingContainer className="grid items-center gap-14 lg:grid-cols-[0.85fr_1.15fr]">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
        >
          <h2 className="text-4xl font-semibold tracking-[-0.06em] sm:text-6xl">
            {plugin.integrationHeading.split('\n').map((line, index) => (
              <span
                key={line}
                className={index === 0 ? 'block' : 'block text-white/45'}
              >
                {line}
              </span>
            ))}
          </h2>
          <p className="mt-6 max-w-xl text-lg leading-8 text-white/45">
            {plugin.integrationDescription}
          </p>
          <div className="mt-9 flex items-center gap-6">
            <Link
              to="/signup"
              className="rounded-full bg-white px-6 py-3 text-sm font-semibold text-black transition-transform hover:-translate-y-0.5"
            >
              Get started
            </Link>
            <Link
              to="/docs/$"
              params={{ _splat: plugin.docsSlug }}
              className="inline-flex items-center gap-2 text-sm text-white/55 hover:text-white"
            >
              Documentation <ArrowRight className="size-4" />
            </Link>
          </div>
        </motion.div>
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
        >
          <CodePanel plugin={plugin} />
        </motion.div>
      </MarketingContainer>
    </section>
  );
}

function CodePanel({ plugin }: { plugin: PluginDefinition }) {
  return (
    <div className="overflow-hidden rounded-3xl border border-white/10 bg-[#0a0a0a] shadow-2xl shadow-indigo-950/20">
      <div className="flex items-center gap-2 border-b border-white/10 bg-white/[0.03] px-5 py-4">
        <span className="size-2.5 rounded-full bg-red-400/70" />
        <span className="size-2.5 rounded-full bg-amber-300/70" />
        <span className="size-2.5 rounded-full bg-emerald-400/70" />
        <span className="ml-2 font-mono text-xs text-white/35">
          {plugin.fileName}
        </span>
      </div>
      <pre className="min-h-72 overflow-x-auto p-6 font-mono text-xs leading-7 text-white/70 sm:text-sm">
        <code>{highlightCode(plugin.code)}</code>
      </pre>
      <div className="grid gap-2 border-t border-white/10 bg-black/25 px-6 py-5 font-mono text-xs text-white/35">
        <span>$ npm run dev</span>
        <span>
          <b className="font-normal text-emerald-300">➜</b> Local:{' '}
          <b className="font-normal text-indigo-300">http://localhost:5173/</b>
        </span>
        <span>
          <b className="font-normal text-emerald-300">➜</b> Tunnel:{' '}
          <b className="font-normal text-cyan-300">
            https://preview.outpipe.app
          </b>
        </span>
      </div>
    </div>
  );
}

function highlightCode(code: string) {
  const tokens = code.split(
    /(\b(?:import|from|export|default|const|function|return|await|new)\b|'[^']*'|"[^"]*"|\b\d+\b)/g,
  );
  let key = 0;
  return tokens.map((token) => {
    const tokenKey = key++;
    if (
      /^(import|from|export|default|const|function|return|await|new)$/.test(
        token,
      )
    ) {
      return (
        <span key={tokenKey} className="text-violet-300">
          {token}
        </span>
      );
    }
    if (/^(?:'[^']*'|"[^"]*")$/.test(token)) {
      return (
        <span key={tokenKey} className="text-emerald-300/80">
          {token}
        </span>
      );
    }
    if (/^\d+$/.test(token)) {
      return (
        <span key={tokenKey} className="text-amber-300">
          {token}
        </span>
      );
    }
    return token;
  });
}
