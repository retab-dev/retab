import { Readable } from "stream";
import * as generated from "./generated_types";
import {ZFieldItem, FieldItem, ZRefObject, RefObject, ZRowList, RowList} from "./generated_types";
export * from "./generated_types";
import * as z from "zod";
import { inferFileInfo } from "./mime";

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

export const ZDocumentExtractRequest = z.object({
    ...(({document, ...rest}) => rest)(generated.ZDocumentExtractRequest.schema.shape),
    documents: z.array(ZMIMEData),
})
export type DocumentExtractRequest = z.input<typeof ZDocumentExtractRequest>;
