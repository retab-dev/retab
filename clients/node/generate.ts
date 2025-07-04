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
  //let split = name.split(/[-_ ]/);
  //return split.length === 1 || (split.length === 2 && (split[1] === "Input" || split[1] === "Output"));
  return true;
}

function findNamedSchema(schema: any): string | undefined {
  let str = "#/components/schemas/";
  if (!("$ref" in schema) || !schema.$ref.startsWith(str)) return undefined;
  let name = schema.$ref.slice(str.length);
  if (!isNamedSchemaName(name)) return undefined;
  return capitalise(camelCase(name));
}

let schemaToTsImports: string[] = [];
function schemaToTs(schema: any, context: any, ignoreNamed: boolean = false): string {
  if (!ignoreNamed && findNamedSchema(schema)) {
    let res = findNamedSchema(schema)!;
    schemaToTsImports.push(res);
    return res;
  }
  switch (schema.type) {
    case "string":
      switch (schema.format) {
        case "date":
          return "Date";
        case "date-time":
          return "Date";
        case "uuid":
          return "string";
        case "binary":
          return "File";
        default:
          if (schema.enum) {
            return schema.enum.map(JSON.stringify).join(" | ");
          }
          return "string";
      }
    case "number":
    case "integer":
      return "number";
    case "boolean":
      return "boolean";
    case "null":
      return "null";
    case "array":
      if (schema.items) {
        return `${schemaToTs(schema.items, context)}[]`;
      } else if (schema.prefixItems) {
        return `[${schema.prefixItems.map((item: any) => schemaToTs(item, context)).join(", ")}]`;
      } else {
        return "any[]";
      }
    case "object":
      if (!schema.properties) return "object";
      const properties = schema.properties;
      const required = schema.required || [];
      const tsProperties = Object.entries(properties).map(([key, value]) => {
        return `  ${key}${required.includes(key) ? "" : "?"}: ${
          schemaToTs(value, context).split("\n")
            .map((line, i) => i === 0 ? line : `  ${line}`).join("\n")
        },`;
      });
      return `{\n${tsProperties.join("\n")}\n}`;
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
    return schemaToTs(res, context);
  }
  if ("allOf" in schema) {
    return schema.allOf.map((s: any) => schemaToTs(s, context)).join(" & ");
  }
  if ("anyOf" in schema) {
    return schema.anyOf.map((s: any) => schemaToTs(s, context)).join(" | ");
  }
  return "any";
}

