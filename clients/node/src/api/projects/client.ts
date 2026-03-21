import * as z from "zod";

import { CompositionClient, RequestOptions } from "../../client.js";
import {
    JSONSchemaInput,
    MIMEDataInput,
    Project,
    RetabParsedChatCompletion,
    ZCreateProjectRequest,
    ZDeleteResponse,
    ZInferenceSettings,
    ZJSONSchema,
    ZMIMEData,
    ZPredictionData,
    ZProject,
    ZRetabParsedChatCompletion,
} from "../../types.js";
import { mimeToBlob } from "../../mime.js";
import { buildListParams, buildProcessMultipartBody, dataPaginatedArray } from "../evals/helpers.js";

const ZProjectDataset = z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    base_json_schema: z.record(z.any()).default({}),
    project_id: z.string(),
}).passthrough();

const ZProjectDatasetDocument = z.object({
    id: z.string(),
    updated_at: z.string(),
    project_id: z.string(),
    dataset_id: z.string(),
    mime_data: z.any(),
    prediction_data: ZPredictionData.default({}),
    extraction_id: z.string().nullable().optional(),
    validation_flags: z.record(z.any()).default({}),
}).passthrough();

const ZProjectSchemaOverrides = z.object({
    descriptionsOverride: z.record(z.string()).nullable().optional(),
    reasoningPromptsOverride: z.record(z.string()).nullable().optional(),
}).passthrough();

const ZProjectDraftIteration = z.object({
    schema_overrides: ZProjectSchemaOverrides.default({}),
    updated_at: z.string().optional(),
    inference_settings: ZInferenceSettings.default({
        model: "retab-small",
        image_resolution_dpi: 192,
        n_consensus: 1,
    }),
}).passthrough();

const ZProjectIteration = z.object({
    id: z.string(),
    updated_at: z.string(),
    inference_settings: ZInferenceSettings.default({
        model: "retab-small",
        image_resolution_dpi: 192,
        n_consensus: 1,
    }),
    schema_overrides: ZProjectSchemaOverrides.default({}),
    parent_id: z.string().nullable().optional(),
    project_id: z.string(),
    dataset_id: z.string(),
    draft: ZProjectDraftIteration.default({}),
    status: z.string().default("draft"),
    finalized_at: z.string().nullable().optional(),
    last_finalize_error: z.string().nullable().optional(),
}).passthrough();

