import { CompositionClient, RequestOptions } from "../../../client.js";
import * as z from "zod";
import { ZInferenceSettings, ZCreateIterationRequest, ZPatchIterationRequest, ZIterationDocumentStatusResponse, ZProcessIterationRequest, dataArray, Iteration, ZIteration, RetabParsedChatCompletion, ZRetabParsedChatCompletion } from "../../../types.js";

export default class APIProjectsIterations extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async create(projectId: string, body: z.input<typeof ZInferenceSettings>, options?: RequestOptions): Promise<Iteration> {
        // Wrap the inference settings in the expected structure
        const createRequest = {
            inference_settings: body
        };
        return this._fetchJson(ZIteration, {
            url: `/v1/projects/${projectId}/iterations`,
            method: "POST",
            body: { ...(await ZCreateIterationRequest.parseAsync(createRequest)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(projectId: string, iterationId: string, body: z.input<typeof ZPatchIterationRequest>, options?: RequestOptions): Promise<Iteration> {
        return this._fetchJson(ZIteration, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}`,
            method: "PATCH",
            body: { ...(await ZPatchIterationRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(projectId: string, params?: { model?: string }, options?: RequestOptions): Promise<Iteration[]> {
        return this._fetchJson(dataArray(ZIteration), {
            url: `/v1/projects/${projectId}/iterations`,
            method: "GET",
            params: { ...(params || {}), ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(projectId: string, iterationId: string, options?: RequestOptions): Promise<Iteration> {
        return this._fetchJson(ZIteration, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(projectId: string, iterationId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/v1/projects/${projectId}/iterations/${iterationId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async status(projectId: string, iterationId: string, options?: RequestOptions): Promise<z.infer<typeof ZIterationDocumentStatusResponse>> {
        return this._fetchJson(ZIterationDocumentStatusResponse, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}/status`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async process(projectId: string, iterationId: string, body?: z.input<typeof ZProcessIterationRequest>, options?: RequestOptions): Promise<Iteration> {
        return this._fetchJson(ZIteration, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}/process`,
            method: "POST",
            body: { ...(body ? await ZProcessIterationRequest.parseAsync(body) : {}), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async process_document(projectId: string, iterationId: string, documentId: string, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}/documents/${documentId}/process`,
            method: "POST",
            body: { stream: false, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
