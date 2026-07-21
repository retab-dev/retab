# frozen_string_literal: true

# Cross-language pagination contract regression test for the Ruby SDK.
#
# Mirrors clients/python/tests/test_pagination_contract.py,
# clients/node/tests/pagination-contract.test.ts,
# clients/go/pagination_contract_test.go,
# clients/php/tests/PaginationContractTest.php,
# clients/dotnet/tests/PaginationContractTests.cs, and
# clients/rust/tests/pagination_contract.rs.
#
# What this test enforces:
#
#   1. Every cursor-paginated list method on a service exposed off
#      `Retab::Client` returns a `Retab::PaginatedList`.
#   2. That `PaginatedList` has its fetch closure wired up, so iterating
#      with `#each` walks every page lazily (not just the first one).
#   3. The registry below covers every list method the SDK ships — a new
#      resource that adds `def list` (or a `def list_*` variant) without
#      registering itself here fails CI, and a registry row whose method
#      disappeared fails too. NON_CURSOR is the explicit allowlist for
#      list methods that legitimately return a non-cursor envelope.
#
# "List method" here means `def list` or any `def list_*` on a resource.
# The `list_*` variants carry the same contract: `usage.list_primitives`,
# `workflows.list_versions` and friends all return `PaginatedList`.

require "test_helper"
require "set"

