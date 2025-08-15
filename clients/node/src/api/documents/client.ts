import { CompositionClient } from "@/client";
import { ZDocumentExtractRequest, DocumentExtractRequest, RetabParsedChatCompletion, ZRetabParsedChatCompletion, ParseRequest, ParseResult, ZParseResult, ZParseRequest, DocumentCreateMessageRequest, DocumentMessage, ZDocumentMessage, ZDocumentCreateMessageRequest, DocumentCreateInputRequest, ZDocumentCreateInputRequest, RetabParsedChatCompletionChunk, ZRetabParsedChatCompletionChunk } from "@/types";


export default class APIDocuments extends CompositionClient {
    constructor(client: CompositionClient) {
        super(client);
    }
    async extract(params: DocumentExtractRequest): Promise<RetabParsedChatCompletion> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url: "/v1/documents/extract",
            method: "POST",
            body: request,
        });
    }
    async extract_stream(params: DocumentExtractRequest): Promise<AsyncGenerator<RetabParsedChatCompletionChunk>> {
        let request = await ZDocumentExtractRequest.parseAsync(params);
        return this._fetchStream(ZRetabParsedChatCompletionChunk, {
            url: "/v1/documents/extract",
            method: "POST",
            body: { ...request, stream: true },
        });
    }
    async parse(params: ParseRequest): Promise<ParseResult> {
        return this._fetchJson(ZParseResult, {
            url: "/v1/documents/parse",
            method: "POST",
            body: await ZParseRequest.parseAsync(params),
        });
    }
    async create_messages(params: DocumentCreateMessageRequest): Promise<DocumentMessage> {
        return this._fetchJson(ZDocumentMessage, {
            url: "/v1/documents/create_messages",
            method: "POST",
            body: await ZDocumentCreateMessageRequest.parseAsync(params),
        });
    }
    async create_inputs(params: DocumentCreateInputRequest): Promise<DocumentMessage> {
        return this._fetchJson(ZDocumentMessage, {
            url: "/v1/documents/create_inputs",
            method: "POST",
            body: await ZDocumentCreateInputRequest.parseAsync(params),
        });
    }
}
