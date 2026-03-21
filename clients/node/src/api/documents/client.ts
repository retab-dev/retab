import * as z from "zod";

import { CompositionClient, RequestOptions } from "../../client.js";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResponse, ZParseResponse, ZParseRequest, RetabParsedChatCompletionChunk, ZRetabParsedChatCompletionChunk, EditRequest, EditResponse, ZEditRequest, ZEditResponse, SplitRequest, ZSplitRequest, ZSplitResponse, ClassifyRequest, ZClassifyRequest, ZClassifyResponse, ZMIMEData, MIMEDataInput } from "../../types.js";
import type { SplitResponse, ClassifyResponse } from "../../generated_types.js";


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
    async sources(
        extraction_id: string,
        {
            file,
            file_id,
            fileId,
        }: {
            file?: MIMEDataInput;
            file_id?: string;
            fileId?: string;
        } = {},
        options?: RequestOptions,
    ): Promise<Record<string, unknown>> {
        const extraction = await this._fetchJson(z.any(), {
            url: `/extractions/${extraction_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });

        let inferredFileId = file_id ?? fileId;
        if (!inferredFileId && file) {
            const parsedFile = await ZMIMEData.parseAsync(file);
            const maybeId = (parsedFile as { id?: string }).id;
            if (typeof maybeId === "string" && maybeId.length > 0) {
                inferredFileId = maybeId;
            }
        }
        if (!inferredFileId) {
            const extractionFileIds = Array.isArray(extraction?.file_ids)
                ? extraction.file_ids.filter((candidate: unknown): candidate is string => typeof candidate === "string" && candidate.length > 0)
                : [];
            if (extractionFileIds.length === 1) {
                inferredFileId = extractionFileIds[0];
            } else if (typeof extraction?.file_id === "string" && extraction.file_id.length > 0) {
                inferredFileId = extraction.file_id;
            }
        }

        if (!inferredFileId) {
            throw new Error("Unable to infer file_id. Provide file_id explicitly, pass a MIMEData with an id, or use an extraction with a single file.");
        }

        let data: Record<string, unknown> = {};
        const completion = extraction?.completion;
        const message = completion?.choices?.[0]?.message;
        if (message?.parsed && typeof message.parsed === "object") {
            data = message.parsed as Record<string, unknown>;
        } else if (typeof message?.content === "string") {
            try {
                data = JSON.parse(message.content) as Record<string, unknown>;
            } catch {
                data = {};
            }
        }

        const ocrResponse = await this._fetchJson(z.any(), {
            url: "/documents/perform_ocr_only",
            method: "POST",
            body: {
                file_id: inferredFileId,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });

        const ocrFileId = ocrResponse?.ocr_file_id;
        if (typeof ocrFileId !== "string" || ocrFileId.length === 0) {
            throw new Error("perform_ocr_only did not return a valid ocr_file_id");
        }

        return this._fetchJson(z.record(z.any()), {
            url: "/documents/compute_field_locations",
            method: "POST",
            body: {
                ocr_file_id: ocrFileId,
                ocr_result: ocrResponse?.ocr_result,
                data,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }
    async extract_stream(params: DocumentExtractRequest, options?: RequestOptions): Promise<AsyncGenerator<RetabParsedChatCompletionChunk>> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchStream(ZRetabParsedChatCompletionChunk, {
            url: "/documents/extractions",
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
}
