import { CompositionClient, RequestOptions } from "../../client.js";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResult, ZParseResult, ZParseRequest, DocumentCreateMessageRequest, DocumentMessage, ZDocumentMessage, ZDocumentCreateMessageRequest, DocumentCreateInputRequest, ZDocumentCreateInputRequest, RetabParsedChatCompletionChunk, ZRetabParsedChatCompletionChunk, EditRequest, EditResponse, ZEditRequest, ZEditResponse } from "../../types.js";


export default class APIDocuments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async extract(params: DocumentExtractRequest, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: "/v1/documents/extract",
            method: "POST",
            body: { ...request, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async extract_stream(params: DocumentExtractRequest, options?: RequestOptions): Promise<AsyncGenerator<RetabParsedChatCompletionChunk>> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchStream(ZRetabParsedChatCompletionChunk, {
            url: "/v1/documents/extract",
            method: "POST",
            body: { ...request, stream: true, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async parse(params: ParseRequest, options?: RequestOptions): Promise<ParseResult> {
        return this._fetchJson(ZParseResult, {
            url: "/v1/documents/parse",
            method: "POST",
            body: { ...(await ZParseRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async create_messages(params: DocumentCreateMessageRequest, options?: RequestOptions): Promise<DocumentMessage> {
        return this._fetchJson(ZDocumentMessage, {
            url: "/v1/documents/create_messages",
            method: "POST",
            body: { ...(await ZDocumentCreateMessageRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async create_inputs(params: DocumentCreateInputRequest, options?: RequestOptions): Promise<DocumentMessage> {
        return this._fetchJson(ZDocumentMessage, {
            url: "/v1/documents/create_inputs",
            method: "POST",
            body: { ...(await ZDocumentCreateInputRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    /**
     * Edit a PDF document by automatically detecting and filling form fields.
     * 
     * This method combines OCR text detection with LLM-based form field inference
     * and filling. It accepts any PDF document and natural language instructions
     * describing the values to fill in.
     * 
     * @param params - EditRequest containing:
     *   - pdf_base64: Base64-encoded PDF file to process
     *   - filling_instructions: Natural language instructions for filling
     *   - model: LLM model for inference (default: "gemini-2.5-pro")
     *   - annotation_level: OCR granularity - "block", "line", or "token" (default: "line")
     * @param options - Optional request options
     * @returns EditResponse containing form schema, filled values, and output PDFs
     */
    async edit(params: EditRequest, options?: RequestOptions): Promise<EditResponse> {
        return this._fetchJson(ZEditResponse, {
            url: "/v1/documents/edit",
            method: "POST",
            body: { ...(await ZEditRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
