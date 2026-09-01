import { SiAngular, SiGo, SiPhp, SiRust, SiTypescript } from 'react-icons/si';

export const sdkDefinitions = [
  {
    name: 'TypeScript',
    packageName: '@outpipe/sdk',
    description: 'A typed foundation for browser, Node.js, and CI tooling.',
    install: 'npm install @outpipe/sdk',
    docsSlug: 'sdk',
    icon: SiTypescript,
    color: 'text-sky-300',
  },
  {
    name: 'Go',
    packageName: 'outpipe.dev/outpipe-go',
    description: 'API and relay helpers for services and command-line tools.',
    install: 'go get outpipe.dev/outpipe-go',
    docsSlug: 'go',
    icon: SiGo,
    color: 'text-cyan-300',
  },
  {
    name: 'Rust',
    packageName: 'outpipe',
    description: 'Async client and protocol types for dependable Rust systems.',
    install: 'cargo add outpipe',
    docsSlug: 'rust',
    icon: SiRust,
    color: 'text-orange-300',
  },
  {
    name: 'PHP',
    packageName: 'outpipe/outpipe-php',
    description: 'A Composer client for Laravel and server-side PHP apps.',
    install: 'composer require outpipe/outpipe-php',
    docsSlug: 'php',
    icon: SiPhp,
    color: 'text-indigo-300',
  },
  {
    name: 'Angular',
    packageName: '@outpipe/angular',
    description: 'Standalone providers and an injectable API service.',
    install: 'npm install @outpipe/angular',
    docsSlug: 'angular',
    icon: SiAngular,
    color: 'text-red-300',
  },
] as const;
