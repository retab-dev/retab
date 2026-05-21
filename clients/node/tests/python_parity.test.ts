import { describe, expect, test } from 'bun:test';
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';

import APIV1 from '../src/api/client.js';
import { AbstractClient } from '../src/client.js';

type RecordedRequest = {
  url: string;
  method: string;
  params?: Record<string, unknown>;
  headers?: Record<string, unknown>;
  bodyMime?: 'application/json' | 'multipart/form-data';
  body?: Record<string, unknown>;
};

type PythonResourceSurface = {
  resource: string;
  methods: string[];
};

type PythonClassDefinition = {
  name: string;
  bases: string[];
  block: string;
};

const PYTHON_CLIENT_FILE = path.resolve(import.meta.dir, '../../python/retab/client.py');
const PYTHON_RESOURCES_DIR = path.resolve(import.meta.dir, '../../python/retab/resources');
const GO_CLIENT_DIR = path.resolve(import.meta.dir, '../../go');
const DOCS_OPENAPI_FILE = path.resolve(
  import.meta.dir,
  '../../../../docs/api-reference/openapi.json'
);
const ALLOWED_NODE_ONLY_METHODS: Record<string, string[]> = {};
const OPENAPI_METHODS = new Set([
  'get',
  'put',
  'post',
  'delete',
  'options',
  'head',
  'patch',
  'trace',
]);

class MockClient extends AbstractClient {
  public requests: RecordedRequest[] = [];

