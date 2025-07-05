import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyImportAnnotationsCsvV1EvaluationsIoEvaluationIdImportAnnotationsCsvPost, ImportAnnotationsCsvResponse } from "@/types";

export default class APIImportAnnotationsCsv extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post(evaluationId: string, { delimiter, lineDelimiter, quote, ...body }: { delimiter?: string, lineDelimiter?: string, quote?: string } & BodyImportAnnotationsCsvV1EvaluationsIoEvaluationIdImportAnnotationsCsvPost): Promise<ImportAnnotationsCsvResponse> {
    let res = await this._fetch({
      url: `/v1/evaluations/io/${evaluationId}/import_annotations_csv`,
      method: "POST",
      params: { "delimiter": delimiter, "line_delimiter": lineDelimiter, "quote": quote },
      body: body,
      bodyMime: "multipart/form-data",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
