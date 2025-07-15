import {ZFieldItem, FieldItem, ZRefObject, RefObject, ZRowList, RowList} from "./generated_types";
export * from "./generated_types";
import * as z from "zod";

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
