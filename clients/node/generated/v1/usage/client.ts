import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APITimeSeriesSub from "./timeSeries/client";
import APIMonthlyCreditsSub from "./monthlyCredits/client";
import APIAutomationSub from "./automation/client";
import APIProcessorSub from "./processor/client";
import APITimeSeriesOrganizationSub from "./timeSeriesOrganization/client";
import APITimeSeriesPreprocessingSub from "./timeSeriesPreprocessing/client";
import APIPreprocessingLogSub from "./preprocessingLog/client";

export default class APIUsage extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  timeSeries = new APITimeSeriesSub(this._client);
  monthlyCredits = new APIMonthlyCreditsSub(this._client);
  automation = new APIAutomationSub(this._client);
  processor = new APIProcessorSub(this._client);
  timeSeriesOrganization = new APITimeSeriesOrganizationSub(this._client);
  timeSeriesPreprocessing = new APITimeSeriesPreprocessingSub(this._client);
  preprocessingLog = new APIPreprocessingLogSub(this._client);

}
