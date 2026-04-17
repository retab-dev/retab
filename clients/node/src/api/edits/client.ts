import { CompositionClient, RequestOptions } from "../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
    ZPaginatedList,
    PaginatedList,
    ZEdit,
    Edit,
} from "../../types.js";
import APIEditsTemplates from "./templates/client.js";

export type EditCreateParams = {
    instructions: string;
    document?: MIMEDataInput;
    template_id?: string;
    model?: string;
    color?: string;
    bust_cache?: boolean;
};

export type EditListParams = {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    filename?: string;
    template_id?: string;
    from_date?: Date;
    to_date?: Date;
};

/**
 * Edits API client — resource-oriented surface for `/v1/edits`.
 *
 * Mirrors the Python `retab.edits` resource. Also exposes an `edits.templates`
 * sub-client for managing reusable PDF form templates.
 */
export default class APIEdits extends CompositionClient {
    public templates: APIEditsTemplates;

    constructor(client: CompositionClient) {
        super(client);
        this.templates = new APIEditsTemplates(this);
    }

    /**
     * Create an edit. Posts to `/edits`.
     *
     * Either `document` OR `template_id` must be provided (not both).
     */
    async create(
        params: EditCreateParams,
        options?: RequestOptions,
    ): Promise<Edit> {
        const body: Record<string, unknown> = {
            instructions: params.instructions,
        };
        if (params.document !== undefined) {
            body["document"] = await ZMIMEData.parseAsync(params.document);
        }
        if (params.template_id !== undefined) {
            body["template_id"] = params.template_id;
        }
        if (params.model !== undefined) {
            body["model"] = params.model;
        }
        if (params.color !== undefined) {
            body["config"] = { color: params.color };
        }
        if (params.bust_cache) {
            body["bust_cache"] = true;
        }
        return this._fetchJson(ZEdit, {
            url: "/edits",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get an edit by ID.
     */
    async get(edit_id: string, options?: RequestOptions): Promise<Edit> {
        return this._fetchJson(ZEdit, {
            url: `/edits/${edit_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List edits with pagination and filtering.
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            filename,
            template_id,
            from_date,
            to_date,
        }: EditListParams = {},
        options?: RequestOptions,
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            filename,
            template_id,
            from_date: from_date?.toISOString(),
            to_date: to_date?.toISOString(),
        };
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined),
        );
        return this._fetchJson(ZPaginatedList, {
            url: "/edits",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete an edit by ID.
     */
    async delete(edit_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/edits/${edit_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
