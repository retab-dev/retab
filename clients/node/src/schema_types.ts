import { z } from "zod";

export const ZPartialSchema = z.lazy(() => (z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
    strict: z.boolean().default(true),
})));
export type PartialSchema = z.infer<typeof ZPartialSchema>;

export const ZSchema = z.lazy(() => (ZPartialSchema.schema).merge(z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
})));
export type Schema = z.infer<typeof ZSchema>;