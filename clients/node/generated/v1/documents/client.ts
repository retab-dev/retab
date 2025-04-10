import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIMonthlyUsageSub from "./monthlyUsage/client";
import APIExtractionsSub from "./extractions/client";
import APILogExtractionSub from "./logExtraction/client";
import APICreateMessagesMockSub from "./createMessagesMock/client";
import APICreateMessagesSub from "./createMessages/client";
import APICreateInputsSub from "./createInputs/client";
import APICorrectImageOrientationSub from "./correctImageOrientation/client";
import APIEmailExtractionSub from "./emailExtraction/client";

export default class APIDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  monthlyUsage = new APIMonthlyUsageSub(this._client);
  extractions = new APIExtractionsSub(this._client);
  logExtraction = new APILogExtractionSub(this._client);
  createMessagesMock = new APICreateMessagesMockSub(this._client);
  createMessages = new APICreateMessagesSub(this._client);
  createInputs = new APICreateInputsSub(this._client);
  correctImageOrientation = new APICorrectImageOrientationSub(this._client);
  emailExtraction = new APIEmailExtractionSub(this._client);

}
