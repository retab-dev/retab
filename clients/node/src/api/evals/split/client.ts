import { CompositionClient, RequestOptions } from "../../../client.js";
import { MIMEDataInput, ZRetabParsedChatCompletion } from "../../../types.js";
import { buildProcessMultipartBody } from "../helpers.js";

export default class APIEvalsSplit extends CompositionClient {
    async process({
        eval_id,
        iteration_id,
        document,
        model,
        image_resolution_dpi,
        n_consensus,
        metadata,
        extraction_id,
        ...rest
    }: {
        eval_id: string;
        iteration_id?: string;
        document?: MIMEDataInput;
        model?: string;
        image_resolution_dpi?: number;
        n_consensus?: number;
        metadata?: Record<string, string>;
        extraction_id?: string;
    }, options?: RequestOptions) {
        if ("documents" in rest) {
            throw new Error("client.evals.split.process(...) accepts only 'document'.");
        }
        const body = await buildProcessMultipartBody({
            document,
            model,
            image_resolution_dpi,
            n_consensus,
            metadata,
            extraction_id,
            extra: options?.body,
        });
        const url = iteration_id ? `/evals/split/extract/${eval_id}/${iteration_id}` : `/evals/split/extract/${eval_id}`;
        return this._fetchJson(ZRetabParsedChatCompletion, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
