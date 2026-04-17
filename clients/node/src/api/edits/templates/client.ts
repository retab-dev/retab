import { CompositionClient, RequestOptions } from "../../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
    ZPaginatedList,
    PaginatedList,
    ZEdit,
    Edit,
} from "../../../types.js";
import {
    ZEditTemplate,
    EditTemplate,
    ZFormField,
    FormField,
} from "../../../generated_types.js";

export type EditTemplateCreateParams = {
    name: string;
    document: MIMEDataInput;
    form_fields: FormField[];
};

export type EditTemplateListParams = {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    name?: string;
};

export type EditTemplateUpdateParams = {
    name?: string | null;
    form_fields?: FormField[] | null;
};

export type EditTemplateFillParams = {
    template_id: string;
    instructions: string;
    model?: string;
    color?: string;
    bust_cache?: boolean;
};

/**
 * Edit Templates API client — resource-oriented surface for
 * `/v1/edits/templates` and `/v1/edits/templates/fill`.
 *
 * Mirrors the Python `retab.edits.templates` resource.
 */
export default class APIEditsTemplates extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Create a template from a PDF document and a set of form fields.
     */
    async create(
        params: EditTemplateCreateParams,
        options?: RequestOptions,
    ): Promise<EditTemplate> {
        const document = await ZMIMEData.parseAsync(params.document);
        const form_fields = await Promise.all(
            params.form_fields.map((f) => ZFormField.parseAsync(f)),
        );
        return this._fetchJson(ZEditTemplate, {
            url: "/edits/templates",
            method: "POST",
            body: {
                name: params.name,
                document,
                form_fields,
                ...(options?.body || {}),
            },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Get a template by ID.
     */
    async get(template_id: string, options?: RequestOptions): Promise<EditTemplate> {
        return this._fetchJson(ZEditTemplate, {
            url: `/edits/templates/${template_id}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List templates with pagination and filtering.
     */
    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            name,
        }: EditTemplateListParams = {},
        options?: RequestOptions,
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            name,
        };
        const cleanParams = Object.fromEntries(
            Object.entries(params).filter(([_, v]) => v !== undefined),
        );
        return this._fetchJson(ZPaginatedList, {
            url: "/edits/templates",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Update a template's name and/or form fields. PATCH `/edits/templates/{id}`.
     */
    async update(
        template_id: string,
        params: EditTemplateUpdateParams,
        options?: RequestOptions,
    ): Promise<EditTemplate> {
        const body: Record<string, unknown> = {};
        if (params.name !== undefined) {
            body["name"] = params.name;
        }
        if (params.form_fields !== undefined) {
            body["form_fields"] =
                params.form_fields === null
                    ? null
                    : await Promise.all(params.form_fields.map((f) => ZFormField.parseAsync(f)));
        }
        return this._fetchJson(ZEditTemplate, {
            url: `/edits/templates/${template_id}`,
            method: "PATCH",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Delete a template by ID.
     */
    async delete(template_id: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/edits/templates/${template_id}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Fill a template, producing an Edit resource.
     *
     * Posts to `/edits/templates/fill`.
     */
    async fill(
        params: EditTemplateFillParams,
        options?: RequestOptions,
    ): Promise<Edit> {
        const body: Record<string, unknown> = {
            template_id: params.template_id,
            instructions: params.instructions,
        };
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
            url: "/edits/templates/fill",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }
}
