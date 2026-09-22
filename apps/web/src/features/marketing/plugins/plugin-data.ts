import { Cable } from 'lucide-react';
import {
  SiAstro,
  SiDrizzle,
  SiExpress,
  SiGraphql,
  SiMongodb,
  SiNestjs,
  SiNextdotjs,
  SiNodedotjs,
  SiPostgresql,
  SiPrisma,
  SiReact,
  SiRedis,
  SiRemix,
  SiSocketdotio,
  SiSolid,
  SiStripe,
  SiSvelte,
  SiSwagger,
  SiTailwindcss,
  SiTrpc,
  SiTypescript,
  SiVercel,
  SiVite,
  SiVuedotjs,
} from 'react-icons/si';

export type PluginId = 'sdk' | 'react' | 'vite' | 'next' | 'nest' | 'express';

export type PluginDefinition = {
  id: PluginId;
  name: string;
  packageName: string;
  eyebrow: string;
  headline: string;
  description: string;
  docsSlug: string;
  install: string;
  fileName: string;
  code: string;
  colorClass: string;
  icon: React.ComponentType<{ className?: string }>;
  features: string[];
  useCases: string[];
  stackDescription: string;
  integrationHeading: string;
  integrationDescription: string;
  technologies: {
    label: string;
    icon: React.ComponentType<{ className?: string }>;
  }[];
};

const typescript = { label: 'TypeScript', icon: SiTypescript };

