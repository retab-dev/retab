import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { DatasetClusteringResponse } from "@/types";

export default class APIClustering extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string, { nClusters, totalSamples }: { nClusters?: number | null, totalSamples?: number | null } = {}): Promise<DatasetClusteringResponse> {
    let res = await this._fetch({
      url: `/v1/db/datasets/${schemaId}/clustering`,
      method: "GET",
      params: { "n_clusters": nClusters, "total_samples": totalSamples },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
