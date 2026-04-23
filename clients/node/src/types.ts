import { Readable } from 'stream';
import * as generated from './generated_types';
import { ZFieldItem, ZRefObject, ZRowList } from './generated_types';
export * from './generated_types';
import * as z from 'zod';
import { inferFileInfo } from './mime';
import fs from 'fs';
import { zodToJsonSchema } from 'zod-to-json-schema';
//export * from "./schema_types";

// Schemas are accessed via `workflows.blocks.get(block_id).resolved_schemas`,
// not via step raw outputs. Step outputs only carry data/payload; user-declared
// block config schemas (`start_json` / `extract` / `function` / `api_call`) live
// on the block itself, and every other block's input/output schema is inferred
// and exposed under `resolved_schemas.input_schemas` / `resolved_schemas.output_schemas`.

export function dataArray<Schema extends z.ZodType<any, any, any>>(
  schema: Schema
): z.ZodType<z.output<Schema>[], z.ZodTypeDef, { data: z.input<Schema>[] }> {
  return z.object({ data: z.array(schema) }).transform((input) => input.data);
}

// Define types for circular references
export type Column = {
  type: 'column';
  size: number;
  items: (
    | Row
    | z.infer<typeof ZFieldItem>
    | z.infer<typeof ZRefObject>
    | z.infer<typeof ZRowList>
  )[];
  name?: string;
};

export type Row = {
  type: 'row';
  name?: string;
  items: (Column | z.infer<typeof ZFieldItem> | z.infer<typeof ZRefObject>)[];
};

export const ZColumn: z.ZodType<Column> = z.lazy(() =>
  z.object({
    type: z.literal('column'),
    size: z.number(),
    items: z.array(z.union([ZRow, ZFieldItem, ZRefObject, ZRowList])),
    name: z.string().optional(),
  })
);

export const ZRow: z.ZodType<Row> = z.lazy(() =>
  z.object({
    type: z.literal('row'),
    name: z.string().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
  })
);

export const ZMIMEData = z
  .union([z.string(), z.instanceof(Buffer), z.instanceof(Readable), generated.ZMIMEData])
  .transform(async (input, ctx) => {
    try {
      if (typeof input === 'object' && input !== null && 'url' in input && 'filename' in input) {
        return input as any;
      }
      return await inferFileInfo(input as any);
    } catch (error: any) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Failed to infer MIME data: ' + error.message,
        fatal: true,
      });
      return z.NEVER;
    }
  });
export type MIMEDataInput = z.input<typeof ZMIMEData>;

export const ZJSONSchema = z
  .union([z.string(), z.record(z.any()), z.instanceof(z.ZodType)])
  .transform(async (input, ctx) => {
    if (input instanceof z.ZodType) {
      return zodToJsonSchema(input, { target: 'openAi' }) as Record<string, any>;
    }
    if (typeof input === 'object') {
      return input;
    }
    try {
      return JSON.parse(await fs.promises.readFile(input, 'utf-8')) as Record<string, any>;
    } catch (error: any) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Error occured when reading JSCON schema: ' + error.message,
        fatal: true,
      });
      return z.NEVER;
    }
  });
export type JSONSchemaInput = z.input<typeof ZJSONSchema>;
export type JSONSchema = z.output<typeof ZJSONSchema>;

export const ZDocumentExtractRequest = z.object({
  // Keep everything except stream and document from generated types
  ...(({ stream, document, metadata, ...rest }) => rest)(
    generated.ZDocumentExtractRequest.schema.shape
  ),
  // Accept a single document (required)
  document: ZMIMEData,
  // Normalize json_schema inputs (paths/zod instances)
  json_schema: ZJSONSchema,
  // Make metadata optional with empty object default
  metadata: z.record(z.string(), z.string()).default({}),
});
export type DocumentExtractRequest = z.input<typeof ZDocumentExtractRequest>;

export const ZExtractionRequest = z.object({
  document: ZMIMEData,
  json_schema: ZJSONSchema,
  model: z.string().default('retab-small'),
  image_resolution_dpi: z.number().min(96).max(300).default(192),
  n_consensus: z.number().min(1).max(16).default(1),
  instructions: z.string().optional(),
  metadata: z.record(z.string(), z.string()).default({}),
  bust_cache: z.boolean().default(false),
});
export type ExtractionRequest = z.input<typeof ZExtractionRequest>;

