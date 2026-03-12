import { CompositionClient, RequestOptions } from "../../client.js";
import { ZMIMEData, MIMEDataInput, ZPaginatedList, PaginatedList } from "../../types.js";
import * as z from "zod";

const ZFile = z.object({
    object: z.literal("file").default("file"),
    id: z.string(),
    filename: z.string(),
    organization_id: z.string(),
    created_at: z.string(),
    updated_at: z.string(),
    page_count: z.number().nullable().optional(),
});
type File = z.infer<typeof ZFile>;

const ZFileLink = z.object({
    download_url: z.string(),
    expires_in: z.string(),
    filename: z.string(),
});
type FileLink = z.infer<typeof ZFileLink>;

const ZUploadFileResponse = z.object({
    fileId: z.string(),
    filename: z.string(),
});
type UploadFileResponse = z.infer<typeof ZUploadFileResponse>;

export default class APIFiles extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    async upload(mimeData: MIMEDataInput, options?: RequestOptions): Promise<UploadFileResponse> {
        const parsed = await ZMIMEData.parseAsync(mimeData);
        return this._fetchJson(ZUploadFileResponse, {
            url: "/files/upload",
            method: "POST",
            body: { mimeData: parsed, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    async list(
        {
            before,
            after,
            limit = 10,
            order = "desc",
            filename,
            mime_type,
            sort_by = "created_at",
        }: {
            before?: string;
            after?: string;
            limit?: number;
            order?: "asc" | "desc";
            filename?: string;
            mime_type?: string;
            sort_by?: string;
        } = {},
        options?: RequestOptions,
    ): Promise<PaginatedList> {
        const params: Record<string, any> = {
            before,
            after,
            limit,
            order,
            filename,
            mime_type,
            sort_by,
        };

        const cleanParams = Object.fromEntries(Object.entries(params).filter(([_, v]) => v !== undefined));

        return this._fetchJson(ZPaginatedList, {
            url: "/files",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    async get(fileId: string, options?: RequestOptions): Promise<File> {
        return this._fetchJson(ZFile, {
            url: `/files/${fileId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    async getDownloadLink(fileId: string, options?: RequestOptions): Promise<FileLink> {
        return this._fetchJson(ZFileLink, {
            url: `/files/${fileId}/download-link`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
