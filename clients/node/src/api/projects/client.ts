import { CompositionClient } from "@/client";
import { BaseProjectInput, Project, ZBaseProject, ZProject } from "@/types";

export default class APIProjects extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async create(body: BaseProjectInput): Promise<Project> {
        return this._fetchJson(ZProject, {
            url: "/v1/projects",
            method: "POST",
            body: await ZBaseProject.parseAsync(body),
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
        return this._fetchJson(ZProject.array(), {
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
