export const FieldUnset = Symbol('FIELD_UNSET');
export type FieldUnset = typeof FieldUnset;

export interface ErrorDetail {
  code: string;
  message: string;
  details?: Record<string, any>;
}

export interface StandardErrorResponse {
  detail: ErrorDetail;
}

export interface StreamingBaseModel {
  streaming_error?: ErrorDetail | null;
}

export interface DocumentPreprocessResponseContent {
  messages: Array<Record<string, any>>;
  json_schema: Record<string, any>;
}

export type HTTPMethod = 'POST' | 'GET' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS' | 'CONNECT' | 'TRACE';

export interface PreparedRequest {
  method: HTTPMethod;
  url: string;
  data?: Record<string, any> | null;
  params?: Record<string, any> | null;
  formData?: Record<string, any> | null;
  files?: Record<string, any> | Array<[string, [string, Buffer, string]]> | null;
  idempotencyKey?: string | null;
  raiseForStatus?: boolean;
}

export interface DeleteResponse {
  success: boolean;
  id: string;
}

export interface ExportResponse {
  success: boolean;
  path: string;
}