  protected async _fetch(params: RecordedRequest): Promise<Response> {
    this.requests.push(params);
    return new Response(JSON.stringify(buildResponse(params)), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }
}

function getNested(target: Record<string, unknown>, resourcePath: string): unknown {
  return resourcePath.split('.').reduce<unknown>((current, segment) => {
    if (!current || typeof current !== 'object') {
      return undefined;
    }
    return (current as Record<string, unknown>)[segment];
  }, target);
}

function getNodePublicPythonMethods(target: Record<string, unknown>): string[] {
  const names = new Set<string>();
  let current: object | null = target;

  while (current && current !== Object.prototype) {
    for (const name of Object.getOwnPropertyNames(current)) {
      if (
        name !== 'constructor' &&
        !name.startsWith('_') &&
        /^(prepare_[a-z0-9_]+|[a-z][a-z0-9_]*)$/.test(name) &&
        typeof target[name as keyof typeof target] === 'function'
      ) {
        names.add(name);
      }
    }
    current = Object.getPrototypeOf(current);
  }

  return Array.from(names).sort();
}

function collectNodeResourcePaths(
  target: Record<string, unknown>,
  pathPrefix = '',
  seen = new Set<unknown>()
): string[] {
  if (seen.has(target)) {
    return [];
  }
  seen.add(target);

  const paths: string[] = [];
  for (const [name, value] of Object.entries(target)) {
    if (
      name.startsWith('_') ||
      typeof value !== 'object' ||
      value === null ||
      Array.isArray(value)
    ) {
      continue;
    }

    const resourcePath = pathPrefix ? `${pathPrefix}.${name}` : name;
    paths.push(resourcePath);
    paths.push(...collectNodeResourcePaths(value as Record<string, unknown>, resourcePath, seen));
  }
  return paths;
}

function buildResponse(params: RecordedRequest): unknown {
  if (params.url === '/extractions/ext-123' && params.method === 'GET') {
    return {
      completion: {
        choices: [
          {
            message: {
              parsed: {
                status: 'approved',
              },
            },
          },
        ],
      },
      file_ids: ['file-123'],
    };
  }
  return {};
}

function findClassDefinition(
  source: string,
  predicate: (name: string, bases: string[]) => boolean
): PythonClassDefinition {
  const lines = source.split('\n');

  for (let index = 0; index < lines.length; index++) {
    const match = /^class\s+([A-Za-z_][A-Za-z0-9_]*)(?:\(([^)]*)\))?:\s*$/.exec(lines[index] ?? '');
    if (!match) {
      continue;
    }
    const [, name, rawBases = ''] = match;
    const bases = rawBases
      .split(',')
      .map((base) => base.trim())
      .filter(Boolean);

    if (!predicate(name, bases)) {
      continue;
    }

    let endIndex = index + 1;
    while (
      endIndex < lines.length &&
      !/^class\s+[A-Za-z_][A-Za-z0-9_]*(?:\(|:)/.test(lines[endIndex] ?? '')
    ) {
      endIndex++;
    }
    return {
      name,
      bases,
      block: lines.slice(index, endIndex).join('\n'),
    };
  }

  throw new Error('Unable to locate matching Python class block');
}

function findMethodBlock(classBlock: string, methodName: string): string {
  const lines = classBlock.split('\n');
  const methodPattern = new RegExp(`^ {4}def ${methodName}\\(`);

  for (let index = 0; index < lines.length; index++) {
    if (!methodPattern.test(lines[index] ?? '')) {
      continue;
    }

    let endIndex = index + 1;
    while (
      endIndex < lines.length &&
      !/^ {4}(def|async def) /.test(lines[endIndex] ?? '') &&
      !/^class\s+[A-Za-z_][A-Za-z0-9_]*\(/.test(lines[endIndex] ?? '')
    ) {
      endIndex++;
    }
    return lines.slice(index, endIndex).join('\n');
  }

  return '';
}

function listPythonMethods(classBlock: string): string[] {
  return Array.from(
    new Set(
      Array.from(classBlock.matchAll(/^ {4}def ([A-Za-z_][A-Za-z0-9_]*)\(/gm))
        .map((match) => match[1])
        .filter((methodName) => methodName !== '__init__' && !methodName.startsWith('_'))
    )
  );
}

function listNestedPythonResources(
  initBlock: string
): Array<{ resourceName: string; className: string }> {
  const seen = new Set<string>();
  const nestedResources: Array<{ resourceName: string; className: string }> = [];

  for (const match of initBlock.matchAll(
    /^ {8}self\.([a-z_][a-z0-9_]*)\s*=\s*([A-Za-z_][A-Za-z0-9_.]*)\(/gm
  )) {
    const resourceName = match[1];
    const className = match[2].split('.').at(-1) ?? match[2];
    const key = `${resourceName}:${className}`;
    if (resourceName.startsWith('_') || seen.has(key)) {
      continue;
    }
    seen.add(key);
    nestedResources.push({ resourceName, className });
  }

  return nestedResources;
}

function resolveRelativePythonImport(currentFile: string, moduleSpecifier: string): string | null {
  const dotPrefixMatch = /^(\.+)(.*)$/.exec(moduleSpecifier);
  if (!dotPrefixMatch) {
    return null;
  }

  const [, dots, remainder] = dotPrefixMatch;
  let baseDirectory = path.dirname(currentFile);
  for (let level = 1; level < dots.length; level++) {
    baseDirectory = path.dirname(baseDirectory);
  }

  const moduleParts = remainder ? remainder.split('.').filter(Boolean) : [];
  const modulePath = path.join(baseDirectory, ...moduleParts);
  const candidates = [
    `${modulePath}.py`,
    path.join(modulePath, 'client.py'),
    path.join(modulePath, '__init__.py'),
  ];

  for (const candidate of candidates) {
    if (existsSync(candidate)) {
      return candidate;
    }
  }

  return null;
}

function buildPythonImportMap(source: string, currentFile: string): Map<string, string> {
  const importMap = new Map<string, string>();

  for (const match of source.matchAll(/^from\s+([.\w]+)\s+import\s+(.+)$/gm)) {
    const [, moduleSpecifier, importedNames] = match;
    if (!moduleSpecifier.startsWith('.')) {
      continue;
    }

    const resolvedFile = resolveRelativePythonImport(currentFile, moduleSpecifier);
    if (!resolvedFile) {
      continue;
    }

    for (const rawImportedName of importedNames.split(',')) {
      const importedName = rawImportedName
        .trim()
        .split(/\s+as\s+/)[0]
        ?.trim();
      if (!importedName) {
        continue;
      }
      importMap.set(importedName, resolvedFile);
    }
  }

  return importMap;
}

function collectPythonClassMethods(
  filePath: string,
  className: string,
  seen = new Set<string>()
): string[] {
  const visitKey = `${filePath}:${className}`;
  if (seen.has(visitKey)) {
    return [];
  }
  seen.add(visitKey);

  const source = readFileSync(filePath, 'utf8');
  const classDefinition = findClassDefinition(source, (name) => name === className);
  const importMap = buildPythonImportMap(source, filePath);
  const methods = new Set(listPythonMethods(classDefinition.block));

  for (const baseClassName of classDefinition.bases) {
    if (baseClassName === 'SyncAPIResource' || baseClassName === 'AsyncAPIResource') {
      continue;
    }

    let baseFilePath: string | null = null;
    try {
      findClassDefinition(source, (name) => name === baseClassName);
      baseFilePath = filePath;
    } catch {
      baseFilePath = importMap.get(baseClassName) ?? null;
    }

    if (!baseFilePath) {
      continue;
    }

    for (const inheritedMethod of collectPythonClassMethods(baseFilePath, baseClassName, seen)) {
      methods.add(inheritedMethod);
    }
  }

  return Array.from(methods).sort();
}

function resolvePythonResourceFile(parentFile: string | null, resourceName: string): string | null {
  const baseDirectory = parentFile ? path.dirname(parentFile) : PYTHON_RESOURCES_DIR;
  const candidates = [
    path.join(baseDirectory, resourceName, 'client.py'),
    path.join(baseDirectory, `${resourceName}.py`),
  ];

  for (const candidate of candidates) {
    if (existsSync(candidate)) {
      return candidate;
    }
  }

  return null;
}

function buildPythonResourceSurfaceFromClass(
  resourceFile: string,
  className: string,
  resourcePath: string
): PythonResourceSurface[] {
  const source = readFileSync(resourceFile, 'utf8');
  const importMap = buildPythonImportMap(source, resourceFile);
  const classDefinition = findClassDefinition(
    source,
    (name, bases) => name === className && bases.includes('SyncAPIResource')
  );
  const initBlock = findMethodBlock(classDefinition.block, '__init__');

  const currentResource: PythonResourceSurface = {
    resource: resourcePath,
    methods: collectPythonClassMethods(resourceFile, classDefinition.name),
  };

  const children = listNestedPythonResources(initBlock).flatMap(
    ({ resourceName: childResourceName, className: childClassName }) => {
      let childResourceFile: string | null = null;
      try {
        findClassDefinition(source, (name) => name === childClassName);
        childResourceFile = resourceFile;
      } catch {
        childResourceFile =
          importMap.get(childClassName) ??
          resolvePythonResourceFile(resourceFile, childResourceName);
      }

      if (!childResourceFile) {
        return [];
      }

      return buildPythonResourceSurfaceFromClass(
        childResourceFile,
        childClassName,
        `${resourcePath}.${childResourceName}`
      );
    }
  );

  return [currentResource, ...children];
}

function buildPythonResourceSurface(
  resourceName: string,
  parentFile: string | null,
  resourcePath = resourceName
): PythonResourceSurface[] {
  const resourceFile = resolvePythonResourceFile(parentFile, resourceName);
  if (!resourceFile) {
    return [];
  }

  const source = readFileSync(resourceFile, 'utf8');
  const classDefinition = findClassDefinition(source, (_name, bases) =>
    bases.includes('SyncAPIResource')
  );
  return buildPythonResourceSurfaceFromClass(resourceFile, classDefinition.name, resourcePath);
}

function buildExpectedPythonSurface(): PythonResourceSurface[] {
  const source = readFileSync(PYTHON_CLIENT_FILE, 'utf8');
  const classDefinition = findClassDefinition(source, (name) => name === 'Retab');
  const initBlock = findMethodBlock(classDefinition.block, '__init__');
  const importMap = buildPythonImportMap(source, PYTHON_CLIENT_FILE);

  return listNestedPythonResources(initBlock)
    .flatMap(({ resourceName, className }) => {
      const resourceFile =
        importMap.get(className) ?? resolvePythonResourceFile(null, resourceName);
      if (!resourceFile) {
        return [];
      }
      return buildPythonResourceSurfaceFromClass(resourceFile, className, resourceName);
    })
    .sort((left, right) => left.resource.localeCompare(right.resource));
}

function toSnakeCase(value: string): string {
  return value
    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1_$2')
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .toLowerCase();
}

function normalizeGoMethodName(value: string): string {
  return toSnakeCase(
    value
      .replaceAll('API', 'Api')
      .replaceAll('CSV', 'Csv')
      .replaceAll('review', 'Hil')
      .replaceAll('HTTP', 'Http')
      .replaceAll('ID', 'Id')
      .replaceAll('JSON', 'Json')
      .replaceAll('MIME', 'Mime')
      .replaceAll('URL', 'Url')
      .replaceAll('YAML', 'Yaml')
  );
}

function readGoClientSource(): string {
  return readdirSync(GO_CLIENT_DIR)
    .filter((filename) => filename.endsWith('.go') && !filename.endsWith('_test.go'))
    .sort()
    .map((filename) => readFileSync(path.join(GO_CLIENT_DIR, filename), 'utf8'))
    .join('\n');
}

function listFilesRecursive(directory: string, extension: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const entryPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      return listFilesRecursive(entryPath, extension);
    }
    return entry.isFile() && entry.name.endsWith(extension) ? [entryPath] : [];
  });
}

function normalizeWorkflowRoutePath(rawPath: string): string {
  const withoutVersion = rawPath.replace(/^\/v1/, '');
  const withoutQuery = withoutVersion.split('?')[0];
  const normalizedParams = withoutQuery
    .replace(/\$\{[^}]+\}/g, '{param}')
    .replace(/\{[^}]+\}/g, '{param}');
  const normalizedSlashes = normalizedParams.replace(/\/+/g, '/');
  return normalizedSlashes.length > 1 ? normalizedSlashes.replace(/\/+$/, '') : normalizedSlashes;
}

