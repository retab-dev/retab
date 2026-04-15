import * as z from "zod";

import {
    ZFileRef,
    ZCategory,
    ZInferenceSettings,
    ZPredictionData,
    ZSubdocument,
} from "../../generated_types.js";

export const ZSplitProjectDraftConfig = z.object({
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    split_config: z.array(ZSubdocument).default([]),
    json_schema: z.record(z.any()).default({}),
    subdocuments: z.array(ZSubdocument).default([]),
}).passthrough();
export type SplitProjectDraftConfig = z.infer<typeof ZSplitProjectDraftConfig>;

export const ZSplitProjectPublishedConfig = ZSplitProjectDraftConfig.extend({
    origin: z.string().default("manual"),
}).passthrough();
export type SplitProjectPublishedConfig = z.infer<typeof ZSplitProjectPublishedConfig>;

export const ZSplitProject = z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    published_config: ZSplitProjectPublishedConfig,
    draft_config: ZSplitProjectDraftConfig,
    is_published: z.boolean().default(false),
}).passthrough();
export type SplitProject = z.infer<typeof ZSplitProject>;

export const ZCreateSplitProjectRequest = z.object({
    name: z.string(),
    split_config: z.array(ZSubdocument).default([]),
}).passthrough();
export type CreateSplitProjectRequest = z.input<typeof ZCreateSplitProjectRequest>;

export const ZPatchSplitProjectRequest = z.object({
    name: z.string().nullable().optional(),
    published_config: ZSplitProjectPublishedConfig.nullable().optional(),
    draft_config: ZSplitProjectDraftConfig.nullable().optional(),
    is_published: z.boolean().nullable().optional(),
}).passthrough();
export type PatchSplitProjectRequest = z.input<typeof ZPatchSplitProjectRequest>;

export const ZSplitBuilderDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    mime_data: ZFileRef,
    prediction_data: ZPredictionData.default({}),
    split_id: z.string().nullable().optional(),
    extraction_id: z.string().nullable().optional(),
}).passthrough();
export type SplitBuilderDocument = z.infer<typeof ZSplitBuilderDocument>;

export const ZSplitDataset = z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    base_split_config: z.array(ZSubdocument).default([]),
    base_json_schema: z.record(z.any()).default({}),
    base_subdocuments: z.array(ZSubdocument).default([]),
    base_inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    project_id: z.string(),
}).passthrough();
export type SplitDataset = z.infer<typeof ZSplitDataset>;

export const ZCreateSplitDatasetRequest = z.object({
    name: z.string(),
    base_split_config: z.array(ZSubdocument).default([]),
    base_inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
}).passthrough();
export type CreateSplitDatasetRequest = z.input<typeof ZCreateSplitDatasetRequest>;

export const ZPatchSplitDatasetRequest = z.object({
    name: z.string().nullable().optional(),
}).passthrough();
export type PatchSplitDatasetRequest = z.input<typeof ZPatchSplitDatasetRequest>;

export const ZSplitDatasetDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    dataset_id: z.string(),
    mime_data: ZFileRef,
    prediction_data: ZPredictionData.default({}),
    extraction_id: z.string().nullable().optional(),
    split_id: z.string().nullable().optional(),
    validation_flags: z.record(z.any()).default({}),
}).passthrough();
export type SplitDatasetDocument = z.infer<typeof ZSplitDatasetDocument>;

export const ZSplitConfigOverrides = z.object({
    descriptions_override: z.record(z.string()).nullable().optional(),
    partition_keys_override: z.record(z.string().nullable()).nullable().optional(),
}).passthrough();
export type SplitConfigOverrides = z.infer<typeof ZSplitConfigOverrides>;

export const ZSplitDraftIteration = z.object({
    split_config_overrides: ZSplitConfigOverrides.default({}),
    updated_at: z.string().optional(),
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
}).passthrough();
export type SplitDraftIteration = z.infer<typeof ZSplitDraftIteration>;

export const ZSplitIteration = z.object({
    id: z.string(),
    updated_at: z.string(),
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    split_config_overrides: ZSplitConfigOverrides.default({}),
    parent_id: z.string().nullable().optional(),
    project_id: z.string(),
    dataset_id: z.string(),
    draft: ZSplitDraftIteration.default({}),
    status: z.string().default("draft"),
    finalized_at: z.string().nullable().optional(),
    last_finalize_error: z.string().nullable().optional(),
}).passthrough();
export type SplitIteration = z.infer<typeof ZSplitIteration>;

export const ZCreateSplitIterationRequest = z.object({
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    split_config_overrides: ZSplitConfigOverrides.default({}),
    project_id: z.string(),
    dataset_id: z.string(),
    parent_id: z.string().nullable().optional(),
}).passthrough();
export type CreateSplitIterationRequest = z.input<typeof ZCreateSplitIterationRequest>;

export const ZPatchSplitIterationRequest = z.object({
    inference_settings: ZInferenceSettings.nullable().optional(),
    split_config_overrides: ZSplitConfigOverrides.nullable().optional(),
    draft: ZSplitDraftIteration.nullable().optional(),
}).passthrough();
export type PatchSplitIterationRequest = z.input<typeof ZPatchSplitIterationRequest>;

export const ZSplitIterationDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    iteration_id: z.string(),
    dataset_id: z.string(),
    dataset_document_id: z.string(),
    mime_data: ZFileRef,
    prediction_data: ZPredictionData.default({}),
    extraction_id: z.string().nullable().optional(),
    split_id: z.string().nullable().optional(),
}).passthrough();
export type SplitIterationDocument = z.infer<typeof ZSplitIterationDocument>;

