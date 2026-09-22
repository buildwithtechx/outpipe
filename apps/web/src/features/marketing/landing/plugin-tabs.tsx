import { Link } from '@tanstack/react-router';
import { useState } from 'react';
import { SiNestjs, SiReact, SiVite } from 'react-icons/si';

const tabs = {
  vite: {
    label: 'Vite',
    file: 'vite.config.ts',
    icon: SiVite,
    color: 'text-[#646cff]',
    pluginId: 'vite',
  },
  react: {
    label: 'React',
    file: 'app.tsx',
    icon: SiReact,
    color: 'text-[#61dafb]',
    pluginId: 'react',
  },
  nest: {
    label: 'NestJS',
    file: 'app.module.ts',
    icon: SiNestjs,
    color: 'text-[#e0234e]',
    pluginId: 'nest',
  },
} as const;

type Tab = keyof typeof tabs;

export function PluginTabs() {
  const [active, setActive] = useState<Tab>('vite');
  const selected = tabs[active];

  return (
    <div className="mt-auto pt-7">
      <div className="mb-3 flex flex-wrap gap-2">
        {(Object.keys(tabs) as Tab[]).map((id) => {
          const tab = tabs[id];
          return (
            <button
              key={id}
              type="button"
              onClick={() => setActive(id)}
              className={`flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm transition-colors ${active === id ? 'bg-white/10 text-white' : 'text-white/35 hover:text-white/65'}`}
            >
              <tab.icon
                className={`size-4 ${active === id ? tab.color : ''}`}
              />
              {tab.label}
            </button>
          );
        })}
      </div>
      <div className="overflow-hidden rounded-2xl border border-white/5 bg-black/40">
        <div className="flex items-center gap-1.5 border-b border-white/5 bg-white/4 px-4 py-2.5">
          <span className="size-2.5 rounded-full bg-red-400/65" />
          <span className="size-2.5 rounded-full bg-amber-300/65" />
          <span className="size-2.5 rounded-full bg-emerald-400/65" />
          <span className="ml-2 font-mono text-[11px] text-white/30">
            {selected.file}
          </span>
        </div>
        <pre className="min-h-40 overflow-x-auto p-4 font-mono text-xs leading-6 text-white/65">
          {active === 'vite' && (
            <code>
              <span className="text-violet-300">import</span> outpipe{' '}
              <span className="text-violet-300">from</span>{' '}
              <span className="text-emerald-300">'@outpipe/vite-plugin'</span>
              {';'}
              {'\n\n'}
              <span className="text-violet-300">export default</span>{' '}
              defineConfig({'{'}
              {'\n  '}plugins: [outpipe()],
              {'\n'}
              {'}'});
            </code>
          )}
          {active === 'react' && (
            <code>
              <span className="text-violet-300">import</span>{' '}
              {'{ OutpipeProvider }'}{' '}
              <span className="text-violet-300">from</span>{' '}
              <span className="text-emerald-300">'@outpipe/react'</span>
              {';'}
              {'\n\n'}
              <span className="text-violet-300">export function</span> App(){' '}
              {'{'}
              {'\n  '}return &lt;OutpipeProvider /&gt;;
              {'\n'}
              {'}'}
            </code>
          )}
          {active === 'nest' && (
            <code>
              <span className="text-violet-300">import</span>{' '}
              {'{ OutpipeModule }'}{' '}
              <span className="text-violet-300">from</span>{' '}
              <span className="text-emerald-300">'@outpipe/nest'</span>
              {';'}
              {'\n\n'}@Module({'{'}
              {'\n  '}imports: [OutpipeModule.forRoot({'{'} localPort:{' '}
              <span className="text-sky-300">3000</span> {'}'})],
              {'\n'}
              {'}'})
            </code>
          )}
        </pre>
      </div>
      <Link
        to="/plugins/$pluginId"
        params={{ pluginId: selected.pluginId }}
        className="mt-3 inline-flex text-sm text-indigo-300 hover:text-indigo-200"
      >
        Explore the {selected.label} integration →
      </Link>
    </div>
  );
}
