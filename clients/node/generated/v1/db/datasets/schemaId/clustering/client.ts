import { AbstractClient, CompositionClient } from '@/client';
import { DatasetClusteringResponse } from "@/types";

export default class APIClustering extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(schemaId: string, { nClusters, totalSamples }: { nClusters?: number | null, totalSamples?: number | null } = {}): Promise<DatasetClusteringResponse> {
    return this._fetch({
      url: `/v1/db/datasets/${schemaId}/clustering`,
      method: "GET",
      params: { "n_clusters": nClusters, "total_samples": totalSamples },
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
  }
  
}
