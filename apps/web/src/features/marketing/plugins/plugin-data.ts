import { Cable } from 'lucide-react';
import {
  SiAngular,
  SiAstro,
  SiDrizzle,
  SiExpress,
  SiGo,
  SiGraphql,
  SiMongodb,
  SiNestjs,
  SiNextdotjs,
  SiNodedotjs,
  SiPhp,
  SiPostgresql,
  SiPrisma,
  SiReact,
  SiRedis,
  SiRemix,
  SiRust,
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

export type PluginId =
  | 'sdk'
  | 'react'
  | 'vite'
  | 'next'
  | 'nest'
  | 'express'
  | 'go'
  | 'rust'
  | 'php'
  | 'angular';

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
  {
    id: 'go',
    name: 'Go SDK',
    packageName: 'github.com/buildwithtechx/outpipe/packages/go',
    eyebrow: 'Go SDK',
    headline: 'Connect Go services\nto Outpipe.',
    description:
      'API and relay helpers for Go services and command-line tools that manage tunnels and local connections.',
    docsSlug: 'go',
    install: 'go get github.com/buildwithtechx/outpipe/packages/go',
    fileName: 'main.go',
    code: 'api, err := client.New(client.Config{\n  BaseURL: "https://api.outpipe.dev",\n  APIKey: os.Getenv("OUTPIPE_API_KEY"),\n})\nif err != nil { log.Fatal(err) }\n\ntunnels, err := api.Tunnels(ctx, os.Getenv("OUTPIPE_ORG_ID"))',
    colorClass: 'text-cyan-300',
    icon: SiGo,
    features: [
      'Authenticated API client',
      'Relay helper',
      'Typed protocol',
      'Go services and CLI tools',
    ],
    useCases: ['Go service previews', 'Developer tooling', 'Tunnel automation'],
    stackDescription:
      'Use the Go module in services and tools that already manage their own processes and local ports.',
    integrationHeading: 'Manage tunnels from\nyour Go service.',
    integrationDescription:
      'Create an API client with your server-side key, or attach the relay helper to a long-running process.',
    technologies: [{ label: 'Go', icon: SiGo }],
  },
  {
    id: 'rust',
    name: 'Rust SDK',
    packageName: 'outpipe',
    eyebrow: 'Rust SDK',
    headline: 'Build reliable tunnels\nwith Rust.',
    description:
      'An async client and protocol types for Rust applications, services, and developer tooling.',
    docsSlug: 'rust',
    install: 'cargo add outpipe',
    fileName: 'main.rs',
    code: 'let api_key = std::env::var("OUTPIPE_API_KEY")?;\nlet client = Client::builder("https://api.outpipe.dev")\n    .api_key(api_key)\n    .build()?;\n\nlet tunnels = client.tunnels("org_123").await?;',
    colorClass: 'text-orange-300',
    icon: SiRust,
    features: [
      'Async API client',
      'Typed protocol',
      'Tunnel operations',
      'Rust services and tools',
    ],
    useCases: ['Service automation', 'Custom daemons', 'Tunnel inspection'],
    stackDescription:
      'Keep Outpipe operations in the same async Rust runtime as the rest of your service.',
    integrationHeading: 'Bring Outpipe into\nyour Rust runtime.',
    integrationDescription:
      'Use the async client to inspect and manage tunnels, with protocol types available for custom relay tooling.',
    technologies: [{ label: 'Rust', icon: SiRust }],
  },
  {
    id: 'php',
    name: 'PHP SDK',
    packageName: 'outpipe/outpipe-php',
    eyebrow: 'PHP SDK',
    headline: 'Connect PHP apps\nto Outpipe.',
    description:
      'A Composer client for PHP and Laravel applications that need authenticated Outpipe API operations.',
    docsSlug: 'php',
    install: 'composer require outpipe/outpipe-php',
    fileName: 'app.php',
    code: "use Outpipe\\Client\\OutpipeClient;\n\n$client = new OutpipeClient(\n    'https://api.outpipe.dev',\n    getenv('OUTPIPE_API_KEY'),\n);\n\n$tunnels = $client->tunnels(getenv('OUTPIPE_ORG_ID'));",
    colorClass: 'text-indigo-300',
    icon: SiPhp,
    features: [
      'Composer package',
      'Authenticated API client',
      'Laravel-friendly setup',
      'Tunnel operations',
    ],
    useCases: [
      'Laravel integrations',
      'Webhook tooling',
      'PHP service dashboards',
    ],
    stackDescription:
      'Use the Composer package alongside your existing PHP application and service container.',
    integrationHeading: 'Manage tunnels from\nyour PHP app.',
    integrationDescription:
      'Create a client at your application boundary and keep its API key on the server.',
    technologies: [{ label: 'PHP', icon: SiPhp }],
  },
  {
    id: 'angular',
    name: 'Angular',
    packageName: '@outpipe/angular',
    eyebrow: 'Angular integration',
    headline: 'Add Outpipe to\nyour Angular app.',
    description:
      'Standalone providers and an injectable API service for Angular applications that manage tunnels.',
    docsSlug: 'angular',
    install: 'npm install @outpipe/angular',
    fileName: 'app.config.ts',
    code: "import { Component, inject } from '@angular/core';\nimport { OutpipeApiService } from '@outpipe/angular';\n\n@Component({ selector: 'app-tunnels', template: '' })\nexport class TunnelsComponent {\n  private outpipe = inject(OutpipeApiService);\n  tunnels = this.outpipe.listTunnels('organization-id');\n}",
    colorClass: 'text-red-300',
    icon: SiAngular,
    features: [
      'Standalone provider',
      'Injectable API service',
      'Angular-first setup',
      'Typed tunnel operations',
    ],
    useCases: ['Developer dashboards', 'Tunnel controls', 'Internal tools'],
    stackDescription:
      'Register the provider with your Angular application and inject the service where tunnel data is needed.',
    integrationHeading: 'Use Outpipe in\nyour Angular app.',
    integrationDescription:
      'Set up the provider once, then use the injectable service for tunnel operations throughout your app.',
    technologies: [{ label: 'Angular', icon: SiAngular }, typescript],
  },
];

export function getPluginDefinition(pluginId: string) {
  return pluginDefinitions.find((plugin) => plugin.id === pluginId);
}
