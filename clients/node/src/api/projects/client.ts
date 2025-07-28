import { CompositionClient } from "@/client";
import { BaseProjectInput, dataArray, ModelProject, ZBaseProject, ZModelProject } from "@/types";
import APIProjectsDocuments from "./documents/client";
import APIProjectsIterations from "./iterations/client";

export default class APIProjects extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    documents = new APIProjectsDocuments(this);
    iterations = new APIProjectsIterations(this);

    async create(body: BaseProjectInput): Promise<ModelProject> {
        return this._fetchJson(ZModelProject, {
            url: "/v1/projects",
            method: "POST",
            body: await ZBaseProject.parseAsync(body),
        });
    }

    async update(projectId: string, body: Partial<BaseProjectInput>): Promise<ModelProject> {
        return this._fetchJson(ZModelProject, {
            url: `/v1/projects/${projectId}`,
            method: "PATCH",
            body: await ZBaseProject.partial().parseAsync(body),
        });
    }

    async list(): Promise<ModelProject[]> {
        return this._fetchJson(dataArray(ZModelProject), {
            url: "/v1/projects",
            method: "GET",
        });
    }

    async get(projectId: string): Promise<ModelProject> {
        return this._fetchJson(ZModelProject, {
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
}
