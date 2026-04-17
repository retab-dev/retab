import { CompositionClient, RequestOptions } from "../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
    ZPaginatedList,
    PaginatedList,
    ZClassification,
    Classification,
} from "../../types.js";
import type { Category } from "../../generated_types.js";

export type ClassificationCategory = Category | { name: string; description?: string };

export type ClassificationCreateParams = {
    document: MIMEDataInput;
    categories: ClassificationCategory[];
    model: string;
    n_consensus?: number;
    bust_cache?: boolean;
    first_n_pages?: number;
    context?: string;
};

export type ClassificationListParams = {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    from_date?: Date;
    to_date?: Date;
};

/**
 * Classifications API client — resource-oriented surface for `/v1/classifications`.
 *
 * Mirrors the Python `retab.classifications` resource.
 */
export default class APIClassifications extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Create a classification. Posts to `/classifications`.
     *
     * @returns Stored `Classification` resource (with `id`, `output`, `consensus`, ...).
     */
    async create(
        params: ClassificationCreateParams,
        options?: RequestOptions,
    ): Promise<Classification> {
        const document = await ZMIMEData.parseAsync(params.document);
        const body: Record<string, unknown> = {
            document,
            categories: params.categories.map((cat) => ({
                name: cat.name,
                description: cat.description ?? "",
            })),
            model: params.model,
        };
        if (params.n_consensus !== undefined) {
            body["n_consensus"] = params.n_consensus;
        }
        if (params.first_n_pages !== undefined) {
            body["first_n_pages"] = params.first_n_pages;
        }
        if (params.context !== undefined) {
            body["context"] = params.context;
        }
        if (params.bust_cache) {
            body["bust_cache"] = true;
        }
        return this._fetchJson(ZClassification, {
            url: "/classifications",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a classification by ID.
     */
    async get(classification_id: string, options?: RequestOptions): Promise<Classification> {
        return this._fetchJson(ZClassification, {
            url: `/classifications/${classification_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List classifications with pagination and filtering.
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            from_date,
            to_date,
        }: ClassificationListParams = {},
        options?: RequestOptions,
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            from_date: from_date?.toISOString(),
            to_date: to_date?.toISOString(),
        };
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined),
        );
        return this._fetchJson(ZPaginatedList, {
            url: "/classifications",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete a classification by ID.
     */
    async delete(classification_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/classifications/${classification_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