export const ZClassifyProjectDraftConfig = z.object({
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    categories: z.array(ZCategory).default([]),
}).passthrough();
export type ClassifyProjectDraftConfig = z.infer<typeof ZClassifyProjectDraftConfig>;

export const ZClassifyProjectPublishedConfig = ZClassifyProjectDraftConfig.extend({
    origin: z.string().default("manual"),
}).passthrough();
export type ClassifyProjectPublishedConfig = z.infer<typeof ZClassifyProjectPublishedConfig>;

export const ZClassifyProject = z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    published_config: ZClassifyProjectPublishedConfig,
    draft_config: ZClassifyProjectDraftConfig,
    is_published: z.boolean().default(false),
}).passthrough();
export type ClassifyProject = z.infer<typeof ZClassifyProject>;

export const ZCreateClassifyProjectRequest = z.object({
    name: z.string(),
    categories: z.array(ZCategory).default([]),
}).passthrough();
export type CreateClassifyProjectRequest = z.input<typeof ZCreateClassifyProjectRequest>;

export const ZPatchClassifyProjectRequest = z.object({
    name: z.string().nullable().optional(),
    published_config: ZClassifyProjectPublishedConfig.nullable().optional(),
    draft_config: ZClassifyProjectDraftConfig.nullable().optional(),
    is_published: z.boolean().nullable().optional(),
}).passthrough();
export type PatchClassifyProjectRequest = z.input<typeof ZPatchClassifyProjectRequest>;

export const ZClassifyBuilderDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    mime_data: ZFileRef,
    prediction_data: ZPredictionData.default({}),
    classification_id: z.string().nullable().optional(),
    extraction_id: z.string().nullable().optional(),
}).passthrough();
export type ClassifyBuilderDocument = z.infer<typeof ZClassifyBuilderDocument>;

export const ZClassifyDataset = z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    base_categories: z.array(ZCategory).default([]),
    base_inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    project_id: z.string(),
}).passthrough();
export type ClassifyDataset = z.infer<typeof ZClassifyDataset>;

export const ZCreateClassifyDatasetRequest = z.object({
    name: z.string(),
    base_categories: z.array(ZCategory).default([]),
    base_inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
}).passthrough();
export type CreateClassifyDatasetRequest = z.input<typeof ZCreateClassifyDatasetRequest>;

export const ZPatchClassifyDatasetRequest = z.object({
    name: z.string().nullable().optional(),
}).passthrough();
export type PatchClassifyDatasetRequest = z.input<typeof ZPatchClassifyDatasetRequest>;

export const ZClassifyDatasetDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    dataset_id: z.string(),
    mime_data: ZFileRef,
    prediction_data: ZPredictionData.default({}),
    classification_id: z.string().nullable().optional(),
    extraction_id: z.string().nullable().optional(),
    validation_flag: z.boolean().nullable().optional(),
}).passthrough();
export type ClassifyDatasetDocument = z.infer<typeof ZClassifyDatasetDocument>;

export const ZCategoryOverrides = z.object({
    description_overrides: z.record(z.string()).default({}),
}).passthrough();
export type CategoryOverrides = z.infer<typeof ZCategoryOverrides>;

export const ZClassifyDraftIteration = z.object({
    category_overrides: ZCategoryOverrides.default({}),
    updated_at: z.string().optional(),
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
}).passthrough();
export type ClassifyDraftIteration = z.infer<typeof ZClassifyDraftIteration>;

export const ZClassifyIteration = z.object({
    id: z.string(),
    updated_at: z.string(),
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    category_overrides: ZCategoryOverrides.default({}),
    parent_id: z.string().nullable().optional(),
    project_id: z.string(),
    dataset_id: z.string(),
    draft: ZClassifyDraftIteration.default({}),
    status: z.string().default("draft"),
    finalized_at: z.string().nullable().optional(),
    last_finalize_error: z.string().nullable().optional(),
}).passthrough();
export type ClassifyIteration = z.infer<typeof ZClassifyIteration>;

export const ZCreateClassifyIterationRequest = z.object({
    inference_settings: ZInferenceSettings.default({ model: "retab-small", image_resolution_dpi: 192, n_consensus: 1 }),
    category_overrides: ZCategoryOverrides.default({}),
    project_id: z.string(),
    dataset_id: z.string(),
    parent_id: z.string().nullable().optional(),
}).passthrough();
export type CreateClassifyIterationRequest = z.input<typeof ZCreateClassifyIterationRequest>;

export const ZPatchClassifyIterationRequest = z.object({
    inference_settings: ZInferenceSettings.nullable().optional(),
    category_overrides: ZCategoryOverrides.nullable().optional(),
    draft: ZClassifyDraftIteration.nullable().optional(),
}).passthrough();
export type PatchClassifyIterationRequest = z.input<typeof ZPatchClassifyIterationRequest>;

export const ZClassifyIterationDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    iteration_id: z.string(),
    dataset_id: z.string(),
    dataset_document_id: z.string(),
    mime_data: ZFileRef,
    prediction_data: ZPredictionData.default({}),
    classification_id: z.string().nullable().optional(),
    extraction_id: z.string().nullable().optional(),
}).passthrough();
export type ClassifyIterationDocument = z.infer<typeof ZClassifyIterationDocument>;
