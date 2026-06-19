// REAL end-to-end tests for workflow graph/runtime sub-resources.
//
// CREDITLESS: these tests only read existing workflow definitions, runs,
// steps, artifacts, reviews, experiments, and saved evals. They never create a
// workflow run, approve/reject a review, run an experiment, or execute a block.

import { describe, expect, test } from 'bun:test';

import { RetabError, type Retab } from '../../src/index.js';
import { LIVE, LIVE_SKIP_REASON, liveClient } from '../live.js';

const d = describe.skipIf(!LIVE);

if (!LIVE) {
  describe('workflow sub-resources e2e', () => {
    test.skip(LIVE_SKIP_REASON, () => {});
  });
}

async function discoverWorkflowId(client: Retab): Promise<string | null> {
  const page = await client.workflows.list({ limit: 25 });
  return page.data[0]?.id ?? null;
}

d('workflow graph sub-resources (live, read-only)', () => {
  test('blocks list/get and edges list/get round-trip with workflow scope', async () => {
    const client = liveClient();
    const workflowId = await discoverWorkflowId(client);
    if (!workflowId) return;

    const blocks = await client.workflows.blocks.list({ workflowId, limit: 5 });
    expect(Array.isArray(blocks.data)).toBe(true);
    expect(blocks.data.length).toBeLessThanOrEqual(5);
    for (const block of blocks.data) {
      expect(block.workflowId).toBe(workflowId);
      expect(typeof block.id).toBe('string');
      expect(block.updatedAt).toBeInstanceOf(Date);
    }

    if (blocks.data.length > 0) {
      const block = blocks.data[0];
      const fetched = await client.workflows.blocks.get(block.id, { workflowId });
      expect(fetched.id).toBe(block.id);
      expect(fetched.workflowId).toBe(workflowId);
    }

    const edges = await client.workflows.edges.list({ workflowId, limit: 5 });
    expect(Array.isArray(edges.data)).toBe(true);
    expect(edges.data.length).toBeLessThanOrEqual(5);
    for (const edge of edges.data) {
      expect(edge.workflowId).toBe(workflowId);
      expect(typeof edge.sourceBlock).toBe('string');
      expect(typeof edge.targetBlock).toBe('string');
      expect(edge.updatedAt).toBeInstanceOf(Date);
    }

    if (edges.data.length > 0) {
      const edge = edges.data[0];
      const fetched = await client.workflows.edges.get(edge.id, { workflowId });
      expect(fetched.id).toBe(edge.id);
      expect(fetched.workflowId).toBe(workflowId);
      expect(fetched.sourceBlock).toBe(edge.sourceBlock);
      expect(fetched.targetBlock).toBe(edge.targetBlock);

      const sourceScoped = await client.workflows.edges.list({
        workflowId,
        sourceBlock: edge.sourceBlock,
        limit: 5,
      });
      for (const scopedEdge of sourceScoped.data) {
        expect(scopedEdge.workflowId).toBe(workflowId);
        expect(scopedEdge.sourceBlock).toBe(edge.sourceBlock);
      }
    }
  });

  test('saved evals and experiments list by workflow and get discovered rows', async () => {
    const client = liveClient();
    const workflowId = await discoverWorkflowId(client);
    if (!workflowId) return;

    const tests = await client.workflows.evals.list({ workflowId, limit: 5 });
    expect(Array.isArray(tests.data)).toBe(true);
    for (const savedTest of tests.data) {
      expect(savedTest.workflowId).toBe(workflowId);
      expect(typeof savedTest.id).toBe('string');
    }
    if (tests.data.length > 0) {
      const fetched = await client.workflows.evals.get(tests.data[0].id);
      expect(fetched.id).toBe(tests.data[0].id);
      expect(fetched.workflowId).toBe(workflowId);
    }

    const experiments = await client.workflows.experiments.list({ workflowId, limit: 5 });
    expect(Array.isArray(experiments.data)).toBe(true);
    for (const experiment of experiments.data) {
      expect(experiment.workflowId).toBe(workflowId);
      expect(typeof experiment.id).toBe('string');
    }
    if (experiments.data.length > 0) {
      const fetched = await client.workflows.experiments.get(experiments.data[0].id);
      expect(fetched.id).toBe(experiments.data[0].id);
      expect(fetched.workflowId).toBe(workflowId);
    }
  });
});

d('workflow runtime sub-resources (live, read-only)', () => {
  test('runs, run-scoped steps, and run-scoped artifacts deserialize from live data', async () => {
    const client = liveClient();
    const runs = await client.workflows.runs.list({ limit: 5 });
    expect(Array.isArray(runs.data)).toBe(true);
    if (runs.data.length === 0) return;

    const run = runs.data[0];
    const fetchedRun = await client.workflows.runs.get(run.id);
    expect(fetchedRun.id).toBe(run.id);
    expect(typeof fetchedRun.workflowId).toBe('string');
    expect(typeof fetchedRun.lifecycle.status).toBe('string');

    const steps = await client.workflows.steps.list({ runId: run.id, limit: 10 });
    expect(Array.isArray(steps.data)).toBe(true);
    for (const step of steps.data) {
      expect(step.runId).toBe(run.id);
      expect(typeof step.stepId).toBe('string');
      expect(typeof step.blockId).toBe('string');
      expect(typeof step.lifecycle.status).toBe('string');
    }
    if (steps.data.length > 0) {
      const step = steps.data[0];
      const fetchedStep = await client.workflows.steps.get(step.stepId, { runId: run.id });
      expect(fetchedStep.stepId).toBe(step.stepId);
      expect(fetchedStep.runId).toBe(run.id);
    }

    const artifacts = await client.workflows.artifacts.list({ runId: run.id, limit: 5 });
    expect(Array.isArray(artifacts.data)).toBe(true);
    for (const artifact of artifacts.data) {
      expect(typeof artifact.id).toBe('string');
      expect(typeof artifact.operation).toBe('string');
    }
  });

  test('reviews list/get remains read-only and artifacts reject unscoped scans', async () => {
    const client = liveClient();
    const reviews = await client.workflows.reviews.list({ limit: 5 });
    expect(Array.isArray(reviews.data)).toBe(true);
    for (const review of reviews.data) {
      expect(typeof review.id).toBe('string');
      expect(typeof review.workflowId).toBe('string');
      expect(typeof review.runId).toBe('string');
      expect(review.createdAt).toBeInstanceOf(Date);
    }
    if (reviews.data.length > 0) {
      const fetched = await client.workflows.reviews.get(reviews.data[0].id);
      expect(fetched.id).toBe(reviews.data[0].id);
      expect(fetched.runId).toBe(reviews.data[0].runId);
    }

    let thrown: unknown;
    try {
      await client.workflows.artifacts.list({ limit: 1 });
    } catch (error) {
      thrown = error;
    }
    expect(thrown).toBeInstanceOf(RetabError);
    expect((thrown as RetabError).status).toBe(400);
  });
});
