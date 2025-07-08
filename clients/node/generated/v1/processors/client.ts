import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIAutomationsSub from "./automations/client";
import APIProcessorIdSub from "./processorId/client";
import { ZPaginatedList, PaginatedList, ZProcessorConfig, ProcessorConfig, ZStoredProcessor, StoredProcessor } from "@/types";

export default class APIProcessors extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  automations = new APIAutomationsSub(this._client);
  processorId = new APIProcessorIdSub(this._client);

  async get({ before, after, limit, order, id, name, modality, model, schemaId, schemaDataId }: { before?: string | null, after?: string | null, limit?: number | null, order?: "asc" | "desc" | null, id?: string | null, name?: string | null, modality?: string | null, model?: string | null, schemaId?: string | null, schemaDataId?: string | null } = {}): Promise<PaginatedList> {
    let res = await this._fetch({
      url: `/v1/processors`,
      method: "GET",
      params: { "before": before, "after": after, "limit": limit, "order": order, "id": id, "name": name, "modality": modality, "model": model, "schema_id": schemaId, "schema_data_id": schemaDataId },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZPaginatedList.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async post({ ...body }: ProcessorConfig): Promise<StoredProcessor> {
    let res = await this._fetch({
      url: `/v1/processors`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZStoredProcessor.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
