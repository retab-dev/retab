import { CompositionClient, RequestOptions } from "../../client.js";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResponse, ZParseResponse, ZParseRequest, DocumentCreateMessageRequest, DocumentMessage, ZDocumentMessage, ZDocumentCreateMessageRequest, DocumentCreateInputRequest, ZDocumentCreateInputRequest, RetabParsedChatCompletionChunk, ZRetabParsedChatCompletionChunk, EditRequest, EditResponse, ZEditRequest, ZEditResponse, SplitRequest, ZSplitRequest, ZSplitResponse, ClassifyRequest, ZClassifyRequest, ZClassifyResponse, GenerateSplitConfigRequest, ZGenerateSplitConfigRequest } from "../../types.js";
import type { SplitResponse, ClassifyResponse, GenerateSplitConfigResponse } from "../../generated_types.js";
import { ZGenerateSplitConfigResponse } from "../../generated_types.js";


export default class APIDocuments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async extract(params: DocumentExtractRequest, options?: RequestOptions): Promise<RetabParsedChatCompletion> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: "/documents/extract",
            method: "POST",
            body: { ...request, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async extract_stream(params: DocumentExtractRequest, options?: RequestOptions): Promise<AsyncGenerator<RetabParsedChatCompletionChunk>> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchStream(ZRetabParsedChatCompletionChunk, {
            url: "/documents/extract",
            method: "POST",
            body: { ...request, stream: true, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async parse(params: ParseRequest, options?: RequestOptions): Promise<ParseResponse> {
        return this._fetchJson(ZParseResponse, {
            url: "/documents/parse",
            method: "POST",
            body: { ...(await ZParseRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async create_messages(params: DocumentCreateMessageRequest, options?: RequestOptions): Promise<DocumentMessage> {
        return this._fetchJson(ZDocumentMessage, {
            url: "/documents/create_messages",
            method: "POST",
            body: { ...(await ZDocumentCreateMessageRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async create_inputs(params: DocumentCreateInputRequest, options?: RequestOptions): Promise<DocumentMessage> {
        return this._fetchJson(ZDocumentMessage, {
            url: "/documents/create_inputs",
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
     *   - instructions: Natural language instructions for filling (required)
     *   - document: MIMEData object, file path, Buffer, or Readable stream (optional, mutually exclusive with template_id)
     *   - model: LLM model for inference (default: "retab-small")
     *   - template_id: Template ID to use for filling with pre-defined form fields (optional, mutually exclusive with document)
     * @param options - Optional request options
     * @returns EditResponse containing form_data (filled fields) and filled_document (MIMEData)
     */
    async edit(params: EditRequest, options?: RequestOptions): Promise<EditResponse> {
        return this._fetchJson(ZEditResponse, {
            url: "/documents/edit",
            method: "POST",
            body: { ...(await ZEditRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    /**
     * Split a document into sections based on provided subdocuments.
     * 
     * This method analyzes a multi-page document and classifies pages into
     * user-defined subdocuments, returning the assigned pages for each section.
     * 
     * @param params - SplitRequest containing:
     *   - document: MIMEData object, file path, Buffer, or Readable stream
     *   - subdocuments: Array of subdocuments with 'name', 'description', and optional 'partition_key'
     *   - model: LLM model for inference (e.g., "retab-small")
     *   - context: Optional business context for the split
     *   - n_consensus: Optional number of split runs to use for consensus scoring
     * @param options - Optional request options
     * @returns SplitResponse containing splits with page lists, optional likelihood/votes,
     * and partitions when partition_key is configured
     * 
     * @example
     * ```typescript
     * const response = await retab.documents.split({
     *     document: "invoice_batch.pdf",
     *     model: "retab-small",
     *     subdocuments: [
     *         { name: "invoice", description: "Invoice documents with billing information" },
     *         { name: "receipt", description: "Receipt documents for payments" },
     *         { name: "contract", description: "Legal contract documents" },
     *     ],
     *     n_consensus: 3,
     * });
     * for (const split of response.splits) {
     *     console.log(`${split.name}: pages ${split.pages.join(', ')} likelihood=${split.likelihood}`);
     * }
     * ```
     */
    async split(params: SplitRequest, options?: RequestOptions): Promise<SplitResponse> {
        return this._fetchJson(ZSplitResponse, {
            url: "/documents/split",
            method: "POST",
            body: { ...(await ZSplitRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    /**
     * Classify a document into one of the provided categories.
     *
     * This method analyzes a document and classifies it into exactly one
     * of the user-defined categories, returning the classification with
     * chain-of-thought reasoning explaining the decision.
     *
     * @param params - ClassifyRequest containing:
     *   - document: MIMEData object, file path, Buffer, or Readable stream
     *   - categories: Array of categories with 'name' and 'description'
     *   - model: LLM model for inference (e.g., "retab-small")
     *   - first_n_pages: (optional) Only use the first N pages for classification. Useful for large documents.
     * @param options - Optional request options
     * @returns ClassifyResponse containing result with reasoning and classification
     *
     * @example
     * ```typescript
     * const response = await retab.documents.classify({
     *     document: "invoice.pdf",
     *     model: "retab-small",
     *     categories: [
     *         { name: "invoice", description: "Invoice documents with billing information" },
     *         { name: "receipt", description: "Receipt documents for payments" },
     *         { name: "contract", description: "Legal contract documents" },
     *     ],
     *     first_n_pages: 3  // Optional: only use first 3 pages
     * });
     * console.log(`Classification: ${response.result.classification}`);
     * console.log(`Reasoning: ${response.result.reasoning}`);
     * ```
     */
    async classify(params: ClassifyRequest, options?: RequestOptions): Promise<ClassifyResponse> {
        return this._fetchJson(ZClassifyResponse, {
            url: "/documents/classify",
            method: "POST",
            body: { ...(await ZClassifyRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
    /**
     * Analyze a document and suggest the subdocuments to use for a later split call.
     *
     * This is useful when you do not know the right split configuration yet.
     * The response can be passed directly into `documents.split(...)`.
     *
     * @param params - GenerateSplitConfigRequest containing:
     *   - document: MIMEData object, file path, Buffer, or Readable stream
     *   - model: LLM model for inference (e.g., "retab-small")
     * @param options - Optional request options
     * @returns GenerateSplitConfigResponse containing suggested subdocuments
     *
     * @example
     * ```typescript
     * const config = await retab.documents.generate_split_config({
     *   document: "property_portfolio.pdf",
     *   model: "retab-small",
     * });
     *
     * const result = await retab.documents.split({
     *   document: "property_portfolio.pdf",
     *   model: "retab-small",
     *   subdocuments: config.subdocuments,
     * });
     * ```
     */
    async generate_split_config(params: GenerateSplitConfigRequest, options?: RequestOptions): Promise<GenerateSplitConfigResponse> {
        return this._fetchJson(ZGenerateSplitConfigResponse, {
            url: "/documents/split/generate_config",
            method: "POST",
            body: { ...(await ZGenerateSplitConfigRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
