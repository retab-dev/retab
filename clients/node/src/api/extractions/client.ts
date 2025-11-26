import { CompositionClient, RequestOptions } from "../../client.js";
import { ZPaginatedList, PaginatedList } from "../../types.js";
import * as z from "zod";

// Response types for extractions API
const ZExtractionCountResponse = z.object({
    count: z.number(),
});
type ExtractionCountResponse = z.infer<typeof ZExtractionCountResponse>;

const ZDownloadResponse = z.object({
    download_url: z.string(),
    filename: z.string(),
    expires_at: z.string(),
});
type DownloadResponse = z.infer<typeof ZDownloadResponse>;

const ZExportToCsvResponse = z.object({
    csv_data: z.string(),
    rows: z.number(),
    columns: z.number(),
});
type ExportToCsvResponse = z.infer<typeof ZExportToCsvResponse>;

const ZDistinctFieldValues = z.record(z.array(z.string()));
type DistinctFieldValues = z.infer<typeof ZDistinctFieldValues>;

// Generic extraction object (flexible since schema varies)
const ZExtraction = z.record(z.any());
type Extraction = z.infer<typeof ZExtraction>;

export default class APIExtractions extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * List extractions with pagination and filtering.
     */
    async list({
        before,
        after,
        limit = 10,
        order = "desc",
        origin_dot_type,
        origin_dot_id,
        from_date,
        to_date,
        human_review_status,
        metadata,
        filename,
    }: {
        before?: string;
        after?: string;
        limit?: number;
        order?: "asc" | "desc";
        origin_dot_type?: string;
        origin_dot_id?: string;
        from_date?: string;
        to_date?: string;
        human_review_status?: string;
        metadata?: Record<string, string>;
        filename?: string;
    } = {}, options?: RequestOptions): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            origin_dot_type,
            origin_dot_id,
            from_date,
            to_date,
            human_review_status,
            filename,
            // Note: metadata must be JSON-serialized as the backend expects a JSON string
            metadata: metadata ? JSON.stringify(metadata) : undefined,
        };

        // Remove undefined values
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZPaginatedList, {
            url: "/v1/extractions",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Count extractions matching filters.
     */
    async count({
        origin_dot_type,
        origin_dot_id,
        human_review_status = "review_required",
        metadata,
    }: {
        origin_dot_type?: string;
        origin_dot_id?: string;
        human_review_status?: string;
        metadata?: Record<string, string>;
    } = {}, options?: RequestOptions): Promise<ExtractionCountResponse> {
        const params: Record<string, any> = {
            origin_dot_type,
            origin_dot_id,
            human_review_status,
            // Note: metadata must be JSON-serialized as the backend expects a JSON string
            metadata: metadata ? JSON.stringify(metadata) : undefined,
        };

        // Remove undefined values
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZExtractionCountResponse, {
            url: "/v1/extractions/count",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Download extractions in various formats. Returns download_url, filename, and expires_at.
     */
    async download({
        order = "desc",
        origin_dot_id,
        from_date,
        to_date,
        human_review_status,
        metadata,
        filename,
        format = "jsonl",
    }: {
        order?: "asc" | "desc";
        origin_dot_id?: string;
        from_date?: string;
        to_date?: string;
        human_review_status?: string;
        metadata?: Record<string, string>;
        filename?: string;
        format?: "jsonl" | "csv" | "xlsx";
    } = {}, options?: RequestOptions): Promise<DownloadResponse> {
        const params: Record<string, any> = {
            order,
            origin_dot_id,
            from_date,
            to_date,
            human_review_status,
            filename,
            format,
            // Note: metadata must be JSON-serialized as the backend expects a JSON string
            metadata: metadata ? JSON.stringify(metadata) : undefined,
        };

        // Remove undefined values
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined)
        );

        return this._fetchJson(ZDownloadResponse, {
            url: "/v1/extractions/download",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Export extractions as CSV. Returns csv_data, rows, and columns.
     */
    async getPayloadForExport({
        project_id,
        extraction_ids,
        json_schema,
        delimiter = ";",
        line_delimiter = "\n",
        quote = '"',
    }: {
        project_id: string;
        extraction_ids: string[];
        json_schema: Record<string, any>;
        delimiter?: string;
        line_delimiter?: string;
        quote?: string;
    }, options?: RequestOptions): Promise<ExportToCsvResponse> {
        return this._fetchJson(ZExportToCsvResponse, {
            url: "/v1/extractions/get_payload_for_export",
            method: "POST",
            body: {
                project_id,
                extraction_ids,
                json_schema,
                ...(options?.body || {}),
            },
            params: {
                delimiter,
                line_delimiter,
                quote,
                ...(options?.params || {}),
            },
            headers: options?.headers,
        });
    }

    /**
     * Update an extraction.
     */
    async update(
        extraction_id: string,
        {
            predictions,
            predictions_draft,
            human_review_status,
            json_schema,
            inference_settings,
            metadata,
        }: {
            predictions?: Record<string, any>;
            predictions_draft?: Record<string, any>;
            human_review_status?: string;
            json_schema?: Record<string, any>;
            inference_settings?: Record<string, any>;
            metadata?: Record<string, string>;
        },
        options?: RequestOptions
    ): Promise<Extraction> {
        const body: Record<string, any> = {};
        if (predictions !== undefined) body.predictions = predictions;
        if (predictions_draft !== undefined) body.predictions_draft = predictions_draft;
        if (human_review_status !== undefined) body.human_review_status = human_review_status;
        if (json_schema !== undefined) body.json_schema = json_schema;
        if (inference_settings !== undefined) body.inference_settings = inference_settings;
        if (metadata !== undefined) body.metadata = metadata;

        return this._fetchJson(ZExtraction, {
            url: `/v1/extractions/${extraction_id}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get an extraction by ID.
     */
    async get(extraction_id: string, options?: RequestOptions): Promise<Extraction> {
        return this._fetchJson(ZExtraction, {
            url: `/v1/extractions/${extraction_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete an extraction by ID.
     */
    async delete(extraction_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/v1/extractions/${extraction_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get distinct values for filterable fields.
     */
    async getDistinctFieldValues(options?: RequestOptions): Promise<DistinctFieldValues> {
        return this._fetchJson(ZDistinctFieldValues, {
            url: "/v1/extractions/fields",
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Download the sample document for an extraction.
     */
    async downloadSampleDocument(extraction_id: string, options?: RequestOptions): Promise<ArrayBuffer> {
        const response = await this._fetch({
            url: `/v1/extractions/${extraction_id}/sample-document`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
        return response.arrayBuffer();
    }
}