const ZProjectIterationDocument = z.object({
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

const ZProjectDatasetDocumentList = z.union([
    z.array(ZProjectDatasetDocument),
    z.object({ data: z.array(ZProjectDatasetDocument) }).passthrough(),
]).transform((value) => Array.isArray(value) ? value : value.data);

const ZProjectIterationDocumentList = z.union([
    z.array(ZProjectIterationDocument),
    z.object({ data: z.array(ZProjectIterationDocument) }).passthrough(),
]).transform((value) => Array.isArray(value) ? value : value.data);

function emitEvalDeprecationWarning(message: string): void {
    process.emitWarning(message, "DeprecationWarning");
}

async function parseOptionalJsonSchema(jsonSchema?: JSONSchemaInput): Promise<Record<string, unknown> | undefined> {
    if (jsonSchema === undefined) {
        return undefined;
    }
    return await ZJSONSchema.parseAsync(jsonSchema);
}

class ProjectIterations extends CompositionClient {
    async create(projectId: string, datasetId: string, body: {
        inference_settings?: unknown;
        schema_overrides?: Record<string, unknown>;
        parent_id?: string;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(ZProjectIteration, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations`,
            method: "POST",
            body: {
                project_id: projectId,
                dataset_id: datasetId,
                ...body,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(projectId: string, datasetId: string, {
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
        return this._fetchJson(dataPaginatedArray(ZProjectIteration), {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations`,
            method: "GET",
            params: {
                ...buildListParams({ before, after, limit, order }),
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    async get(projectId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZProjectIteration, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async updateDraft(projectId: string, datasetId: string, iterationId: string, body: {
        inference_settings?: unknown;
        schema_overrides?: Record<string, unknown>;
        draft?: Record<string, unknown>;
    }, options?: RequestOptions) {
        const requestBody = body.draft ? body : {
            draft: {
                ...(body.inference_settings !== undefined ? { inference_settings: body.inference_settings } : {}),
                ...(body.schema_overrides !== undefined ? { schema_overrides: body.schema_overrides } : {}),
            },
        };

        return this._fetchJson(ZProjectIteration, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "PATCH",
            body: {
                ...requestBody,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(projectId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZDeleteResponse, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async finalize(projectId: string, datasetId: string, iterationId: string, options?: RequestOptions) {
        return this._fetchJson(ZProjectIteration, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/finalize`,
            method: "POST",
            body: { ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getSchema(projectId: string, datasetId: string, iterationId: string, {
        useDraft = false,
    }: {
        useDraft?: boolean;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(z.record(z.any()), {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/schema`,
            method: "GET",
            params: {
                ...(useDraft ? { use_draft: true } : {}),
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    async processDocuments(projectId: string, datasetId: string, iterationId: string, datasetDocumentId: string, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/processDocumentsFromDatasetId`,
            method: "POST",
            body: {
                dataset_document_id: datasetDocumentId,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDocument(projectId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZProjectIterationDocument, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(projectId: string, datasetId: string, iterationId: string, {
        limit = 1000,
        offset = 0,
    }: {
        limit?: number;
        offset?: number;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(ZProjectIterationDocumentList, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents`,
            method: "GET",
            params: {
                limit,
                offset,
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    async updateDocument(projectId: string, datasetId: string, iterationId: string, documentId: string, body: {
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZProjectIterationDocument, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "PATCH",
            body: {
                ...body,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(projectId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZDeleteResponse, {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/iteration-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getMetrics(projectId: string, datasetId: string, iterationId: string, {
        forceRefresh = false,
    }: {
        forceRefresh?: boolean;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(z.any(), {
            url: `/projects/${projectId}/datasets/${datasetId}/iterations/${iterationId}/metrics`,
            method: "GET",
            params: {
                ...(forceRefresh ? { force_refresh: true } : {}),
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    async update_draft(projectId: string, datasetId: string, iterationId: string, body: {
        inference_settings?: unknown;
        schema_overrides?: Record<string, unknown>;
        draft?: Record<string, unknown>;
    }, options?: RequestOptions) {
        return this.updateDraft(projectId, datasetId, iterationId, body, options);
    }

    async get_schema(projectId: string, datasetId: string, iterationId: string, { use_draft = false }: { use_draft?: boolean } = {}, options?: RequestOptions) {
        return this.getSchema(projectId, datasetId, iterationId, { useDraft: use_draft }, options);
    }

    async process_documents(projectId: string, datasetId: string, iterationId: string, datasetDocumentId: string, options?: RequestOptions) {
        return this.processDocuments(projectId, datasetId, iterationId, datasetDocumentId, options);
    }

    async get_document(projectId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this.getDocument(projectId, datasetId, iterationId, documentId, options);
    }

    async list_documents(projectId: string, datasetId: string, iterationId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this.listDocuments(projectId, datasetId, iterationId, { limit, offset }, options);
    }

    async update_document(projectId: string, datasetId: string, iterationId: string, documentId: string, body: {
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this.updateDocument(projectId, datasetId, iterationId, documentId, body, options);
    }

    async delete_document(projectId: string, datasetId: string, iterationId: string, documentId: string, options?: RequestOptions) {
        return this.deleteDocument(projectId, datasetId, iterationId, documentId, options);
    }

    async get_metrics(projectId: string, datasetId: string, iterationId: string, { force_refresh = false }: { force_refresh?: boolean } = {}, options?: RequestOptions) {
        return this.getMetrics(projectId, datasetId, iterationId, { forceRefresh: force_refresh }, options);
    }
}

class ProjectDatasets extends CompositionClient {
    public iterations: ProjectIterations;

    constructor(client: CompositionClient) {
        super(client);
        this.iterations = new ProjectIterations(this);
    }

    async create(projectId: string, body: {
        name: string;
        base_json_schema?: JSONSchemaInput;
    }, options?: RequestOptions) {
        const baseJsonSchema = await parseOptionalJsonSchema(body.base_json_schema);

        return this._fetchJson(ZProjectDataset, {
            url: `/projects/${projectId}/datasets`,
            method: "POST",
            body: {
                name: body.name,
                ...(baseJsonSchema !== undefined ? { base_json_schema: baseJsonSchema } : {}),
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(projectId: string, {
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
        return this._fetchJson(dataPaginatedArray(ZProjectDataset), {
            url: `/projects/${projectId}/datasets`,
            method: "GET",
            params: {
                ...buildListParams({ before, after, limit, order }),
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    async get(projectId: string, datasetId: string, options?: RequestOptions) {
        return this._fetchJson(ZProjectDataset, {
            url: `/projects/${projectId}/datasets/${datasetId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(projectId: string, datasetId: string, body: { name?: string }, options?: RequestOptions) {
        return this._fetchJson(ZProjectDataset, {
            url: `/projects/${projectId}/datasets/${datasetId}`,
            method: "PATCH",
            body: {
                ...body,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(projectId: string, datasetId: string, options?: RequestOptions) {
        return this._fetchJson(ZDeleteResponse, {
            url: `/projects/${projectId}/datasets/${datasetId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async duplicate(projectId: string, datasetId: string, { name }: { name?: string } = {}, options?: RequestOptions) {
        return this._fetchJson(ZProjectDataset, {
            url: `/projects/${projectId}/datasets/${datasetId}/duplicate`,
            method: "POST",
            body: {
                ...(name !== undefined ? { name } : {}),
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async addDocument(projectId: string, datasetId: string, body: {
        mime_data: MIMEDataInput;
        prediction_data?: unknown;
    }, options?: RequestOptions) {
        return this._fetchJson(ZProjectDatasetDocument, {
            url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents`,
            method: "POST",
            body: {
                mime_data: await ZMIMEData.parseAsync(body.mime_data),
                project_id: projectId,
                dataset_id: datasetId,
                ...(body.prediction_data !== undefined ? { prediction_data: body.prediction_data } : {}),
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDocument(projectId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZProjectDatasetDocument, {
            url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async listDocuments(projectId: string, datasetId: string, {
        limit = 1000,
        offset = 0,
    }: {
        limit?: number;
        offset?: number;
    } = {}, options?: RequestOptions) {
        return this._fetchJson(ZProjectDatasetDocumentList, {
            url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents`,
            method: "GET",
            params: {
                limit,
                offset,
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    async updateDocument(projectId: string, datasetId: string, documentId: string, body: {
        validation_flags?: Record<string, unknown>;
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this._fetchJson(ZProjectDatasetDocument, {
            url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "PATCH",
            body: {
                ...body,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async deleteDocument(projectId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this._fetchJson(ZDeleteResponse, {
            url: `/projects/${projectId}/datasets/${datasetId}/dataset-documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async add_document(projectId: string, datasetId: string, body: {
        mime_data: MIMEDataInput;
        prediction_data?: unknown;
    }, options?: RequestOptions) {
        return this.addDocument(projectId, datasetId, body, options);
    }

    async get_document(projectId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this.getDocument(projectId, datasetId, documentId, options);
    }

    async list_documents(projectId: string, datasetId: string, { limit = 1000, offset = 0 }: { limit?: number; offset?: number } = {}, options?: RequestOptions) {
        return this.listDocuments(projectId, datasetId, { limit, offset }, options);
    }

    async update_document(projectId: string, datasetId: string, documentId: string, body: {
        validation_flags?: Record<string, unknown>;
        prediction_data?: unknown;
        extraction_id?: string;
    }, options?: RequestOptions) {
        return this.updateDocument(projectId, datasetId, documentId, body, options);
    }

    async delete_document(projectId: string, datasetId: string, documentId: string, options?: RequestOptions) {
        return this.deleteDocument(projectId, datasetId, documentId, options);
    }
}

export default class APIProjects extends CompositionClient {
    public datasets: ProjectDatasets;

    constructor(client: CompositionClient) {
        super(client);
        this.datasets = new ProjectDatasets(this);
    }

    async create(body: { name: string; json_schema: JSONSchemaInput }, options?: RequestOptions): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: "/projects",
            method: "POST",
            body: { ...(await ZCreateProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(options?: RequestOptions): Promise<Project[]> {
        return this._fetchJson(dataPaginatedArray(ZProject), {
            url: "/projects",
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(projectId: string, options?: RequestOptions): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: `/projects/${projectId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async prepare_update(projectId: string, body: { name?: string; json_schema?: JSONSchemaInput }, options?: RequestOptions) {
        const jsonSchema = await parseOptionalJsonSchema(body.json_schema);

        return {
            url: `/projects/${projectId}`,
            method: "PATCH",
            body: {
                ...(body.name !== undefined ? { name: body.name } : {}),
                ...(jsonSchema !== undefined ? { json_schema: jsonSchema } : {}),
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        };
    }

    async _update(projectId: string, body: { name?: string; json_schema?: JSONSchemaInput }, options?: RequestOptions): Promise<Project> {
        const jsonSchema = await parseOptionalJsonSchema(body.json_schema);

        return this._fetchJson(ZProject, {
            url: `/projects/${projectId}`,
            method: "PATCH",
            body: {
                ...(body.name !== undefined ? { name: body.name } : {}),
                ...(jsonSchema !== undefined ? { json_schema: jsonSchema } : {}),
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(projectId: string, options?: RequestOptions) {
        return this._fetchJson(ZDeleteResponse, {
            url: `/projects/${projectId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async publish(projectId: string, origin?: string, options?: RequestOptions): Promise<Project> {
        const params = origin ? { origin, ...(options?.params || {}) } : options?.params;

        return this._fetchJson(ZProject, {
            url: `/projects/${projectId}/publish`,
            method: "POST",
            params,
            headers: options?.headers,
        });
    }

    async extract({
        project_id,
        iteration_id,
        document,
        documents,
        model,
        image_resolution_dpi,
        n_consensus,
        metadata,
        extraction_id,
    }: {
        project_id: string;
        iteration_id?: string;
        document?: MIMEDataInput;
        documents?: MIMEDataInput[];
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        emitEvalDeprecationWarning(
            "client.projects.extract(...) is deprecated; use client.evals.extract.process(...) instead."
        );
        const url = iteration_id ? `/projects/extract/${project_id}/${iteration_id}` : `/projects/extract/${project_id}`;
        const body = await buildProcessMultipartBody({
            document,
            documents,
            model,
            image_resolution_dpi,
            n_consensus,
            metadata,
            extraction_id,
            extra: options?.body,
        });

        return this._fetchJson(ZRetabParsedChatCompletion, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async split({
        project_id,
        document,
        model,
        image_resolution_dpi,
        n_consensus,
        metadata,
        extraction_id,
    }: {
        project_id: string;
        document: MIMEDataInput;
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        emitEvalDeprecationWarning(
            "client.projects.split(...) is deprecated; use client.evals.split.process(...) instead."
        );
        const parsedDocument = await ZMIMEData.parseAsync(document);

        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: `/projects/split/${project_id}`,
            method: "POST",
            body: {
                document: mimeToBlob(parsedDocument),
                ...(model !== undefined ? { model } : {}),
                ...(image_resolution_dpi !== undefined ? { image_resolution_dpi } : {}),
                ...(n_consensus !== undefined ? { n_consensus } : {}),
                ...(metadata !== undefined ? { metadata: JSON.stringify(metadata) } : {}),
                ...(extraction_id !== undefined ? { extraction_id } : {}),
                ...(options?.body || {}),
            },
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