export const ZLegacyExtractionConsensus = z.object({
  choices: z.array(z.record(z.any())).default([]),
  likelihoods: z.record(z.any()).nullable().optional(),
});
export type LegacyExtractionConsensus = z.infer<typeof ZLegacyExtractionConsensus>;

export const ZLegacyExtractionRecord = generated.ZExtraction.transform((raw) => {
  // Legacy fields (inference_settings, original_model, predictions, consensus_details,
  // likelihoods) may be present on older Extraction payloads but are no longer typed
  // on the current ZExtraction schema. Access them through an untyped view.
  const legacyRaw = raw as any;
  const inferenceSettings = legacyRaw.inference_settings ?? {
    model: legacyRaw.original_model ?? 'retab-small',
    image_resolution_dpi: 192,
    n_consensus: 1,
    chunking_keys: undefined,
  };
  return {
    ...raw,
    model: legacyRaw.model ?? inferenceSettings.model,
    image_resolution_dpi:
      legacyRaw.image_resolution_dpi ?? inferenceSettings.image_resolution_dpi,
    n_consensus: legacyRaw.n_consensus ?? inferenceSettings.n_consensus,
    chunking_keys: legacyRaw.chunking_keys ?? inferenceSettings.chunking_keys,
    output: legacyRaw.output ?? legacyRaw.predictions ?? {},
    consensus: {
      choices: (legacyRaw.consensus?.choices ?? legacyRaw.consensus_details ?? []).map(
        (choice: any) => {
          const data = choice?.data;
          return data && typeof data === 'object' && !Array.isArray(data) ? data : choice;
        }
      ),
      likelihoods: legacyRaw.consensus?.likelihoods ?? legacyRaw.likelihoods ?? null,
    },
  };
});
export type LegacyExtractionRecord = z.output<typeof ZLegacyExtractionRecord>;

export const ZRetabParsedChatCompletion = generated.ZRetabParsedChatCompletion.transform(
  (completion) => ({
    ...completion,
    data: completion.choices?.[0]?.message?.parsed ?? null,
    text: completion.choices?.[0]?.message?.content ?? null,
  })
);
export type RetabParsedChatCompletion = z.output<typeof ZRetabParsedChatCompletion>;

export const ZParseRequest = z.object({
  ...generated.ZParseRequest.schema.shape,
  document: ZMIMEData,
});
export type ParseRequest = z.input<typeof ZParseRequest>;

// Parse resource (stored document parse record returned by /v1/parses).
// Mirrors retab/types/parses.py on the Python SDK.
export const ZParseOutput = z.object({
  pages: z.array(z.string()),
  text: z.string(),
});
export type ParseOutput = z.infer<typeof ZParseOutput>;

