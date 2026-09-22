#!/usr/bin/env bash

set -euo pipefail

mapfile -d '' go_files < <(find . -type f -name '*.go' \
  -not -path './node_modules/*' -not -path './.git/*' -print0)
if ((${#go_files[@]})); then
  gofmt -w "${go_files[@]}"
fi

if [[ -f packages/rust/Cargo.toml ]] && command -v cargo >/dev/null 2>&1; then
  cargo fmt --manifest-path packages/rust/Cargo.toml --all
fi

if [[ -f packages/php/composer.json ]]; then
  if [[ -x packages/php/vendor/bin/php-cs-fixer ]]; then
    packages/php/vendor/bin/php-cs-fixer fix packages/php/src packages/php/tests
  elif command -v php-cs-fixer >/dev/null 2>&1; then
    php-cs-fixer fix packages/php/src packages/php/tests
  else
    mapfile -d '' php_files < <(find packages/php/src packages/php/tests -type f -name '*.php' -print0)
    for php_file in "${php_files[@]}"; do
      php -l "$php_file" >/dev/null
    done
    printf '%s\n' 'PHP formatter unavailable; syntax checked PHP files instead.' >&2
  fi
fi

npm run format:fix
