import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyHandleEndpointProcessingV1EndpointsEndpointIdPost, AutomationLog } from "@/types";

export default class APIEndpointId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(endpointId: string, { idempotencyKey, ...body }: { idempotencyKey?: string | null } & BodyHandleEndpointProcessingV1EndpointsEndpointIdPost): Promise<AutomationLog> {
    let res = await this._fetch({
      url: `/v1/endpoints/${endpointId}`,
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