class PaginationContractTest < Minitest::Test
  # Resource registry: every list method the SDK exposes off
  # `Retab::Client.<service>...`. Format:
  #   [service_label, list_path, sample_item, invoke_proc]
  #
  # `service` is the discovery label: the dotted accessor path for a bare
  # `list`, and the accessor path plus method name for a `list_*` variant.
  # `list_path` is the URL the SDK hits — used to scope the WebMock stub
  # so each row covers exactly one route. `sample_item` is a minimal JSON
  # blob that the resource model accepts. `invoke_proc` calls the list
  # method with the smallest valid argument set.
  REGISTRY = [
    {
      service: :classifications,
      path: "/v1/classifications",
      sample: "{}",
      invoke: -> (c) { c.classifications.list }
    },
    {
      service: "edits.templates",
      path: "/v1/edits/templates",
      sample: "{}",
      invoke: -> (c) { c.edits.templates.list }
    },
    {
      service: :edits,
      path: "/v1/edits",
      sample: "{}",
      invoke: -> (c) { c.edits.list }
    },
    {
      service: "workflows.experiments.results",
      path: "/v1/workflows/experiments/results",
      sample: "{}",
      invoke: -> (c) { c.workflows.experiments.results.list(run_id: "run_x") }
    },
    {
      service: "workflows.experiments.runs",
      path: "/v1/workflows/experiments/runs",
      sample: "{}",
      invoke: -> (c) { c.workflows.experiments.runs.list }
    },
    {
      service: :extractions,
      path: "/v1/extractions",
      sample: "{}",
      invoke: -> (c) { c.extractions.list }
    },
    {
      service: :files,
      path: "/v1/files",
      sample: "{}",
      invoke: -> (c) { c.files.list }
    },
    {
      service: :parses,
      path: "/v1/parses",
      sample: "{}",
      invoke: -> (c) { c.parses.list }
    },
    {
      service: :partitions,
      path: "/v1/partitions",
      sample: "{}",
      invoke: -> (c) { c.partitions.list }
    },
    {
      service: :splits,
      path: "/v1/splits",
      sample: "{}",
      invoke: -> (c) { c.splits.list }
    },
    {
      service: "workflows.artifacts",
      path: "/v1/workflows/artifacts",
      sample: "{}",
      invoke: -> (c) { c.workflows.artifacts.list(run_id: "run_x") }
    },
    {
      service: "workflows.blocks.executions",
      path: "/v1/workflows/blocks/executions",
      sample: "{}",
      invoke: -> (c) { c.workflows.blocks.executions.list(run_id: "run_x", block_id: "blk_x") }
    },
    {
      service: "workflows.blocks",
      path: "/v1/workflows/blocks",
      sample: "{}",
      invoke: -> (c) { c.workflows.blocks.list(workflow_id: "wf_x") }
    },
    {
      service: "workflows.blocks.list_versions",
      path: "/v1/workflows/blocks/versions",
      sample: "{}",
      invoke: -> (c) { c.workflows.blocks.list_versions(workflow_id: "wf_x") }
    },
    {
      service: "workflows.edges",
      path: "/v1/workflows/edges",
      sample: "{}",
      invoke: -> (c) { c.workflows.edges.list(workflow_id: "wf_x") }
    },
    {
      service: "workflows.edges.list_versions",
      path: "/v1/workflows/edges/versions",
      sample: "{}",
      invoke: -> (c) { c.workflows.edges.list_versions(workflow_id: "wf_x") }
    },
    {
      service: "workflows.experiments",
      path: "/v1/workflows/experiments",
      sample: "{}",
      invoke: -> (c) { c.workflows.experiments.list(workflow_id: "wf_x") }
    },
    {
      service: "workflows.reviews.versions",
      path: "/v1/workflows/reviews/versions",
      sample: "{}",
      invoke: -> (c) { c.workflows.reviews.versions.list(review_id: "rev_x") }
    },
    {
      service: "workflows.reviews",
      path: "/v1/workflows/reviews",
      sample: "{}",
      invoke: -> (c) { c.workflows.reviews.list }
    },
    {
      service: "workflows.runs",
      path: "/v1/workflows/runs",
      sample: "{}",
      invoke: -> (c) { c.workflows.runs.list }
    },
    {
      service: "workflows.steps",
      path: "/v1/workflows/steps",
      sample: "{}",
      invoke: -> (c) { c.workflows.steps.list }
    },
    {
      service: "workflows.evals.results",
      path: "/v1/workflows/evals/results",
      sample: "{}",
      invoke: -> (c) { c.workflows.evals.results.list(run_id: "run_x") }
    },
    {
      service: "workflows.evals.runs",
      path: "/v1/workflows/evals/runs",
      sample: "{}",
      invoke: -> (c) { c.workflows.evals.runs.list }
    },
    {
      service: "workflows.evals",
      path: "/v1/workflows/evals",
      sample: "{}",
      invoke: -> (c) { c.workflows.evals.list(workflow_id: "wf_x") }
    },
    {
      service: "workflows.list_versions",
      path: "/v1/workflows/versions",
      sample: "{}",
      invoke: -> (c) { c.workflows.list_versions(workflow_id: "wf_x") }
    },
    {
      service: :workflows,
      path: "/v1/workflows",
      sample: "{}",
      invoke: -> (c) { c.workflows.list }
    },
    {
      service: "usage.list_blocks",
      path: "/v1/usage/blocks",
      sample: "{}",
      invoke: -> (c) { c.usage.list_blocks }
    },
    {
      service: "usage.list_primitives",
      path: "/v1/usage/primitives",
      sample: "{}",
      invoke: -> (c) { c.usage.list_primitives }
    },
    {
      service: "usage.list_runs",
      path: "/v1/usage/runs",
      sample: "{}",
      invoke: -> (c) { c.usage.list_runs }
    }
  ].freeze

  # List methods that legitimately return a non-cursor envelope, so the
  # PaginatedList contract does not apply:
  #   * tables.list                 -> {tables: [...]}, no cursor envelope
  #   * secrets.list_secrets        -> Retab::SecretListResponse, unpaginated
  #   * secrets.list_secret_value   -> Retab::SecretValueResponse; "list" is
  #                                    the API verb here, not a collection
  #   * *.list_diff                 -> a single *VersionDiff object
  # If you add an entry here, document the reason on the same line.
  NON_CURSOR = [
    "tables",
    "secrets.list_secrets",
    "secrets.list_secret_value",
    "workflows.list_diff",
    "workflows.blocks.list_diff",
    "workflows.edges.list_diff"
  ].freeze

  def setup
    @client = Retab::Client.new(api_key: "sk_test_123")
  end

  # --- Discovery test ------------------------------------------------------
  #
  # Walk every service accessor on `Retab::Client` and find every `list` /
  # `list_*` method. Assert that each one appears in REGISTRY or in
  # NON_CURSOR, and that every REGISTRY row still matches a real method. A
  # new service that adds a list method without registering itself fails
  # the first assertion; a row left behind after a rename fails the second.

  def test_registry_covers_every_list_method
    registered = REGISTRY.map { |row| row[:service].to_s }.to_set
    known = registered + NON_CURSOR.to_set
    discovered = collect_list_methods(@client).to_set

    missing = (discovered - known).to_a.sort
    assert_empty(
      missing,
      "These list methods are not in REGISTRY or NON_CURSOR: #{missing.inspect}. " \
        "Add them to test/test_pagination_contract.rb."
    )

    stale = (registered - discovered).to_a.sort
    assert_empty(
      stale,
      "These REGISTRY rows no longer match a list method on the client: #{stale.inspect}. " \
        "Remove or rename them in test/test_pagination_contract.rb."
    )
  end

  def collect_list_methods(root)
    out = []
    visited = Set.new
    walk_resources(root, [], out, visited)
    out
  end

  def walk_resources(resource, path, out, visited)
    return if visited.include?(resource.object_id)
    visited << resource.object_id

    if path.any?
      resource.public_methods(false).each do |meth|
        name = meth.to_s
        next unless list_method?(name)

        out << (name == "list" ? path.join(".") : [*path, name].join("."))
      end
    end

    resource.public_methods(false).sort.each do |meth|
      next if list_method?(meth.to_s) || meth.to_s.start_with?("_")
      method_object = resource.method(meth)
      next unless method_object.parameters.empty?

      child = resource.public_send(meth)
      next unless child.class.name&.start_with?("Retab::")

      walk_resources(child, [*path, meth.to_s], out, visited)
    rescue StandardError
      next
    end
  end

  def list_method?(name)
    name == "list" || name.start_with?("list_")
  end

  # --- Closure-wired test --------------------------------------------------
  #
  # For each registered list method, stub the underlying HTTP path so the
  # first call returns a page with `list_metadata.after = 'cursor-2'` and
  # the second call returns a terminal page (`after = nil`). Walk the
  # result with `#to_a` and confirm two requests were issued — proving the
  # auto-paging closure is wired and re-issues with the swapped cursor.

  REGISTRY.each do |row|
    define_method("test_#{row[:service].to_s.tr(".", "_")}_pagination_walks_every_page") do
      first_body = JSON.generate(
        "data" => [JSON.parse(row[:sample])],
        "list_metadata" => {"before" => nil, "after" => "cursor-2"}
      )
      second_body = JSON.generate(
        "data" => [JSON.parse(row[:sample])],
        "list_metadata" => {"before" => nil, "after" => nil}
      )

      stub = stub_request(:get, %r{\Ahttps://api\.retab\.com#{Regexp.escape(row[:path])}(\?|\z)})
        .to_return(
          {body: first_body, status: 200},
          {body: second_body, status: 200}
        )

      result = row[:invoke].call(@client)
      assert_kind_of(
        Retab::PaginatedList,
        result,
        "#{row[:service]} returned #{result.class}, expected Retab::PaginatedList"
      )

      all_items = result.to_a
      assert_equal(
        2,
        all_items.length,
        "#{row[:service]} yielded #{all_items.length} items across 2 pages; closure is not wired"
      )

      assert_requested(stub, times: 2)
    end
  end
end
