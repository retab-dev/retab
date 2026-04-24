import { CompositionClient, RequestOptions } from "../../client.js";
import { MIMEDataInput, ZPaginatedList, PaginatedList } from "../../types.js";
import * as z from "zod";
import crypto from "crypto";
import fs from "fs";
import path from "path";

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
    storageUrl: z.string().optional(),
}).transform((value) => ({
    ...value,
    file_id: value.fileId,
    storage_url: value.storageUrl,
}));
type UploadFileResponse = z.infer<typeof ZUploadFileResponse>;

const ZCreateUploadResponse = z.object({
    fileId: z.string(),
    filename: z.string(),
    uploadUrl: z.string(),
    uploadMethod: z.string().default("PUT"),
    uploadHeaders: z.record(z.string()).default({}),
    storageUrl: z.string(),
    expiresAt: z.string(),
}).transform((value) => ({
    ...value,
    file_id: value.fileId,
    upload_url: value.uploadUrl,
    upload_method: value.uploadMethod,
    upload_headers: value.uploadHeaders,
    storage_url: value.storageUrl,
    expires_at: value.expiresAt,
}));
type CreateUploadResponse = z.infer<typeof ZCreateUploadResponse>;

async function isLocalFilePath(input: MIMEDataInput): Promise<boolean> {
    if (typeof input !== "string" || input.startsWith("https://") || input.startsWith("data:")) {
        return false;
    }
    return fs.promises.stat(input).then((stat) => stat.isFile()).catch(() => false);
}

function contentTypeForFilename(filename: string): string {
    if (filename.endsWith(".pdf")) return "application/pdf";
    if (filename.endsWith(".png")) return "image/png";
    if (filename.endsWith(".jpg") || filename.endsWith(".jpeg")) return "image/jpeg";
    if (filename.endsWith(".txt")) return "text/plain";
    return "application/octet-stream";
}

async function fileBodyForPath(filePath: string): Promise<Blob | fs.ReadStream> {
    const bunFile = (globalThis as typeof globalThis & { Bun?: { file?: (filePath: string) => Blob } }).Bun?.file?.(filePath);
    if (bunFile) {
        return bunFile;
    }
    return fs.createReadStream(filePath);
}

async function fileSha256(filePath: string): Promise<string> {
    return new Promise((resolve, reject) => {
        const hash = crypto.createHash("sha256");
        const stream = fs.createReadStream(filePath);
        stream.on("data", (chunk) => hash.update(chunk));
        stream.on("error", reject);
        stream.on("end", () => resolve(hash.digest("hex")));
    });
}

export default class APIFiles extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    prepare_upload(filename: string, contentType: string, sizeBytes: number, sha256?: string): {
        url: string;
        method: string;
        body: Record<string, string | number>;
    } {
        const body: Record<string, string | number> = {
            filename,
            content_type: contentType,
            size_bytes: sizeBytes,
        };
        if (sha256 !== undefined) {
            body.sha256 = sha256;
        }
        return {
            url: "/files/upload",
            method: "POST",
            body,
        };
    }

    prepare_complete_upload(fileId: string, sha256?: string): {
        url: string;
        method: string;
        body: Record<string, string>;
    } {
        const body: Record<string, string> = {};
        if (sha256 !== undefined) {
            body.sha256 = sha256;
        }
        return {
            url: `/files/upload/${fileId}/complete`,
            method: "POST",
            body,
        };
    }

    async upload(mimeData: MIMEDataInput, options?: RequestOptions): Promise<UploadFileResponse> {
        if (await isLocalFilePath(mimeData)) {
            const filePath = mimeData as string;
            const stat = await fs.promises.stat(filePath);
            const filename = path.basename(filePath);
            const contentType = contentTypeForFilename(filename);
            const sha256 = await fileSha256(filePath);
            const uploadRequest = this.prepare_upload(filename, contentType, stat.size, sha256);
            const session = await this._fetchJson(ZCreateUploadResponse, {
                ...uploadRequest,
                params: options?.params,
                headers: options?.headers,
            }) as CreateUploadResponse;
            const uploadResponse = await fetch(session.upload_url, {
                method: session.upload_method,
                headers: session.upload_headers,
                body: await fileBodyForPath(filePath) as RequestInit["body"],
                // Required by Node fetch for streamed request bodies. Ignored by Bun.
                duplex: "half",
            } as RequestInit & { duplex: "half" });
            if (!uploadResponse.ok) {
                throw new Error(`Direct file upload failed: ${uploadResponse.status} ${await uploadResponse.text()}`);
            }
            const completeRequest = this.prepare_complete_upload(session.file_id, sha256);
            return this._fetchJson(ZUploadFileResponse, {
                ...completeRequest,
                params: options?.params,
                headers: options?.headers,
            });
        }
        throw new Error("files.upload only accepts local file paths");
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

    async get_download_link(fileId: string, options?: RequestOptions): Promise<FileLink> {
        return this.getDownloadLink(fileId, options);
    }
}
