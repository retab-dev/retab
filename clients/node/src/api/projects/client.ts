import { CompositionClient, RequestOptions } from "../../client.js";
import { mimeToBlob } from "../../mime.js";
import { BaseProjectInput, dataArray, Project, ZBaseProject, ZProjectLoose as ZProject, ZCreateProjectRequest, CreateProjectRequest, MIMEDataInput, ZMIMEData, RetabParsedChatCompletion, ZRetabParsedChatCompletion } from "../../types.js";
import APIProjectsDocuments from "./documents/client";
import APIProjectsIterations from "./iterations/client";

export default class APIProjects extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    documents = new APIProjectsDocuments(this);
    iterations = new APIProjectsIterations(this);

    async create(body: CreateProjectRequest, options?: RequestOptions): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: "/v1/projects",
            method: "POST",
            body: { ...(await ZCreateProjectRequest.parseAsync(body)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async update(projectId: string, body: Partial<BaseProjectInput>, options?: RequestOptions): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: `/v1/projects/${projectId}`,
            method: "PATCH",
            body: { ...(await ZBaseProject.partial().parseAsync(body)), ...(options?.body || {}) },
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

    async extract({
        project_id,
        iteration_id,
        document,
        documents,
        temperature,
        seed,
        store
    }: {
        project_id: string,
        iteration_id?: string,
        document?: MIMEDataInput,
        documents?: MIMEDataInput[],
        temperature?: number,
        seed?: number,
        store?: boolean,
    }, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        if (!document && (!documents || documents.length === 0)) {
            throw new Error("Either 'document' or 'documents' must be provided.");
        }
        const url = iteration_id ? `/v1/projects/extract/${project_id}/${iteration_id}` : `/v1/projects/extract/${project_id}`;

        // Only include optional parameters if they are provided
        const bodyParams: any = {
            documents: (await ZMIMEData.array().parseAsync([...document ? [document] : [], ...(documents || [])])).map(mimeToBlob)
        };

        if (temperature !== undefined) bodyParams.temperature = temperature;
        if (seed !== undefined) bodyParams.seed = seed;
        if (store !== undefined) bodyParams.store = store;

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
