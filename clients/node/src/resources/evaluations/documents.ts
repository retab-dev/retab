import { SyncAPIResource, AsyncAPIResource } from '../../resource.js';
import { PreparedRequest, DeleteResponse } from '../../types/standards.js';
import { MIMEData } from '../../types/mime.js';
import { prepareMimeDocument } from '../../utils/mime.js';
import { RetabParsedChatCompletion } from '../../types/documents/extractions.js';

// Evaluation-specific types
export interface DocumentItem {
  mime_data: MIMEData;
  annotation: Record<string, any>;
  annotation_metadata?: Record<string, any> | null;
}

export interface EvaluationDocument {
  id: string;
  evaluation_id: string;
  mime_data: MIMEData;
  annotation: Record<string, any>;
  annotation_metadata?: Record<string, any> | null;
  created_at: string;
  updated_at: string;
}

export interface PatchEvaluationDocumentRequest {
  annotation: Record<string, any>;
}

class DocumentsMixin {
  public prepareGet(evaluationId: string, documentId: string): PreparedRequest {
    return {
      method: 'GET',
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}`,
    };
  }

  public prepareCreate(evaluationId: string, document: MIMEData, annotation: Record<string, any>): PreparedRequest {
    const documentItem: DocumentItem = {
      mime_data: document,
      annotation,
      annotation_metadata: null,
    };
    return {
      method: 'POST',
      url: `/v1/evaluations/${evaluationId}/documents`,
      data: documentItem,
    };
  }

  public prepareList(evaluationId: string): PreparedRequest {
    return {
      method: 'GET',
      url: `/v1/evaluations/${evaluationId}/documents`,
    };
  }

  public prepareUpdate(evaluationId: string, documentId: string, annotation: Record<string, any>): PreparedRequest {
    const updateRequest: PatchEvaluationDocumentRequest = { annotation };
    return {
      method: 'PATCH',
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}`,
      data: updateRequest,
    };
  }

  public prepareDelete(evaluationId: string, documentId: string): PreparedRequest {
    return {
      method: 'DELETE',
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}`,
    };
  }

  public prepareLlmAnnotate(evaluationId: string, documentId: string): PreparedRequest {
    return {
      method: 'POST',
      url: `/v1/evaluations/${evaluationId}/documents/${documentId}/llm-annotate`,
      data: { stream: false },
    };
  }
}

export class Documents extends SyncAPIResource {
  private mixin = new DocumentsMixin();

  async create(
    evaluationId: string,
    document: string | Buffer | MIMEData,
    annotation: Record<string, any>
  ): Promise<EvaluationDocument> {
    const mimeData = prepareMimeDocument(document);
    const preparedRequest = this.mixin.prepareCreate(evaluationId, mimeData, annotation);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument;
  }

  async get(evaluationId: string, documentId: string): Promise<EvaluationDocument> {
    const preparedRequest = this.mixin.prepareGet(evaluationId, documentId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument;
  }

  async list(evaluationId: string): Promise<EvaluationDocument[]> {
    const preparedRequest = this.mixin.prepareList(evaluationId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument[];
  }

  async update(
    evaluationId: string,
    documentId: string,
    annotation: Record<string, any>
  ): Promise<EvaluationDocument> {
    const preparedRequest = this.mixin.prepareUpdate(evaluationId, documentId, annotation);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument;
  }

  async delete(evaluationId: string, documentId: string): Promise<DeleteResponse> {
    const preparedRequest = this.mixin.prepareDelete(evaluationId, documentId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as DeleteResponse;
  }

  async llmAnnotate(evaluationId: string, documentId: string): Promise<RetabParsedChatCompletion> {
    const preparedRequest = this.mixin.prepareLlmAnnotate(evaluationId, documentId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as RetabParsedChatCompletion;
  }
}

export class AsyncDocuments extends AsyncAPIResource {
  private mixin = new DocumentsMixin();

  async create(
    evaluationId: string,
    document: string | Buffer | MIMEData,
    annotation: Record<string, any>
  ): Promise<EvaluationDocument> {
    const mimeData = prepareMimeDocument(document);
    const preparedRequest = this.mixin.prepareCreate(evaluationId, mimeData, annotation);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument;
  }

  async get(evaluationId: string, documentId: string): Promise<EvaluationDocument> {
    const preparedRequest = this.mixin.prepareGet(evaluationId, documentId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument;
  }

  async list(evaluationId: string): Promise<EvaluationDocument[]> {
    const preparedRequest = this.mixin.prepareList(evaluationId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument[];
  }

  async update(
    evaluationId: string,
    documentId: string,
    annotation: Record<string, any>
  ): Promise<EvaluationDocument> {
    const preparedRequest = this.mixin.prepareUpdate(evaluationId, documentId, annotation);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as EvaluationDocument;
  }

  async delete(evaluationId: string, documentId: string): Promise<DeleteResponse> {
    const preparedRequest = this.mixin.prepareDelete(evaluationId, documentId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as DeleteResponse;
  }

  async llmAnnotate(evaluationId: string, documentId: string): Promise<RetabParsedChatCompletion> {
    const preparedRequest = this.mixin.prepareLlmAnnotate(evaluationId, documentId);
    const response = await this._client._preparedRequest(preparedRequest);
    return response as RetabParsedChatCompletion;
  }
}