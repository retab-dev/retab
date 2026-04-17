import { describe, expect, test } from 'bun:test';
import { existsSync, readFileSync } from 'node:fs';
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
const ALLOWED_NODE_ONLY_METHODS: Record<string, string[]> = {};

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
  if (params.url === '/documents/perform_ocr_only' && params.method === 'POST') {
    return {
      ocr_file_id: 'ocr-123',
      ocr_result: {
        pages: [],
      },
    };
  }
  if (params.url === '/documents/compute_field_locations' && params.method === 'POST') {
    return {
      locations: {
        status: [{ page: 1, x: 10, y: 20 }],
      },
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

  test('documents.sources follows the same extraction -> ocr -> field-locations flow', async () => {
    const mockClient = new MockClient();
    const api = new APIV1(mockClient);

    const result = await api.documents.sources('ext-123');

    expect(result).toEqual({
      locations: {
        status: [{ page: 1, x: 10, y: 20 }],
      },
    });
    expect(mockClient.requests.map((request) => `${request.method} ${request.url}`)).toEqual([
      'GET /extractions/ext-123',
      'POST /documents/perform_ocr_only',
      'POST /documents/compute_field_locations',
    ]);
    expect(mockClient.requests[1]?.body).toEqual({ file_id: 'file-123' });
    expect(mockClient.requests[2]?.body).toEqual({
      ocr_file_id: 'ocr-123',
      ocr_result: {
        pages: [],
      },
      data: {
        status: 'approved',
      },
    });
  });
});
