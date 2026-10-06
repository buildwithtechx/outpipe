# Web deployment on Cloudflare

Connect the `outpipe` Worker to this repository with `main` as the production branch.
Keep the root directory at `/` so npm installs the complete workspace.

| Setting | Command |
| --- | --- |
| Build | `npm run build --workspace apps/web` |
| Production deploy | `npx wrangler deploy --cwd apps/web` |

Wrangler runs inside `apps/web` to use its configuration and the Vite-generated
deployment configuration.

Disable **Builds for Preview branches** in Cloudflare's **Previews Base** build
settings. Only the production branch deploys automatically.
