import fs from 'fs';
import crypto from 'crypto';

// Hash generation utilities
export function generateBlake2bHashFromString(input: string): string {
  // Using sha256 as Node.js doesn't have built-in blake2b
  return crypto.createHash('sha256').update(input).digest('hex').substring(0, 16);
}


export function generateSchemaDataId(jsonSchema: Record<string, any>): string {
  const cleanedSchema = cleanSchema(
    JSON.parse(JSON.stringify(jsonSchema)),
    true,
    ['description', 'default', 'title', 'required', 'examples', 'deprecated', 'readOnly', 'writeOnly']
  );
  return 'sch_data_id_' + generateBlake2bHashFromString(
    JSON.stringify(cleanedSchema, Object.keys(cleanedSchema).sort())
  );
}

export function generateSchemaId(jsonSchema: Record<string, any>): string {
  return 'sch_id_' + generateBlake2bHashFromString(
    JSON.stringify(jsonSchema, Object.keys(jsonSchema).sort())
  );
}

// Schema cleaning utilities
export function cleanSchema(
  schema: Record<string, any>,
  removeCustomFields: boolean = true,
  fieldsToRemove: string[] = []
): Record<string, any> {
  const cleaned = { ...schema };

  // Remove specified fields
  for (const field of fieldsToRemove) {
    delete cleaned[field];
  }

  // Remove custom fields (starting with X-)
  if (removeCustomFields) {
    for (const key of Object.keys(cleaned)) {
      if (key.startsWith('X-')) {
        delete cleaned[key];
      }
    }
  }

  // Recursively clean nested objects
  for (const [key, value] of Object.entries(cleaned)) {
    if (typeof value === 'object' && value !== null) {
      if (Array.isArray(value)) {
        cleaned[key] = value.map(item => 
          typeof item === 'object' && item !== null 
            ? cleanSchema(item, removeCustomFields, fieldsToRemove)
            : item
        );
      } else {
        cleaned[key] = cleanSchema(value, removeCustomFields, fieldsToRemove);
      }
    }
  }

  return cleaned;
}

// JSON Schema loading utilities
export function loadJsonSchema(jsonSchema: Record<string, any> | string): Record<string, any> {
  if (typeof jsonSchema === 'string') {
    if (fs.existsSync(jsonSchema)) {
      return JSON.parse(fs.readFileSync(jsonSchema, 'utf-8'));
    }
    return JSON.parse(jsonSchema);
  }
  return jsonSchema;
}

// Schema expansion utilities
export function expandRefs(schema: Record<string, any>): Record<string, any> {
  const expanded = { ...schema };
  const definitions = expanded.$defs || expanded.definitions || {};

  function resolveRef(obj: any): any {
    if (typeof obj === 'object' && obj !== null) {
      if (obj.$ref && typeof obj.$ref === 'string') {
        const refPath = obj.$ref;
        if (refPath.startsWith('#/$defs/')) {
          const defName = refPath.replace('#/$defs/', '');
          if (definitions[defName]) {
            return resolveRef(definitions[defName]);
          }
        }
        return obj;
      }

      if (Array.isArray(obj)) {
        return obj.map(resolveRef);
      }

      const resolved: any = {};
      for (const [key, value] of Object.entries(obj)) {
        resolved[key] = resolveRef(value);
      }
      return resolved;
    }
    return obj;
  }

  return resolveRef(expanded);
}

// Reasoning schema creation
export function createReasoningSchema(schema: Record<string, any>): Record<string, any> {
  const enhanced = JSON.parse(JSON.stringify(schema));

  function addReasoningFields(obj: any, path: string = ''): any {
    if (typeof obj === 'object' && obj !== null && !Array.isArray(obj)) {
      const enhanced = { ...obj };

      // Add reasoning field for this level
      const reasoningFieldName = path ? `reasoning___${path.split('.').pop()}` : 'reasoning___root';
      enhanced.properties = enhanced.properties || {};
      enhanced.properties[reasoningFieldName] = {
        type: 'string',
        description: 'Reasoning for extraction decisions at this level',
        'X-ReasoningPrompt': 'Explain your reasoning for the extracted data',
      };

      // Recursively process properties
      if (enhanced.properties) {
        for (const [key, value] of Object.entries(enhanced.properties)) {
          if (key.startsWith('reasoning___')) continue;
          enhanced.properties[key] = addReasoningFields(value, path ? `${path}.${key}` : key);
        }
      }

      // Handle array items
      if (enhanced.items) {
        enhanced.items = addReasoningFields(enhanced.items, `${path}.*`);
      }

      return enhanced;
    }
    return obj;
  }

  return addReasoningFields(enhanced);
}

