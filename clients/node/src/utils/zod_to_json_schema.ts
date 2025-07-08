import { z } from 'zod';

/**
 * Convert a Zod schema to JSON Schema.
 * This is a basic implementation that handles common Zod types.
 */
export function zodToJsonSchema(zodSchema: z.ZodType<any>): Record<string, any> {
  function convertZodType(schema: z.ZodType<any>): any {
    const def = (schema as any)._def;

    switch (def.typeName) {
      case 'ZodString':
        const stringSchema: any = { type: 'string' };
        if (def.description) {
          stringSchema.description = def.description;
        }
        return stringSchema;

      case 'ZodNumber':
        const numberSchema: any = { type: 'number' };
        if (def.description) {
          numberSchema.description = def.description;
        }
        return numberSchema;

      case 'ZodBoolean':
        const booleanSchema: any = { type: 'boolean' };
        if (def.description) {
          booleanSchema.description = def.description;
        }
        return booleanSchema;

      case 'ZodArray':
        return {
          type: 'array',
          items: convertZodType(def.type),
          ...(def.description && { description: def.description })
        };

      case 'ZodObject':
        const properties: Record<string, any> = {};
        const required: string[] = [];

        for (const [key, value] of Object.entries(def.shape())) {
          properties[key] = convertZodType(value as z.ZodType<any>);
          
          // Check if the field is required (not optional)
          const zodValue = value as any;
          if (!zodValue._def.typeName?.includes('Optional') && 
              zodValue._def.typeName !== 'ZodOptional') {
            required.push(key);
          }
        }

        const objectSchema: any = {
          type: 'object',
          properties,
          additionalProperties: false
        };

        if (required.length > 0) {
          objectSchema.required = required;
        }

        if (def.description) {
          objectSchema.description = def.description;
        }

        return objectSchema;

      case 'ZodOptional':
        return convertZodType(def.innerType);

      case 'ZodNullable':
        const innerSchema = convertZodType(def.innerType);
        return {
          ...innerSchema,
          nullable: true
        };

      case 'ZodUnion':
        return {
          anyOf: def.options.map((option: z.ZodType<any>) => convertZodType(option))
        };

      case 'ZodEnum':
        return {
          type: 'string',
          enum: def.values,
          ...(def.description && { description: def.description })
        };

      case 'ZodLiteral':
        return {
          type: typeof def.value,
          const: def.value,
          ...(def.description && { description: def.description })
        };

      case 'ZodRecord':
        return {
          type: 'object',
          additionalProperties: convertZodType(def.valueType),
          ...(def.description && { description: def.description })
        };

      default:
        // Fallback for unknown types
        return {
          type: 'string',
          description: def.description || `Unsupported Zod type: ${def.typeName}`
        };
    }
  }

  return convertZodType(zodSchema);
}

/**
 * Extract descriptions from Zod schema and convert them to field descriptions.
 */
export function extractZodDescriptions(zodSchema: z.ZodType<any>): Record<string, string> {
  const descriptions: Record<string, string> = {};

  function extractFromType(schema: z.ZodType<any>, path: string = ''): void {
    const def = (schema as any)._def;

    if (def.typeName === 'ZodObject') {
      for (const [key, value] of Object.entries(def.shape())) {
        const currentPath = path ? `${path}.${key}` : key;
        const zodValue = value as any;
        
        if (zodValue._def.description) {
          descriptions[currentPath] = zodValue._def.description;
        }
        
        extractFromType(zodValue, currentPath);
      }
    } else if (def.typeName === 'ZodOptional' || def.typeName === 'ZodNullable') {
      extractFromType(def.innerType, path);
    } else if (def.typeName === 'ZodArray') {
      extractFromType(def.type, path);
    }
  }

  extractFromType(zodSchema);
  return descriptions;
}