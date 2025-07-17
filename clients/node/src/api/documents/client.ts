import { CompositionClient } from "@/client";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResult, ZParseResult, ZParseRequest } from "@/types";


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
    async parse(params: ParseRequest): Promise<ParseResult> {
        return await this._fetchJson(ZParseResult, {
            url: "/v1/documents/parse",
            method: "POST",
            body: ZParseRequest.parse(params),
        });
    }
}