function processSchema(schema: any) {

  let classStructure: Record<string, any> = {};

  for (const path of Object.keys(schema.paths)) {
    if (!path.startsWith("/v1")) continue;
    // @ts-ignore
    const methods = schema.paths[path];
    const multipleMethods = Object.keys(methods).length > 1;
    for (const method of Object.keys(methods)) {
      // @ts-ignore
      const operation = methods[method];
      let arrayPath = path.split("/").filter(Boolean).map(i => i.startsWith("{") ? i.slice(1, -1) : i).map(i => camelCase(i));
      arrayPath.push(method);
      schemaToTsImports = [];
      let functionDef = `async ${arrayPath[arrayPath.length - 1]}(`;
      let functionParams = [];
      let otherParamsNames = [];
      let otherParamsTypes = [];
      let otherParamTypesOptional = true;
      let bodyType = undefined;
      for (let param of operation.parameters || []) {
        if (param.in === "path") {
          functionParams.push(`${camelCase(param.name)}: ${schemaToTs(param.schema, schema)}`);
        } else {
          otherParamsNames.push(`${camelCase(param.name)}`);
          otherParamsTypes.push(`${camelCase(param.name)}${param.required ? "" : "?"}: ${schemaToTs(param.schema, schema)}`);
          if (param.required) otherParamTypesOptional = false;
        }
      }
      if (operation.requestBody) {
        otherParamsNames.push("...body");
        if (Object.keys(operation.requestBody.content).length !== 1) {
          throw new Error("Multiple content types are not supported");
        }
        let contentType = Object.keys(operation.requestBody.content)[0];
        bodyType = schemaToTs(operation.requestBody.content[contentType].schema, schema);
      }
      if (otherParamsNames.length > 0) {
        let types = [];
        if (otherParamsTypes.length > 0) {
          types.push(`{ ${otherParamsTypes.join(", ")} }`);
        }
        if (bodyType) {
          types.push(bodyType);
        }
        functionParams.push(`{ ${otherParamsNames.join(", ")} }: ${types.join(" & ")}${!bodyType && otherParamTypesOptional ? " = {}" : ""}`);
      }
      let returnTypes = Object.entries(operation.responses["200"].content).map(([contentType, value]: [string, any]) => {
        if (contentType === "application/json") {
          return schemaToTs(value.schema, schema);
        }
        if (contentType === "application/stream+json") {
          return `AsyncGenerator<${schemaToTs(value.schema, schema)}>`;
        }
        throw new Error(`Unsupported content type ${contentType}`);
      });
      functionDef += functionParams.join(", ") + `): Promise<${returnTypes.join(" | ")}> {\n`;
      functionDef += `  let res = await this._fetch({\n`;
      functionDef += `    url: \`${path.split("/").map(i => i.startsWith("{") ? `\${${camelCase(i.slice(1, -1))}}` : i).join("/")}\`,\n`;
      functionDef += `    method: "${method.toUpperCase()}",\n`;
      let queryParams = (operation.parameters || []).filter((p: any) => p.in === "query");
      if (queryParams.length > 0) {
        functionDef += `    params: { ${queryParams.map((p: any) => `${JSON.stringify(p.name)}: ${camelCase(p.name)}`).join(", ")} },\n`;
      }
      let headerParams = (operation.parameters || []).filter((p: any) => p.in === "header");
      if (headerParams.length > 0) {
        functionDef += `    headers: { ${headerParams.map((p: any) => `${JSON.stringify(p.name)}: ${camelCase(p.name)}`).join(", ")} },\n`;
      }
      if (bodyType) {
        functionDef += `    body: body,\n`;
        functionDef += `    bodyMime: "${Object.keys(operation.requestBody.content)[0]}",\n`;
      }
      if (operation.security) {
        functionDef += `    auth: [${operation.security.map((s: any) => JSON.stringify(Object.keys(s)[0])).join(", ")}],\n`;
      }
      functionDef += `  });\n`;
      Object.entries(operation.responses["200"].content).forEach(([contentType, value]) => {
        if (contentType === "application/json") {
          functionDef += `  if (res.headers.get("Content-Type") === "application/json") return res.json();\n`;
          return;
        }
        if (contentType === "application/stream+json") {
          functionDef += `  if (res.headers.get("Content-Type") === "application/stream+json") return streamResponse(res);\n`;
          return;
        }
      });
      functionDef += `  throw new Error("Bad content type");\n`;


      functionDef += `}\n`;
      arrayPath.reduce((acc, v, i) => {
        if (i === arrayPath.length - 1) {
          acc[v] = [functionDef, schemaToTsImports];
        }
        if (!acc[v]) {
          acc[v] = {};
        }
        return acc[v];
      }, classStructure);
    }
  }

  function generateClass(name: string, path: string, structure: Record<string, any>) {
    let typeImports = [];
    let imports = ["import { AbstractClient, CompositionClient, streamResponse } from '@/client';"];
    let classDef = `export default class API${capitalise(name)} extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }\n\n`;
    for (const [key, value] of Object.entries(structure)) {
      if (!Array.isArray(value)) {
        imports.push(`import API${capitalise(key)}Sub from "./${key}/client";`);
        classDef += `  ${key} = new API${capitalise(key)}Sub(this._client);\n`;
        generateClass(key, path + "/" + key, value);
      }
    }
    classDef += "\n";
    for (const [key, value] of Object.entries(structure)) {
      if (Array.isArray(value)) {
        let [code, codeImports] = value;
        typeImports.push(...codeImports);
        classDef += code.split("\n").map((line: string) => `  ${line}`).join("\n") + "\n";
      }
    }
    classDef += `}\n`;
    if (typeImports.length > 0) {
      imports.push(`import { ${typeImports.join(", ")} } from "@/types";`);
    }
    if (imports.length > 0) {
      classDef = imports.join("\n") + "\n\n" + classDef;
    }
    fs.mkdirSync(path, { recursive: true });
    fs.writeFileSync(`${path}/client.ts`, classDef);
  }

  generateClass("generated", "generated", classStructure);

  let schemas = "";
  for (const [key, value] of Object.entries(schema.components.schemas)) {
    if (isNamedSchemaName(key)) {
      schemas += `export type ${capitalise(camelCase(key))} = ${schemaToTs(value, schema, true)};\n\n`;
    }
  }
  fs.writeFileSync("generated/types.ts", schemas);
}

fetch("https://api.uiform.com/openapi.json").then(r => r.json()).then(processSchema);