// Validation utilities
export function validateCurrency(currencyCode: any): string | null {
  if (currencyCode == null) return null;
  const code = String(currencyCode).trim();
  if (!code) return null;

  // Basic currency code validation (ISO 4217)
  const validCurrencies = ['USD', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY'];
  return validCurrencies.includes(code.toUpperCase()) ? code.toUpperCase() : null;
}

export function validateCountryCode(value: any): string | null {
  if (value == null) return null;
  const code = String(value).trim();
  if (!code) return null;

  // Basic country code validation (ISO 3166)
  const validCountries = ['US', 'CA', 'GB', 'FR', 'DE', 'JP', 'AU', 'IT', 'ES', 'NL'];
  return validCountries.includes(code.toUpperCase()) ? code.toUpperCase() : null;
}

export function validateEmailRegex(value: any): string | null {
  if (value == null) return null;
  const email = String(value).trim();
  if (!email) return null;

  const pattern = /^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$/;
  return pattern.test(email) ? email.toLowerCase() : null;
}

export function validatePhoneNumber(value: any): string | null {
  if (value == null) return null;
  const phone = String(value).trim();
  if (!phone) return null;

  // Basic phone number validation
  const pattern = /^\+?[1-9]\d{1,14}$/;
  return pattern.test(phone.replace(/\s|-/g, '')) ? phone : null;
}

// Filter auxiliary fields from JSON
export function filterAuxiliaryFieldsJson(jsonData: any): any {
  // Handle string input by parsing it first
  let data;
  if (typeof jsonData === 'string') {
    try {
      data = JSON.parse(jsonData);
    } catch (e) {
      // If parsing fails, return the string as-is (it might not be JSON)
      return jsonData;
    }
  } else {
    data = jsonData;
  }
  
  if (typeof data === 'object' && data !== null) {
    if (Array.isArray(data)) {
      return data.map(filterAuxiliaryFieldsJson);
    }

    const filtered: any = {};
    for (const [key, value] of Object.entries(data)) {
      if (!key.startsWith('reasoning___') && !key.startsWith('quote___') && !key.startsWith('X-')) {
        filtered[key] = filterAuxiliaryFieldsJson(value);
      }
    }
    return filtered;
  }
  return data;
}

// Convert JSON Schema to TypeScript interface
export function jsonSchemaToTypescriptInterface(
  schema: Record<string, any>,
  addFieldDescription: boolean = true
): string {
  function getTypeScriptType(property: any): string {
    if (property.type === 'string') return 'string';
    if (property.type === 'number' || property.type === 'integer') return 'number';
    if (property.type === 'boolean') return 'boolean';
    if (property.type === 'array') {
      const itemType = property.items ? getTypeScriptType(property.items) : 'any';
      return `${itemType}[]`;
    }
    if (property.type === 'object') {
      return 'Record<string, any>';
    }
    return 'any';
  }

  function generateInterface(obj: any, interfaceName: string = 'Schema'): string {
    let result = `interface ${interfaceName} {\n`;
    
    if (obj.properties) {
      for (const [key, property] of Object.entries(obj.properties)) {
        const prop = property as any;
        const optional = !obj.required || !obj.required.includes(key) ? '?' : '';
        const type = getTypeScriptType(prop);
        
        if (addFieldDescription && prop.description) {
          result += `  /** ${prop.description} */\n`;
        }
        result += `  ${key}${optional}: ${type};\n`;
      }
    }
    
    result += '}';
    return result;
  }

  return generateInterface(schema);
}

// Convert JSON Schema to NLP data structure
export function jsonSchemaToNlpDataStructure(schema: Record<string, any>): string {
  function processObject(obj: any, indent: number = 0): string {
    const spaces = '  '.repeat(indent);
    let result = '';

    if (obj.properties) {
      for (const [key, property] of Object.entries(obj.properties)) {
        const prop = property as any;
        const optional = !obj.required || !obj.required.includes(key);
        const suffix = optional ? ' (optional)' : ' (required)';
        
        result += `${spaces}- ${key}${suffix}`;
        if (prop.description) {
          result += `: ${prop.description}`;
        }
        result += '\n';

        if (prop.type === 'object' && prop.properties) {
          result += processObject(prop, indent + 1);
        } else if (prop.type === 'array' && prop.items) {
          result += `${spaces}  Items:\n`;
          if (prop.items.type === 'object') {
            result += processObject(prop.items, indent + 2);
          } else {
            result += `${spaces}    Type: ${prop.items.type || 'any'}\n`;
          }
        }
      }
    }

    return result;
  }

  return processObject(schema);
}

// Convert JSON Schema to strict OpenAI schema
export function jsonSchemaToStrictOpenaiSchema(schema: Record<string, any>): Record<string, any> {
  const strict = JSON.parse(JSON.stringify(schema));

  function makeStrict(obj: any): any {
    if (typeof obj === 'object' && obj !== null) {
      if (obj.type === 'object' && obj.properties) {
        // Ensure all properties are required for strict mode
        obj.required = Object.keys(obj.properties);
        obj.additionalProperties = false;
        
        for (const property of Object.values(obj.properties)) {
          makeStrict(property);
        }
      } else if (obj.type === 'array' && obj.items) {
        makeStrict(obj.items);
      }
    }
    return obj;
  }

  return makeStrict(strict);
}

// Unflatten dictionary utility
export function unflattenDict(flatDict: Record<string, any>, separator: string = '.'): Record<string, any> {
  const result: Record<string, any> = {};

  for (const [key, value] of Object.entries(flatDict)) {
    const keys = key.split(separator);
    let current = result;

    for (let i = 0; i < keys.length - 1; i++) {
      const k = keys[i];
      if (!(k in current)) {
        current[k] = {};
      }
      current = current[k];
    }

    current[keys[keys.length - 1]] = value;
  }

  return result;
}

// Schema to TypeScript type utility
export function schemaToTsType(
  schema: Record<string, any>,
  definitions: Record<string, any> = {},
  visited: Set<string> = new Set(),
  depth: number = 0,
  maxDepth: number = 10,
  addFieldDescription: boolean = true
): string {
  if (depth > maxDepth) return 'any';

  if (schema.$ref) {
    const refName = schema.$ref.replace('#/$defs/', '');
    if (visited.has(refName)) return refName;
    if (definitions[refName]) {
      visited.add(refName);
      return schemaToTsType(definitions[refName], definitions, visited, depth + 1, maxDepth, addFieldDescription);
    }
    return 'any';
  }

  switch (schema.type) {
    case 'string':
      return 'string';
    case 'number':
    case 'integer':
      return 'number';
    case 'boolean':
      return 'boolean';
    case 'array':
      if (schema.items) {
        const itemType = schemaToTsType(schema.items, definitions, visited, depth + 1, maxDepth, addFieldDescription);
        return `${itemType}[]`;
      }
      return 'any[]';
    case 'object':
      if (schema.properties) {
        const props = Object.entries(schema.properties).map(([key, prop]: [string, any]) => {
          const optional = !schema.required || !schema.required.includes(key) ? '?' : '';
          const propType = schemaToTsType(prop, definitions, visited, depth + 1, maxDepth, addFieldDescription);
          return `${key}${optional}: ${propType}`;
        });
        return `{ ${props.join('; ')} }`;
      }
      return 'Record<string, any>';
    default:
      return 'any';
  }
}