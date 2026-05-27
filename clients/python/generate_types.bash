#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Execute as a module so stdlib imports like `types` do not get shadowed by
# `retab/types` when Python initializes `sys.path`.
cd "$script_dir"
python3 -m retab.generate_types > ../node/src/generated_types.ts
