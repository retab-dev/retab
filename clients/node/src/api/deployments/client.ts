import { CompositionClient } from "../../client.js";
import { mimeToBlob } from "../../mime.js";
import { MIMEDataInput, ZMIMEData, RetabParsedChatCompletion, ZRetabParsedChatCompletion } from "../../types.js";

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
    }): Promise<RetabParsedChatCompletion> {
        if (!document && (!documents || documents.length === 0)) {
            throw new Error("Either 'document' or 'documents' must be provided.");
        }
        let url = `/v1/deployments/extract/${project_id}/${iteration_id}`;

        // Only include optional parameters if they are provided
        const bodyParams: any = {
            documents: (await ZMIMEData.array().parseAsync([...document ? [document] : [], ...(documents || [])])).map(mimeToBlob)
        };

        if (temperature !== undefined) bodyParams.temperature = temperature;
        if (seed !== undefined) bodyParams.seed = seed;
        if (store !== undefined) bodyParams.store = store;

        return this._fetchJson(ZRetabParsedChatCompletion, {
            url,
            method: "POST",
            body: bodyParams,
            bodyMime: "multipart/form-data",
        });
    }
}