export const ZParse = z
  .object({
    id: z.string(),
    file: generated.ZFileRef,
    model: z.string(),
    table_parsing_format: z.union([
      z.literal('markdown'),
      z.literal('yaml'),
      z.literal('html'),
      z.literal('json'),
    ]),
    image_resolution_dpi: z.number(),
    instructions: z.string().nullable().optional(),
    output: ZParseOutput,
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type Parse = z.infer<typeof ZParse>;

// ---------------------------------------------------------------------------
// Canonical resource types (stored records returned by the new resource
// routes). These mirror `retab/types/{classifications,splits,edits,extractions}.py`
// from the Python SDK.
// ---------------------------------------------------------------------------

// Classification resource
export const ZClassificationDecision = z.object({
  reasoning: z.string(),
  category: z.string(),
});
export type ClassificationDecision = z.infer<typeof ZClassificationDecision>;

export const ZClassificationConsensus = z.object({
  choices: z.array(ZClassificationDecision).default([]),
  likelihood: z.number().nullable().optional(),
});
export type ClassificationConsensus = z.infer<typeof ZClassificationConsensus>;

export const ZClassification = z
  .object({
    id: z.string(),
    file: generated.ZFileRef,
    model: z.string(),
    categories: z.array(generated.ZCategory),
    n_consensus: z.number().default(1),
    instructions: z.string().nullable().optional(),
    output: ZClassificationDecision,
    consensus: ZClassificationConsensus.default({ choices: [] }),
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type Classification = z.infer<typeof ZClassification>;

// Split resource
export const ZSplitSubdocument = z.object({
  name: z.string(),
  description: z.string().default(""),
  allow_multiple_instances: z.boolean().default(false),
});
export type SplitSubdocument = z.infer<typeof ZSplitSubdocument>;

export const ZSplitResult = z.object({
  name: z.string(),
  pages: z.array(z.number()),
});
export type SplitResult = z.infer<typeof ZSplitResult>;

export const ZSplitSubdocumentLikelihood = z.object({
  name: z.number().nullable().optional(),
  pages: z.array(z.number()).default([]),
});
export type SplitSubdocumentLikelihood = z.infer<typeof ZSplitSubdocumentLikelihood>;

export const ZSplitConsensus = z.object({
  likelihoods: z.array(ZSplitSubdocumentLikelihood).nullable().optional(),
  choices: z.array(z.array(ZSplitResult)).default([]),
});
export type SplitConsensus = z.infer<typeof ZSplitConsensus>;

export const ZSplit = z
  .object({
    id: z.string(),
    file: generated.ZFileRef,
    model: z.string(),
    subdocuments: z.array(ZSplitSubdocument),
    n_consensus: z.number().default(1),
    instructions: z.string().nullable().optional(),
    output: z.array(ZSplitResult),
    consensus: ZSplitConsensus.nullable().optional(),
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type Split = z.infer<typeof ZSplit>;

export const ZProcessingRequestOrigin = z.object({
  type: z.string(),
  id: z.string().nullable().optional(),
});
export type ProcessingRequestOrigin = z.infer<typeof ZProcessingRequestOrigin>;

// Partitions resource
export const ZPartitionChunk = z.object({
  key: z.string(),
  pages: z.array(z.number()).default([]),
});
export type PartitionChunk = z.infer<typeof ZPartitionChunk>;

export const ZPartitionChunkLikelihood = z.object({
  key: z.number().nullable().optional(),
  pages: z.array(z.number()).default([]),
});
export type PartitionChunkLikelihood = z.infer<typeof ZPartitionChunkLikelihood>;

export const ZPartitionConsensus = z.object({
  choices: z.array(z.array(ZPartitionChunk)).default([]),
  likelihoods: z.array(ZPartitionChunkLikelihood).nullable().optional(),
});
export type PartitionConsensus = z.infer<typeof ZPartitionConsensus>;

export const ZPartition = z
  .object({
    id: z.string(),
    file: generated.ZFileRef,
    model: z.string(),
    key: z.string(),
    instructions: z.string().default(""),
    n_consensus: z.number().default(1),
    output: z.array(ZPartitionChunk).default([]),
    consensus: ZPartitionConsensus.default({ choices: [] }),
    origin: ZProcessingRequestOrigin.nullable().optional(),
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type Partition = z.infer<typeof ZPartition>;

// Edit resource (canonical stored record from /v1/edits). NOTE: the generated
// `ZEditResponse` already exists and represents the one-shot legacy response
// (`form_data` + `filled_document`). This canonical `ZEdit` mirrors
// `retab/types/edits.py::Edit`.
export const ZEditResult = z.object({
  form_data: z.array(generated.ZFormField),
  filled_document: generated.ZMIMEData,
});
export type EditResult = z.infer<typeof ZEditResult>;

export const ZEdit = z
  .object({
    id: z.string(),
    file: generated.ZFileRef,
    model: z.string(),
    instructions: z.string(),
    config: generated.ZEditConfig,
    template_id: z.string().nullable().optional(),
    data: ZEditResult,
    usage: generated.ZRetabUsage,
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type Edit = z.infer<typeof ZEdit>;

// Extraction resource (new v2 shape from /v1/extractions). Mirrors
// `retab/types/extractions.py::Extraction`. The generated `ZExtraction` is the
// legacy Mongo shape and is kept available for backward
// compatibility as `LegacyExtractionRecord`.
export const ZExtractionConsensus = z.object({
  choices: z.array(z.record(z.any())).default([]),
  likelihoods: z.record(z.any()).nullable().optional(),
});
export type ExtractionConsensus = z.infer<typeof ZExtractionConsensus>;

export const ZExtractionV2 = z
  .object({
    id: z.string(),
    file: generated.ZFileRef,
    model: z.string(),
    json_schema: z.record(z.any()),
    n_consensus: z.number().default(1),
    image_resolution_dpi: z.number().default(192),
    instructions: z.string().nullable().optional(),
    output: z.record(z.any()),
    consensus: ZExtractionConsensus.default({ choices: [] }),
    origin: ZProcessingRequestOrigin.nullable().optional(),
    metadata: z.record(z.string()).default({}),
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
    organization_id: z.string().nullable().optional(),
  })
  .passthrough();
export type ExtractionV2 = z.infer<typeof ZExtractionV2>;

export const ZGenerateSchemaRequest = z.object({
  ...generated.ZGenerateSchemaRequest.schema.shape,
  documents: ZMIMEData.array(),
});
export type GenerateSchemaRequest = z.input<typeof ZGenerateSchemaRequest>;

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

function normalizeClassifyDecision(
  value: unknown
): z.input<typeof generated.ZClassifyDecision> | undefined {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    return undefined;
  }

  const record = value as Record<string, unknown>;
  if (typeof record.reasoning === 'string' && typeof record.category === 'string') {
    return {
      reasoning: record.reasoning,
      category: record.category,
    };
  }

  if (typeof record.reasoning === 'string' && typeof record.classification === 'string') {
    return {
      reasoning: record.reasoning,
      category: record.classification,
    };
  }

  if (typeof record.classification === 'object' && record.classification !== null) {
    return normalizeClassifyDecision(record.classification);
  }

  return undefined;
}

function normalizeClassifyChoices(
  value: unknown
): z.input<typeof generated.ZClassifyChoice>[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value.flatMap((choice) => {
    const normalized = normalizeClassifyDecision(choice);
    return normalized ? [{ classification: normalized }] : [];
  });
}

function normalizeClassifyResponsePayload(payload: unknown): unknown {
  if (typeof payload !== 'object' || payload === null || Array.isArray(payload)) {
    return payload;
  }

  const record = payload as Record<string, unknown>;
  const classification =
    normalizeClassifyDecision(record.classification) ?? normalizeClassifyDecision(record.result);
  if (!classification) {
    return payload;
  }

  const consensusRecord =
    typeof record.consensus === 'object' &&
    record.consensus !== null &&
    !Array.isArray(record.consensus)
      ? (record.consensus as Record<string, unknown>)
      : undefined;

  const likelihood = consensusRecord?.likelihood ?? record.likelihood;
  const choices = normalizeClassifyChoices(consensusRecord?.choices ?? record.votes);

  return {
    classification,
    consensus: {
      choices,
      ...(likelihood !== undefined ? { likelihood } : {}),
    },
    usage: record.usage,
  };
}

export const ZClassifyResponse = z.preprocess(
  (payload) => normalizeClassifyResponsePayload(payload),
  z.object({
    ...generated.ZClassifyResponse.schema.shape,
    consensus: generated.ZClassifyResponse.schema.shape.consensus.default({ choices: [] }),
  })
);
export type ClassifyResponse = z.infer<typeof ZClassifyResponse>;

export const ZEditRequest = z.object({
  ...generated.ZEditRequest.schema.shape,
  document: ZMIMEData.nullable().optional(),
  config: generated.ZEditConfig.optional().default({ color: '#000080' }),
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

export const ZWorkflowRunStep = z
  .object({
    run_id: z.string(),
    organization_id: z.string(),
    block_id: z.string(),
    step_id: z.string(),
    block_type: z.string(),
    block_label: z.string(),
    status: z.string(),
    started_at: z.string().nullable().optional(),
    completed_at: z.string().nullable().optional(),
    duration_ms: z.number().nullable().optional(),
    error: z.string().nullable().optional(),
    artifact: generated.ZStepArtifactRef.nullable().optional(),
    handle_outputs: z.record(z.string(), z.any()).nullable().optional(),
    handle_inputs: z.record(z.string(), z.any()).nullable().optional(),
    requires_human_review: z.boolean().nullable().optional(),
    human_reviewed_at: z.string().nullable().optional(),
    human_review_approved: z.boolean().nullable().optional(),
    retry_count: z.number().nullable().optional(),
    loop_id: z.string().nullable().optional(),
    iteration: z.number().nullable().optional(),
    created_at: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowRunStep = z.infer<typeof ZWorkflowRunStep>;

export const ZWorkflow = z
  .object({
    id: z.string(),
    name: z.string().default('Untitled Workflow'),
    description: z.string().default(''),
    organization_id: z.string().nullable().optional(),
    published: z
      .object({
        snapshot_id: z.string().nullable().optional(),
        published_at: z.string().nullable().optional(),
      })
      .nullable()
      .default(null),
    email_trigger: z
      .object({
        allowed_senders: z.array(z.string()).default([]),
        allowed_domains: z.array(z.string()).default([]),
      })
      .default({
        allowed_senders: [],
        allowed_domains: [],
      }),
    created_at: z.string(),
    updated_at: z.string(),
  })
  .passthrough();
export type Workflow = z.infer<typeof ZWorkflow>;

// ---------------------------------------------------------------------------
// Workflow graph types (blocks, edges, subflows)
// ---------------------------------------------------------------------------

export const ZResolvedSchemas = z
  .object({
    input_schemas: z.record(z.string(), z.any()).default({}),
    output_schemas: z.record(z.string(), z.any()).default({}),
    _field_ref_drift: z.record(z.any()).nullable().optional(),
  })
  .passthrough();
export type ResolvedSchemas = z.infer<typeof ZResolvedSchemas>;

export const ZWorkflowBlock = z
  .object({
    id: z.string(),
    workflow_id: z.string(),
    organization_id: z.string(),
    draft_version: z.string().nullable().optional(),
    type: z.string(),
    label: z.string().default(''),
    position_x: z.number().default(0),
    position_y: z.number().default(0),
    width: z.number().nullable().optional(),
    height: z.number().nullable().optional(),
    config: z.record(z.any()).nullable().optional(),
    parent_id: z.string().nullable().optional(),
    resolved_schemas: ZResolvedSchemas.nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowBlock = z.infer<typeof ZWorkflowBlock>;

export const ZWorkflowEdgeDoc = z
  .object({
    id: z.string(),
    workflow_id: z.string(),
    organization_id: z.string(),
    draft_version: z.string().nullable().optional(),
    source_block: z.string(),
    target_block: z.string(),
    source_handle: z.string().nullable().optional(),
    target_handle: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowEdgeDoc = z.infer<typeof ZWorkflowEdgeDoc>;

export const ZWorkflowSubflow = z
  .object({
    id: z.string(),
    workflow_id: z.string(),
    organization_id: z.string(),
    draft_version: z.string().nullable().optional(),
    type: z.string(),
    label: z.string().default(''),
    position_x: z.number().default(0),
    position_y: z.number().default(0),
    width: z.number().default(400),
    height: z.number().default(300),
    config: z.record(z.any()).nullable().optional(),
    child_block_ids: z.array(z.string()).default([]),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowSubflow = z.infer<typeof ZWorkflowSubflow>;

export const ZWorkflowWithEntities = z
  .object({
    workflow: ZWorkflow,
    blocks: z.array(ZWorkflowBlock).default([]),
    edges: z.array(ZWorkflowEdgeDoc).default([]),
    subflows: z.array(ZWorkflowSubflow).default([]),
  })
  .passthrough();
export type WorkflowWithEntities = z.infer<typeof ZWorkflowWithEntities>;

export type WorkflowRunStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'error'
  | 'waiting_for_human'
  | 'cancelled';

export type WorkflowRunTriggerType =
  | 'manual'
  | 'api'
  | 'schedule'
  | 'webhook'
  | 'email'
  | 'restart';

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

export const ZWorkflowRunExportResponse = z
  .object({
    csv_data: z.string(),
    rows: z.number(),
    columns: z.number(),
  })
  .passthrough();
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
      `Workflow run ${run.id} ${run.status === 'cancelled' ? 'was cancelled' : 'failed'}${run.error ? `: ${run.error}` : ''}`
    );
    this.name = 'WorkflowRunError';
    this.run = run;
  }
}

/**
 * Throw a {@link WorkflowRunError} if the run did not succeed.
 * Modelled after `httpx.Response.raise_for_status()`.
 */
export function raiseForStatus(run: generated.WorkflowRun): void {
  if (run.status === 'error' || run.status === 'cancelled') throw new WorkflowRunError(run);
}

export const ZModel = z.lazy(() =>
  z.object({
    id: z.string(),
    created: z.number(),
    object: z.literal('model'),
    owned_by: z.string(),
  })
);
export type Model = z.infer<typeof ZModel>;
