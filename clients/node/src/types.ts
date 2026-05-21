import { Readable } from 'stream';
import * as generated from './generated_types';
import { ZFieldItem, ZRefObject, ZRowList } from './generated_types';
export * from './generated_types';
import * as z from 'zod';
import { inferFileInfo, withMimeDataProperties } from './mime';
import fs from 'fs';
import { zodToJsonSchema } from 'zod-to-json-schema';
//export * from "./schema_types";

// Graph-derived block schemas are exposed through
// `workflows.getResolvedSchemas(workflowId)` and
// `workflows.blocks.getResolvedSchemas(workflowId, blockId)`. They are not
// embedded in public block objects.

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
        return withMimeDataProperties(input as generated.MIMEData);
      }
      return withMimeDataProperties(await inferFileInfo(input as any));
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
export type MIMEData = generated.MIMEData & { readonly id?: string };

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
    image_resolution_dpi: legacyRaw.image_resolution_dpi ?? inferenceSettings.image_resolution_dpi,
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
  })
  .strip();
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
  })
  .strip();
export type Classification = z.infer<typeof ZClassification>;

// Split resource
export const ZSplitSubdocument = z.object({
  name: z.string(),
  description: z.string().default(''),
  partition_key: z.string().nullable().optional(),
  allow_overlap: z.boolean().default(true),
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
  })
  .strip();
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
    instructions: z.string().default(''),
    n_consensus: z.number().default(1),
    allow_overlap: z.boolean().default(true),
    output: z.array(ZPartitionChunk).default([]),
    consensus: ZPartitionConsensus.default({ choices: [] }),
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
  })
  .strip();
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
  })
  .strip();
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
    metadata: z.record(z.string()).default({}),
    usage: generated.ZRetabUsage.nullable().optional(),
    created_at: z.string().nullable().optional(),
  })
  .strip();
export type ExtractionV2 = z.infer<typeof ZExtractionV2>;

export const ZGenerateSchemaRequest = z.object({
  ...generated.ZGenerateSchemaRequest.schema.shape,
  documents: ZMIMEData.array(),
});
export type GenerateSchemaRequest = z.input<typeof ZGenerateSchemaRequest>;

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

