# Release & Publishing Guide

Outpipe is a monorepo containing the core relay server, CLI, desktop applications, and native SDKs across multiple programming languages. Each target is published via dedicated GitHub Actions workflows using specific Git tag conventions.

---

## Tag Conventions & CI Workflows

| Target / Ecosystem                               | Tag Pattern      | Example Tag          | Workflow File                            | Destination                       |
| :----------------------------------------------- | :--------------- | :------------------- | :--------------------------------------- | :-------------------------------- |
| **Full Platform Release** (CLI & Desktop)        | `v*`             | `v0.1.0`             | `.github/workflows/release.yml`          | GitHub Releases, SBOM, SHA256SUMS |
| **npm Packages** (`@outpipe/sdk`, `react`, etc.) | `packages-v*`    | `packages-v0.1.0`    | `.github/workflows/publish-packages.yml` | npm Registry (OIDC Provenance)    |
| **Go SDK Module** (`packages/go`)                | `packages/go/v*` | `packages/go/v0.1.0` | `.github/workflows/publish-go.yml`       | `proxy.golang.org`                |
| **Rust SDK Crate** (`packages/rust`)             | `rust-v*`        | `rust-v0.1.0`        | `.github/workflows/publish-rust.yml`     | crates.io                         |
| **PHP SDK Package** (`packages/php`)             | `v*`             | `v0.1.0`             | Packagist Webhook                        | Packagist                         |

---

## 1. Platform Release (CLI & Desktop Apps)

Creating and pushing a standard `v*` tag triggers `.github/workflows/release.yml`:

```bash
VERSION="0.1.0"
git tag "v${VERSION}"
git push origin "v${VERSION}"
```

### Artifacts built by this workflow

1. **Server & Daemon Binaries**: `outpipe-api`, `outpipe`, `outpipe-cron`, `outpipe-check` for Linux (`amd64`).
2. **CLI Binaries**: `outpipe-cli` for Linux (`amd64`), macOS (`amd64`, `arm64`), and Windows (`amd64`).
3. **Desktop Installers (Tauri)**:
   - Debian / Ubuntu (`.deb`)
   - macOS (`.dmg`)
   - Windows (`.exe` / NSIS)
4. **Security & Integrity**:
   - SHA-256 checksums (`SHA256SUMS`)
   - Software Bill of Materials in SPDX format (`outpipe-source.spdx.json`)
   - Automatic GitHub Release notes with downloadable assets.

---

## 2. npm Packages (`@outpipe/*`)

To publish TypeScript/JavaScript packages (`@outpipe/sdk`, `@outpipe/react`, `@outpipe/next`, `@outpipe/nest`, `@outpipe/express`, `@outpipe/vite-plugin`, `@outpipe/angular`):

```bash
VERSION="0.1.0"
git tag "packages-v${VERSION}"
git push origin "packages-v${VERSION}"
```

### npm trusted publishing mechanism

- Uses **npm Trusted Publishing** via GitHub Actions OIDC (`--provenance`).
- No long-lived `NPM_TOKEN` secret is needed.
- Configure the repository and workflow as a trusted publisher on npmjs.com for every public `@outpipe/*` package:
  - Publisher: GitHub Actions
  - Repository: `buildwithtechx/outpipe`
  - Workflow: `publish-packages.yml`
  - Environment: none
- The workflow builds all workspace packages (`npm run build:packages`), validates dry-run package packs, and publishes each workspace.

---

## 3. Go SDK Module

Because the Go SDK is located in a monorepo subdirectory (`packages/go`), the Go toolchain requires the tag to include the relative directory prefix:

```bash
VERSION="0.1.0"
git tag "packages/go/v${VERSION}"
git push origin "packages/go/v${VERSION}"
```

### Go publishing process

- Runs `go vet ./...` and `go test ./...` in `packages/go`.
- Requests immediate module indexing on `proxy.golang.org`:

  ```bash
  GOPROXY=proxy.golang.org go list -m "github.com/buildwithtechx/outpipe/packages/go@v0.1.0"
  ```

- Allows users to install with: `go get github.com/buildwithtechx/outpipe/packages/go@v0.1.0`.

---

## 4. Rust SDK Crate

To publish the `outpipe` crate to crates.io:

```bash
VERSION="0.1.0"
git tag "rust-v${VERSION}"
git push origin "rust-v${VERSION}"
```

### Rust publishing process

- Requires repository secret `CRATES_IO_TOKEN` containing a crates.io publish token.
- Runs formatting (`cargo fmt --check`), tests (`cargo test`), dry-run validation (`cargo publish --dry-run`), and finally publishes to crates.io.

---

## 5. PHP SDK Package

The root `composer.json` maps autoloading for `outpipe/outpipe-php` to `packages/php`.

- **Initial Setup**: Submit `https://github.com/buildwithtechx/outpipe` once at `https://packagist.org/packages/submit`, then enable the Packagist GitHub webhook for push events. Packagist will discover new tags automatically.
- When a `v*` tag is pushed to GitHub, Packagist indexes the new release.
- Pre-release verification:

  ```bash
  cd packages/php
  composer validate --strict --no-check-publish
  composer test
  composer check
  ```

---

## Complete Release Checklist (One-Shot Publish)

To release a coordinated version across all platforms and package registries simultaneously:

```bash
VERSION="0.1.0"

# 1. Run full local checks
npm run fmt
npm run typecheck
go test -race ./...
npm run test:go-sdk
npm run test:rust-sdk

# 2. Tag all release targets
git tag "v${VERSION}"
git tag "packages-v${VERSION}"
git tag "packages/go/v${VERSION}"
git tag "rust-v${VERSION}"

# 3. Push all tags to GitHub
git push origin "v${VERSION}" "packages-v${VERSION}" "packages/go/v${VERSION}" "rust-v${VERSION}"
```

---

## Manual Workflow Dispatch

All package publishing workflows (`publish-packages.yml`, `publish-go.yml`, `publish-rust.yml`) can also be dispatched manually without creating a tag:

1. Navigate to **GitHub Repository** -> **Actions**.
2. Select the target workflow from the sidebar.
3. Click **Run workflow** -> Select `main` or target branch -> Click **Run workflow**.
