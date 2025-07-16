import { CompositionClient } from "@/client";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion } from "@/types";


export default class APIDocuments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async extract(params: DocumentExtractRequest): Promise<RetabParsedChatCompletion> {
        return await this._fetchJson(ZRetabParsedChatCompletion, {
            url: "/v1/documents/extract",
            method: "POST",
            body: await ZDocumentExtractRequest.parseAsync(params),
        });
    }
}
