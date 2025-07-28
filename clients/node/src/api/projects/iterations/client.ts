import { CompositionClient } from "@/client";
import * as z from "zod";
import { ZInferenceSettings, ZPatchIterationRequest, dataArray, ZIterationsIteration, IterationsIteration } from "@/types";

export default class APIProjectsIterations extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async create(projectId: string, body: z.input<typeof ZInferenceSettings>): Promise<IterationsIteration> {
        return this._fetchJson(ZIterationsIteration, {
            url: `/v1/projects/${projectId}/iterations`,
            method: "POST",
            body: await ZInferenceSettings.parseAsync(body),
        });
    }

    async update(projectId: string, iterationId: string, body: z.input<typeof ZPatchIterationRequest>): Promise<IterationsIteration> {
        return this._fetchJson(ZIterationsIteration, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}`,
            method: "PATCH",
            body: await ZPatchIterationRequest.parseAsync(body),
        });
    }

    async list(projectId: string, params?: {model?: string}): Promise<IterationsIteration[]> {
        return this._fetchJson(dataArray(ZIterationsIteration), {
            url: `/v1/projects/${projectId}/iterations`,
            method: "GET",
            params: params,
        });
    }

    async get(projectId: string, iterationId: string): Promise<IterationsIteration> {
        return this._fetchJson(ZIterationsIteration, {
            url: `/v1/projects/${projectId}/iterations/${iterationId}`,
            method: "GET",
        });
    }

    async delete(projectId: string, iterationId: string): Promise<void> {
        return this._fetchJson({
            url: `/v1/projects/${projectId}/iterations/${iterationId}`,
            method: "DELETE",
        });
    }
}
