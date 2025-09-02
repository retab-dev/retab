import { CompositionClient, RequestOptions } from "../../../client.js";
import * as z from "zod";
import { DocumentItemInput, ZDocumentItem, ZProjectDocument, ProjectDocument, ZPatchProjectDocumentRequest, dataArray } from "../../../types.js";

export default class APIProjectsDocuments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async create(projectId: string, body: DocumentItemInput, options?: RequestOptions): Promise<ProjectDocument> {
        return this._fetchJson(ZProjectDocument, {
            url: `/v1/projects/${projectId}/documents`,
            method: "POST",
            body: { ...(await ZDocumentItem.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(projectId: string, documentId: string, body: z.input<typeof ZPatchProjectDocumentRequest>, options?: RequestOptions): Promise<ProjectDocument> {
        return this._fetchJson(ZProjectDocument, {
            url: `/v1/projects/${projectId}/documents/${documentId}`,
            method: "PATCH",
            body: { ...(await ZPatchProjectDocumentRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(projectId: string, options?: RequestOptions): Promise<ProjectDocument[]> {
        return this._fetchJson(dataArray(ZProjectDocument), {
            url: `/v1/projects/${projectId}/documents`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async get(projectId: string, documentId: string, options?: RequestOptions): Promise<ProjectDocument> {
        return this._fetchJson(ZProjectDocument, {
            url: `/v1/projects/${projectId}/documents/${documentId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async delete(projectId: string, documentId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/v1/projects/${projectId}/documents/${documentId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
