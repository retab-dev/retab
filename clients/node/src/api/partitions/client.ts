import { CompositionClient, RequestOptions } from "../../client.js";
import {
    ZMIMEData,
    MIMEDataInput,
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
}
