import { describe, expect, test } from 'bun:test';
import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

type OpenApiDocument = {
  paths: Record<string, Record<string, unknown>>;
};

type SourceRoute = {
  sourcePath: string;
  route: string;
  normalizedRoute: string;
};

type SourceRouteMethod = SourceRoute & {
  method: string;
  normalizedKey: string;
};

const TEST_DIR = dirname(fileURLToPath(import.meta.url));
const SDK_ROOT = resolve(TEST_DIR, '..');
const API_SOURCE_ROOT = join(SDK_ROOT, 'src');
const WORKFLOW_SOURCE_ROOT = join(API_SOURCE_ROOT, 'workflows');
const OPENAPI_PATH = resolve(SDK_ROOT, '..', '..', '..', 'docs', 'api-reference', 'openapi.json');

const PATH_PARAMETER_NAMES: Record<string, string> = {
  artifactId: 'artifact_id',
  blockId: 'block_id',
  classificationId: 'classification_id',
  classification_id: 'classification_id',
  edgeId: 'edge_id',
  editId: 'edit_id',
  edit_id: 'edit_id',
  experimentId: 'experiment_id',
  extractionId: 'extraction_id',
  extraction_id: 'extraction_id',
  fileId: 'file_id',
  jobId: 'job_id',
  job_id: 'job_id',
  parseId: 'parse_id',
  parse_id: 'parse_id',
  partitionId: 'partition_id',
  resultId: 'result_id',
  reviewId: 'review_id',
  runId: 'run_id',
  splitId: 'split_id',
  split_id: 'split_id',
  stepId: 'step_id',
  templateId: 'template_id',
  template_id: 'template_id',
  testId: 'test_id',
  versionId: 'version_id',
  workflowId: 'workflow_id',
};

const APPROVED_NON_REFERENCE_ROUTES = new Set([
  // Existing SDK conveniences that are not part of the generated public API
  // reference. Keep explicit so new non-reference routes cannot appear silently.
  '/v1/extractions/stream',
]);

const FORBIDDEN_ROUTE_SUBSTRINGS = [
  '/workflow-reviews',
  '/workflow_reviews',
  '/workflows_reviews',
  '/workflows/review-versions',
  '/workflow-review-versions',
  '/workflows/reviews/append',
  '/workflows/reviews/versions/append',
  '/workflows/reviews/{review_id}/versions',
  '/workflows/reviews/${reviewId}/versions',
  '/workflows/blocks/executions/runs',
];

const FORBIDDEN_METHOD_NAMES = [
  'appendVersion',
  'append_version',
  'createReviewVersion',
  'create_review_version',
  'getReviewVersion',
  'get_review_version',
  'listReviewVersions',
  'list_review_versions',
  'workflowsReviewsAppendVersion',
  'workflows_reviews_append_version',
  'workflowsReviewsVersionsCreate',
  'workflows_reviews_versions_create',
  'workflowsBlockExecutionsCreate',
  'workflows_blocks_executions_create',
  'workflowsTestsCreate',
  'workflows_tests_create',
  'workflowsTestsRunsCreate',
  'workflows_tests_runs_create',
];

const CANONICAL_WORKFLOW_REVIEW_SIMULATION_AND_TEST_ROUTES = [
  'GET /v1/workflows/reviews',
  'GET /v1/workflows/reviews/{review_id}',
  'POST /v1/workflows/reviews/{review_id}/approve',
  'POST /v1/workflows/reviews/{review_id}/reject',
  'GET /v1/workflows/reviews/versions',
  'POST /v1/workflows/reviews/versions',
  'GET /v1/workflows/reviews/versions/{version_id}',
  'GET /v1/workflows/blocks/executions',
  'POST /v1/workflows/blocks/executions',
  'GET /v1/workflows/tests',
  'POST /v1/workflows/tests',
  'GET /v1/workflows/tests/{test_id}',
  'PATCH /v1/workflows/tests/{test_id}',
  'DELETE /v1/workflows/tests/{test_id}',
  'GET /v1/workflows/tests/runs',
  'POST /v1/workflows/tests/runs',
  'GET /v1/workflows/tests/runs/{run_id}',
  'POST /v1/workflows/tests/runs/{run_id}/cancel',
  'GET /v1/workflows/tests/results',
  'GET /v1/workflows/tests/results/{result_id}',
];

