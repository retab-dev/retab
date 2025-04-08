import { AbstractClient, CompositionClient } from '@/client';
import APIMonthlyUsage from "./monthlyUsage/client";
import APIExtractions from "./extractions/client";
import APILogExtraction from "./logExtraction/client";
import APICreateMessagesMock from "./createMessagesMock/client";
import APICreateMessages from "./createMessages/client";
import APICreateInputs from "./createInputs/client";
import APICorrectImageOrientation from "./correctImageOrientation/client";
import APIEmailExtraction from "./emailExtraction/client";

export default class APIDocuments extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  monthlyUsage = new APIMonthlyUsage(this);
  extractions = new APIExtractions(this);
  logExtraction = new APILogExtraction(this);
  createMessagesMock = new APICreateMessagesMock(this);
  createMessages = new APICreateMessages(this);
  createInputs = new APICreateInputs(this);
  correctImageOrientation = new APICorrectImageOrientation(this);
  emailExtraction = new APIEmailExtraction(this);

}
