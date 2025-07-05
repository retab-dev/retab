import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { EvaluateSchemaRequest, EvaluateSchemaResponse } from "@/types";

export default class APIEvaluate extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: EvaluateSchemaRequest): Promise<EvaluateSchemaResponse> {
    let res = await this._fetch({
      url: `/v1/schemas/evaluate`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
