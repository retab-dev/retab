import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import APIProcessSub from "./process/client";
import APIDistancesSub from "./distances/client";
import { ZRetabTypesPredictionsPredictionDataOutput, RetabTypesPredictionsPredictionDataOutput, ZPatchIterationDocumentPredictionRequest, PatchIterationDocumentPredictionRequest } from "@/types";

export default class APIDocumentId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  process = new APIProcessSub(this._client);
  distances = new APIDistancesSub(this._client);

  async get(evaluationId: string, iterationId: string, documentId: string): Promise<RetabTypesPredictionsPredictionDataOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/documents/${documentId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesPredictionsPredictionDataOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
  async patch(evaluationId: string, iterationId: string, documentId: string, { ...body }: PatchIterationDocumentPredictionRequest): Promise<RetabTypesPredictionsPredictionDataOutput> {
    let res = await this._fetch({
      url: `/v1/evaluations/${evaluationId}/iterations/${iterationId}/documents/${documentId}`,
      method: "PATCH",
      body: body,
      bodyMime: "application/json",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZRetabTypesPredictionsPredictionDataOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
