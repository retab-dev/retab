import { CompositionClient } from "@/client";
import { BaseProjectInput, dataArray, Project, ZBaseProject, ZProject, ZCreateProjectRequest, CreateProjectRequest } from "@/types";
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
}
