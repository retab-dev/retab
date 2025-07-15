import fs from "fs";

function capitalise(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function camelCase(str: string): string {
  return str.split(/[-_ ]/).map((word, index) => {
    if (word.length === 0) return "";
    if (index === 0) return word.charAt(0).toLowerCase() + word.slice(1);
    return word.charAt(0).toUpperCase() + word.slice(1);
  }).join("");
}

function isNamedSchemaName(name: string): boolean {
  return name[0] !== name[0].toLowerCase();
}

function findNamedSchema(schema: any): string | undefined {
  if (schema.$ref) {
    let name = schema.$ref.split("/").pop();
    if (name && isNamedSchemaName(name)) {
      return name;
    }
  }
  return undefined;
}
function indentLastLines(str: string) {
  return str.split("\n").map((line, i) => i === 0 ? line : `  ${line}`).join("\n");
}

function schemaToZod(schema: any, context: any, ignoreNamed: boolean = false): string {
  if (!ignoreNamed && findNamedSchema(schema)) {
    let res = findNamedSchema(schema)!;
    return "Z" + res;
  }
  switch (schema.type) {
    case "string":
      switch (schema.format) {
        case "date":
          return "DateOrISO";
        case "date-time":
          return "DateOrISO";
        case "uuid":
          return "z.string().uuid()";
        case "binary":
          return "z.instanceof(File)";
        default:
          if (schema.enum) {
            return `z.enum([${schema.enum.map(JSON.stringify).join(", ")}])`;
          }
          return "z.string()";
      }
    case "number":
    case "integer":
      return "z.number()";
    case "boolean":
      return "z.boolean()";
    case "null":
      return "z.null()";
    case "array":
      if (schema.items) {
        return `z.array(${schemaToZod(schema.items, context)})`;
      } else if (schema.prefixItems) {
        return `z.tuple([${schema.prefixItems.map((item: any) => schemaToZod(item, context)).join(", ")}])`;
      } else {
        return "z.array(z.any())";
      }
    case "object":
      if (!schema.properties) return "z.object({})";
      const properties = schema.properties;
      const required = schema.required || [];
      const zProperties = Object.entries(properties).map(([key, value]) => {
        return `  ${key}: ${indentLastLines(schemaToZod(value, context))}${required.includes(key) ? "" : ".optional()"},`;
      });
      return `z.object({\n${zProperties.join("\n")}\n})`;
  }
  if ("$ref" in schema) {
    const ref = schema.$ref.split("/");
    if (ref[0] !== "#") {
      throw new Error("External references are not supported");
    }
    let res = context;
    for (const part of ref.slice(1)) {
      if (res[part]) {
        res = res[part];
      } else {
        throw new Error(`Reference ${schema.$ref} not found`);
      }
    }
    return schemaToZod(res, context);
  }
  if ("allOf" in schema) {
    return `z.intersection(${schema.allOf.map((s: any) => schemaToZod(s, context)).join(", ")})`;
  }
  if ("anyOf" in schema) {
    let results = schema.anyOf.map((s: any) => schemaToZod(s, context));
    return `z.union([${results.join(", ")}])`;
  }
  return "z.any()";
}

let data = JSON.parse(fs.readFileSync("types.json", "utf-8"));
let schemas = "import * as z from 'zod'\nimport { DateOrISO } from '@/client';\n\n";
let duplicates = new Set<string>();
for (const [key, value] of Object.entries(data)) {
  let path = key.split(".").slice(2);
  let name = path.pop()!;
  if (duplicates.has(name)) continue;
  duplicates.add(name);
  schemas += `export const Z${name} = z.lazy(() => ${schemaToZod(value, value, true)});\n`;
  schemas += `export type ${name} = z.infer<typeof Z${name}>;\n\n`;
}
fs.writeFileSync("generated_types.ts", schemas);