function listClientFiles(root: string): string[] {
  return readdirSync(root, { withFileTypes: true }).flatMap((entry) => {
    const path = join(root, entry.name);
    if (entry.isDirectory()) {
      return listClientFiles(path);
    }
    return entry.isFile() && entry.name.endsWith('.ts') && !entry.name.endsWith('.d.ts')
      ? [path]
      : [];
  });
}

function readWorkflowSources(): Array<{ path: string; text: string }> {
  return listClientFiles(WORKFLOW_SOURCE_ROOT).map((path) => ({
    path,
    text: readFileSync(path, 'utf8'),
  }));
}

function readApiSources(): Array<{ path: string; text: string }> {
  return listClientFiles(API_SOURCE_ROOT).map((path) => ({
    path,
    text: readFileSync(path, 'utf8'),
  }));
}

function routePathParameterForExpression(expression: string): string {
  for (const [variableName, parameterName] of Object.entries(PATH_PARAMETER_NAMES)) {
    if (expression.includes(variableName)) {
      return `{${parameterName}}`;
    }
  }
  throw new Error(`No OpenAPI path-parameter mapping for SDK route expression: ${expression}`);
}

function normalizeSdkRoute(route: string): string {
  const pathWithoutQuery = route.split('?')[0] as string;
  const normalizedDynamicSegments = pathWithoutQuery.replace(/\$\{([^}]+)\}/g, (_, expression) =>
    routePathParameterForExpression(expression)
  );
  if (normalizedDynamicSegments.startsWith('/v1/')) {
    return normalizedDynamicSegments;
  }
  return `/v1${normalizedDynamicSegments}`;
}

function normalizeRouteShape(route: string): string {
  return route.replace(/\{[^}]+}/g, '{}');
}

