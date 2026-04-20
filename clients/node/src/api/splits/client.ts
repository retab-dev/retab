import { CompositionClient, RequestOptions } from "../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
    ZPaginatedList,
    PaginatedList,
    ZSplit,
    Split,
} from "../../types.js";

export type SplitSubdocument = {
    name: string;
    description?: string;
    allow_multiple_instances?: boolean;
};

export type SplitCreateParams = {
    document: MIMEDataInput;
    subdocuments: SplitSubdocument[];
    model: string;
    n_consensus?: number;
    bust_cache?: boolean;
    instructions?: string;
};

export type SplitListParams = {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    from_date?: Date;
    to_date?: Date;
};

/**
 * Splits API client — resource-oriented surface for `/v1/splits`.
 *
 * Mirrors the Python `retab.splits` resource.
 */
export default class APISplits extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Create a split. Posts to `/splits`.
     */
    async create(
        params: SplitCreateParams,
        options?: RequestOptions,
    ): Promise<Split> {
        const document = await ZMIMEData.parseAsync(params.document);
        const body: Record<string, unknown> = {
            document,
            subdocuments: params.subdocuments.map((sd) => ({
                name: sd.name,
                description: sd.description ?? "",
                allow_multiple_instances: sd.allow_multiple_instances ?? false,
            })),
            model: params.model,
        };
        if (params.n_consensus !== undefined) {
            body["n_consensus"] = params.n_consensus;
        }
        if (params.instructions !== undefined) {
            body["instructions"] = params.instructions;
        }
        if (params.bust_cache) {
            body["bust_cache"] = true;
        }
        return this._fetchJson(ZSplit, {
            url: "/splits",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a split by ID.
     */
    async get(split_id: string, options?: RequestOptions): Promise<Split> {
        return this._fetchJson(ZSplit, {
            url: `/splits/${split_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List splits with pagination and filtering.
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            from_date,
            to_date,
        }: SplitListParams = {},
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
            url: "/splits",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete a split by ID.
     */
    async delete(split_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/splits/${split_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
