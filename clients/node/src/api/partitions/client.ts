import { CompositionClient, RequestOptions } from "../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
    ZPaginatedList,
    PaginatedList,
    ZPartition,
    Partition,
} from "../../types.js";

export type PartitionCreateParams = {
    document: MIMEDataInput;
    key: string;
    instructions: string;
    model: string;
    n_consensus?: number;
    bust_cache?: boolean;
};

export type PartitionListParams = {
    before?: string;
    after?: string;
    limit?: number;
    order?: "asc" | "desc";
    filename?: string;
    from_date?: Date;
    to_date?: Date;
};

export default class APIPartitions extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }

    /**
     * Create partitions. Posts to `/partitions`.
     */
    async create(
        params: PartitionCreateParams,
        options?: RequestOptions,
    ): Promise<Partition> {
        const document = await ZMIMEData.parseAsync(params.document);
        const body: Record<string, unknown> = {
            document,
            key: params.key,
            instructions: params.instructions,
            model: params.model,
            n_consensus: params.n_consensus ?? 1,
        };
        if (params.bust_cache) {
            body["bust_cache"] = true;
        }
        return this._fetchJson(ZPartition, {
            url: "/partitions",
            method: "POST",
            body: { ...body, ...(options?.body || {}) },
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * Retrieve a persisted partition resource by id.
     *
     * Typically used to fetch the partition referenced by a workflow step's
     * ``step.artifact`` (``operation === "partition"``).
     */
    async get(
        partitionId: string,
        options?: RequestOptions,
    ): Promise<Partition> {
        return this._fetchJson(ZPartition, {
            url: `/partitions/${partitionId}`,
            method: "GET",
            params: options?.params,
            headers: options?.headers,
        });
    }

    /**
     * List partitions with pagination and filtering.
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
        }: PartitionListParams = {},
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
            url: "/partitions",
            method: "GET",
            params: { ...cleanParams, ...(options?.params || {}) },
            headers: options?.headers,
        });
    }

    /**
     * Delete a partition by ID.
     */
    async delete(partitionId: string, options?: RequestOptions): Promise<void> {
        return this._fetchJson({
            url: `/partitions/${partitionId}`,
            method: "DELETE",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
