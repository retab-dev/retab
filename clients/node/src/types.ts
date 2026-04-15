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

export const ZRetabParsedChatCompletion = generated.ZRetabParsedChatCompletion.transform((completion) => ({
    ...completion,
    data: completion.choices?.[0]?.message?.parsed ?? null,
    text: completion.choices?.[0]?.message?.content ?? null,
}));
export type RetabParsedChatCompletion = z.output<typeof ZRetabParsedChatCompletion>;

export const ZParseRequest = z.object({
    ...generated.ZParseRequest.schema.shape,
    document: ZMIMEData,
});
export type ParseRequest = z.input<typeof ZParseRequest>;

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
    config: generated.ZEditConfig.optional().default({ color: "#000080" }),
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

export const ZWorkflowRunStep = z.object({
    run_id: z.string(),
    organization_id: z.string(),
    node_id: z.string(),
    step_id: z.string(),
    node_type: z.string(),
    node_label: z.string(),
    status: z.string(),
    started_at: z.string().nullable().optional(),
    completed_at: z.string().nullable().optional(),
    duration_ms: z.number().nullable().optional(),
    error: z.string().nullable().optional(),
    handle_outputs: z.record(z.string(), z.any()).nullable().optional(),
    handle_inputs: z.record(z.string(), z.any()).nullable().optional(),
    input_document: generated.ZFileRef.nullable().optional(),
    output_document: generated.ZFileRef.nullable().optional(),
    split_documents: z.record(z.string(), generated.ZFileRef).nullable().optional(),
    requires_human_review: z.boolean().nullable().optional(),
    human_reviewed_at: z.string().nullable().optional(),
    human_review_approved: z.boolean().nullable().optional(),
    retry_count: z.number().nullable().optional(),
    loop_id: z.string().nullable().optional(),
    iteration: z.number().nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
}).passthrough();
export type WorkflowRunStep = z.infer<typeof ZWorkflowRunStep>;

export const ZWorkflow = z.object({
    id: z.string(),
    name: z.string().default("Untitled Workflow"),
    description: z.string().default(""),
    organization_id: z.string().nullable().optional(),
    published: z.object({
        snapshot_id: z.string().nullable().optional(),
        published_at: z.string().nullable().optional(),
    }).nullable().default(null),
    email_trigger: z.object({
        allowed_senders: z.array(z.string()).default([]),
        allowed_domains: z.array(z.string()).default([]),
    }).default({
        allowed_senders: [],
        allowed_domains: [],
    }),
    created_at: z.string(),
    updated_at: z.string(),
}).passthrough();
export type Workflow = z.infer<typeof ZWorkflow>;

// ---------------------------------------------------------------------------
// Workflow graph types (blocks, edges, subflows)
// ---------------------------------------------------------------------------

export const ZWorkflowBlock = z.object({
    id: z.string(),
    workflow_id: z.string(),
    organization_id: z.string(),
    draft_version: z.string().nullable().optional(),
    type: z.string(),
    label: z.string().default(""),
    position_x: z.number().default(0),
    position_y: z.number().default(0),
    width: z.number().nullable().optional(),
    height: z.number().nullable().optional(),
    config: z.record(z.any()).nullable().optional(),
    parent_id: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
}).passthrough();
export type WorkflowBlock = z.infer<typeof ZWorkflowBlock>;

export const ZWorkflowEdgeDoc = z.object({
    id: z.string(),
    workflow_id: z.string(),
    organization_id: z.string(),
    draft_version: z.string().nullable().optional(),
    source_block: z.string(),
    target_block: z.string(),
    source_handle: z.string().nullable().optional(),
    target_handle: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
}).passthrough();
export type WorkflowEdgeDoc = z.infer<typeof ZWorkflowEdgeDoc>;

export const ZWorkflowSubflow = z.object({
    id: z.string(),
    workflow_id: z.string(),
    organization_id: z.string(),
    draft_version: z.string().nullable().optional(),
    type: z.string(),
    label: z.string().default(""),
    position_x: z.number().default(0),
    position_y: z.number().default(0),
    width: z.number().default(400),
    height: z.number().default(300),
    config: z.record(z.any()).nullable().optional(),
    child_block_ids: z.array(z.string()).default([]),
    updated_at: z.string().nullable().optional(),
}).passthrough();
export type WorkflowSubflow = z.infer<typeof ZWorkflowSubflow>;

export const ZWorkflowWithEntities = z.object({
    workflow: ZWorkflow,
    blocks: z.array(ZWorkflowBlock).default([]),
    edges: z.array(ZWorkflowEdgeDoc).default([]),
    subflows: z.array(ZWorkflowSubflow).default([]),
}).passthrough();
export type WorkflowWithEntities = z.infer<typeof ZWorkflowWithEntities>;

export type WorkflowRunStatus =
    | "pending"
    | "running"
    | "completed"
    | "error"
    | "waiting_for_human"
    | "cancelled";

export type WorkflowRunTriggerType =
    | "manual"
    | "api"
    | "schedule"
    | "webhook"
    | "email"
    | "restart";

export type WorkflowBlockCreateRequest = {
    id: string;
    type: string;
    label?: string;
    positionX?: number;
    positionY?: number;
    width?: number;
    height?: number;
    config?: Record<string, unknown>;
    parentId?: string;
};

export type WorkflowBlockUpdateRequest = {
    label?: string;
    positionX?: number;
    positionY?: number;
    width?: number;
    height?: number;
    config?: Record<string, unknown>;
    parentId?: string;
};

export type WorkflowEdgeCreateRequest = {
    id: string;
    sourceBlock: string;
    targetBlock: string;
    sourceHandle?: string;
    targetHandle?: string;
};

export const ZWorkflowRunExportResponse = z.object({
    csv_data: z.string(),
    rows: z.number(),
    columns: z.number(),
}).passthrough();
export type WorkflowRunExportResponse = z.infer<typeof ZWorkflowRunExportResponse>;

// ---------------------------------------------------------------------------
// Workflow run utility functions
// ---------------------------------------------------------------------------

/**
 * Error thrown by {@link raiseForStatus} when a workflow run has failed.
 */
export class WorkflowRunError extends Error {
    public readonly run: generated.WorkflowRun;
    constructor(run: generated.WorkflowRun) {
        super(
            `Workflow run ${run.id} ${run.status === "cancelled" ? "was cancelled" : "failed"}${run.error ? `: ${run.error}` : ""}`
        );
        this.name = "WorkflowRunError";
        this.run = run;
    }
}

/**
 * Throw a {@link WorkflowRunError} if the run did not succeed.
 * Modelled after `httpx.Response.raise_for_status()`.
 */
export function raiseForStatus(run: generated.WorkflowRun): void {
    if (run.status === "error" || run.status === "cancelled") throw new WorkflowRunError(run);
}

export const ZModel = z.lazy(() => (z.object({
    id: z.string(),
    created: z.number(),
    object: z.literal("model"),
    owned_by: z.string(),
})));
export type Model = z.infer<typeof ZModel>;
