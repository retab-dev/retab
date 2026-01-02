import { CompositionClient, RequestOptions } from "../../client.js";
import {
    EditRequest,
    EditResponse,
    ZEditRequest,
    ZEditResponse,
} from "../../types.js";
import APIEditTemplates from "./templates/client.js";

export default class APIEdit extends CompositionClient {
    public templates: APIEditTemplates;

    constructor(client: CompositionClient) {
        super(client);
        this.templates = new APIEditTemplates(this);
    }

    /**
     * Edit a document by detecting and filling form fields with provided instructions.
     *
     * This method performs:
     * 1. Detection to identify form field bounding boxes
     * 2. LLM inference to name and describe detected fields
     * 3. LLM-based form filling using the provided instructions
     * 4. Returns the filled document with form field values populated
     *
     * Either `document` OR `template_id` must be provided, but not both.
     *
     * Supported document formats:
     * - PDF: Native form field detection and filling
     * - DOCX/DOC: Native editing to preserve styles and formatting
     * - PPTX/PPT: Native editing for presentations
     * - XLSX/XLS: Native editing for spreadsheets
     *
     * @param params - EditRequest containing:
     *   - instructions: Natural language instructions for filling (required)
     *   - document: MIMEData object, file path, Buffer, or Readable stream (optional, mutually exclusive with template_id)
     *   - model: LLM model for inference (default: "retab-small")
     *   - template_id: Template ID to use for filling with pre-defined form fields (optional, mutually exclusive with document)
     * @param options - Optional request options
     * @returns EditResponse containing form_data (filled fields) and filled_document (MIMEData)
     */
    async fill_document(params: EditRequest, options?: RequestOptions): Promise<EditResponse> {
        return this._fetchJson(ZEditResponse, {
            url: "/v1/edit/fill-document",
            method: "POST",
            body: { ...(await ZEditRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