export const pluginDefinitions: PluginDefinition[] = [
  {
    id: 'sdk',
    name: 'TypeScript SDK',
    packageName: '@outpipe/sdk',
    eyebrow: 'Framework-neutral foundation',
    headline: 'Bring public access\nto any TypeScript service.',
    description:
      'A small browser and Node.js client for applications that need direct control over authentication, tunnel creation, status, and shutdown.',
    docsSlug: 'sdk',
    install: 'npm install @outpipe/sdk',
    fileName: 'server.ts',
    code: "import { OutpipeClient } from '@outpipe/sdk';\n\nconst client = new OutpipeClient({\n  apiKey: process.env.OUTPIPE_API_KEY,\n});\n\nawait client.openTunnel({ protocol: 'http', localPort: 3000 });",
    colorClass: 'text-cyan-300',
    icon: Cable,
    features: [
      'Browser and Node.js support',
      'Typed protocol lifecycle',
      'Reconnect-aware state',
      'Fetch-based transport',
    ],
    useCases: [
      'Preview environments',
      'Custom developer tooling',
      'CI pipeline jobs',
    ],
    stackDescription:
      'Use the same typed tunnel lifecycle in a browser, Node.js service, or your own developer tool.',
    integrationHeading: 'Embed a tunnel in any\nTypeScript service.',
    integrationDescription:
      'Use the framework-neutral client when your runtime owns the server, the workflow, or the developer experience around it.',
    technologies: [typescript, { label: 'Node.js', icon: SiNodedotjs }],
  },
  {
    id: 'react',
    name: 'React',
    packageName: '@outpipe/react',
    eyebrow: 'React integration',
    headline: 'Share React previews\nwith your team.',
    description:
      'Provider and hooks for showing connection state, public URLs, and tunnel controls directly in your React application.',
    docsSlug: 'react',
    install: 'npm install @outpipe/react',
    fileName: 'app.tsx',
    code: "import { OutpipeProvider, useTunnel } from '@outpipe/react';\n\nfunction PreviewStatus() {\n  const { tunnel, status } = useTunnel();\n  return <span>{tunnel?.publicUrl ?? status}</span>;\n}",
    colorClass: 'text-sky-300',
    icon: SiReact,
    features: [
      'Provider and hooks',
      'Reactive tunnel status',
      'Typed lifecycle actions',
      'Works with React 18+',
    ],
    useCases: [
      'Preview dashboards',
      'Internal developer portals',
      'Live connection controls',
    ],
    stackDescription:
      'Keep tunnel state close to the UI, whether your app uses a client router, server framework, or custom provider.',
    integrationHeading: 'Make tunnel state part of your\nReact app.',
    integrationDescription:
      'Wrap your application once, then use typed hooks for status, public URLs, reconnects, and lifecycle actions wherever the UI needs them.',
    technologies: [
      { label: 'React', icon: SiReact },
      typescript,
      { label: 'Next.js', icon: SiNextdotjs },
    ],
  },
  {
    id: 'vite',
    name: 'Vite',
    packageName: '@outpipe/vite-plugin',
    eyebrow: 'Vite integration',
    headline: 'Share your Vite app\nwithout extra config.',
    description:
      'The development server integration opens a tunnel when Vite is ready and keeps the local target aligned with the running server.',
    docsSlug: 'vite',
    install: 'npm install -D @outpipe/vite-plugin',
    fileName: 'vite.config.ts',
    code: "import { defineConfig } from 'vite';\nimport react from '@vitejs/plugin-react';\nimport outpipe from '@outpipe/vite-plugin';\n\nexport default defineConfig({\n  plugins: [react(), outpipe()],\n});",
    colorClass: 'text-indigo-300',
    icon: SiVite,
    features: [
      'Starts with the dev server',
      'Dynamic port awareness',
      'HMR-friendly lifecycle',
      'React, Vue, Svelte, and Solid',
    ],
    useCases: ['Design reviews', 'Webhook callbacks', 'Remote QA sessions'],
    stackDescription:
      'React, Vue, Svelte, Solid, Astro, and more can share the same Vite development workflow.',
    integrationHeading: 'Integrate with any\nVite application.',
    integrationDescription:
      'Whether you are building with React, Vue, Svelte, Solid, or Astro, the plugin starts with your dev server and keeps the public endpoint aligned.',
    technologies: [
      { label: 'React', icon: SiReact },
      { label: 'Vue', icon: SiVuedotjs },
      { label: 'Svelte', icon: SiSvelte },
      { label: 'Solid', icon: SiSolid },
      { label: 'Astro', icon: SiAstro },
      { label: 'Remix', icon: SiRemix },
    ],
  },
  {
    id: 'next',
    name: 'Next.js',
    packageName: '@outpipe/next',
    eyebrow: 'Next.js integration',
    headline: 'Put your Next.js app\nwithin reach.',
    description:
      'A lifecycle wrapper for Next.js development and server workflows, with the same tunnel controls as the CLI and SDK.',
    docsSlug: 'next',
    install: 'npm install @outpipe/next',
    fileName: 'next.config.ts',
    code: "import withOutpipe from '@outpipe/next';\n\nexport default withOutpipe({\n  reactStrictMode: true,\n});",
    colorClass: 'text-white',
    icon: SiNextdotjs,
    features: [
      'App and Pages Router support',
      'Development lifecycle hooks',
      'Server-friendly configuration',
      'Typed configuration',
    ],
    useCases: ['Preview deployments', 'OAuth callback testing', 'Client demos'],
    stackDescription:
      'Keep your framework, data layer, styling system, and deployment workflow exactly where they are.',
    integrationHeading: 'Give every Next.js preview\na public edge.',
    integrationDescription:
      'Add the wrapper to your existing Next.js configuration and share local pages, API routes, and OAuth callbacks without extra proxy setup.',
    technologies: [
      { label: 'Vercel', icon: SiVercel },
      { label: 'Prisma', icon: SiPrisma },
      { label: 'Tailwind', icon: SiTailwindcss },
      typescript,
      { label: 'tRPC', icon: SiTrpc },
      { label: 'Drizzle', icon: SiDrizzle },
    ],
  },
  {
    id: 'nest',
    name: 'NestJS',
    packageName: '@outpipe/nest',
    eyebrow: 'NestJS integration',
    headline: 'Expose your NestJS API\nin one call.',
    description:
      'A Nest module and service that make tunnel lifecycle part of your application bootstrap and shutdown flow.',
    docsSlug: 'nest',
    install: 'npm install @outpipe/nest',
    fileName: 'app.module.ts',
    code: "import { OutpipeModule } from '@outpipe/nest';\n\n@Module({\n  imports: [OutpipeModule.forRoot({ localPort: 3000 })],\n})\nexport class AppModule {}",
    colorClass: 'text-rose-300',
    icon: SiNestjs,
    features: [
      'Module-based setup',
      'Lifecycle-aware service',
      'Automatic shutdown cleanup',
      'TypeScript-first API',
    ],
    useCases: ['Webhook development', 'Team API previews', 'Staging callbacks'],
    stackDescription:
      'Pair NestJS with the databases, APIs, observability, and billing tools your service already depends on.',
    integrationHeading: 'One module. A reachable\nNestJS server.',
    integrationDescription:
      'Start your Nest application normally, then attach the tunnel to its lifecycle so local APIs and webhooks are available to the people testing them.',
    technologies: [
      { label: 'PostgreSQL', icon: SiPostgresql },
      { label: 'MongoDB', icon: SiMongodb },
      { label: 'Redis', icon: SiRedis },
      { label: 'GraphQL', icon: SiGraphql },
      { label: 'Stripe', icon: SiStripe },
      { label: 'Swagger', icon: SiSwagger },
    ],
  },
  {
    id: 'express',
    name: 'Express',
    packageName: '@outpipe/express',
    eyebrow: 'Express integration',
    headline: 'Put your Express server\non a public URL.',
    description:
      'A lightweight lifecycle wrapper for Express servers, designed for APIs, webhooks, and services that already own their HTTP process.',
    docsSlug: 'express',
    install: 'npm install @outpipe/express',
    fileName: 'server.ts',
    code: "import express from 'express';\nimport { outpipeTunnel } from '@outpipe/express';\n\nconst app = express();\noutpipeTunnel(app, { localPort: 3000 });\napp.listen(3000);",
    colorClass: 'text-amber-300',
    icon: SiExpress,
    features: [
      'Minimal middleware setup',
      'Idempotent lifecycle',
      'Status endpoint helpers',
      'Express 4 and 5 support',
    ],
    useCases: [
      'Webhook inspection',
      'Partner API demos',
      'Local service sharing',
    ],
    stackDescription:
      'Expose an existing Express process without replacing your middleware, data, or production conventions.',
    integrationHeading: 'Expose an Express server\nwithout rewiring it.',
    integrationDescription:
      'Keep your routes, middleware, and local process intact while Outpipe provides the public URL and connection lifecycle around it.',
    technologies: [
      { label: 'MongoDB', icon: SiMongodb },
      { label: 'PostgreSQL', icon: SiPostgresql },
      { label: 'Redis', icon: SiRedis },
      { label: 'Socket.IO', icon: SiSocketdotio },
      { label: 'Stripe', icon: SiStripe },
    ],
  },
];

export function getPluginDefinition(pluginId: string) {
  return pluginDefinitions.find((plugin) => plugin.id === pluginId);
}
