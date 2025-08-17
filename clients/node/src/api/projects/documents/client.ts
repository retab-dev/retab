import { CompositionClient } from "../../../client.js";
import * as z from "zod";
import { DocumentItemInput, ZDocumentItem, ZProjectDocument, ProjectDocument, ZPatchProjectDocumentRequest, dataArray } from "../../../types.js";

export default class APIProjectsDocuments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async create(projectId: string, body: DocumentItemInput): Promise<ProjectDocument> {
        return this._fetchJson(ZProjectDocument, {
            url: `/v1/projects/${projectId}/documents`,
            method: "POST",
            body: await ZDocumentItem.parseAsync(body),
        });
    }

    async update(projectId: string, documentId: string, body: z.input<typeof ZPatchProjectDocumentRequest>): Promise<ProjectDocument> {
        return this._fetchJson(ZProjectDocument, {
            url: `/v1/projects/${projectId}/documents/${documentId}`,
            method: "PATCH",
            body: await ZPatchProjectDocumentRequest.parseAsync(body),
        });
    }

    async list(projectId: string): Promise<ProjectDocument[]> {
        return this._fetchJson(dataArray(ZProjectDocument), {
            url: `/v1/projects/${projectId}/documents`,
            method: "GET",
        });
    }

    async get(projectId: string, documentId: string): Promise<ProjectDocument> {
        return this._fetchJson(ZProjectDocument, {
            url: `/v1/projects/${projectId}/documents/${documentId}`,
            method: "GET",
        });
    }

    async delete(projectId: string, documentId: string): Promise<void> {
        return this._fetchJson({
            url: `/v1/projects/${projectId}/documents/${documentId}`,
            method: "DELETE",
        });
    }
}
