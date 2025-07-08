import * as fs from 'fs';
import { generateBlake2bHashFromString } from './hash.js';

/**
 * Recursively sort all object keys to match Python's json.dumps(sort_keys=True).
 */
function sortKeysRecursively(obj: any): any {
  if (Array.isArray(obj)) {
    return obj.map(sortKeysRecursively);
  } else if (obj !== null && typeof obj === 'object') {
    return Object.keys(obj)
      .sort()
      .reduce((result: any, key) => {
        result[key] = sortKeysRecursively(obj[key]);
        return result;
      }, {});
  }
  return obj;
}

/**
 * Clean a schema by removing specified fields and custom fields.
 */
export function cleanSchema(
  jsonSchema: Record<string, any>,
  options: {
    removeCustomFields?: boolean;
    fieldsToRemove?: string[];
  } = {}
): Record<string, any> {
  if (!jsonSchema || typeof jsonSchema !== 'object') {
    throw new Error('Invalid schema provided');
  }
  const { removeCustomFields = false, fieldsToRemove = [] } = options;
  const cleaned = JSON.parse(JSON.stringify(jsonSchema)); // Deep clone

  function cleanObject(obj: any): any {
    if (typeof obj !== 'object' || obj === null) {
      return obj;
    }

    if (Array.isArray(obj)) {
      return obj.map(cleanObject);
    }

    const result: any = {};
    for (const [key, value] of Object.entries(obj)) {
      // Skip custom fields (starting with X-)
      if (removeCustomFields && key.startsWith('X-')) {
        continue;
      }
      
      // Skip specified fields to remove
      if (fieldsToRemove.includes(key)) {
        continue;
      }

      result[key] = cleanObject(value);
    }
    return result;
  }

  return cleanObject(cleaned);
}

/**
 * Generate a schema data ID (hash of schema data, ignoring prompt/description/default fields).
 * Uses Python-compatible JSON formatting with spaces and recursive key sorting.
 */
export function generateSchemaDataId(jsonSchema: Record<string, any>): string {
  const cleanedSchema = cleanSchema(jsonSchema, {
    removeCustomFields: true,
    fieldsToRemove: [
      'description',
      'default',
      'title',
      'required',
      'examples',
      'deprecated',
      'readOnly',
      'writeOnly'
    ]
  });

  const sortedSchema = sortKeysRecursively(cleanedSchema);
  const hashInput = pythonJsonStringify(sortedSchema);
  return 'sch_data_id_' + generateBlake2bHashFromString(hashInput);
}

/**
 * Generate a schema ID (hash of the complete schema).
 * Uses Python-compatible JSON formatting (sort_keys=True, default separators).
 */
export function generateSchemaId(jsonSchema: Record<string, any>): string {
  if (!jsonSchema || typeof jsonSchema !== 'object') {
    throw new Error('Invalid schema provided');
  }
  const sortedSchema = sortKeysRecursively(jsonSchema);
  const hashInput = pythonJsonStringify(sortedSchema);
  return 'sch_id_' + generateBlake2bHashFromString(hashInput);
}

/**
 * Stringify JSON to match Python's json.dumps() format exactly.
 * Python uses separators (',', ': ') by default.
 */
function pythonJsonStringify(obj: any): string {
  if (obj === null) return 'null';
  if (typeof obj === 'boolean') return obj.toString();
  if (typeof obj === 'number') return obj.toString();
  if (typeof obj === 'string') return `"${obj}"`;
  
  if (Array.isArray(obj)) {
    const items = obj.map(pythonJsonStringify);
    return `[${items.join(', ')}]`;
  }
  
  if (typeof obj === 'object') {
    const pairs = Object.keys(obj)
      .sort() // Ensure sorted keys
      .map(key => `"${key}": ${pythonJsonStringify(obj[key])}`);
    return `{${pairs.join(', ')}}`;
  }
  
  return 'null';
}

/**
 * Navigate a JSON path in a schema and get the value at that path.
 */
export function getValueAtPath(obj: any, path: string): any {
  const parts = path.split('.');
  let current = obj;
  
  for (const part of parts) {
    if (current && typeof current === 'object' && part in current) {
      current = current[part];
    } else {
      return null;
    }
  }
  
  return current;
}

/**
 * Navigate a JSON path in a schema and set the value at that path.
 */
export function setValueAtPath(obj: any, path: string, value: any): void {
  const parts = path.split('.');
  let current = obj;
  
  for (let i = 0; i < parts.length - 1; i++) {
    const part = parts[i];
    if (!current[part] || typeof current[part] !== 'object') {
      current[part] = {};
    }
    current = current[part];
  }
  
  current[parts[parts.length - 1]] = value;
}

/**
 * Find attribute in schema using a pattern-based search.
 */
export function getPatternAttribute(
  jsonSchema: Record<string, any>,
  pattern: string,
  attribute: 'X-FieldPrompt' | 'X-ReasoningPrompt' | 'type'
): string | null {
  // Simple implementation - look for pattern in properties
  if (jsonSchema.properties && typeof jsonSchema.properties === 'object') {
    for (const [key, value] of Object.entries(jsonSchema.properties)) {
      if (key.includes(pattern) || pattern === key) {
        if (typeof value === 'object' && value !== null && attribute in value) {
          return (value as any)[attribute];
        }
      }
    }
  }
  
  return null;
}

/**
 * Set attribute in schema using a pattern-based search.
 */
export function setPatternAttribute(
  jsonSchema: Record<string, any>,
  pattern: string,
  attribute: 'X-FieldPrompt' | 'X-ReasoningPrompt' | 'X-SystemPrompt' | 'description',
  value: string
): void {
  // Simple implementation - set in root or in matching properties
  if (attribute === 'X-SystemPrompt') {
    jsonSchema[attribute] = value;
    return;
  }

  if (jsonSchema.properties && typeof jsonSchema.properties === 'object') {
    for (const [key, propValue] of Object.entries(jsonSchema.properties)) {
      if (key.includes(pattern) || pattern === key) {
        if (typeof propValue === 'object' && propValue !== null) {
          (propValue as any)[attribute] = value;
        }
      }
    }
  }
}

/**
 * Load a JSON schema from either a dictionary, JSON string, or file path.
 * 
 * @param jsonSchema - Either a dictionary containing the schema, a JSON string, or a path to a JSON file
 * @returns The loaded JSON schema
 * @throws Error if the schema file contains invalid JSON or doesn't exist
 */
export function loadJsonSchema(jsonSchema: Record<string, any> | string): Record<string, any> {
  if (typeof jsonSchema === 'string') {
    // Check if it's a file path or JSON string
    if (jsonSchema.trim().startsWith('{') || jsonSchema.trim().startsWith('[')) {
      // It's a JSON string
      try {
        return JSON.parse(jsonSchema);
      } catch (error) {
        throw new Error(`Invalid JSON string: ${error}`);
      }
    } else {
      // It's a file path
      try {
        const fileContent = fs.readFileSync(jsonSchema, 'utf-8');
        return JSON.parse(fileContent);
      } catch (error) {
        if ((error as any).code === 'ENOENT') {
          throw new Error(`Schema file not found: ${jsonSchema}`);
        }
        throw new Error(`Error reading or parsing schema file: ${error}`);
      }
    }
  }
  
  // It's already an object
  return jsonSchema;
}