import { AbstractClient, CompositionClient } from '@/client';
import APIFileId from "./fileId/client";

export default class APIFileScores extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  fileId = new APIFileId(this);

}
