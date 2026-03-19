import * as z from "zod";

import { mimeToBlob } from "../../mime.js";
import { MIMEDataInput, ZMIMEData } from "../../types.js";

export function cleanObject<T extends Record<string, unknown>>(value: T): T {
    return Object.fromEntries(
        Object.entries(value).filter(([_, item]) => item !== undefined)
    ) as T;
}

export function buildListParams({
    before,
    after,
    limit,
    order,
    fields,
    extra,
}: {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    fields?: string;
    extra?: Record<string, unknown>;
} = {}): Record<string, unknown> {
    return cleanObject({
        before,
        after,
        limit,
        order,
        fields,
        ...(extra || {}),
    });
}

export async function buildProcessMultipartBody({
    document,
    documents,
    model,
    image_resolution_dpi,
    n_consensus,
    metadata,
    extraction_id,
    extra,
}: {
    document?: MIMEDataInput;
    documents?: MIMEDataInput[];
    model?: string;
    image_resolution_dpi?: number;
    n_consensus?: number;
    metadata?: Record<string, string>;
    extraction_id?: string;
    extra?: Record<string, unknown>;
}): Promise<Record<string, unknown>> {
    if (!document && !documents?.length) {
        throw new Error("Either document or documents must be provided");
    }
    if (document && documents?.length) {
        throw new Error("Provide either document or documents, not both");
    }

    const body: Record<string, unknown> = cleanObject({
        model,
        image_resolution_dpi,
        n_consensus,
        metadata: metadata ? JSON.stringify(metadata) : undefined,
        extraction_id,
        ...(extra || {}),
    });

    if (document) {
        const parsedDocument = await ZMIMEData.parseAsync(document);
        body.document = mimeToBlob(parsedDocument);
        return body;
    }

    const parsedDocuments = await Promise.all((documents || []).map((item) => ZMIMEData.parseAsync(item)));
    body.documents = parsedDocuments.map((item) => mimeToBlob(item));
    return body;
}

export function dataPaginatedArray<Schema extends z.ZodTypeAny>(schema: Schema): z.ZodEffects<
    z.ZodObject<{
        data: z.ZodArray<Schema>;
    } & {
        list_metadata: z.ZodTypeAny;
    }>,
    z.output<Schema>[],
    {
        data: z.input<Schema>[];
        list_metadata?: unknown;
    }
> {
    return z.object({
        data: z.array(schema),
        list_metadata: z.any().optional(),
    }).transform((input) => input.data);
}