function extractSourceRoutes(sources: Array<{ path: string; text: string }>): SourceRoute[] {
  return sources.flatMap(({ path, text }) =>
    Array.from(text.matchAll(/path:\s*(['"`])([^'"`]+)\1/g)).map((match) => {
      const route = match[2] as string;
      return {
        sourcePath: path,
        route,
        normalizedRoute: normalizeSdkRoute(route),
      };
    })
  );
}

function extractSourceRouteMethods(
  sources: Array<{ path: string; text: string }>
): SourceRouteMethod[] {
  return sources.flatMap(({ path, text }) => {
    const pathThenMethod = Array.from(
      text.matchAll(/path:\s*(['"`])([^'"`]+)\1,\s*method:\s*(['"`])([A-Z]+)\3/g)
    ).map((match) => ({ route: match[2] as string, method: match[4] as string }));
    const methodThenPath = Array.from(
      text.matchAll(/method:\s*(['"`])([A-Z]+)\1,\s*path:\s*(['"`])([^'"`]+)\3/g)
    ).map((match) => ({ route: match[4] as string, method: match[2] as string }));

    return [...pathThenMethod, ...methodThenPath].map(({ route, method }) => {
      const normalizedRoute = normalizeSdkRoute(route);
      return {
        sourcePath: path,
        route,
        normalizedRoute,
        method,
        normalizedKey: `${method} ${normalizedRoute}`,
      };
    });
  });
}

function extractMethodNames(sources: Array<{ path: string; text: string }>): string[] {
  return sources.flatMap(({ text }) =>
    Array.from(text.matchAll(/^\s+(?:async\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*\(/gm)).map(
      (match) => match[1] as string
    )
  );
}

function readOpenApi(): OpenApiDocument {
  return JSON.parse(readFileSync(OPENAPI_PATH, 'utf8')) as OpenApiDocument;
}

function openApiRouteMethodKeys(openApi: OpenApiDocument): Set<string> {
  return new Set(
    Object.entries(openApi.paths).flatMap(([path, methods]) =>
      Object.keys(methods).map((method) => `${method.toUpperCase()} ${path}`)
    )
  );
}

describe('Node SDK workflow API design consistency', () => {
  const apiSources = readApiSources();
  const sources = readWorkflowSources();
  const apiSourceRoutes = extractSourceRoutes(apiSources);
  const sourceRoutes = extractSourceRoutes(sources);
  const sourceRouteMethods = extractSourceRouteMethods(sources);
  const openApi = readOpenApi();
  const openApiRoutes = new Set(Object.keys(openApi.paths));
  const normalizedOpenApiRoutes = new Set(Object.keys(openApi.paths).map(normalizeRouteShape));
  const openApiRouteMethods = openApiRouteMethodKeys(openApi);

  test('every SDK API route string is in public OpenAPI or explicitly approved', () => {
    const unknownRoutes = apiSourceRoutes
      .filter(
        ({ normalizedRoute }) => !normalizedOpenApiRoutes.has(normalizeRouteShape(normalizedRoute))
      )
      .filter(({ normalizedRoute }) => !APPROVED_NON_REFERENCE_ROUTES.has(normalizedRoute))
      .map(({ sourcePath, route, normalizedRoute }) => ({
        sourcePath,
        route,
        normalizedRoute,
      }));

    expect(unknownRoutes).toEqual([]);
  });

  test('every SDK workflow route string exists in the public OpenAPI document', () => {
    const unknownRoutes = sourceRoutes
      .filter(({ normalizedRoute }) => !openApiRoutes.has(normalizedRoute))
      .map(({ sourcePath, route, normalizedRoute }) => ({
        sourcePath,
        route,
        normalizedRoute,
      }));

    expect(unknownRoutes).toEqual([]);
  });

  test('every SDK workflow route and method pair exists in the public OpenAPI document', () => {
    const unknownRouteMethods = sourceRouteMethods
      .filter(({ normalizedKey }) => !openApiRouteMethods.has(normalizedKey))
      .map(({ sourcePath, route, method, normalizedRoute }) => ({
        sourcePath,
        method,
        route,
        normalizedRoute,
      }));

    expect(unknownRouteMethods).toEqual([]);
  });

  test('canonical review, block execution, and test routes are the SDK routes in use', () => {
    const sdkRouteMethods = new Set(sourceRouteMethods.map(({ normalizedKey }) => normalizedKey));

    for (const routeMethod of CANONICAL_WORKFLOW_REVIEW_SIMULATION_AND_TEST_ROUTES) {
      expect(openApiRouteMethods.has(routeMethod)).toBe(true);
      expect(sdkRouteMethods.has(routeMethod)).toBe(true);
    }
  });

  test('removed workflow route substrings are absent from SDK route strings', () => {
    const forbiddenRoutes = sourceRoutes.flatMap(({ sourcePath, route, normalizedRoute }) =>
      FORBIDDEN_ROUTE_SUBSTRINGS.filter(
        (substring) => route.includes(substring) || normalizedRoute.includes(substring)
      ).map((substring) => ({
        sourcePath,
        route,
        normalizedRoute,
        substring,
      }))
    );

    expect(forbiddenRoutes).toEqual([]);
  });

  test('removed wrapper-style workflow method names are absent from SDK workflow clients', () => {
    const methodNames = extractMethodNames(sources);
    const forbiddenNames = methodNames.filter(
      (name) =>
        FORBIDDEN_METHOD_NAMES.includes(name) ||
        /^workflows[A-Z]/.test(name) ||
        /^workflows_/.test(name)
    );

    expect(forbiddenNames).toEqual([]);
  });
});
