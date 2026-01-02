import { CompositionClient, RequestOptions } from "../../../client.js";
import {
    ZPaginatedList,
    PaginatedList,
    InferFormSchemaRequest,
    InferFormSchemaResponse,
    ZInferFormSchemaRequest,
    ZInferFormSchemaResponse,
    MIMEDataInput,
    ZMIMEData,
    EditResponse,
    ZEditResponse,
} from "../../../types.js";
import {
    EditTemplate,
    ZEditTemplate,
    FormField,
} from "../../../generated_types.js";

export default class APIEditTemplates extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * List edit templates with pagination and optional filtering.
     *
     * @param params - Pagination and filter parameters
     * @param options - Optional request options
     * @returns PaginatedList of EditTemplate objects
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            filename,
            mime_type,
            include_embeddings = false,
            sort_by = "created_at",
        }: {
            before?: string;
            after?: string;
            limit?: number;
            order?: "asc" | "desc";
            filename?: string;
            mime_type?: string;
            include_embeddings?: boolean;
            sort_by?: string;
        } = {},
        options?: RequestOptions
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            filename,
            mime_type,
            include_embeddings,
            sort_by,
        };

        // Remove undefined values
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZPaginatedList, {
            url: "/v1/edit/templates",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Get an edit template by ID.
     *
     * @param template_id - The ID of the template to retrieve
     * @param options - Optional request options
     * @returns EditTemplate object
     */
    async get(template_id: string, options?: RequestOptions): Promise<EditTemplate> {
        return this._fetchJson(ZEditTemplate, {
            url: `/v1/edit/templates/${template_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Create a new edit template.
     *
     * @param params - CreateEditTemplateRequest containing:
     *   - name: Name of the template
     *   - document: The document to use as a template (MIMEData, file path, Buffer, or Readable)
     *   - form_fields: Array of form field definitions
     * @param options - Optional request options
     * @returns EditTemplate object
     */
    async create(
        {
            name,
            document,
            form_fields,
        }: {
            name: string;
            document: MIMEDataInput;
            form_fields: FormField[];
        },
        options?: RequestOptions
    ): Promise<EditTemplate> {
        const parsedDocument = await ZMIMEData.parseAsync(document);

        return this._fetchJson(ZEditTemplate, {
            url: "/v1/edit/templates",
            method: "POST",
            body: {
                name,
                document: parsedDocument,
                form_fields,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Update an existing edit template.
     *
     * @param template_id - The ID of the template to update
     * @param params - UpdateEditTemplateRequest containing optional fields to update:
     *   - name: New name for the template
     *   - form_fields: Updated array of form field definitions
     * @param options - Optional request options
     * @returns EditTemplate object
     */
    async update(
        template_id: string,
        {
            name,
            form_fields,
        }: {
            name?: string;
            form_fields?: FormField[];
        } = {},
        options?: RequestOptions
    ): Promise<EditTemplate> {
        const body: Record<string, any> = {};
        if (name !== undefined) body.name = name;
        if (form_fields !== undefined) body.form_fields = form_fields;

        return this._fetchJson(ZEditTemplate, {
            url: `/v1/edit/templates/${template_id}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete an edit template by ID.
     *
     * @param template_id - The ID of the template to delete
     * @param options - Optional request options
     */
    async delete(template_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/v1/edit/templates/${template_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Duplicate an existing edit template.
     *
     * @param template_id - The ID of the template to duplicate
     * @param params - DuplicateEditTemplateRequest containing optional:
     *   - name: Name for the duplicated template (defaults to "{original_name} (copy)")
     * @param options - Optional request options
     * @returns EditTemplate object (the new duplicated template)
     */
    async duplicate(
        template_id: string,
        { name }: { name?: string } = {},
        options?: RequestOptions
    ): Promise<EditTemplate> {
        const body: Record<string, any> = {};
        if (name !== undefined) body.name = name;

        return this._fetchJson(ZEditTemplate, {
            url: `/v1/edit/templates/${template_id}/duplicate`,
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Generate (infer) form schema from a PDF document.
     *
     * This method combines computer vision for precise bounding box detection
     * with LLM for semantic field naming (key, description) and type classification.
     *
     * Supported document formats:
     * - PDF: Direct processing (only PDF is supported)
     *
     * @param params - InferFormSchemaRequest containing:
     *   - document: MIMEData object, file path, Buffer, or Readable stream
     *   - model: LLM model for field naming (default: "retab-small")
     *   - instructions: Optional instructions to guide form field detection
     * @param options - Optional request options
     * @returns InferFormSchemaResponse containing:
     *   - form_schema: The detected form schema with field keys, descriptions, types, and bounding boxes
     *   - annotated_pdf: PDF with bounding boxes drawn around detected fields for visual verification
     *   - field_count: Number of fields detected
     */
    async generate(
        params: InferFormSchemaRequest,
        options?: RequestOptions
    ): Promise<InferFormSchemaResponse> {
        return this._fetchJson(ZInferFormSchemaResponse, {
            url: "/v1/edit/templates/generate",
            method: "POST",
            body: { ...(await ZInferFormSchemaRequest.parseAsync(params)), ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Fill a PDF form using a pre-defined template.
     *
     * This method uses a template's pre-defined form fields to fill a PDF form,
     * skipping the field detection step for faster processing.
     *
     * Use cases:
     * - Batch processing of the same form with different data
     * - Faster form filling when field detection is already done
     * - Consistent field mapping across multiple fills
     *
     * @param params - Fill request containing:
     *   - template_id: The template ID to use for filling
     *   - instructions: Instructions describing how to fill the form fields
     *   - model: LLM model for inference (default: "retab-small")
     * @param options - Optional request options
     * @returns EditResponse containing:
     *   - form_data: List of form fields with filled values
     *   - filled_document: The filled PDF document as MIMEData
     */
    async fill(
        {
            template_id,
            instructions,
            model = "retab-small",
        }: {
            template_id: string;
            instructions: string;
            model?: string;
        },
        options?: RequestOptions
    ): Promise<EditResponse> {
        return this._fetchJson(ZEditResponse, {
            url: "/v1/edit/templates/fill",
            method: "POST",
            body: {
                template_id,
                instructions,
                model,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }
}

