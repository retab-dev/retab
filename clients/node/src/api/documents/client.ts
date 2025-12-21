import { CompositionClient, RequestOptions } from "../../client.js";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResult, ZParseResult, ZParseRequest, DocumentCreateMessageRequest, DocumentMessage, ZDocumentMessage, ZDocumentCreateMessageRequest, DocumentCreateInputRequest, ZDocumentCreateInputRequest, RetabParsedChatCompletionChunk, ZRetabParsedChatCompletionChunk, EditRequest, EditResponse, ZEditRequest, ZEditResponse, SplitRequest, ZSplitRequest, ZSplitResponse } from "../../types.js";
import type { SplitResponse } from "../../generated_types.js";


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
     * Either `document` OR `template_id` must be provided, but not both.
     * - When `document` is provided: OCR + LLM inference to detect and fill form fields
     * - When `template_id` is provided: Uses pre-defined form fields from the template (PDF only)
     * 
     * @param params - EditRequest containing:
     *   - filling_instructions: Natural language instructions for filling (required)
     *   - document: MIMEData object, file path, Buffer, or Readable stream (optional, mutually exclusive with template_id)
     *   - model: LLM model for inference (default: "retab-small")
     *   - template_id: Template ID to use for filling with pre-defined form fields (optional, mutually exclusive with document)
     * @param options - Optional request options
     * @returns EditResponse containing form_data (filled fields) and filled_document (MIMEData)
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
    /**
     * Split a document into sections based on provided categories.
     * 
     * This method analyzes a multi-page document and classifies pages into 
     * user-defined categories, returning the page ranges for each section.
     * 
     * @param params - SplitRequest containing:
     *   - document: MIMEData object, file path, Buffer, or Readable stream
     *   - categories: Array of categories with 'name' and 'description'
     *   - model: LLM model for inference (e.g., "retab-small")
     * @param options - Optional request options
     * @returns SplitResponse containing splits array with name, start_page, and end_page for each section
     * 
     * @example
     * ```typescript
     * const response = await retab.documents.split({
     *     document: "invoice_batch.pdf",
     *     model: "retab-small",
     *     categories: [
     *         { name: "invoice", description: "Invoice documents with billing information" },
     *         { name: "receipt", description: "Receipt documents for payments" },
     *         { name: "contract", description: "Legal contract documents" },
     *     ]
     * });
     * for (const split of response.splits) {
     *     console.log(`${split.name}: pages ${split.start_page}-${split.end_page}`);
     * }
     * ```
     */
    async split(params: SplitRequest, options?: RequestOptions): Promise<SplitResponse> {
        return this._fetchJson(ZSplitResponse, {
            url: "/v1/documents/split",
            method: "POST",
            body: { ...(await ZSplitRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