function normalizeClassifyChoices(value: unknown): z.input<typeof generated.ZClassifyChoice>[] {
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

// ---------------------------------------------------------------------------
// BREAKING CHANGES (workflow step artifact + StepStatus lifecycle cutover)
// ---------------------------------------------------------------------------
// All step shapes (`StepStatus` / `StepExecutionResponse` /
// `WorkflowRunStep`) now share `StepCore`. Step state is a single
// discriminated `lifecycle` payload. `status` and `terminal` are removed.
// `iteration_context` is replaced by a flat `loop_containers:
// ContainerContextData[]`.
// Step state is now read from `step.lifecycle`. Callers that read
// `step.metadata.evaluations` should fetch the backing record via `step.artifact`.
// ---------------------------------------------------------------------------
const ZHandlePayloadRecord = z.preprocess(
  (value) => (value === null ? {} : value),
  z.record(z.string(), generated.ZHandlePayload).default({})
);

export const ZWorkflowRunStep = z.object({
  // StepCore fields
  block_id: z.string(),
  step_id: z.string().default(''),
  block_type: z.string(),
  block_label: z.string(),
  lifecycle: generated.ZStepLifecycle,
  started_at: z.string().nullable().optional(),
  completed_at: z.string().nullable().optional(),
  loop_containers: z.array(generated.ZContainerContextData).default([]),
  model: z.string().nullable().optional(),
  // WorkflowRunStep extras
  run_id: z.string(),
  artifact: generated.ZStepArtifactRef.nullable().optional(),
  handle_outputs: ZHandlePayloadRecord,
  handle_inputs: ZHandlePayloadRecord,
  retry_count: z.number().default(0),
  created_at: z.string().nullable().optional(),
});
export type WorkflowRunStep = z.infer<typeof ZWorkflowRunStep>;

export const ZStepExecutionResponse = z.object({
  block_id: z.string(),
  step_id: z.string().default(''),
  block_type: z.string(),
  block_label: z.string(),
  lifecycle: generated.ZStepLifecycle,
  started_at: z.string().nullable().optional(),
  completed_at: z.string().nullable().optional(),
  model: z.string().nullable().optional(),
  loop_containers: z.array(generated.ZContainerContextData).default([]),
  artifact: generated.ZStepArtifactRef.nullable().optional(),
  handle_outputs: ZHandlePayloadRecord,
  handle_inputs: ZHandlePayloadRecord,
});
export type StepExecutionResponse = z.infer<typeof ZStepExecutionResponse>;

export const ZWorkflowArtifact = z
  .object({
    operation: z.string(),
    id: z.string(),
  })
  .passthrough();
export type WorkflowArtifact = z.infer<typeof ZWorkflowArtifact>;

export const ZWorkflow = z
  .object({
    id: z.string(),
    name: z.string().default('Untitled Workflow'),
    description: z.string().default(''),
    published: z
      .object({
        version_id: z.string().nullable().optional(),
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
// Workflow graph types (blocks, edges)
// ---------------------------------------------------------------------------

export const ZResolvedSchemas = z
  .object({
    input_schemas: z.record(z.string(), z.any()).default({}),
    output_schemas: z.record(z.string(), z.any()).default({}),
    field_ref_drift: z.record(z.any()).nullable().optional(),
  })
  .passthrough();
export type ResolvedSchemas = z.infer<typeof ZResolvedSchemas>;

export const ZWorkflowBlock = z.object({
  id: z.string(),
  workflow_id: z.string(),
  draft_version: z.string().nullable().optional(),
  type: z.string(),
  label: z.string().default(''),
  position_x: z.number().default(0),
  position_y: z.number().default(0),
  width: z.number().nullable().optional(),
  height: z.number().nullable().optional(),
  config: z.record(z.any()).nullable().optional(),
  field_ref_snapshot: z.record(z.string(), z.string()).nullable().optional(),
  parent_id: z.string().nullable().optional(),
  updated_at: z.string().nullable().optional(),
});
export type WorkflowBlock = z.infer<typeof ZWorkflowBlock>;

export const ZWorkflowEdgeDoc = z
  .object({
    id: z.string(),
    workflow_id: z.string(),
    draft_version: z.string().nullable().optional(),
    source_block: z.string(),
    target_block: z.string(),
    source_handle: z.string().nullable().optional(),
    target_handle: z.string().nullable().optional(),
    updated_at: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowEdgeDoc = z.infer<typeof ZWorkflowEdgeDoc>;

export const ZWorkflowWithEntities = z
  .object({
    workflow: ZWorkflow,
    blocks: z.array(ZWorkflowBlock).default([]),
    edges: z.array(ZWorkflowEdgeDoc).default([]),
  })
  .passthrough();
export type WorkflowWithEntities = z.infer<typeof ZWorkflowWithEntities>;

export const ZWorkflowResolvedSchemasResponse = z
  .object({
    workflow_id: z.string(),
    draft_version: z.string().nullable().optional(),
    schemas: z.record(z.string(), ZResolvedSchemas).default({}),
  })
  .passthrough();
export type WorkflowResolvedSchemasResponse = z.infer<typeof ZWorkflowResolvedSchemasResponse>;

export const ZBlockResolvedSchemasResponse = z
  .object({
    workflow_id: z.string(),
    block_id: z.string(),
    draft_version: z.string().nullable().optional(),
    schema: ZResolvedSchemas,
  })
  .passthrough();
export type BlockResolvedSchemasResponse = z.infer<typeof ZBlockResolvedSchemasResponse>;

export const ZDeclarativePlanSummary = z
  .object({
    add: z.number().default(0),
    change: z.number().default(0),
    destroy: z.number().default(0),
    replace: z.number().default(0),
    noop: z.number().default(0),
    total: z.number().default(0),
    has_changes: z.boolean().default(false),
  })
  .passthrough();
export type DeclarativePlanSummary = z.infer<typeof ZDeclarativePlanSummary>;

export const ZDeclarativePlanFieldChange = z
  .object({
    path: z.array(z.union([z.string(), z.number()])),
    path_display: z.string(),
    action: z.string(),
    before: z.any().optional().nullable(),
    after: z.any().optional().nullable(),
    before_sensitive: z.boolean().default(false),
    after_sensitive: z.boolean().default(false),
    unified_diff: z.string().optional().nullable(),
  })
  .passthrough();
export type DeclarativePlanFieldChange = z.infer<typeof ZDeclarativePlanFieldChange>;

export const ZDeclarativePlanChange = z
  .object({
    before: z.any().optional().nullable(),
    after: z.any().optional().nullable(),
    before_sensitive: z.any().default({}),
    after_sensitive: z.any().default({}),
    field_changes: z.array(ZDeclarativePlanFieldChange).default([]),
  })
  .passthrough();
export type DeclarativePlanChange = z.infer<typeof ZDeclarativePlanChange>;

export const ZDeclarativePlanResourceChange = z
  .object({
    address: z.string(),
    target: z.string(),
    target_id: z.string(),
    name: z.string(),
    type: z.string(),
    actions: z.array(z.string()),
    summary: z.string(),
    change: ZDeclarativePlanChange,
    path: z.string().optional().nullable(),
  })
  .passthrough();
export type DeclarativePlanResourceChange = z.infer<typeof ZDeclarativePlanResourceChange>;

export const ZDeclarativeValidationResponse = z
  .object({
    workflow_id: z.string(),
    block_count: z.number(),
    edge_count: z.number(),
    is_valid: z.boolean(),
    diagnostics: z.record(z.any()),
  })
  .passthrough();
export type DeclarativeValidationResponse = z.infer<typeof ZDeclarativeValidationResponse>;

export const ZDeclarativePlanResponse = z
  .object({
    workflow_id: z.string(),
    action: z.string(),
    block_count: z.number(),
    edge_count: z.number(),
    diagnostics: z.record(z.any()),
    format_version: z.string().default('workflows-plan/v1'),
    summary: ZDeclarativePlanSummary.default({
      add: 0,
      change: 0,
      destroy: 0,
      replace: 0,
      noop: 0,
      total: 0,
      has_changes: false,
    }),
    resource_changes: z.array(ZDeclarativePlanResourceChange).default([]),
    rendered_plan: z.string().default('No changes. Infrastructure is up-to-date.'),
  })
  .passthrough();
export type DeclarativePlanResponse = z.infer<typeof ZDeclarativePlanResponse>;

export const ZDeclarativeApplyResponse = z
  .object({
    workflow_id: z.string(),
    created: z.boolean(),
    block_count: z.number(),
    edge_count: z.number(),
    diagnostics: z.record(z.any()),
    format_version: z.string().default('workflows-plan/v1'),
    summary: ZDeclarativePlanSummary.default({
      add: 0,
      change: 0,
      destroy: 0,
      replace: 0,
      noop: 0,
      total: 0,
      has_changes: false,
    }),
    resource_changes: z.array(ZDeclarativePlanResourceChange).default([]),
    rendered_plan: z.string().default('No changes. Infrastructure is up-to-date.'),
  })
  .passthrough();
export type DeclarativeApplyResponse = z.infer<typeof ZDeclarativeApplyResponse>;

export const ZDeclarativeExportResponse = z
  .object({
    workflow_id: z.string(),
    yaml_definition: z.string(),
  })
  .passthrough();
export type DeclarativeExportResponse = z.infer<typeof ZDeclarativeExportResponse>;

export type WorkflowRunStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'error'
  | 'awaiting_review'
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
// Workflow diagnose-graph response (POST /workflows/{id}/diagnose-graph)
// ---------------------------------------------------------------------------

export const ZWorkflowDiagnosisIssue = z
  .object({
    severity: z.enum(['error', 'warning', 'info']),
    code: z.string(),
    message: z.string(),
    block_id: z.string().nullable().optional(),
  })
  .passthrough();
export type WorkflowDiagnosisIssue = z.infer<typeof ZWorkflowDiagnosisIssue>;

export const ZWorkflowDiagnosisStats = z
  .object({
    total_blocks: z.number().default(0),
    total_edges: z.number().default(0),
    block_types: z.record(z.number()).default({}),
    start_document_blocks: z.number().default(0),
  })
  .passthrough();
export type WorkflowDiagnosisStats = z.infer<typeof ZWorkflowDiagnosisStats>;

export const ZWorkflowDiagnosisResponse = z
  .object({
    is_valid: z.boolean(),
    issues: z.array(ZWorkflowDiagnosisIssue).default([]),
    suggestions: z.array(z.string()).default([]),
    stats: ZWorkflowDiagnosisStats.default({
      total_blocks: 0,
      total_edges: 0,
      block_types: {},
      start_document_blocks: 0,
    }),
  })
  .passthrough();
export type WorkflowDiagnosisResponse = z.infer<typeof ZWorkflowDiagnosisResponse>;

// ---------------------------------------------------------------------------
// Block simulation (POST /workflows/runs/{run_id}/steps/{block_id}/simulate)
// ---------------------------------------------------------------------------

export const ZStepArtifactRef = z
  .object({
    operation: z.string(),
    id: z.string(),
  })
  .passthrough();
export type StepArtifactRef = z.infer<typeof ZStepArtifactRef>;

export const ZBlockSimulation = z
  .object({
    id: z.string(),
    workflow_id: z.string(),
    run_id: z.string(),
    block_id: z.string(),
    block_type: z.string(),
    success: z.boolean(),
    handle_inputs: z.record(z.any()).nullable().optional(),
    artifact: ZStepArtifactRef.nullable().optional(),
    handle_outputs: z.record(z.any()).nullable().optional(),
    routing_decision: z.array(z.string()).nullable().optional(),
    error: z.string().nullable().optional(),
    duration_ms: z.number().nullable().optional(),
    skipped: z.boolean().default(false),
    created_at: z.string().nullable().optional(),
    block_config: z.record(z.any()).nullable().optional(),
    step_id: z.string().nullable().optional(),
    available_iterations: z.array(z.record(z.any())).nullable().optional(),
  })
  .passthrough();
export type BlockSimulation = z.infer<typeof ZBlockSimulation>;

// ---------------------------------------------------------------------------
// Workflow reviews (served under /workflows/reviews)
//
// A review record attached to a gated workflow block run. Actor-neutral by
// construction: a proposal authored by a model, an agent, or a human all flow
// through the SAME types and methods. `Actor.kind` is descriptive data — code
// never branches on it.
// ---------------------------------------------------------------------------

/** Who performed a review action. `kind` is descriptive only — never branched on. */
export const ZActor = z
  .object({
    kind: z.enum(['model', 'agent', 'human']),
    id: z.string(),
    display_name: z.string(),
  })
  .passthrough();
export type Actor = z.infer<typeof ZActor>;

/** One immutable snapshot of the block output in the version history. */
export const ZOutputVersion = z
  .object({
    parent_id: z.string().nullable().default(null),
    author: ZActor,
    origin: z.enum(['model_output', 'agent_created', 'human_created']),
    snapshot: z.record(z.unknown()),
    note: z.string().nullable().default(null),
    created_at: z.string(),
  })
  .passthrough();
export type OutputVersion = z.infer<typeof ZOutputVersion>;

/** A verdict recorded against a specific output version. */
export const ZReviewDecision = z
  .object({
    verdict: z.enum(['approved', 'rejected']),
    version_id: z.string(),
    decided_by: ZActor,
    decided_at: z.string(),
    reason: z.string().nullable().default(null),
  })
  .passthrough();
export type ReviewDecision = z.infer<typeof ZReviewDecision>;

/** The full versioned review for one gated block run.
 *
 * The backend strips `organization_id` and `runtime_block_id` from public
 * responses, so neither is modeled here.
 */
export const ZReviewOverlay = z
  .object({
    _id: z.string(),
    workflow_id: z.string(),
    workflow_version_id: z.string(),
    workflow_run_id: z.string(),
    block_id: z.string(),
    block_run_id: z.string(),
    block_type: z.enum(['extract', 'classifier', 'split', 'for_each']),
    triggered_by: z.record(z.unknown()),
    awaiting_since: z.string(),
    priority: z.number(),
    versions_by_id: z.record(ZOutputVersion),
    decision: ZReviewDecision.nullable().default(null),
  })
  .strip();
export type ReviewOverlay = z.infer<typeof ZReviewOverlay>;

/** A lightweight review summary returned by `reviews.list(...)` — no version history. */
export const ZReviewQueueItem = z
  .object({
    _id: z.string(),
    workflow_id: z.string(),
    workflow_version_id: z.string(),
    workflow_run_id: z.string(),
    block_id: z.string(),
    block_run_id: z.string(),
    block_type: z.enum(['extract', 'classifier', 'split', 'for_each']),
    triggered_by: z.record(z.unknown()),
    awaiting_since: z.string(),
    priority: z.number(),
  })
  .strip();
export type ReviewQueueItem = z.infer<typeof ZReviewQueueItem>;

/** Boundary resource IDs for page navigation. */
export const ZListMetadata = z
  .object({
    before: z.string().nullable(),
    after: z.string().nullable(),
  })
  .passthrough();
export type ListMetadata = z.infer<typeof ZListMetadata>;

/** Envelope returned by `reviews.list(...)`.
 *
 * Pages are cursored via `list_metadata.before` / `list_metadata.after`.
 * `has_more` is a derived convenience: `list_metadata.after !== null`.
 */
export const ZReviewQueueResponse = z
  .object({
    data: z.array(ZReviewQueueItem).default([]),
    list_metadata: ZListMetadata,
  })
  .passthrough();
export type ReviewQueueResponse = z.infer<typeof ZReviewQueueResponse>;

/** Status of the Temporal resume signal sent after a decision is committed.
 *
 * `"resumed"`: signal succeeded; the workflow run advanced past the gate.
 * `"skipped"`: defensive — there was no decision to resume (should not occur).
 * `"failed"`: the signal raised; the decision IS committed but the run did
 *   NOT advance. Inspect `resume_error` for context and retry via your own
 *   reconciliation path.
 */
export const ZResumeStatus = z.enum(['resumed', 'skipped', 'failed']);
export type ResumeStatus = z.infer<typeof ZResumeStatus>;

/**
 * Submission status returned by `reviews.approve/reject(...)`.
 *
 * - `"accepted"`: decision committed AND the Temporal resume signal succeeded.
 * - `"already_applied"`: idempotent replay of an identical `(verdict, version_id)`.
 * - `"accepted_pending_resume"`: decision committed, but the Temporal resume
 *   signal failed. The workflow has NOT advanced. Inspect `resume_error` and
 *   reconcile out-of-band.
 */
export const ZSubmissionStatus = z.enum(['accepted', 'already_applied', 'accepted_pending_resume']);
export type SubmissionStatus = z.infer<typeof ZSubmissionStatus>;

/** Envelope returned by `reviews.approve/reject(...)`. */
export const ZSubmitDecisionResponse = z
  .object({
    submission_status: ZSubmissionStatus,
    review: ZReviewOverlay,
    resume_status: ZResumeStatus.default('resumed'),
    resume_error: z.string().nullable().default(null),
  })
  .passthrough();
export type SubmitDecisionResponse = z.infer<typeof ZSubmitDecisionResponse>;
