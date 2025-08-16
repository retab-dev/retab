import { CompositionClient } from "@/client";
import { mimeToBlob } from "@/mime";
import { BaseProjectInput, dataArray, Project, ZBaseProject, ZProject, ZCreateProjectRequest, CreateProjectRequest, MIMEDataInput, ZMIMEData, ZSchema } from "@/types";
import APIProjectsDocuments from "./documents/client";
import APIProjectsIterations from "./iterations/client";

export default class APIProjects extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    documents = new APIProjectsDocuments(this);
    iterations = new APIProjectsIterations(this);

    async create(body: CreateProjectRequest): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: "/v1/projects",
            method: "POST",
            body: await ZCreateProjectRequest.parseAsync(body),
        });
    }

    async update(projectId: string, body: Partial<BaseProjectInput>): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: `/v1/projects/${projectId}`,
            method: "PATCH",
            body: await ZBaseProject.partial().parseAsync(body),
        });
    }

    async list(): Promise<Project[]> {
        return this._fetchJson(dataArray(ZProject), {
            url: "/v1/projects",
            method: "GET",
        });
    }

    async get(projectId: string): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: `/v1/projects/${projectId}`,
            method: "GET",
        });
    }

    async delete(projectId: string): Promise<void> {
        return this._fetchJson({
            url: `/v1/projects/${projectId}`,
            method: "DELETE",
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
        iteration_id: string,
        document?: MIMEDataInput,
        documents?: MIMEDataInput[],
        temperature?: number,
        seed?: number,
        store?: boolean,
    }) {
        if (!document && (!documents || documents.length === 0)) {
            throw new Error("Either 'document' or 'documents' must be provided.");
        }
        let url = `/v1/projects/extract/${project_id}/${iteration_id}`;
        return this._fetchJson(ZSchema, {
            url,
            method: "POST",
            body: {
                temperature, seed, store,
                documents: (await ZMIMEData.array().parseAsync([...document ? [document] : [], documents || []])).map(mimeToBlob)
            },
            bodyMime: "multipart/form-data",
        });
    }
}
