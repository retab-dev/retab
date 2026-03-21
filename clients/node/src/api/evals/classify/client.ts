import { CompositionClient, RequestOptions } from "../../../client.js";
import { MIMEDataInput, ZClassifyResponse } from "../../../types.js";
import { buildProcessMultipartBody } from "../helpers.js";

export default class APIEvalsClassify extends CompositionClient {
    async process({
        eval_id,
        iteration_id,
        document,
        model,
        metadata,
        ...rest
    }: {
        eval_id: string;
        iteration_id?: string;
        document?: MIMEDataInput;
        model?: string;
        metadata?: Record<string, string>;
    }, options?: RequestOptions) {
        if ("documents" in rest) {
            throw new Error("client.evals.classify.process(...) accepts only 'document'.");
        }
        const body = await buildProcessMultipartBody({
            document,
            model,
            metadata,
            extra: options?.body,
        });
        const url = iteration_id ? `/evals/classify/extract/${eval_id}/${iteration_id}` : `/evals/classify/extract/${eval_id}`;
        return this._fetchJson(ZClassifyResponse, {
            url,
            method: "POST",
            body,
            bodyMime: "multipart/form-data",
            params: options?.params,
            headers: options?.headers,
        });
    }
}
