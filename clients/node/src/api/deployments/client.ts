import { CompositionClient } from "@/client";
import { mimeToBlob } from "@/mime";
import { MIMEDataInput, ZMIMEData, ZSchema } from "@/types";

export default class APIDeployments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    
    async extract({
        project_id,
        iteration_id,
        document,
        documents,
        temperature,
        seed,
        store
    }: {
        project_id: string,
        iteration_id: string,
        document?: MIMEDataInput,
        documents?: MIMEDataInput[],
        temperature?: number,
        seed?: number,
        store?: boolean,
    }) {
        if (!document && (!documents || documents.length === 0)) {
            throw new Error("Either 'document' or 'documents' must be provided.");
        }
        let url = `/v1/deployments/extract/${project_id}/${iteration_id}`;
        return await this._fetchJson(ZSchema, {
            url,
            method: "POST",
            body: {
                temperature, seed, store,
                documents: (await ZMIMEData.array().parseAsync([...document ? [document] : [], documents || []])).map(mimeToBlob)
            },
            bodyMime: "multipart/form-data",
        });
    }
}
