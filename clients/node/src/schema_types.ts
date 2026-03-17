import { z } from "zod";

export const ZSchemaGenerationResponse = z.lazy(() => (z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
    strict: z.boolean().default(true),
})));
export type SchemaGenerationResponse = z.infer<typeof ZSchemaGenerationResponse>;

export const ZSchema = z.lazy(() => (ZSchemaGenerationResponse.schema).merge(z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
})));
export type Schema = z.infer<typeof ZSchema>;