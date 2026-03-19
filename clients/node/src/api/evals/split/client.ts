import * as z from "zod";

import { CompositionClient, RequestOptions } from "../../../client.js";
import { MIMEDataInput, ZMIMEData, ZRetabParsedChatCompletion } from "../../../types.js";
import { buildListParams, buildProcessMultipartBody, cleanObject, dataPaginatedArray } from "../helpers.js";
import {
    ZCreateSplitDatasetRequest,
    ZCreateSplitIterationRequest,
    ZCreateSplitProjectRequest,
    ZPatchSplitDatasetRequest,
    ZPatchSplitIterationRequest,
    ZPatchSplitProjectRequest,
    ZSplitBuilderDocument,
    ZSplitDataset,
    ZSplitDatasetDocument,
    ZSplitIteration,
    ZSplitIterationDocument,
    ZSplitProject,
} from "../schemas.js";

class SplitTemplates extends CompositionClient {
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
        return this._fetchJson(dataPaginatedArray(ZSplitProject), {
            url: "/evals/split/templates",
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order, fields }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async listBuilderDocumentPreviews(templateIds: string[], options?: RequestOptions) {
        const response = await this._fetchJson(z.object({ data: z.record(z.array(z.any())) }), {
            url: "/evals/split/templates/builder-documents/previews",
            method: "GET",
            params: { template_ids: templateIds.join(","), ...(options?.params || {}) },
            headers: options?.headers,
        });
        return response.data;
    }

    async get(templateId: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitProject, {
            url: `/evals/split/templates/${templateId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listBuilderDocuments(templateId: string, options?: RequestOptions) {
        return this._fetchJson(z.array(ZSplitBuilderDocument), {
            url: `/evals/split/templates/${templateId}/builder-documents`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async clone(templateId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        const response = await this._fetchJson(z.object({ project: ZSplitProject }), {
            url: `/evals/split/templates/${templateId}/clone`,
            method: "POST",
            body: { ...cleanObject({ name }), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
        return response.project;
    }
}

class SplitIterations extends CompositionClient {
    async create(evalId: string, datasetId: string, body: {
        inference_settings?: unknown;
        split_config_overrides?: Record<string, unknown>;
        parent_id?: string;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(ZSplitIteration, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations`,
            method: "POST",
            body: { ...(await ZCreateSplitIterationRequest.parseAsync({ project_id: evalId, dataset_id: datasetId, ...body })), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(evalId: string, datasetId: string, {
        before,
        after,
        limit = 10,
        order = "desc",
    }: { before?: string; after?: string; limit?: number; order?: "asc" | "desc" } = {}, options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZSplitIteration), {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations`,
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitIteration, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async updateDraft(evalId: string, datasetId: string, iterationId: string, body: {
        inference_settings?: unknown;
        split_config_overrides?: Record<string, unknown>;
        draft?: Record<string, unknown>;
    }, options?: RequestOptions) {
        return this._fetchJson(ZSplitIteration, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "PATCH",
            body: { ...(await ZPatchSplitIterationRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async finalize(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitIteration, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/finalize`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getSchema(evalId: string, datasetId: string, iterationId: string, { useDraft = false }: { useDraft?: boolean } = {}, options?: RequestOptions) {
        return this._fetchJson(z.record(z.any()), {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/schema`,
            method: "GET",
            params: { ...(useDraft ? { use_draft: true } : {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async processDocuments(evalId: string, datasetId: string, iterationId: string, datasetDocumentId: string, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/processDocumentsFromDatasetId`,
            method: "POST",
            body: { dataset_document_id: datasetDocumentId, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitIterationDocument, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(evalId: string, datasetId: string, iterationId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this._fetchJson(z.array(ZSplitIterationDocument), {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents`,
            method: "GET",
            params: { limit, offset, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async updateDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, body: {
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZSplitIterationDocument, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getMetrics(evalId: string, datasetId: string, iterationId: string, { forceRefresh = false }: { forceRefresh?: boolean } = {}, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/metrics`,
            method: "GET",
            params: { ...(forceRefresh ? { force_refresh: true } : {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async processDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}/process`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

class SplitDatasets extends CompositionClient {
    public iterations: SplitIterations;

    constructor(client: CompositionClient) {
        super(client);
        this.iterations = new SplitIterations(this);
    }

    async create(evalId: string, body: z.input<typeof ZCreateSplitDatasetRequest>, options?: RequestOptions) {
        return this._fetchJson(ZSplitDataset, {
            url: `/evals/split/${evalId}/datasets`,
            method: "POST",
            body: { ...(await ZCreateSplitDatasetRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(evalId: string, { before, after, limit = 10, order = "desc" }: { before?: string; after?: string; limit?: number; order?: "asc" | "desc" } = {}, options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZSplitDataset), {
            url: `/evals/split/${evalId}/datasets`,
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(evalId: string, datasetId: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitDataset, {
            url: `/evals/split/${evalId}/datasets/${datasetId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(evalId: string, datasetId: string, body: z.input<typeof ZPatchSplitDatasetRequest>, options?: RequestOptions) {
        return this._fetchJson(ZSplitDataset, {
            url: `/evals/split/${evalId}/datasets/${datasetId}`,
            method: "PATCH",
            body: { ...(await ZPatchSplitDatasetRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, datasetId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/split/${evalId}/datasets/${datasetId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async duplicate(evalId: string, datasetId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        return this._fetchJson(ZSplitDataset, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/duplicate`,
            method: "POST",
            body: { ...cleanObject({ name }), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async addDocument(evalId: string, datasetId: string, body: { mime_data: MIMEDataInput; prediction_data?: unknown }, options?: RequestOptions) {
        return this._fetchJson(ZSplitDatasetDocument, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/dataset-documents`,
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
        return this._fetchJson(ZSplitDatasetDocument, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(evalId: string, datasetId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this._fetchJson(z.array(ZSplitDatasetDocument), {
            url: `/evals/split/${evalId}/datasets/${datasetId}/dataset-documents`,
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
        return this._fetchJson(ZSplitDatasetDocument, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/split/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async processDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: `/evals/split/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}/process`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

export default class APIEvalsSplit extends CompositionClient {
    public datasets: SplitDatasets;
    public templates: SplitTemplates;

    constructor(client: CompositionClient) {
        super(client);
        this.datasets = new SplitDatasets(this);
        this.templates = new SplitTemplates(this);
    }

    async create(body: z.input<typeof ZCreateSplitProjectRequest>, options?: RequestOptions) {
        return this._fetchJson(ZSplitProject, {
            url: "/evals/split",
            method: "POST",
            body: { ...(await ZCreateSplitProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZSplitProject), {
            url: "/evals/split",
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(evalId: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitProject, {
            url: `/evals/split/${evalId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(evalId: string, body: z.input<typeof ZPatchSplitProjectRequest>, options?: RequestOptions) {
        return this._fetchJson(ZSplitProject, {
            url: `/evals/split/${evalId}`,
            method: "PATCH",
            body: { ...(await ZPatchSplitProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/split/${evalId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async publish(evalId: string, origin?: string, options?: RequestOptions) {
        return this._fetchJson(ZSplitProject, {
            url: `/evals/split/${evalId}/publish`,
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
    }: {
        eval_id: string;
        iteration_id?: string;
        document: MIMEDataInput;
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions) {
        const body = await buildProcessMultipartBody({
            document,
            model,
            image_resolution_dpi,
            n_consensus,
            metadata,
            extraction_id,
            extra: options?.body,
        });
        const url = iteration_id ? `/evals/split/extract/${eval_id}/${iteration_id}` : `/evals/split/extract/${eval_id}`;
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
