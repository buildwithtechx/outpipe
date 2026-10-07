# Outpipe Web

Vite + React + TanStack Router single-page app for the Outpipe dashboard,
marketing site, and docs. It talks to the Outpipe API over HTTP; there is no
server-side rendering.

## Getting Started

```bash
npm install
npm run dev
```

## Production

```bash
npm run build
```

This emits static assets to `dist/`, served as a single-page app. `wrangler.jsonc`
points at `./dist` with SPA fallback handling, so client routes resolve to
`index.html`.

```bash
npm run deploy
```

## Conventions

- Routes live in `src/routes/` and are generated into `src/routeTree.gen.ts`
  via `npm run generate-routes` (TanStack Router Vite plugin + `tsr`).
- Path alias `#/*` maps to `src/*`.
- Client environment variables live in `src/env.ts` and use the `VITE_` prefix.
- Shared query state goes through the singleton `queryClient` in
  `src/integrations/tanstack-query/root-provider.tsx`.
- Docs content lives in `content/docs/` and is bundled client-side with
  `fumadocs-mdx`; the docs route loader reads pages directly, no server
  functions involved.

## Env

Copy `.env.example` to `.env` and set `VITE_OUTPIPE_API_BASE_URL` to the API
origin. PostHog is optional via `VITE_OUTPIPE_POSTHOG_KEY` /
`VITE_OUTPIPE_POSTHOG_HOST`.
