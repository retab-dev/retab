import * as z from "zod";

import { CompositionClient, RequestOptions } from "../../../client.js";
import { MIMEDataInput, ZCategory, ZClassifyResponse, ZMIMEData } from "../../../types.js";
import { buildListParams, buildProcessMultipartBody, cleanObject, dataPaginatedArray } from "../helpers.js";
import {
    ZClassifyBuilderDocument,
    ZClassifyDataset,
    ZClassifyDatasetDocument,
    ZClassifyIteration,
    ZClassifyIterationDocument,
    ZClassifyProject,
    ZCreateClassifyDatasetRequest,
    ZCreateClassifyIterationRequest,
    ZCreateClassifyProjectRequest,
    ZPatchClassifyDatasetRequest,
    ZPatchClassifyIterationRequest,
    ZPatchClassifyProjectRequest,
} from "../schemas.js";

class ClassifyTemplates extends CompositionClient {
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
        return this._fetchJson(dataPaginatedArray(ZClassifyProject), {
            url: "/evals/classify/templates",
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order, fields }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async listBuilderDocumentPreviews(templateIds: string[], options?: RequestOptions) {
        const response = await this._fetchJson(z.object({ data: z.record(z.array(z.any())) }), {
            url: "/evals/classify/templates/builder-documents/previews",
            method: "GET",
            params: { template_ids: templateIds.join(","), ...(options?.params || {}) },
            headers: options?.headers,
        });
        return response.data;
    }

    async get(templateId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyProject, {
            url: `/evals/classify/templates/${templateId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listBuilderDocuments(templateId: string, options?: RequestOptions) {
        return this._fetchJson(z.array(ZClassifyBuilderDocument), {
            url: `/evals/classify/templates/${templateId}/builder-documents`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async clone(templateId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        const response = await this._fetchJson(z.object({ project: ZClassifyProject }), {
            url: `/evals/classify/templates/${templateId}/clone`,
            method: "POST",
            body: { ...cleanObject({ name }), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
        return response.project;
    }
}

class ClassifyIterations extends CompositionClient {
    async create(evalId: string, datasetId: string, body: {
        inference_settings?: unknown;
        category_overrides?: Record<string, unknown>;
        parent_id?: string;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(ZClassifyIteration, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations`,
            method: "POST",
            body: { ...(await ZCreateClassifyIterationRequest.parseAsync({ project_id: evalId, dataset_id: datasetId, ...body })), ...(options?.body || {}) },
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
        return this._fetchJson(dataPaginatedArray(ZClassifyIteration), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations`,
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyIteration, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async updateDraft(evalId: string, datasetId: string, iterationId: string, body: {
        inference_settings?: unknown;
        category_overrides?: Record<string, unknown>;
        draft?: Record<string, unknown>;
    }, options?: RequestOptions) {
        return this._fetchJson(ZClassifyIteration, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "PATCH",
            body: { ...(await ZPatchClassifyIterationRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async finalize(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyIteration, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/finalize`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getCategories(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(z.array(ZCategory), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/categories`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getSchema(evalId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/schema`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async processDocuments(evalId: string, datasetId: string, iterationId: string, datasetDocumentId: string, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/processDocumentsFromDatasetId`,
            method: "POST",
            body: { dataset_document_id: datasetDocumentId, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyIterationDocument, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(evalId: string, datasetId: string, iterationId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this._fetchJson(z.array(ZClassifyIterationDocument), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents`,
            method: "GET",
            params: { limit, offset, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async updateDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, body: {
        prediction_data?: unknown;
        classification_id?: string;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZClassifyIterationDocument, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getMetrics(evalId: string, datasetId: string, iterationId: string, { forceRefresh = false }: { forceRefresh?: boolean } = {}, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/metrics`,
            method: "GET",
            params: { ...(forceRefresh ? { force_refresh: true } : {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async processDocument(evalId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyResponse, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}/process`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

class ClassifyDatasets extends CompositionClient {
    public iterations: ClassifyIterations;

    constructor(client: CompositionClient) {
        super(client);
        this.iterations = new ClassifyIterations(this);
    }

    async create(evalId: string, body: z.input<typeof ZCreateClassifyDatasetRequest>, options?: RequestOptions) {
        return this._fetchJson(ZClassifyDataset, {
            url: `/evals/classify/${evalId}/datasets`,
            method: "POST",
            body: { ...(await ZCreateClassifyDatasetRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(evalId: string, { before, after, limit = 10, order = "desc" }: { before?: string; after?: string; limit?: number; order?: "asc" | "desc" } = {}, options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZClassifyDataset), {
            url: `/evals/classify/${evalId}/datasets`,
            method: "GET",
            params: { ...buildListParams({ before, after, limit, order }), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(evalId: string, datasetId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyDataset, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(evalId: string, datasetId: string, body: z.input<typeof ZPatchClassifyDatasetRequest>, options?: RequestOptions) {
        return this._fetchJson(ZClassifyDataset, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}`,
            method: "PATCH",
            body: { ...(await ZPatchClassifyDatasetRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, datasetId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/classify/${evalId}/datasets/${datasetId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async duplicate(evalId: string, datasetId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        return this._fetchJson(ZClassifyDataset, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/duplicate`,
            method: "POST",
            body: { ...cleanObject({ name }), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async addDocument(evalId: string, datasetId: string, body: { mime_data: MIMEDataInput; prediction_data?: unknown }, options?: RequestOptions) {
        return this._fetchJson(ZClassifyDatasetDocument, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/dataset-documents`,
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
        return this._fetchJson(ZClassifyDatasetDocument, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(evalId: string, datasetId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this._fetchJson(z.array(ZClassifyDatasetDocument), {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/dataset-documents`,
            method: "GET",
            params: { limit, offset, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async updateDocument(evalId: string, datasetId: string, documentId: string, body: {
        validation_flag?: boolean | null;
        prediction_data?: unknown;
        classification_id?: string;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZClassifyDatasetDocument, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/classify/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async processDocument(evalId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyResponse, {
            url: `/evals/classify/${evalId}/datasets/${datasetId}/dataset-documents/${documentId}/process`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

export default class APIEvalsClassify extends CompositionClient {
    public datasets: ClassifyDatasets;
    public templates: ClassifyTemplates;

    constructor(client: CompositionClient) {
        super(client);
        this.datasets = new ClassifyDatasets(this);
        this.templates = new ClassifyTemplates(this);
    }

    async create(body: z.input<typeof ZCreateClassifyProjectRequest>, options?: RequestOptions) {
        return this._fetchJson(ZClassifyProject, {
            url: "/evals/classify",
            method: "POST",
            body: { ...(await ZCreateClassifyProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(options?: RequestOptions) {
        return this._fetchJson(dataPaginatedArray(ZClassifyProject), {
            url: "/evals/classify",
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(evalId: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyProject, {
            url: `/evals/classify/${evalId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(evalId: string, body: z.input<typeof ZPatchClassifyProjectRequest>, options?: RequestOptions) {
        return this._fetchJson(ZClassifyProject, {
            url: `/evals/classify/${evalId}`,
            method: "PATCH",
            body: { ...(await ZPatchClassifyProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(evalId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/evals/classify/${evalId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async publish(evalId: string, origin?: string, options?: RequestOptions) {
        return this._fetchJson(ZClassifyProject, {
            url: `/evals/classify/${evalId}/publish`,
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
        metadata,
    }: {
        eval_id: string;
        iteration_id?: string;
        document: MIMEDataInput;
        model?: string;
        metadata?: Record<string, string>;
    }, options?: RequestOptions) {
        const body = await buildProcessMultipartBody({
            document,
            model,
            metadata,
            extra: options?.body,
        });
        const url = iteration_id ? `/evals/classify/extract/${eval_id}/${iteration_id}` : `/evals/classify/extract/${eval_id}`;
        return this._fetchJson(ZClassifyResponse, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
