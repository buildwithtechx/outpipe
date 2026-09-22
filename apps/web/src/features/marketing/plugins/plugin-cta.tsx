import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { motion } from 'motion/react';
import { MarketingContainer } from '#/components/layout';
import type { PluginDefinition } from './plugin-data';

export function PluginCta({ plugin }: { plugin: PluginDefinition }) {
  return (
    <section className="py-16 sm:py-20">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.5 }}
      >
        <MarketingContainer className="text-center">
          <div className="mx-auto flex size-16 items-center justify-center rounded-2xl border border-white/10 bg-white/[0.04] text-3xl text-white">
            <plugin.icon />
          </div>
          <h2 className="mt-7 text-4xl font-semibold tracking-[-0.06em] sm:text-6xl">
            Put your {plugin.name} app on the web.
          </h2>
          <p className="mx-auto mt-5 max-w-xl text-lg text-white/45">
            Create a public endpoint for your local server in a few seconds.
          </p>
          <Link
            to="/signup"
            className="mt-9 inline-flex items-center gap-2 rounded-full bg-white px-7 py-4 text-sm font-semibold text-[#080b14] transition-transform hover:-translate-y-0.5"
          >
            Create your tunnel <ArrowRight className="size-4" />
          </Link>
        </MarketingContainer>
      </motion.div>
    </section>
  );
}
