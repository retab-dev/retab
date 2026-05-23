# frozen_string_literal: true

# Cross-language pagination contract regression test for the Ruby SDK.
#
# Mirrors clients/python/tests/test_pagination_contract.py,
# clients/node/tests/pagination-contract.test.ts, and
# clients/go/pagination_contract_test.go.
#
# What this test enforces:
#
#   1. Every `def list` method on a service exposed off `Retab::Client`
#      returns a `Retab::PaginatedList`.
#   2. That `PaginatedList` has its fetch closure wired up, so iterating
#      with `#each` walks every page lazily (not just the first one).
#   3. The registry below covers every list method the SDK ships — a new
#      resource that adds `def list` without registering itself here
#      fails CI. KNOWN_BYPASS is the explicit allowlist for resources
#      that legitimately bypass the central helper (currently empty).
#
# See .notes/blueprints/sdk-pagination-contract.md for the full
# cross-language contract.

require "test_helper"
require "set"

class PaginationContractTest < Minitest::Test
  # Resource registry: every list method the SDK exposes off
  # `Retab::Client.<service>.list(...)`. Format:
  #   [service_accessor, list_path, sample_item, invoke_proc]
  #
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
      service: :edit_templates,
      path: "/v1/edits/templates",
      sample: "{}",
      invoke: -> (c) { c.edit_templates.list }
    },
    {
      service: :edits,
      path: "/v1/edits",
      sample: "{}",
      invoke: -> (c) { c.edits.list }
    },
    {
      service: :experiment_run_results,
      path: "/v1/workflows/experiments/results",
      sample: "{}",
      invoke: -> (c) { c.experiment_run_results.list(run_id: "run_x") }
    },
    {
      service: :experiment_runs,
      path: "/v1/workflows/experiments/runs",
      sample: "{}",
      invoke: -> (c) { c.experiment_runs.list }
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
      service: :jobs,
      path: "/v1/jobs",
      sample: "{}",
      invoke: -> (c) { c.jobs.list }
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
      service: :workflow_artifacts,
      path: "/v1/workflows/artifacts",
      sample: "{}",
      invoke: -> (c) { c.workflow_artifacts.list(run_id: "run_x") }
    },
    {
      service: :workflow_block_executions,
      path: "/v1/workflows/blocks/executions",
      sample: "{}",
      invoke: -> (c) { c.workflow_block_executions.list(run_id: "run_x", block_id: "blk_x") }
    },
    {
      service: :workflow_blocks,
      path: "/v1/workflows/blocks",
      sample: "{}",
      invoke: -> (c) { c.workflow_blocks.list(workflow_id: "wf_x") }
    },
    {
      service: :workflow_edges,
      path: "/v1/workflows/edges",
      sample: "{}",
      invoke: -> (c) { c.workflow_edges.list(workflow_id: "wf_x") }
    },
    {
      service: :workflow_experiments,
      path: "/v1/workflows/experiments",
      sample: "{}",
      invoke: -> (c) { c.workflow_experiments.list(workflow_id: "wf_x") }
    },
    {
      service: :workflow_review_versions,
      path: "/v1/workflows/reviews/versions",
      sample: "{}",
      invoke: -> (c) { c.workflow_review_versions.list(review_id: "rev_x") }
    },
    {
      service: :workflow_reviews,
      path: "/v1/workflows/reviews",
      sample: "{}",
      invoke: -> (c) { c.workflow_reviews.list }
    },
    {
      service: :workflow_runs,
      path: "/v1/workflows/runs",
      sample: "{}",
      invoke: -> (c) { c.workflow_runs.list }
    },
    {
      service: :workflow_steps,
      path: "/v1/workflows/steps",
      sample: "{}",
      invoke: -> (c) { c.workflow_steps.list }
    },
    {
      service: :workflow_test_run_results,
      path: "/v1/workflows/tests/results",
      sample: "{}",
      invoke: -> (c) { c.workflow_test_run_results.list(run_id: "run_x") }
    },
    {
      service: :workflow_test_runs,
      path: "/v1/workflows/tests/runs",
      sample: "{}",
      invoke: -> (c) { c.workflow_test_runs.list }
    },
    {
      service: :workflow_tests,
      path: "/v1/workflows/tests",
      sample: "{}",
      invoke: -> (c) { c.workflow_tests.list(workflow_id: "wf_x") }
    },
    {
      service: :workflows,
      path: "/v1/workflows",
      sample: "{}",
      invoke: -> (c) { c.workflows.list }
    }
  ].freeze

  # Resources that legitimately bypass the central `request_page` helper.
  # Empty today — Ruby's response envelope is uniform across every route.
  # If you add an entry here, document the reason and update the blueprint.
  KNOWN_BYPASS = [].freeze

  def setup
    @client = Retab::Client.new(api_key: "sk_test_123")
  end

  # --- Discovery test ------------------------------------------------------
  #
  # Walk every service accessor on `Retab::Client` and find every `list`
  # method. Assert that each one appears in the REGISTRY (or in
  # KNOWN_BYPASS). A new service that adds `def list` without registering
  # itself here will fail this assertion.

  def test_registry_covers_every_list_method
    registered = REGISTRY.map { |row| row[:service] }.to_set + KNOWN_BYPASS.to_set

    client_methods = (Retab::Client.instance_methods - Retab::BaseClient.instance_methods - Object.instance_methods)
    services_with_list = []

    client_methods.each do |meth|
      next if meth.to_s.start_with?("_")
      service = @client.send(meth)
      next unless service.respond_to?(:list)
      services_with_list << meth
    end

    missing = services_with_list.reject { |s| registered.include?(s) }
    assert_empty(
      missing,
      "These services expose a `list` method but are not in REGISTRY or KNOWN_BYPASS: #{missing.inspect}. " \
        "Add them to test/test_pagination_contract.rb."
    )
  end

  # --- Closure-wired test --------------------------------------------------
  #
  # For each registered list method, stub the underlying HTTP path so the
  # first call returns a page with `list_metadata.after = 'cursor-2'` and
  # the second call returns a terminal page (`after = nil`). Walk the
  # result with `#to_a` and confirm two requests were issued — proving the
  # auto-paging closure is wired and re-issues with the swapped cursor.

  REGISTRY.each do |row|
    define_method("test_#{row[:service]}_pagination_walks_every_page") do
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
        "#{row[:service]}.list returned #{result.class}, expected Retab::PaginatedList"
      )

      all_items = result.to_a
      assert_equal(
        2,
        all_items.length,
        "#{row[:service]}.list#each yielded #{all_items.length} items across 2 pages; closure is not wired"
      )

      assert_requested(stub, times: 2)
    end
  end
end
