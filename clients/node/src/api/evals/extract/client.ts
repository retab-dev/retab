import * as z from "zod";

import { CompositionClient, RequestOptions } from "../../../client.js";
import { MIMEDataInput, ZCreateProjectRequest, ZMIMEData, ZProject, ZRetabParsedChatCompletion, ZRetabParsedChatCompletionChunk, ZInferenceSettings, ZPredictionData } from "../../../types.js";
import { buildListParams, buildProcessMultipartBody, cleanObject, dataPaginatedArray } from "../helpers.js";

const ZDataset = z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    base_json_schema: z.record(z.any()).default({}),
    project_id: z.string(),
}).passthrough();

const ZDatasetDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    dataset_id: z.string(),
    mime_data: z.any(),
    prediction_data: ZPredictionData.default({}),
    extraction_id: z.string().nullable().optional(),
    validation_flags: z.record(z.any()).default({}),
}).passthrough();

const ZSchemaOverrides = z.object({
    descriptionsOverride: z.record(z.string()).nullable().optional(),
    reasoningPromptsOverride: z.record(z.string()).nullable().optional(),
}).passthrough();

const ZDraftIteration = z.object({
    schema_overrides: ZSchemaOverrides.default({}),
    updated_at: z.string().optional(),
    inference_settings: ZInferenceSettings.default({ model: "retab-small", reasoning_effort: "minimal", image_resolution_dpi: 192, n_consensus: 1 }),
}).passthrough();

const ZIteration = z.object({
    id: z.string(),
    updated_at: z.string(),
    inference_settings: ZInferenceSettings.default({ model: "retab-small", reasoning_effort: "minimal", image_resolution_dpi: 192, n_consensus: 1 }),
    schema_overrides: ZSchemaOverrides.default({}),
    parent_id: z.string().nullable().optional(),
    project_id: z.string(),
    dataset_id: z.string(),
    draft: ZDraftIteration.default({}),
    status: z.string().default("draft"),
    finalized_at: z.string().nullable().optional(),
    last_finalize_error: z.string().nullable().optional(),
}).passthrough();

const ZIterationDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    iteration_id: z.string(),
    dataset_id: z.string(),
    dataset_document_id: z.string(),
    mime_data: z.any(),
    prediction_data: ZPredictionData.default({}),
    extraction_id: z.string().nullable().optional(),
}).passthrough();

const ZBuilderDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    mime_data: z.any(),
    prediction_data: ZPredictionData.default({}),
    extraction_id: z.string().nullable().optional(),
}).passthrough();

