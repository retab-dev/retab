import { Readable } from "stream";
import * as generated from "./generated_types";
import {ZFieldItem, FieldItem, ZRefObject, RefObject, ZRowList, RowList} from "./generated_types";
export * from "./generated_types";
import * as z from "zod";
import { inferFileInfo } from "./mime";
import fs from "fs";
import { zodToJsonSchema } from 'zod-to-json-schema';

export function dataArray<Schema extends z.ZodType<any, any, any>>(schema: Schema): z.ZodType<
    z.output<Schema>[],
    any,
    {data: z.input<Schema>[]}
> {
    return z.object({data: z.array(schema)}).transform((input) => input.data);
}

export const ZColumn: z.ZodType<{
    type: "column";
    size: number;
    items: (z.infer<typeof ZRow> | FieldItem | RefObject | RowList)[];
    name?: string | undefined;
}> = z.lazy(() => z.object({
    type: z.literal("column"),
    size: z.number(),
    items: z.array(z.union([ZRow, ZFieldItem, ZRefObject, ZRowList])),
    name: z.string().optional(),
}));
export type Column = z.infer<typeof ZColumn>;

export const ZRow: z.ZodType<{
    type: "row";
    name?: string | undefined;
    items: (z.infer<typeof ZColumn> | FieldItem | RefObject)[];
}> = z.lazy(() => z.object({
    type: z.literal("row"),
    name: z.string().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
}));
export type Row = z.infer<typeof ZRow>;

export const ZMIMEData = z.union([
    z.string(),
    z.instanceof(Buffer),
    z.instanceof(Readable),
    generated.ZMIMEData,
]).transform(async (input, ctx) => {
    try {
        if (typeof input === "object" && "url" in input && "filename" in input) {
            return input;
        }
        return await inferFileInfo(input);
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
        return zodToJsonSchema(input, {target: "openAi"}) as Record<string, any>;
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
    ...(({document, stream, ...rest}) => rest)(generated.ZDocumentExtractRequest.schema.shape),
    documents: z.array(ZMIMEData),
    json_schema: ZJSONSchema,
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

export const ZBaseProject = z.object({
    ...generated.ZBaseProject.schema.shape,
    json_schema: ZJSONSchema,
});
export type BaseProjectInput = z.input<typeof ZBaseProject>;
export type BaseProject = z.output<typeof ZBaseProject>;

export const ZDocumentItem = z.object({
    ...generated.ZDocumentItem.schema.shape,
    mime_data: ZMIMEData,
});
export type DocumentItemInput = z.input<typeof ZDocumentItem>;
export type DocumentItem = z.output<typeof ZDocumentItem>;
