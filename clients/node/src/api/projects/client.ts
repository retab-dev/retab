import { CompositionClient, RequestOptions } from "../../client.js";
import { mimeToBlob } from "../../mime.js";
import { dataArray, Project, ZProject, ZCreateProjectRequest, CreateProjectRequest, MIMEDataInput, ZMIMEData, RetabParsedChatCompletion, ZRetabParsedChatCompletion } from "../../types.js";

export default class APIProjects extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async create(body: CreateProjectRequest, options?: RequestOptions): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: "/v1/projects",
            method: "POST",
            body: { ...(await ZCreateProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(options?: RequestOptions): Promise<Project[]> {
        return this._fetchJson(dataArray(ZProject), {
            url: "/v1/projects",
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(projectId: string, options?: RequestOptions): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: `/v1/projects/${projectId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(projectId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/v1/projects/${projectId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async publish(projectId: string, origin?: string, options?: RequestOptions): Promise<Project> {
        const params = origin ? { origin, ...(options?.params || {}) } : options?.params;

        return this._fetchJson(ZProject, {
            url: `/v1/projects/${projectId}/publish`,
            method: "POST",
            params,
            headers: options?.headers,
        });
    }

    async extract({
        project_id,
        iteration_id,
        document,
        model,
        image_resolution_dpi,
        n_consensus,
        temperature,
        metadata,
        extraction_id
    }: {
        project_id: string,
        iteration_id?: string,
        document: MIMEDataInput,
        model?: string,
        image_resolution_dpi?: number,
        n_consensus?: number,
        temperature?: number,
        metadata?: Record<string, string>,
        extraction_id?: string,
    }, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        const url = iteration_id ? `/v1/projects/extract/${project_id}/${iteration_id}` : `/v1/projects/extract/${project_id}`;

        // Parse and convert document to blob for multipart form upload
        const parsedDocument = await ZMIMEData.parseAsync(document);

        // Only include optional parameters if they are provided
        const bodyParams: any = {
            document: mimeToBlob(parsedDocument)
        };

        if (model !== undefined) bodyParams.model = model;
        if (image_resolution_dpi !== undefined) bodyParams.image_resolution_dpi = image_resolution_dpi;
        if (n_consensus !== undefined) bodyParams.n_consensus = n_consensus;
        if (temperature !== undefined) bodyParams.temperature = temperature;
        // Note: metadata must be JSON-serialized since multipart forms only accept primitive types
        if (metadata !== undefined) bodyParams.metadata = JSON.stringify(metadata);
        if (extraction_id !== undefined) bodyParams.extraction_id = extraction_id;

        return this._fetchJson(ZRetabParsedChatCompletion, {
            url,
            method: "POST",
            body: { ...bodyParams, ...(options?.body || {}) },
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