class ExtractTemplates extends CompositionClient {
    async list({
        before,
        after,
        limit = 10,
        order = "desc",
        fields,
    }: {
        before?: string;
        after?: string;
        limit?: number;
        order?: "asc" | "desc";
        fields?: string;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZProject), {
            url: "/evals/extract/templates",
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order, fields }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async listBuilderDocumentPreviews(templateIds: string[], options?: RequestOptions): Promise<Record<string, unknown[]>> {
        const response = await this._fetchJson(z.object({ data: z.record(z.array(z.any())) }), {
            url: "/evals/extract/templates/builder-documents/previews",
            method: "GET",
            params: { template_ids: templateIds.join(","), ...(options?.params || {}) },
            headers: options?.headers,
        });
        return response.data;
    }

    async get(templateId: string, options?: RequestOptions) {
        return this._fetchJson(ZProject, {
            url: `/evals/extract/templates/${templateId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listBuilderDocuments(templateId: string, options?: RequestOptions) {
        return this._fetchJson(z.array(ZBuilderDocument), {
            url: `/evals/extract/templates/${templateId}/builder-documents`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async clone(templateId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        const response = await this._fetchJson(z.object({ project: ZProject }), {
            url: `/evals/extract/templates/${templateId}/clone`,
            method: "POST",
            body: { ...cleanObject({ name }), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
        return response.project;
    }
}

class ExtractIterations extends CompositionClient {
    async create(evalId: string, datasetId: string, body: {
        inference_settings?: unknown;
        schema_overrides?: Record<string, unknown>;
        parent_id?: string;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(ZIteration, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations`,
            method: "POST",
            body: { project_id: evalId, dataset_id: datasetId, ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(evalId: string, datasetId: string, {
        before,
        after,
        limit = 10,
        order = "desc",
    }: {
        before?: string;
        after?: string;
        limit?: number;
        order?: "asc" | "desc";
    } = {}, options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZIteration), {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations`,
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZIteration, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async updateDraft(evalId: string, datasetId: string, iterationId: string, body: {
        inference_settings?: unknown;
        schema_overrides?: Record<string, unknown>;
        draft?: Record<string, unknown>;
    }, options?: RequestOptions) {
        return this._fetchJson(ZIteration, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async finalize(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZIteration, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/finalize`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getSchema(evalId: string, datasetId: string, iterationId: string, { useDraft = false }: { useDraft?: boolean } = {}, options?: RequestOptions) {
        return this._fetchJson(z.record(z.any()), {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/schema`,
            method: "GET",
            params: { ...(useDraft ? { use_draft: true } : {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async processDocuments(evalId: string, datasetId: string, iterationId: string, datasetDocumentId: string, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/processDocumentsFromDatasetId`,
            method: "POST",
            body: { dataset_document_id: datasetDocumentId, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZIterationDocument, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(evalId: string, datasetId: string, iterationId: string, {
        limit = 1000,
        offset = 0,
    }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this._fetchJson(z.array(ZIterationDocument), {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents`,
            method: "GET",
            params: { limit, offset, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async updateDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, body: {
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZIterationDocument, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getMetrics(evalId: string, datasetId: string, iterationId: string, { forceRefresh = false }: { forceRefresh?: boolean } = {}, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/metrics`,
            method: "GET",
            params: { ...(forceRefresh ? { force_refresh: true } : {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async processDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}/process`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

class ExtractDatasets extends CompositionClient {
    public iterations: ExtractIterations;

    constructor(client: CompositionClient) {
        super(client);
        this.iterations = new ExtractIterations(this);
    }

    async create(evalId: string, body: { name: string; base_json_schema?: Record<string, unknown> }, options?: RequestOptions) {
        return this._fetchJson(ZDataset, {
            url: `/evals/extract/${evalId}/datasets`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(evalId: string, {
        before,
        after,
        limit = 10,
        order = "desc",
    }: {
        before?: string;
        after?: string;
        limit?: number;
        order?: "asc" | "desc";
    } = {}, options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZDataset), {
            url: `/evals/extract/${evalId}/datasets`,
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(evalId: string, datasetId: string, options?: RequestOptions) {
        return this._fetchJson(ZDataset, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(evalId: string, datasetId: string, body: { name?: string }, options?: RequestOptions) {
        return this._fetchJson(ZDataset, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, datasetId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/extract/${evalId}/datasets/${datasetId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async duplicate(evalId: string, datasetId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        return this._fetchJson(ZDataset, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/duplicate`,
            method: "POST",
            body: { ...cleanObject({ name }), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async addDocument(evalId: string, datasetId: string, body: {
        mime_data: MIMEDataInput;
        prediction_data?: unknown;
    }, options?: RequestOptions) {
        return this._fetchJson(ZDatasetDocument, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/dataset-documents`,
            method: "POST",
            body: {
                mime_data: await ZMIMEData.parseAsync(body.mime_data),
                project_id: evalId,
                dataset_id: datasetId,
                ...(body.prediction_data !== undefined ? { prediction_data: body.prediction_data } : {}),
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZDatasetDocument, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(evalId: string, datasetId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this._fetchJson(z.array(ZDatasetDocument), {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/dataset-documents`,
            method: "GET",
            params: { limit, offset, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async updateDocument(evalId: string, datasetId: string, documentId: string, body: {
        validation_flags?: Record<string, unknown>;
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZDatasetDocument, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/extract/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async processDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: `/evals/extract/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}/process`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

export default class APIEvalsExtract extends CompositionClient {
    public datasets: ExtractDatasets;
    public templates: ExtractTemplates;

    constructor(client: CompositionClient) {
        super(client);
        this.datasets = new ExtractDatasets(this);
        this.templates = new ExtractTemplates(this);
    }

    async create(body: { name: string; json_schema: Record<string, unknown> }, options?: RequestOptions) {
        return this._fetchJson(ZProject, {
            url: "/evals/extract",
            method: "POST",
            body: { ...(await ZCreateProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZProject), {
            url: "/evals/extract",
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(evalId: string, options?: RequestOptions) {
        return this._fetchJson(ZProject, {
            url: `/evals/extract/${evalId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(evalId: string, body: { name?: string; json_schema?: Record<string, unknown> }, options?: RequestOptions) {
        return this._fetchJson(ZProject, {
            url: `/evals/extract/${evalId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/extract/${evalId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async publish(evalId: string, origin?: string, options?: RequestOptions) {
        return this._fetchJson(ZProject, {
            url: `/evals/extract/${evalId}/publish`,
            method: "POST",
            params: { ...(origin ? { origin } : {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async process({
        eval_id,
        iteration_id,
        document,
        model,
        image_resolution_dpi,
        n_consensus,
        metadata,
        extraction_id,
        ...rest
    }: {
        eval_id: string;
        iteration_id?: string;
        document?: MIMEDataInput;
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions) {
        if ("documents" in rest) {
            throw new Error("client.evals.extract.process(...) accepts only 'document'.")
        }
        const body = await buildProcessMultipartBody({
            document,
            model,
            image_resolution_dpi,
            n_consensus,
            metadata,
            extraction_id,
            extra: options?.body,
        });
        const url = iteration_id ? `/evals/extract/extract/${eval_id}/${iteration_id}` : `/evals/extract/extract/${eval_id}`;
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async extract(params: {
        eval_id: string;
        iteration_id?: string;
        document?: MIMEDataInput;
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions) {
        process.emitWarning(
            "client.evals.extract.extract(...) is deprecated; use client.evals.extract.process(...) instead.",
            "DeprecationWarning"
        );
        return this.process(params, options);
    }

    async processStream({
        eval_id,
        iteration_id,
        document,
        model,
        image_resolution_dpi,
        n_consensus,
        metadata,
        extraction_id,
        ...rest
    }: {
        eval_id: string;
        iteration_id?: string;
        document?: MIMEDataInput;
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions) {
        if ("documents" in rest) {
            throw new Error("client.evals.extract.processStream(...) accepts only 'document'.")
        }
        const body = await buildProcessMultipartBody({
            document,
            model,
            image_resolution_dpi,
            n_consensus,
            metadata,
            extraction_id,
            extra: options?.body,
        });
        const url = iteration_id ? `/evals/extract/extract/${eval_id}/${iteration_id}/stream` : `/evals/extract/extract/${eval_id}/stream`;
        return this._fetchStream(ZRetabParsedChatCompletionChunk, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