function routeKey(method: string, rawPath: string): string | null {
  const normalizedPath = normalizeWorkflowRoutePath(rawPath);
  if (!normalizedPath.startsWith('/workflows')) {
    return null;
  }
  return `${method.toLowerCase()} ${normalizedPath}`;
}

function stripQuotedLiteral(rawValue: string): string {
  return rawValue.replace(/^f?['"`]/, '').replace(/['"`]$/, '');
}

function collectDocsWorkflowRouteKeys(): string[] {
  const spec = JSON.parse(readFileSync(DOCS_OPENAPI_FILE, 'utf8')) as {
    paths: Record<string, Record<string, unknown>>;
  };
  const routes = new Set<string>();

  for (const [rawPath, pathItem] of Object.entries(spec.paths)) {
    for (const method of Object.keys(pathItem)) {
      if (!OPENAPI_METHODS.has(method)) {
        continue;
      }
      const key = routeKey(method, rawPath);
      if (key) {
        routes.add(key);
      }
    }
  }

  return Array.from(routes).sort();
}

function collectPythonWorkflowRouteKeys(): string[] {
  const routes = new Set<string>();
  const workflowResourcesDirectory = path.join(PYTHON_RESOURCES_DIR, 'workflows');

  for (const filePath of listFilesRecursive(workflowResourcesDirectory, '.py')) {
    const source = readFileSync(filePath, 'utf8');
    for (const match of source.matchAll(/PreparedRequest\(([\s\S]*?)\)/g)) {
      const block = match[1];
      const methodMatch = /\bmethod\s*=\s*["']([A-Z]+)["']/.exec(block);
      const urlMatch = /\burl\s*=\s*(f?["'][^"']+["'])/.exec(block);
      if (!methodMatch || !urlMatch) {
        continue;
      }
      const key = routeKey(methodMatch[1], stripQuotedLiteral(urlMatch[1]));
      if (key) {
        routes.add(key);
      }
    }
  }

  return Array.from(routes).sort();
}

function collectNodeWorkflowRouteKeys(): string[] {
  const routes = new Set<string>();
  const workflowSourceDirectory = path.resolve(import.meta.dir, '../src/api/workflows');

  for (const filePath of listFilesRecursive(workflowSourceDirectory, '.ts')) {
    const source = readFileSync(filePath, 'utf8');
    for (const match of source.matchAll(
      /\burl:\s*(`[^`]+`|'[^']+'|"[^"]+")[\s\S]{0,350}?\bmethod:\s*['"]([A-Z]+)['"]/g
    )) {
      const key = routeKey(match[2], stripQuotedLiteral(match[1]));
      if (key) {
        routes.add(key);
      }
    }
  }

  return Array.from(routes).sort();
}

function goURLExpressionToRoutePaths(method: string, expression: string): string[] {
  if (expression.includes('reviewPath(')) {
    if (method === 'get') {
      return ['/workflows/reviews/{param}'];
    }
    if (method === 'post' && expression.includes('action')) {
      return ['/workflows/reviews/{param}/approve', '/workflows/reviews/{param}/reject'];
    }
  }

  const expressionWithPathParams = expression.replace(/url\.PathEscape\([^)]+\)/g, '"{param}"');
  const pathFromStringLiterals = Array.from(expressionWithPathParams.matchAll(/"([^"]*)"/g))
    .map((match) => match[1])
    .join('');

  return pathFromStringLiterals ? [pathFromStringLiterals] : [];
}

function collectGoWorkflowRouteKeys(): string[] {
  const routes = new Set<string>();
  const source = readFileSync(path.join(GO_CLIENT_DIR, 'workflows.go'), 'utf8');

  const recordRoute = (method: string, expression: string): void => {
    for (const rawPath of goURLExpressionToRoutePaths(method, expression)) {
      const key = routeKey(method, rawPath);
      if (key) {
        routes.add(key);
      }
    }
  };

  for (const match of source.matchAll(
    /s\.client\.do\([^\n]*?http\.Method([A-Za-z]+)\s*,\s*([^,\n]+)/g
  )) {
    recordRoute(match[1].toLowerCase(), match[2].trim());
  }

  for (const match of source.matchAll(/PreparedRequest\s*\{([\s\S]*?)\n\s*\}/g)) {
    const block = match[1];
    const methodMatch = /\bMethod:\s*http\.Method([A-Za-z]+)/.exec(block);
    const urlMatch = /\bURL:\s*([^,\n]+)/.exec(block);
    if (!methodMatch || !urlMatch) {
      continue;
    }
    recordRoute(methodMatch[1].toLowerCase(), urlMatch[1].trim());
  }

  return Array.from(routes).sort();
}

function collectGoServiceFields(
  source: string
): Map<string, Array<{ name: string; type: string }>> {
  const fieldsByService = new Map<string, Array<{ name: string; type: string }>>();

  for (const match of source.matchAll(
    /type\s+([A-Za-z_][A-Za-z0-9_]*Service)\s+struct\s*\{([\s\S]*?)\n\}/g
  )) {
    const [, serviceName, body] = match;
    const fields: Array<{ name: string; type: string }> = [];

    for (const fieldMatch of body.matchAll(
      /^\s*([A-Z][A-Za-z0-9_]*)\s+\*([A-Za-z_][A-Za-z0-9_]*Service)\b/gm
    )) {
      fields.push({ name: fieldMatch[1], type: fieldMatch[2] });
    }

    fieldsByService.set(serviceName, fields);
  }

  return fieldsByService;
}

function collectGoServiceMethods(source: string): Map<string, string[]> {
  const methodsByService = new Map<string, Set<string>>();

  for (const match of source.matchAll(
    /func\s+\([^)]+\s+\*([A-Za-z_][A-Za-z0-9_]*Service)\)\s+([A-Z][A-Za-z0-9_]*)\s*\(/g
  )) {
    const [, serviceName, methodName] = match;
    if (!methodsByService.has(serviceName)) {
      methodsByService.set(serviceName, new Set());
    }
    methodsByService.get(serviceName)?.add(normalizeGoMethodName(methodName));
  }

  return new Map(
    Array.from(methodsByService.entries()).map(([serviceName, methodNames]) => [
      serviceName,
      Array.from(methodNames).sort(),
    ])
  );
}

function collectGoRootServices(source: string): Array<{ name: string; type: string }> {
  const match = /type\s+Client\s+struct\s*\{([\s\S]*?)\n\}/.exec(source);
  if (!match) {
    throw new Error('Unable to locate Go Client struct');
  }

  return Array.from(
    match[1].matchAll(/^\s*([A-Z][A-Za-z0-9_]*)\s+\*([A-Za-z_][A-Za-z0-9_]*Service)\b/gm)
  ).map((fieldMatch) => ({ name: fieldMatch[1], type: fieldMatch[2] }));
}

function buildGoResourceSurface(): PythonResourceSurface[] {
  const source = readGoClientSource();
  const fieldsByService = collectGoServiceFields(source);
  const methodsByService = collectGoServiceMethods(source);

  const visit = (
    serviceName: string,
    resourcePath: string,
    seen: Set<string>
  ): PythonResourceSurface[] => {
    const visitKey = `${serviceName}:${resourcePath}`;
    if (seen.has(visitKey)) {
      return [];
    }
    seen.add(visitKey);

    const current: PythonResourceSurface = {
      resource: resourcePath,
      methods: methodsByService.get(serviceName) ?? [],
    };
    const children = (fieldsByService.get(serviceName) ?? []).flatMap(({ name, type }) =>
      visit(type, `${resourcePath}.${normalizeGoMethodName(name)}`, seen)
    );

    return [current, ...children];
  };

  return collectGoRootServices(source)
    .flatMap(({ name, type }) => visit(type, normalizeGoMethodName(name), new Set()))
    .sort((left, right) => left.resource.localeCompare(right.resource));
}

describe('python sdk parity surface', () => {
  test('node client exposes the full python resource surface', () => {
    const api = new APIV1(new MockClient()) as unknown as Record<string, unknown>;
    const expectedPythonSurface = buildExpectedPythonSurface();
    const expectedResourcePaths = expectedPythonSurface.map(({ resource }) => resource).sort();
    const nodeResourcePaths = collectNodeResourcePaths(api).sort();

    expect(expectedPythonSurface.length).toBeGreaterThan(0);
    expect(nodeResourcePaths).toEqual(expectedResourcePaths);

    for (const { resource, methods } of expectedPythonSurface) {
      const nodeResource = getNested(api, resource);
      expect(nodeResource, `missing resource ${resource}`).toBeDefined();

      for (const method of methods) {
        expect(
          typeof (nodeResource as Record<string, unknown>)[method],
          `missing method ${resource}.${method}`
        ).toBe('function');
      }

      const expectedMethodSet = new Set([
        ...methods,
        ...(ALLOWED_NODE_ONLY_METHODS[resource] ?? []),
      ]);
      const nodeMethods = getNodePublicPythonMethods(nodeResource as Record<string, unknown>);
      const extraMethods = nodeMethods.filter((method) => !expectedMethodSet.has(method));

      expect(
        extraMethods,
        `unexpected python-style methods on ${resource}: ${extraMethods.join(', ')}`
      ).toEqual([]);
    }
  });

  test('go client exposes the full python operation surface', () => {
    const expectedPythonSurface = buildExpectedPythonSurface();
    const expectedResourcePaths = expectedPythonSurface.map(({ resource }) => resource).sort();
    const goSurface = buildGoResourceSurface();
    const goResourcePaths = goSurface.map(({ resource }) => resource).sort();
    const goMethodsByResource = new Map(
      goSurface.map(({ resource, methods }) => [resource, new Set(methods)])
    );

    expect(expectedPythonSurface.length).toBeGreaterThan(0);
    expect(goResourcePaths).toEqual(expectedResourcePaths);

    for (const { resource, methods } of expectedPythonSurface) {
      const goMethods = goMethodsByResource.get(resource) ?? new Set();
      const expectedOperationMethods = methods.filter((method) => !method.startsWith('prepare_'));
      const missingMethods = expectedOperationMethods.filter((method) => !goMethods.has(method));

      expect(
        missingMethods,
        `missing Go methods on ${resource}: ${missingMethods.join(', ')}`
      ).toEqual([]);
    }
  });

  test('workflow route shapes match docs openapi across sdks', () => {
    const docsRoutes = collectDocsWorkflowRouteKeys();

    expect(docsRoutes.length).toBeGreaterThan(0);
    expect(collectPythonWorkflowRouteKeys()).toEqual(docsRoutes);
    expect(collectNodeWorkflowRouteKeys()).toEqual(docsRoutes);
    expect(collectGoWorkflowRouteKeys()).toEqual(docsRoutes);
  });
});
