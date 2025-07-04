import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyHandleEndpointProcessingV1ProcessorsAutomationsEndpointsProcessEndpointIdPost, AutomationLog } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(endpointId: string, { id, idempotencyKey, ...body }: { id: string, idempotencyKey?: string | null } & BodyHandleEndpointProcessingV1ProcessorsAutomationsEndpointsProcessEndpointIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/processors/automations/endpoints/process/${endpointId}`,
      method: "POST",
      params: { "id": id },
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
