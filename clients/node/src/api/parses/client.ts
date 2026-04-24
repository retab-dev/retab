import { CompositionClient, RequestOptions } from "../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
    ZPaginatedList,
    PaginatedList,
    ZParse,
    Parse,
} from "../../types.js";

type TableParsingFormat = "markdown" | "yaml" | "html" | "json";

export type ParseCreateParams = {
    document: MIMEDataInput;
    model: string;
    table_parsing_format?: TableParsingFormat;
    image_resolution_dpi?: number;
    instructions?: string;
    bust_cache?: boolean;
};

export type ParseListParams = {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    filename?: string;
    from_date?: Date;
    to_date?: Date;
};

export default class APIParses extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Create a parse record. Mirrors Python `client.parses.create(...)`.
     *
     * Persists the parse result and returns the stored `Parse` resource
     * (with `id`, nested `output: { pages, text }`, timestamps, etc.).
     */
    async create(params: ParseCreateParams, options?: RequestOptions): Promise<Parse> {
        const document = await ZMIMEData.parseAsync(params.document);
        const body: Record<string, unknown> = {
            document,
            model: params.model,
        };
        if (params.table_parsing_format !== undefined) {
            body["table_parsing_format"] = params.table_parsing_format;
        }
        if (params.image_resolution_dpi !== undefined) {
            body["image_resolution_dpi"] = params.image_resolution_dpi;
        }
        if (params.instructions !== undefined) {
            body["instructions"] = params.instructions;
        }
        if (params.bust_cache) {
            body["bust_cache"] = true;
        }
        return this._fetchJson(ZParse, {
            url: "/parses",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a parse record by ID.
     */
    async get(parse_id: string, options?: RequestOptions): Promise<Parse> {
        return this._fetchJson(ZParse, {
            url: `/parses/${parse_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List parse records with pagination and filtering.
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            filename,
            from_date,
            to_date,
        }: ParseListParams = {},
        options?: RequestOptions,
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            filename,
            from_date: from_date?.toISOString(),
            to_date: to_date?.toISOString(),
        };

        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined),
        );

        return this._fetchJson(ZPaginatedList, {
            url: "/parses",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete a parse record by ID.
     */
    async delete(parse_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/parses/${parse_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
