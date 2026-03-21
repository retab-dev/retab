import { CompositionClient, RequestOptions } from "../../../client.js";
import {
    EditResponse,
    InferFormSchemaRequest,
    InferFormSchemaResponse,
    ZEditResponse,
    ZInferFormSchemaRequest,
    ZInferFormSchemaResponse,
} from "../../../types.js";

export default class APIEditTemplates extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
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
            url: "/edit/templates/generate",
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
     *   - color: Hex color code for filled text (default: "#000080")
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
            color,
        }: {
            template_id: string;
            instructions: string;
            model?: string;
            color?: string;
        },
        options?: RequestOptions
    ): Promise<EditResponse> {
        const body: Record<string, unknown> = {
            template_id,
            instructions,
            model,
        };
        if (color !== undefined) {
            body.config = { color };
        }

        return this._fetchJson(ZEditResponse, {
            url: "/edit/templates/fill",
            method: "POST",
            body: {
                ...body,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
