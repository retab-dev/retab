import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import APIImportDocumentsSub from "./importDocuments/client";
import APIImportAnnotationsCsvSub from "./importAnnotationsCsv/client";
import APIExportDocumentsSub from "./exportDocuments/client";

export default class APIEvaluationId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  importDocuments = new APIImportDocumentsSub(this._client);
  importAnnotationsCsv = new APIImportAnnotationsCsvSub(this._client);
  exportDocuments = new APIExportDocumentsSub(this._client);

}
