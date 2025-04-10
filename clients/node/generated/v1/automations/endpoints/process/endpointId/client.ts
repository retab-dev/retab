import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyHandleEndpointProcessingV1AutomationsEndpointsProcessEndpointIdPost, AutomationLog } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(endpointId: string, { idempotencyKey, ...body }: { idempotencyKey?: string | null } & BodyHandleEndpointProcessingV1AutomationsEndpointsProcessEndpointIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/automations/endpoints/process/${endpointId}`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
