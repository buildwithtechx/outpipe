# Web deployment on Cloudflare

Connect the `outpipe` Worker to this repository with `main` as the production branch.
Keep the root directory at `/` so npm installs the complete workspace.

| Setting | Command |
| --- | --- |
| Build | `npm run build --workspace apps/web` |
| Production deploy | `npx wrangler deploy --cwd apps/web` |
| Preview deploy | `npx wrangler preview --cwd apps/web` |

Wrangler must run inside `apps/web` to find its configuration and the Vite-generated
deployment configuration. Running `npx wrangler preview` from the repository root
reports a missing `previews` block even when the web configuration includes it.

Keep `previews: {}` in `apps/web/wrangler.jsonc` to enable branch previews.
After changing dashboard commands, trigger a fresh build from the branch. Retrying
an earlier build can reuse the earlier command settings.
