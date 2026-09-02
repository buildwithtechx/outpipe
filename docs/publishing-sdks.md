# Publishing SDKs

The SDKs use their native registries. Run the compatibility and security
checks before creating a release tag.

## TypeScript and Angular

The `packages-v*` tag starts `.github/workflows/publish-packages.yml`.
Publishing uses npm Trusted Publishing through GitHub Actions OIDC; it does
not use an `NPM_TOKEN` secret. Configure the same repository and workflow as a
trusted publisher for every public npm package:

- Repository: `buildwithtechx/outpipe`
- Workflow: `publish-packages.yml`
- Environment: none

## Rust

The `rust-v*` tag starts `.github/workflows/publish-rust.yml`. Add a repository
secret named `CRATES_IO_TOKEN` containing a crates.io publish token. The job
runs formatting, tests, and `cargo publish --dry-run` before publishing
`packages/rust` to crates.io.

## Go

Go modules are published by Git tags rather than uploaded to a registry. The
Go SDK module path is `github.com/buildwithtechx/outpipe/packages/go`.

For version `v0.1.0`, create and push this tag:

```sh
git tag packages/go/v0.1.0
git push origin packages/go/v0.1.0
GOPROXY=proxy.golang.org go list -m github.com/buildwithtechx/outpipe/packages/go@v0.1.0
```

The tag must include `packages/go/` because the module is in a repository
subdirectory.

## PHP and Laravel

The repository root now contains the Composer manifest for
`outpipe/outpipe-php`, with its autoload paths mapped to `packages/php`. Submit
`https://github.com/buildwithtechx/outpipe` once at
`https://packagist.org/packages/submit`, then enable the Packagist GitHub
webhook for push events. Packagist will discover new Composer tags
automatically.

Public Packagist requires `composer.json` at the repository root; it does not
publish packages that exist only in a repository subdirectory.

Before creating a PHP release tag, run:

```sh
cd packages/php
composer validate --strict --no-check-publish
composer test
composer check
```

## Release checklist

1. Update the SDK version and changelog entry.
2. Run the repository compatibility, security, and conformance checks.
3. Publish TypeScript packages with `packages-v*`.
4. Publish Rust with `rust-v*`.
5. Publish Go with `packages/go/v*`.
6. Push the PHP tag after Packagist synchronization is enabled.
