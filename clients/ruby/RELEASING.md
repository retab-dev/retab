# Releasing the `retab` gem

Step-by-step routine for cutting a new release of the [`retab`](https://rubygems.org/gems/retab) Ruby gem.

The gem is **spec-driven**: most of `lib/retab/*.rb` and `rbi/retab/*.rbi` is regenerated from `open-source/docs/api-reference/openapi.json` via the [`retab-ruby-emitter`](../../../.oagen-workspace/retab-ruby-emitter/) workspace. Only a handful of hand-maintained runtime files survive regeneration — every file carrying an `@oagen-ignore-file` marker (mime.rb, gemspec, base_client.rb, errors.rb, types/, util.rb, version.rb, test_helper.rb, smoke test).

---

## One-time setup

You only do this once per machine.

### 1. Ruby toolchain

System Ruby on macOS is 2.6 — too old. The gem requires Ruby ≥ 3.0 (uses `Data.define` and Zeitwerk 2.6+).

```bash
brew install ruby
export PATH="/opt/homebrew/opt/ruby/bin:$PATH"
ruby --version   # expect 3.x or 4.x
```

Persist the PATH change by appending to `~/.zshrc`:

```bash
echo 'export PATH="/opt/homebrew/opt/ruby/bin:$PATH"' >> ~/.zshrc
```

### 2. rubygems.org account + MFA

The `retab` namespace requires multi-factor authentication on every push.

1. Sign in to <https://rubygems.org>.
2. **Edit Profile → Multi-factor authentication** → enable **TOTP** (authenticator app).
3. Generate an **API key** at <https://rubygems.org/profile/api_keys> with **Push rubygem** scope. Save the key string.

### 3. API key on disk

```bash
mkdir -p ~/.gem
printf -- "---\n:rubygems_api_key: rubygems_XXXX_paste_yours_here_XXXX\n" > ~/.gem/credentials
chmod 600 ~/.gem/credentials
```

`chmod 600` is mandatory — `gem` refuses to read the file otherwise.

### 4. Bundler deps for the gem

```bash
cd open-source/sdk/clients/ruby
bundle config set --local path '.bundle'
bundle install
```

---

## Per-release routine

Everything below assumes you're in the repo root at `/Users/sachaichbiah/Local/retab` (or your local clone equivalent).

### 1. Pick the new version

`retab.gemspec` carries `spec.version = '<n.n.n>'`. The gemspec is `@oagen-ignore-file`, so edit it directly — regeneration won't overwrite it.

```bash
# Example: 0.1.0 → 0.1.1
sed -i '' "s/spec.version = '0.1.0'/spec.version = '0.1.1'/" \
  open-source/sdk/clients/ruby/retab.gemspec
```

For semver discipline:

| Change type | Bump |
|---|---|
| Spec-driven regen only, no shape changes | patch (`0.1.0 → 0.1.1`) |
| New endpoints, additional optional fields | minor (`0.1.0 → 0.2.0`) |
| Removed/renamed fields, breaking signature changes | major (`0.1.0 → 1.0.0`) |

### 2. Regenerate from the spec

Only if the OpenAPI spec or the emitter changed since the last release. Otherwise skip to step 3.

```bash
cd .oagen-workspace/retab-gen
node --experimental-strip-types --no-warnings \
  ../oagen/dist/cli/index.mjs generate \
  --config ruby-only.config.ts \
  --spec ../../open-source/docs/api-reference/openapi.json \
  --lang ruby --namespace retab \
  --output ../../open-source/sdk/clients/ruby
```

Expected output:

```
Ignored 3 files (@oagen-ignore-file)
Generated 666 files in ../../open-source/sdk/clients/ruby
```

`Ignored 3` confirms the hand-maintained `mime.rb`, `test_mime_smoke.rb`, and `retab.gemspec` survived. If the count is wrong, investigate before continuing.

### 3. Run the test suite

```bash
cd open-source/sdk/clients/ruby
bundle exec ruby -Ilib -Itest test/test_mime_smoke.rb
bundle exec ruby -Ilib -Itest -e "Dir['test/**/test_*.rb'].each { |f| require File.expand_path(f) }"
```

Expected:

- Mime smoke: `5 runs, 11 assertions, 0 failures, 0 errors, 0 skips`
- Full suite: `458 runs, 1844+ assertions, 0 failures, 0 errors, 0 skips`

If either fails, stop and fix before publishing. Common breakages and where the fix lives:

| Symptom | Fix location |
|---|---|
| `Could not find <gem>` | `Gemfile` or `retab.gemspec` deps |
| `cannot load such file -- <stdlib>` (e.g. `base64`) | Add the gem to `retab.gemspec` (stdlib gems were extracted in Ruby 3.4+) |
| `wrong number of arguments` in `MimeData.new` | `lib/retab/mime.rb` — keep both positional-hash and kwarg forms working |
| `undefined method 'last_response='` on a model | Model class missing `attr_accessor :last_response` (inherits from `Retab::Types::BaseModel`) |

### 4. Brand check

Sanity-check that nothing slipped in from the WorkOS upstream emitter:

```bash
grep -rni "workos" lib/ rbi/ *.gemspec 2>/dev/null | head
# expect: no output
grep -rn "^module Retab" lib/ | wc -l
# expect: hundreds — every spec-derived .rb file
grep "spec.name" retab.gemspec
# expect: spec.name = 'retab'
```

### 5. Build the gem

```bash
cd open-source/sdk/clients/ruby
gem build retab.gemspec
```

Expected:

```
Successfully built RubyGem
  Name: retab
  Version: <new-version>
  File: retab-<new-version>.gem
```

### 6. Push to rubygems.org

```bash
gem push retab-<new-version>.gem --otp <6-digit-code-from-authenticator>
```

The OTP code expires every ~30 seconds. If `gem push` says "Please enter OTP code" with no `--otp` flag, your account is in **UI + API MFA** mode and you must pass it explicitly.

Expected:

```
Pushing gem to https://rubygems.org...
Successfully registered gem: retab (<new-version>)
```

Visit <https://rubygems.org/gems/retab> to confirm the new version appears.

### 7. Tag the release

```bash
git tag retab-ruby-v<new-version>
git push origin retab-ruby-v<new-version>
```

Tag prefix is `retab-ruby-` so it doesn't collide with future Python/Node/Go/Rust release tags from the same monorepo.

### 8. Clean up local artifacts

```bash
rm open-source/sdk/clients/ruby/retab-*.gem
```

The built `.gem` file is not committed — it's regenerable from `gem build` at any time.

---

## Hot-fixing a published release

If a published version has a bug, **never re-push the same version**. RubyGems rejects re-pushes by design.

1. Bump the patch number in `retab.gemspec` (`0.1.0 → 0.1.1`).
2. Fix the bug.
3. Run the full per-release routine from step 3.

If a release is critically broken (e.g. ships secrets, breaks consumer builds):

```bash
gem yank retab -v <bad-version> --otp <code>
```

`yank` deletes the version from rubygems.org. Use sparingly — every yank breaks any `Gemfile.lock` that pinned to that version.

---

## Troubleshooting

### `Rubygem requires owners to enable MFA. You must enable MFA before pushing new version.`

Your rubygems.org account doesn't have MFA enabled. See **One-time setup → step 2**.

### `You have enabled multifactor authentication but no OTP code provided.`

Add `--otp <code>` to the `gem push` invocation. Code is the current 6-digit value from your authenticator app.

### `You don't have permission to push to this gem.`

Your API key is valid but you're not a listed owner of the `retab` gem. Existing owners can add you:

```bash
gem owner retab --add <your-email>
```

### `gem build` fails with `<file> is not in the list of files`

The gemspec's `spec.files = Dir[...]` glob doesn't match a file the gem needs at runtime. Either widen the glob or add the file under one of the listed roots (`lib/`, `rbi/`, root for `README.md` / `LICENSE`).

### `bundle install` fails with `Could not find compatible versions`

A transitive dep needs a newer Ruby. Check `ruby --version` against `spec.required_ruby_version` in the gemspec.

---

## Quick reference

```bash
# Full release in one block — fill in <version> and <otp>
export PATH="/opt/homebrew/opt/ruby/bin:$PATH"
cd /Users/sachaichbiah/Local/retab/open-source/sdk/clients/ruby

# 1. Bump version
sed -i '' "s/spec.version = '.*'/spec.version = '<version>'/" retab.gemspec

# 2. Regenerate (skip if no spec/emitter changes)
( cd ../../../../.oagen-workspace/retab-gen && \
  node --experimental-strip-types --no-warnings \
  ../oagen/dist/cli/index.mjs generate \
  --config ruby-only.config.ts \
  --spec ../../open-source/docs/api-reference/openapi.json \
  --lang ruby --namespace retab \
  --output ../../open-source/sdk/clients/ruby )

# 3. Test
bundle install
bundle exec ruby -Ilib -Itest test/test_mime_smoke.rb
bundle exec ruby -Ilib -Itest -e "Dir['test/**/test_*.rb'].each { |f| require File.expand_path(f) }"

# 4. Build + push
gem build retab.gemspec
gem push retab-<version>.gem --otp <otp>

# 5. Tag
cd ../../../..
git tag retab-ruby-v<version>
git push origin retab-ruby-v<version>
```
