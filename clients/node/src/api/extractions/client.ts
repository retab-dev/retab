import { CompositionClient, RequestOptions } from "../../client.js";
import { ZPaginatedList, PaginatedList } from "../../types.js";
import * as z from "zod";

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
            origin_type,
            origin_id,
            from_date,
            to_date,
            metadata,
            filename,
        }: {
            before?: string;
            after?: string;
            limit?: number;
            order?: "asc" | "desc";
            origin_type?: string;
            origin_id?: string;
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
            origin_type,
            origin_id,
            from_date: from_date?.toISOString(),
            to_date: to_date?.toISOString(),
            filename,
            // Note: metadata must be JSON-serialized as the backend expects a JSON string
            metadata: metadata ? JSON.stringify(metadata) : undefined,
        };

        const cleanParams = Object.fromEntries(Object.entries(params).filter(([_, v]) => v !== undefined));

        return this._fetchJson(ZPaginatedList, {
            url: "/extractions",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Get an extraction by ID.
     */
    async get(extraction_id: string, options?: RequestOptions): Promise<Extraction> {
        return this._fetchJson(ZExtraction, {
            url: `/extractions/${extraction_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get extraction result enriched with per-leaf source provenance.
     *
     * Each extracted leaf value is wrapped as {value, source} where source
     * contains citation content, surrounding context, and a format-specific anchor.
     *
     * @param extraction_id - ID of the extraction to source
     */
    async sources(
        extraction_id: string,
        options?: RequestOptions,
    ): Promise<Record<string, any>> {
        return this._fetchJson(z.record(z.any()), {
            url: `/extractions/${extraction_id}/sources`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

}
