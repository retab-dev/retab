import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { AlignDictsRequest, AlignDictsResponse } from "@/types";

export default class APIAlignDicts extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: AlignDictsRequest): Promise<AlignDictsResponse> {
    let res = await this._fetch({
      url: `/v1/consensus/align_dicts`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
