import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIExtractionsSub from "./extractions/client";
import APIExtractSub from "./extract/client";
import APILogExtractionSub from "./logExtraction/client";
import APICreateMessagesMockSub from "./createMessagesMock/client";
import APICreateMessagesSub from "./createMessages/client";
import APICreateInputsSub from "./createInputs/client";
import APICorrectImageOrientationSub from "./correctImageOrientation/client";
import APIEmailExtractionSub from "./emailExtraction/client";
import APIConvertToPdfSub from "./convertToPdf/client";
import APIPerformOcrOnlySub from "./performOcrOnly/client";
import APIComputeFieldLocationsSub from "./computeFieldLocations/client";
import APIPerformOcrSub from "./performOcr/client";
import APIParseSub from "./parse/client";

export default class APIDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  extractions = new APIExtractionsSub(this._client);
  extract = new APIExtractSub(this._client);
  logExtraction = new APILogExtractionSub(this._client);
  createMessagesMock = new APICreateMessagesMockSub(this._client);
  createMessages = new APICreateMessagesSub(this._client);
  createInputs = new APICreateInputsSub(this._client);
  correctImageOrientation = new APICorrectImageOrientationSub(this._client);
  emailExtraction = new APIEmailExtractionSub(this._client);
  convertToPdf = new APIConvertToPdfSub(this._client);
  performOcrOnly = new APIPerformOcrOnlySub(this._client);
  computeFieldLocations = new APIComputeFieldLocationsSub(this._client);
  performOcr = new APIPerformOcrSub(this._client);
  parse = new APIParseSub(this._client);

}
