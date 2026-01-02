import { Readable } from "stream";
import * as generated from "./generated_types";
import { ZFieldItem, ZRefObject, ZRowList } from "./generated_types";
export * from "./generated_types";
import * as z from "zod";
import { inferFileInfo } from "./mime";
import fs from "fs";
import { zodToJsonSchema } from 'zod-to-json-schema';
//export * from "./schema_types";

export function dataArray<Schema extends z.ZodType<any, any, any>>(schema: Schema): z.ZodType<
    z.output<Schema>[],
    z.ZodTypeDef,
    { data: z.input<Schema>[] }
> {
    return z.object({ data: z.array(schema) }).transform((input) => input.data);
}

// Define types for circular references
export type Column = {
    type: "column";
    size: number;
    items: (Row | z.infer<typeof ZFieldItem> | z.infer<typeof ZRefObject> | z.infer<typeof ZRowList>)[];
    name?: string;
};

export type Row = {
    type: "row";
    name?: string;
    items: (Column | z.infer<typeof ZFieldItem> | z.infer<typeof ZRefObject>)[];
};

export const ZColumn: z.ZodType<Column> = z.lazy(() => z.object({
    type: z.literal("column"),
    size: z.number(),
    items: z.array(z.union([ZRow, ZFieldItem, ZRefObject, ZRowList])),
    name: z.string().optional(),
}));

export const ZRow: z.ZodType<Row> = z.lazy(() => z.object({
    type: z.literal("row"),
    name: z.string().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
}));

export const ZMIMEData = z.union([
    z.string(),
    z.instanceof(Buffer),
    z.instanceof(Readable),
    generated.ZMIMEData,
]).transform(async (input, ctx) => {
    try {
        if (typeof input === "object" && input !== null && "url" in input && "filename" in input) {
            return input as any;
        }
        return await inferFileInfo(input as any);
    } catch (error: any) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Failed to infer MIME data: " + error.message,
            fatal: true,
        });
        return z.NEVER;
    }
})
export type MIMEDataInput = z.input<typeof ZMIMEData>;

export const ZJSONSchema = z.union([
    z.string(),
    z.record(z.any()),
    z.instanceof(z.ZodType),
]).transform(async (input, ctx) => {
    if (input instanceof z.ZodType) {
        return zodToJsonSchema(input, { target: "openAi" }) as Record<string, any>;
    }
    if (typeof input === "object") {
        return input;
    }
    try {
        return JSON.parse(await fs.promises.readFile(input, "utf-8")) as Record<string, any>;
    } catch (error: any) {
        ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: "Error occured when reading JSCON schema: " + error.message,
            fatal: true,
        });
        return z.NEVER;
    }
})
export type JSONSchemaInput = z.input<typeof ZJSONSchema>;
export type JSONSchema = z.output<(typeof ZJSONSchema)>;

export const ZDocumentExtractRequest = z.object({
    // Keep everything except stream and document from generated types
    ...(({ stream, document, metadata, ...rest }) => rest)(generated.ZDocumentExtractRequest.schema.shape),
    // Accept a single document (required)
    document: ZMIMEData,
    // Normalize json_schema inputs (paths/zod instances)
    json_schema: ZJSONSchema,
    // Make metadata optional with empty object default
    metadata: z.record(z.string(), z.string()).default({}),
})
export type DocumentExtractRequest = z.input<typeof ZDocumentExtractRequest>;

export const ZParseRequest = z.object({
    ...generated.ZParseRequest.schema.shape,
    document: ZMIMEData,
});
export type ParseRequest = z.input<typeof ZParseRequest>;

export const ZDocumentCreateMessageRequest = z.object({
    ...generated.ZDocumentCreateMessageRequest.schema.shape,
    document: ZMIMEData,
});
export type DocumentCreateMessageRequest = z.input<typeof ZDocumentCreateMessageRequest>;

export const ZDocumentCreateInputRequest = z.object({
    ...generated.ZDocumentCreateInputRequest.schema.shape,
    document: ZMIMEData,
    json_schema: ZJSONSchema,
});
export type DocumentCreateInputRequest = z.input<typeof ZDocumentCreateInputRequest>;

export const ZGenerateSchemaRequest = z.object({
    ...generated.ZGenerateSchemaRequest.schema.shape,
    documents: ZMIMEData.array(),
});
export type GenerateSchemaRequest = z.input<typeof ZGenerateSchemaRequest>;

export const ZCreateProjectRequest = z.object({
    ...generated.ZCreateProjectRequest.schema.shape,
    json_schema: ZJSONSchema,
});
export type CreateProjectRequest = z.input<typeof ZCreateProjectRequest>;


export const ZSplitRequest = z.object({
    ...generated.ZSplitRequest.schema.shape,
    document: ZMIMEData,
});
export type SplitRequest = z.input<typeof ZSplitRequest>;

export const ZClassifyRequest = z.object({
    ...generated.ZClassifyRequest.schema.shape,
    document: ZMIMEData,
});
export type ClassifyRequest = z.input<typeof ZClassifyRequest>;

export const ZEditRequest = z.object({
    ...generated.ZEditRequest.schema.shape,
    document: ZMIMEData.nullable().optional(),
});
export type EditRequest = z.input<typeof ZEditRequest>;

export const ZEditResponse = generated.ZEditResponse;
export type EditResponse = z.infer<typeof ZEditResponse>;

export const ZInferFormSchemaRequest = z.object({
    ...generated.ZInferFormSchemaRequest.schema.shape,
    document: ZMIMEData,
});
export type InferFormSchemaRequest = z.input<typeof ZInferFormSchemaRequest>;

export const ZInferFormSchemaResponse = generated.ZInferFormSchemaResponse;
export type InferFormSchemaResponse = z.infer<typeof ZInferFormSchemaResponse>;

export const ZModel = z.lazy(() => (z.object({
    id: z.string(),
    created: z.number(),
    object: z.literal("model"),
    owned_by: z.string(),
})));
export type Model = z.infer<typeof ZModel>;


