import { CompositionClient, RequestOptions } from "../../client.js";
import { ZPaginatedList, PaginatedList } from "../../types.js";
import * as z from "zod";

const ZDownloadResponse = z.object({
    download_url: z.string(),
    filename: z.string(),
    expires_at: z.string(),
});

type DownloadResponse = z.infer<typeof ZDownloadResponse>;

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
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            origin_dot_type,
            origin_dot_id,
            from_date,
            to_date,
            metadata,
            filename,
        }: {
            before?: string;
            after?: string;
            limit?: number;
            order?: "asc" | "desc";
            origin_dot_type?: string;
            origin_dot_id?: string;
            from_date?: Date;
            to_date?: Date;
            metadata?: Record<string, string>;
            filename?: string;
        } = {},
        options?: RequestOptions,
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            origin_dot_type,
            origin_dot_id,
            from_date: from_date?.toISOString(),
            to_date: to_date?.toISOString(),
            filename,
            // Note: metadata must be JSON-serialized as the backend expects a JSON string
            metadata: metadata ? JSON.stringify(metadata) : undefined,
        };

        const cleanParams = Object.fromEntries(Object.entries(params).filter(([_, v]) => v !== undefined));

        return this._fetchJson(ZPaginatedList, {
            url: "/v1/extractions",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Download extractions in various formats.
     */
    async download(
        {
            order = "desc",
            origin_dot_id,
            from_date,
            to_date,
            metadata,
            filename,
            format = "jsonl",
        }: {
            order?: "asc" | "desc";
            origin_dot_id?: string;
            from_date?: Date;
            to_date?: Date;
            metadata?: Record<string, string>;
            filename?: string;
            format?: "jsonl" | "csv" | "xlsx";
        } = {},
        options?: RequestOptions,
    ): Promise<DownloadResponse> {
        const params: Record<string, any> = {
            order,
            origin_dot_id,
            from_date: from_date?.toISOString(),
            to_date: to_date?.toISOString(),
            filename,
            format,
            // Note: metadata must be JSON-serialized as the backend expects a JSON string
            metadata: metadata ? JSON.stringify(metadata) : undefined,
        };

        const cleanParams = Object.fromEntries(Object.entries(params).filter(([_, v]) => v !== undefined));

        return this._fetchJson(ZDownloadResponse, {
            url: "/v1/extractions/download",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
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
            json_schema,
            inference_settings,
            metadata,
        }: {
            predictions?: Record<string, any>;
            predictions_draft?: Record<string, any>;
            json_schema?: Record<string, any>;
            inference_settings?: Record<string, any>;
            metadata?: Record<string, string>;
        },
        options?: RequestOptions,
    ): Promise<Extraction> {
        const body: Record<string, any> = {};
        if (predictions !== undefined) body.predictions = predictions;
        if (predictions_draft !== undefined) body.predictions_draft = predictions_draft;
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
}
