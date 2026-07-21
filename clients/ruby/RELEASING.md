# Releasing the `retab` gem

Routine for cutting a new release of the [`retab`](https://rubygems.org/gems/retab) Ruby gem.

The gem is **spec-driven**: most of `lib/retab/*.rb` and `rbi/retab/*.rbi` is regenerated from
`public/docs/api-reference/openapi.json` by the `retab-ruby-emitter` package (at
`factory/generators/oagen/retab-ruby-emitter/` in the parent monorepo — `public/sdk` is a
submodule, so that path is outside this repo), driven through Bazel. Only the hand-maintained runtime files survive regeneration — every file carrying an
`@oagen-ignore-file` marker (17 of them today: `base_client.rb`, `configuration.rb`, `errors.rb`,
`mime.rb`, `multipart.rb`, `paginated_list.rb`, `types/*.rb`, `util.rb`, `version.rb`, `Gemfile`,
`retab.gemspec`, and the hand-written tests).

Publishing is scripted end to end. You should not be running `sed`, `gem build`, or `gem push` by
hand — the publish script does version selection, build, verification, and push.

---

## One-time setup

### 1. Ruby toolchain

`.ruby-version` pins **3.4.9** and `publish_packages_ruby.sh` selects it via rbenv. System Ruby on
macOS is 2.6 and is too old (the gem needs ≥ 3.0 for `Data.define` and Zeitwerk 2.6+).

```bash
rbenv install 3.4.9   # skip if already installed
rbenv versions        # expect 3.4.9 listed
```

Inside `public/sdk/clients/ruby`, `.ruby-version` makes rbenv select 3.4.9 automatically. If
`ruby -v` reports 2.6 there, `.ruby-version` is missing — restore it before doing anything else.

### 2. rubygems.org account + MFA

The `retab` namespace requires multi-factor authentication on every push.

1. Sign in to <https://rubygems.org>.
2. **Edit Profile → Multi-factor authentication** → enable **TOTP**.
3. Generate an **API key** at <https://rubygems.org/profile/api_keys> with **Push rubygem** scope.

### 3. API key on disk

```bash
mkdir -p ~/.gem
printf -- "---\n:rubygems_api_key: rubygems_XXXX_paste_yours_here_XXXX\n" > ~/.gem/credentials
chmod 600 ~/.gem/credentials
```

`chmod 600` is mandatory — `gem` refuses to read the file otherwise. Never commit this file.

### 4. Bundler deps

```bash
cd public/sdk/clients/ruby
bundle install
```

---

## Per-release routine

### 1. Regenerate from the spec (only if the spec or emitter changed)

Generation is a Bazel-owned artifact surface; do not invoke the emitter by hand.

```bash
# What the surface is and which targets own it
skynet artifact sdk-ruby

# Regenerate the committed tree, then assert no drift
factory/skynet/bin/bazel.sh run  //public/sdk:generated_ruby_sdk_update
factory/skynet/bin/bazel.sh test //public/sdk:check_ruby
```

The update step rsyncs the generated tree over `clients/ruby/`. It cannot delete the non-generated
scaffolding: `factory/bazel/update_generated_sdk_tree.sh` builds rsync `protect` rules from
`factory/bazel/sdk_release_scaffolding.json` plus the committed `.oagen-manifest.json`, so an update
can only remove files the generator actually claims. `//factory/bazel:sdk_release_scaffolding_check`
asserts those files are present.

### 2. Run the test suite

```bash
cd public/sdk/clients/ruby
bundle exec ruby -Ilib -Itest test/test_mime_smoke.rb
bundle exec ruby -Ilib -Itest \
  -e 'Dir["test/**/*_test.rb", "test/**/test_*.rb"].uniq.sort.each { |f| require_relative f }'
```

Expected (2026-07-21):

- Mime smoke: `6 runs, 14 assertions, 0 failures, 0 errors, 0 skips`
- Full suite: `645 runs, 2532 assertions, 0 failures, 0 errors, 0 skips`

Counts grow as the spec grows; failures are what matter.

The publish script runs both suites itself and aborts on failure, so this step is a fast local
check rather than the gate.

| Symptom | Fix location |
|---|---|
| `Could not find <gem>` | `Gemfile` or `retab.gemspec` deps |
| `cannot load such file -- <stdlib>` (e.g. `base64`) | Add the gem to `retab.gemspec` (stdlib gems were extracted in Ruby 3.4+) |
| `wrong number of arguments` in `MimeData.new` | `lib/retab/mime.rb` — keep both positional-hash and kwarg forms working |
| `undefined method 'last_response='` on a model | Model class missing `attr_accessor :last_response` (inherits from `Retab::Types::BaseModel`) |
| `list methods are not in REGISTRY or NON_CURSOR` | A resource gained a `list`/`list_*` method — register it in `test/test_pagination_contract.rb` |

### 3. Brand check

Sanity-check that nothing slipped in from the WorkOS upstream emitter:

```bash
grep -rni "workos" lib/ rbi/ retab.gemspec | head   # expect: no output
grep "spec.name" retab.gemspec                      # expect: spec.name = 'retab'
```

### 4. Publish

```bash
export RUBYGEMS_OTP=<6-digit-code-from-authenticator>   # optional; prompted otherwise
cube publish_packages ruby
```

That dispatches to `tools/dev/scripts/subscripts/publish_packages/publish_packages_ruby.sh`, which:

1. Selects Ruby 3.4.9 via rbenv.
2. Runs the publish preflight — `ensure_publish_checkout_ready` (must be on a branch with an
   upstream, synced with the remote) and `ensure_release_scaffolding_present` (the release
   scaffolding registry check).
3. `bundle install`, then runs the mime smoke test and the full suite. Any failure aborts.
4. Picks the version: reads `spec.version` from `retab.gemspec`, queries rubygems for published
   `retab` versions, and takes the next free patch in that major.minor. **You do not bump by hand.**
   When it bumps, it writes both `retab.gemspec` and `lib/retab/version.rb`.
5. `gem build`, verifies with `gem specification`, and `gem push` (passing `--otp $RUBYGEMS_OTP`
   when set).

Useful flags: `--test false` skips the suites (use only when you just ran them).

Confirm the new version at <https://rubygems.org/gems/retab>.

### 5. Commit the version bump

The script leaves the bumped `retab.gemspec` and `lib/retab/version.rb` uncommitted by design — the
publish is the source of truth for what version exists. Commit them, or the next release's
clean-tree preflight fails.

```bash
git add public/sdk/clients/ruby/retab.gemspec public/sdk/clients/ruby/lib/retab/version.rb
```

`public/sdk` is a submodule: commit and push there first, then bump the parent pointer.

### 6. Tag the release

```bash
git tag retab-ruby-v<new-version>
git push origin retab-ruby-v<new-version>
```

Tag prefix is `retab-ruby-` so it doesn't collide with the other per-language release tags.

### 7. Clean up local artifacts

```bash
rm -f public/sdk/clients/ruby/retab-*.gem
```

The built `.gem` is not committed — `gem build` regenerates it.

---

## Hot-fixing a published release

Never re-push a published version; RubyGems rejects it by design. Fix the bug and re-run
`cube publish_packages ruby` — it picks the next free patch automatically.

If a release is critically broken (ships secrets, breaks consumer builds):

```bash
gem yank retab -v <bad-version> --otp <code>
```

Use sparingly — every yank breaks any `Gemfile.lock` pinned to that version.

---

## Troubleshooting

### `publish-preflight: branch <x> has no upstream` / checkout not clean

`ensure_publish_checkout_ready` refuses to publish from a detached HEAD, an unpushed branch, or a
tree whose HEAD differs from its upstream. Push the branch first.

### `publish-preflight: missing release scaffolding`

A file listed in `factory/bazel/sdk_release_scaffolding.json` under `clients.ruby.required` is
gone — most likely deleted by an SDK update. Restore it from git history and re-run. This is the
guard that exists because `Gemfile` and `retab.gemspec` were deleted this way four separate times.

### `Rubygem requires owners to enable MFA.`

Your rubygems.org account doesn't have MFA enabled. See **One-time setup → step 2**.

### `You have enabled multifactor authentication but no OTP code provided.`

Set `RUBYGEMS_OTP` to the current 6-digit code before invoking the publish script.

### `You don't have permission to push to this gem.`

Your API key is valid but you're not an owner of `retab`. An existing owner can add you:

```bash
gem owner retab --add <your-email>
```

### `gem build` fails with `<file> is not in the list of files`

The gemspec's `spec.files = Dir[...]` glob doesn't match a file the gem needs at runtime. Widen the
glob or move the file under one of the listed roots (`lib/`, `rbi/`, root for `README.md` /
`LICENSE`).

### `bundle install` fails with `Could not find compatible versions`

A transitive dep needs a newer Ruby. Check `ruby -v` (should be 3.4.9 via `.ruby-version`) against
`spec.required_ruby_version` in the gemspec.
