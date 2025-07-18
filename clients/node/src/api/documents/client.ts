import { CompositionClient } from "@/client";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResult, ZParseResult, ZParseRequest, DocumentCreateMessageRequest, DocumentMessage, ZDocumentMessage, ZDocumentCreateMessageRequest, DocumentCreateInputRequest, ZDocumentCreateInputRequest } from "@/types";


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
            body: await ZParseRequest.parseAsync(params),
        });
    }
    async createMessages(params: DocumentCreateMessageRequest): Promise<DocumentMessage> {
        return await this._fetchJson(ZDocumentMessage, {
            url: "/v1/documents/create_messages",
            method: "POST",
            body: await ZDocumentCreateMessageRequest.parseAsync(params),
        });
    }
    async createInputs(params: DocumentCreateInputRequest): Promise<DocumentMessage> {
        return await this._fetchJson(ZDocumentMessage, {
            url: "/v1/documents/create_inputs",
            method: "POST",
            body: await ZDocumentCreateInputRequest.parseAsync(params),
        });
    }
}
