import { fileURLToPath } from "node:url";

import { describe, expect, test } from "bun:test";

import APIV1 from "../src/api/client.js";
import { AbstractClient } from "../src/client.js";

type RecordedRequest = {
    url: string;
    method: string;
    params?: Record<string, unknown>;
    headers?: Record<string, unknown>;
    bodyMime?: "application/json" | "multipart/form-data";
    body?: Record<string, unknown>;
};

class MockClient extends AbstractClient {
    public requests: RecordedRequest[] = [];

    protected async _fetch(params: RecordedRequest): Promise<Response> {
        this.requests.push(params);
        return new Response(JSON.stringify(buildResponse(params)), {
            status: 200,
            headers: {
                "Content-Type": "application/json",
            },
        });
    }
}

function findRequest(requests: RecordedRequest[], method: string, url: string): RecordedRequest {
    const request = requests.find((candidate) => candidate.method === method && candidate.url === url);
    if (!request) {
        throw new Error(`Missing request: ${method} ${url}`);
    }
    return request;
}

function buildResponse(params: RecordedRequest): unknown {
    if (params.url === "/evals/extract" && params.method === "POST") {
        return buildExtractEval(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract" && params.method === "GET") {
        return buildExtractEval({ name: "Extract Eval" });
    }
    if (params.url === "/evals/extract" && params.method === "GET") {
        return {
            data: [buildExtractEval({ name: "Extract Eval" })],
            list_metadata: {},
        };
    }
    if (params.url === "/evals/extract/eval-extract" && params.method === "PATCH") {
        return buildExtractEval({ name: (params.body as Record<string, unknown>).name || "Extract Eval v2" });
    }
    if (params.url === "/evals/extract/eval-extract/publish" && params.method === "POST") {
        return buildExtractEval({ name: "Extract Eval", is_published: true });
    }
    if (params.url === "/evals/extract/eval-extract/datasets" && params.method === "POST") {
        return buildExtractDataset(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract" && params.method === "GET") {
        return buildExtractDataset({ name: "Extract Dataset" });
    }
    if (params.url === "/evals/extract/eval-extract/datasets" && params.method === "GET") {
        return {
            data: [buildExtractDataset({ name: "Extract Dataset" })],
            list_metadata: {},
        };
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract" && params.method === "PATCH") {
        return buildExtractDataset({ name: (params.body as Record<string, unknown>).name || "Extract Dataset v2" });
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/duplicate" && params.method === "POST") {
        return buildExtractDataset({ name: (params.body as Record<string, unknown>).name || "Extract Dataset Copy" });
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents" && params.method === "POST") {
        return buildExtractDatasetDocument(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents/doc-extract" && params.method === "GET") {
        return buildExtractDatasetDocument({});
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents" && params.method === "GET") {
        return [buildExtractDatasetDocument({})];
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents/doc-extract" && params.method === "PATCH") {
        return buildExtractDatasetDocument(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents/doc-extract/process" && params.method === "POST") {
        return buildParsedCompletion();
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations" && params.method === "POST") {
        return buildExtractIteration(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract" && params.method === "GET") {
        return buildExtractIteration({});
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations" && params.method === "GET") {
        return {
            data: [buildExtractIteration({})],
            list_metadata: {},
        };
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract" && params.method === "PATCH") {
        return buildExtractIteration(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/schema" && params.method === "GET") {
        return {
            type: "object",
            properties: {
                status: {
                    type: "string",
                    description: "Review status",
                },
            },
        };
    }
    if (
        params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/iteration-documents/processDocumentsFromDatasetId" &&
        params.method === "POST"
    ) {
        return { success: true, id: "iter-doc-extract" };
    }
    if (
        params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/iteration-documents/iter-doc-extract" &&
        params.method === "GET"
    ) {
        return buildExtractIterationDocument({});
    }
    if (
        params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/iteration-documents" &&
        params.method === "GET"
    ) {
        return [buildExtractIterationDocument({})];
    }
    if (
        params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/iteration-documents/iter-doc-extract" &&
        params.method === "PATCH"
    ) {
        return buildExtractIterationDocument(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/metrics" && params.method === "GET") {
        return buildMetrics();
    }
    if (
        params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/iteration-documents/iter-doc-extract/process" &&
        params.method === "POST"
    ) {
        return buildParsedCompletion();
    }
    if (params.url === "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/finalize" && params.method === "POST") {
        return buildExtractIteration({ status: "completed" });
    }
    if (params.url === "/evals/split" && params.method === "POST") {
        return buildSplitEval(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/classify" && params.method === "POST") {
        return buildClassifyEval(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/split/eval-split/datasets" && params.method === "POST") {
        return buildSplitDataset(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/classify/eval-classify/datasets" && params.method === "POST") {
        return buildClassifyDataset(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/split/eval-split/datasets/dataset-split/iterations" && params.method === "POST") {
        return buildSplitIteration(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/classify/eval-classify/datasets/dataset-classify/iterations" && params.method === "POST") {
        return buildClassifyIteration(params.body as Record<string, unknown>);
    }
    if (params.url === "/evals/classify/eval-classify/datasets/dataset-classify/iterations/iter-classify/categories" && params.method === "GET") {
        return [
            { name: "invoice", description: "Invoice document" },
            { name: "receipt", description: "Receipt document" },
        ];
    }
    if (params.url === "/evals/extract/templates" && params.method === "GET") {
        return {
            data: [buildExtractEval({ name: "Template" })],
            list_metadata: {},
        };
    }
    if (params.url === "/evals/extract/extract/eval-extract" && params.method === "POST") {
        return buildParsedCompletion();
    }
    if (params.url === "/evals/split/extract/eval-split" && params.method === "POST") {
        return buildParsedCompletion();
    }
    if (params.url === "/evals/classify/extract/eval-classify" && params.method === "POST") {
        return {
            result: {
                reasoning: "matched category",
                classification: "invoice",
            },
            usage: {
                credits: 1,
            },
        };
    }

    throw new Error(`Unhandled mock request: ${params.method} ${params.url}`);
}

function buildInferenceSettings(): Record<string, unknown> {
    return {
        model: "retab-small",
        reasoning_effort: "minimal",
        image_resolution_dpi: 192,
        n_consensus: 1,
    };
}

function buildBaseMimeData(filename = "document.pdf"): Record<string, unknown> {
    return {
        id: "file-1",
        filename,
        mime_type: "application/pdf",
    };
}

function buildParsedCompletion(): Record<string, unknown> {
    return {
        id: "chatcmpl-1",
        object: "chat.completion",
        created: 1710000000,
        model: "retab-small",
        choices: [
            {
                index: 0,
                finish_reason: "stop",
                message: {
                    role: "assistant",
                    content: "{\"status\":\"ok\"}",
                    parsed: { status: "ok" },
                },
            },
        ],
        extraction_id: "ext-1",
    };
}

function buildExtractEval(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "eval-extract",
        name: typeof body.name === "string" ? body.name : "Extract Eval",
        updated_at: "2026-03-18T00:00:00Z",
        published_config: {
            inference_settings: buildInferenceSettings(),
            json_schema: (body.json_schema as Record<string, unknown>) || {},
            origin: "manual",
        },
        draft_config: {
            inference_settings: buildInferenceSettings(),
            json_schema: (body.json_schema as Record<string, unknown>) || {},
        },
        is_published: body.is_published === true,
        is_schema_generated: true,
    };
}

function buildExtractDataset(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "dataset-extract",
        name: typeof body.name === "string" ? body.name : "Extract Dataset",
        updated_at: "2026-03-18T00:00:00Z",
        base_json_schema: (body.base_json_schema as Record<string, unknown>) || {
            type: "object",
            properties: {
                status: { type: "string" },
            },
        },
        project_id: "eval-extract",
    };
}

function buildExtractDatasetDocument(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "doc-extract",
        updated_at: "2026-03-18T00:00:00Z",
        project_id: "eval-extract",
        dataset_id: "dataset-extract",
        mime_data: buildBaseMimeData("dataset-document.pdf"),
        prediction_data: body.prediction_data || { prediction: { status: "ok" } },
        extraction_id: typeof body.extraction_id === "string" ? body.extraction_id : "ext-1",
        validation_flags: body.validation_flags || { status: "reviewed" },
    };
}

function buildExtractIteration(body: Record<string, unknown>): Record<string, unknown> {
    const draft = (body.draft as Record<string, unknown> | undefined) || {
        schema_overrides: {
            descriptionsOverride: {
                status: "Review status",
            },
        },
        inference_settings: buildInferenceSettings(),
    };

    return {
        id: "iter-extract",
        updated_at: "2026-03-18T00:00:00Z",
        inference_settings: buildInferenceSettings(),
        schema_overrides: (body.schema_overrides as Record<string, unknown>) || {},
        parent_id: body.parent_id || null,
        project_id: "eval-extract",
        dataset_id: "dataset-extract",
        draft,
        status: typeof body.status === "string" ? body.status : "draft",
        finalized_at: typeof body.status === "string" && body.status === "completed" ? "2026-03-18T00:00:00Z" : null,
        last_finalize_error: null,
    };
}

function buildExtractIterationDocument(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "iter-doc-extract",
        updated_at: "2026-03-18T00:00:00Z",
        project_id: "eval-extract",
        iteration_id: "iter-extract",
        dataset_id: "dataset-extract",
        dataset_document_id: "doc-extract",
        mime_data: buildBaseMimeData("iteration-document.pdf"),
        prediction_data: body.prediction_data || { prediction: { status: "iteration" } },
        extraction_id: typeof body.extraction_id === "string" ? body.extraction_id : "ext-2",
    };
}

function buildMetrics(): Record<string, unknown> {
    return {
        overall_metrics: {
            accuracy: 0.9,
            similarity: 0.95,
            total_error_rate: 0.1,
            true_positive_rate: 0.85,
            true_negative_rate: 1.0,
            false_positive_rate: 0.0,
            false_negative_rate: 0.15,
            mismatched_value_rate: 0.05,
            accuracy_per_field: { status: 0.9 },
            similarity_per_field: { status: 0.95 },
            total_documents: 1,
            total_fields_compared: 1,
        },
        document_metrics: [
            {
                document_id: "iter-doc-extract",
                filename: "iteration-document.pdf",
                true_positives: [],
                true_negatives: [],
                false_positives: [],
                false_negatives: [],
                mismatched_values: [],
                field_similarities: { status: 0.95 },
                key_mappings: { status: "status" },
            },
        ],
    };
}

function buildSplitEval(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "eval-split",
        name: typeof body.name === "string" ? body.name : "Split Eval",
        updated_at: "2026-03-18T00:00:00Z",
        published_config: {
            inference_settings: buildInferenceSettings(),
            split_config: body.split_config || [],
            json_schema: {},
            subdocuments: body.split_config || [],
            origin: "manual",
        },
        draft_config: {
            inference_settings: buildInferenceSettings(),
            split_config: body.split_config || [],
            json_schema: {},
            subdocuments: body.split_config || [],
        },
        is_published: false,
    };
}

function buildClassifyEval(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "eval-classify",
        name: typeof body.name === "string" ? body.name : "Classify Eval",
        updated_at: "2026-03-18T00:00:00Z",
        published_config: {
            inference_settings: buildInferenceSettings(),
            categories: body.categories || [],
            origin: "manual",
        },
        draft_config: {
            inference_settings: buildInferenceSettings(),
            categories: body.categories || [],
        },
        is_published: false,
    };
}

function buildSplitDataset(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "dataset-split",
        name: typeof body.name === "string" ? body.name : "Split Dataset",
        updated_at: "2026-03-18T00:00:00Z",
        base_split_config: body.base_split_config || [],
        base_json_schema: {},
        base_subdocuments: body.base_split_config || [],
        base_inference_settings: buildInferenceSettings(),
        project_id: "eval-split",
    };
}

function buildClassifyDataset(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "dataset-classify",
        name: typeof body.name === "string" ? body.name : "Classify Dataset",
        updated_at: "2026-03-18T00:00:00Z",
        base_categories: body.base_categories || [],
        base_inference_settings: buildInferenceSettings(),
        project_id: "eval-classify",
    };
}

function buildSplitIteration(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "iter-split",
        updated_at: "2026-03-18T00:00:00Z",
        inference_settings: body.inference_settings || buildInferenceSettings(),
        split_config_overrides: body.split_config_overrides || {},
        parent_id: body.parent_id || null,
        project_id: "eval-split",
        dataset_id: "dataset-split",
        draft: {
            split_config_overrides: body.split_config_overrides || {},
            inference_settings: body.inference_settings || buildInferenceSettings(),
        },
        status: "draft",
        finalized_at: null,
        last_finalize_error: null,
    };
}

function buildClassifyIteration(body: Record<string, unknown>): Record<string, unknown> {
    return {
        id: "iter-classify",
        updated_at: "2026-03-18T00:00:00Z",
        inference_settings: body.inference_settings || buildInferenceSettings(),
        category_overrides: body.category_overrides || {},
        parent_id: body.parent_id || null,
        project_id: "eval-classify",
        dataset_id: "dataset-classify",
        draft: {
            category_overrides: body.category_overrides || {},
            inference_settings: body.inference_settings || buildInferenceSettings(),
        },
        status: "draft",
        finalized_at: null,
        last_finalize_error: null,
    };
}

const payslipPath = fileURLToPath(new URL("./data/payslip/payslip.pdf", import.meta.url));

describe("Node SDK evals", () => {
    test("exposes eval category services and builds process requests", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        expect(api.evals.extract.datasets.iterations).toBeDefined();
        expect(api.evals.extract.templates).toBeDefined();
        expect(api.evals.split.datasets.iterations).toBeDefined();
        expect(api.evals.split.templates).toBeDefined();
        expect(api.evals.classify.datasets.iterations).toBeDefined();
        expect(api.evals.classify.templates).toBeDefined();

        const extractResponse = await api.evals.extract.process({
            eval_id: "eval-extract",
            document: payslipPath,
            metadata: { source: "unit-test" },
        });
        expect(extractResponse.choices[0].message.parsed).toEqual({ status: "ok" });

        const splitResponse = await api.evals.split.process({
            eval_id: "eval-split",
            document: payslipPath,
            model: "retab-small",
        });
        expect(splitResponse.choices[0].message.parsed).toEqual({ status: "ok" });

        const classifyResponse = await api.evals.classify.process({
            eval_id: "eval-classify",
            document: payslipPath,
        });
        expect(classifyResponse.result.classification).toBe("invoice");

        const extractRequest = mockClient.requests[0];
        expect(extractRequest.url).toBe("/evals/extract/extract/eval-extract");
        expect(extractRequest.bodyMime).toBe("multipart/form-data");
        expect(extractRequest.body?.document).toBeInstanceOf(Blob);
        expect(extractRequest.body?.metadata).toBe(JSON.stringify({ source: "unit-test" }));

        const splitRequest = mockClient.requests[1];
        expect(splitRequest.url).toBe("/evals/split/extract/eval-split");
        expect(splitRequest.bodyMime).toBe("multipart/form-data");
        expect(splitRequest.body?.document).toBeInstanceOf(Blob);

        const classifyRequest = mockClient.requests[2];
        expect(classifyRequest.url).toBe("/evals/classify/extract/eval-classify");
        expect(classifyRequest.bodyMime).toBe("multipart/form-data");
        expect(classifyRequest.body?.document).toBeInstanceOf(Blob);
    });

    test("creates split and classify nested resources with category-specific payloads", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const splitEval = await api.evals.split.create({
            name: "Split Eval",
            split_config: [{ name: "Invoice", description: "Invoice pages" }],
        });
        expect(splitEval.draft_config.split_config[0].name).toBe("Invoice");

        const splitDataset = await api.evals.split.datasets.create("eval-split", {
            name: "Split Dataset",
            base_split_config: [{ name: "Invoice", description: "Invoice pages" }],
        });
        expect(splitDataset.base_split_config[0].name).toBe("Invoice");

        const splitIteration = await api.evals.split.datasets.iterations.create("eval-split", "dataset-split", {
            split_config_overrides: {
                descriptions_override: { Invoice: "Updated description" },
            },
        });
        expect(splitIteration.project_id).toBe("eval-split");

        const classifyEval = await api.evals.classify.create({
            name: "Classify Eval",
            categories: [{ name: "invoice", description: "Invoice document" }],
        });
        expect(classifyEval.draft_config.categories[0].name).toBe("invoice");

        const classifyDataset = await api.evals.classify.datasets.create("eval-classify", {
            name: "Classify Dataset",
            base_categories: [{ name: "invoice", description: "Invoice document" }],
        });
        expect(classifyDataset.base_categories[0].name).toBe("invoice");

        const classifyIteration = await api.evals.classify.datasets.iterations.create("eval-classify", "dataset-classify", {
            category_overrides: {
                description_overrides: { invoice: "Updated category description" },
            },
        });
        expect(classifyIteration.dataset_id).toBe("dataset-classify");

        const categories = await api.evals.classify.datasets.iterations.getCategories(
            "eval-classify",
            "dataset-classify",
            "iter-classify"
        );
        expect(categories.map((category) => category.name)).toEqual(["invoice", "receipt"]);

        const splitEvalRequest = mockClient.requests[0];
        expect(splitEvalRequest.url).toBe("/evals/split");
        expect(splitEvalRequest.body?.split_config).toEqual([{ name: "Invoice", description: "Invoice pages" }]);

        const splitDatasetRequest = mockClient.requests[1];
        expect(splitDatasetRequest.url).toBe("/evals/split/eval-split/datasets");
        expect(splitDatasetRequest.body?.base_split_config).toEqual([{ name: "Invoice", description: "Invoice pages" }]);

        const splitIterationRequest = mockClient.requests[2];
        expect(splitIterationRequest.url).toBe("/evals/split/eval-split/datasets/dataset-split/iterations");
        expect(splitIterationRequest.body?.project_id).toBe("eval-split");
        expect(splitIterationRequest.body?.dataset_id).toBe("dataset-split");

        const classifyEvalRequest = mockClient.requests[3];
        expect(classifyEvalRequest.url).toBe("/evals/classify");
        expect(classifyEvalRequest.body?.categories).toEqual([{ name: "invoice", description: "Invoice document" }]);

        const classifyDatasetRequest = mockClient.requests[4];
        expect(classifyDatasetRequest.url).toBe("/evals/classify/eval-classify/datasets");
        expect(classifyDatasetRequest.body?.base_categories).toEqual([{ name: "invoice", description: "Invoice document" }]);

        const classifyIterationRequest = mockClient.requests[5];
        expect(classifyIterationRequest.url).toBe("/evals/classify/eval-classify/datasets/dataset-classify/iterations");
        expect(classifyIterationRequest.body?.project_id).toBe("eval-classify");
        expect(classifyIterationRequest.body?.dataset_id).toBe("dataset-classify");
    });

    test("lists extract templates from the evals namespace", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const templates = await api.evals.extract.templates.list();
        expect(templates).toHaveLength(1);
        expect(templates[0].id).toBe("eval-extract");
        expect(mockClient.requests[0].url).toBe("/evals/extract/templates");
    });

    test("rejects multi-document extract requests", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        await expect(
            api.evals.extract.process({
                eval_id: "eval-extract",
                documents: [payslipPath],
            } as never)
        ).rejects.toThrow("accepts only 'document'");
    });

    test("runs the extract lifecycle through nested eval resources", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const createdEval = await api.evals.extract.create({
            name: "Extract Eval",
            json_schema: {
                type: "object",
                properties: {
                    status: { type: "string" },
                },
            },
        });
        const fetchedEval = await api.evals.extract.get("eval-extract");
        const listedEvals = await api.evals.extract.list();
        const updatedEval = await api.evals.extract.update("eval-extract", {
            name: "Extract Eval v2",
        });
        const publishedEval = await api.evals.extract.publish("eval-extract", "iter-extract");
        const processedEval = await api.evals.extract.process({
            eval_id: "eval-extract",
            document: payslipPath,
            model: "retab-small",
            metadata: { source: "lifecycle" },
        });

        const dataset = await api.evals.extract.datasets.create("eval-extract", {
            name: "Extract Dataset",
            base_json_schema: {
                type: "object",
                properties: {
                    status: { type: "string" },
                },
            },
        });
        const fetchedDataset = await api.evals.extract.datasets.get("eval-extract", "dataset-extract");
        const listedDatasets = await api.evals.extract.datasets.list("eval-extract");
        const updatedDataset = await api.evals.extract.datasets.update("eval-extract", "dataset-extract", {
            name: "Extract Dataset v2",
        });
        const duplicatedDataset = await api.evals.extract.datasets.duplicate("eval-extract", "dataset-extract", {
            name: "Extract Dataset Copy",
        });
        const datasetDocument = await api.evals.extract.datasets.addDocument("eval-extract", "dataset-extract", {
            mime_data: payslipPath,
            prediction_data: { prediction: { status: "ok" } },
        });
        const fetchedDatasetDocument = await api.evals.extract.datasets.getDocument("eval-extract", "dataset-extract", "doc-extract");
        const listedDatasetDocuments = await api.evals.extract.datasets.listDocuments("eval-extract", "dataset-extract");
        const updatedDatasetDocument = await api.evals.extract.datasets.updateDocument("eval-extract", "dataset-extract", "doc-extract", {
            validation_flags: { status: "reviewed" },
            extraction_id: "ext-updated",
        });
        const processedDatasetDocument = await api.evals.extract.datasets.processDocument("eval-extract", "dataset-extract", "doc-extract");

        const iteration = await api.evals.extract.datasets.iterations.create("eval-extract", "dataset-extract");
        const fetchedIteration = await api.evals.extract.datasets.iterations.get("eval-extract", "dataset-extract", "iter-extract");
        const listedIterations = await api.evals.extract.datasets.iterations.list("eval-extract", "dataset-extract");
        const updatedIteration = await api.evals.extract.datasets.iterations.updateDraft("eval-extract", "dataset-extract", "iter-extract", {
            draft: {
                schema_overrides: {
                    descriptionsOverride: {
                        status: "Review status",
                    },
                },
            },
        });
        const schema = await api.evals.extract.datasets.iterations.getSchema(
            "eval-extract",
            "dataset-extract",
            "iter-extract",
            { useDraft: true }
        );
        const processDocumentsResult = await api.evals.extract.datasets.iterations.processDocuments(
            "eval-extract",
            "dataset-extract",
            "iter-extract",
            "doc-extract"
        );
        const iterationDocument = await api.evals.extract.datasets.iterations.getDocument(
            "eval-extract",
            "dataset-extract",
            "iter-extract",
            "iter-doc-extract"
        );
        const listedIterationDocuments = await api.evals.extract.datasets.iterations.listDocuments(
            "eval-extract",
            "dataset-extract",
            "iter-extract"
        );
        const updatedIterationDocument = await api.evals.extract.datasets.iterations.updateDocument(
            "eval-extract",
            "dataset-extract",
            "iter-extract",
            "iter-doc-extract",
            {
                extraction_id: "ext-iter-updated",
            }
        );
        const metrics = await api.evals.extract.datasets.iterations.getMetrics(
            "eval-extract",
            "dataset-extract",
            "iter-extract",
            { forceRefresh: true }
        );
        const processedIterationDocument = await api.evals.extract.datasets.iterations.processDocument(
            "eval-extract",
            "dataset-extract",
            "iter-extract",
            "iter-doc-extract"
        );
        const finalizedIteration = await api.evals.extract.datasets.iterations.finalize(
            "eval-extract",
            "dataset-extract",
            "iter-extract"
        );

        expect(createdEval.id).toBe("eval-extract");
        expect(fetchedEval.name).toBe("Extract Eval");
        expect(listedEvals[0].id).toBe("eval-extract");
        expect(updatedEval.name).toBe("Extract Eval v2");
        expect(publishedEval.is_published).toBe(true);
        expect(processedEval.choices[0].message.parsed).toEqual({ status: "ok" });

        expect(dataset.project_id).toBe("eval-extract");
        expect(fetchedDataset.id).toBe("dataset-extract");
        expect(listedDatasets[0].id).toBe("dataset-extract");
        expect(updatedDataset.name).toBe("Extract Dataset v2");
        expect(duplicatedDataset.name).toBe("Extract Dataset Copy");
        expect(datasetDocument.extraction_id).toBe("ext-1");
        expect(fetchedDatasetDocument.id).toBe("doc-extract");
        expect(listedDatasetDocuments[0].validation_flags.status).toBe("reviewed");
        expect(updatedDatasetDocument.extraction_id).toBe("ext-updated");
        expect(processedDatasetDocument.choices[0].message.parsed).toEqual({ status: "ok" });

        expect(iteration.id).toBe("iter-extract");
        expect(fetchedIteration.dataset_id).toBe("dataset-extract");
        expect(listedIterations[0].status).toBe("draft");
        expect(updatedIteration.draft.schema_overrides.descriptionsOverride).toEqual({ status: "Review status" });
        expect(schema.properties.status.description).toBe("Review status");
        expect(processDocumentsResult).toEqual({ success: true, id: "iter-doc-extract" });
        expect(iterationDocument.dataset_document_id).toBe("doc-extract");
        expect(listedIterationDocuments[0].id).toBe("iter-doc-extract");
        expect(updatedIterationDocument.extraction_id).toBe("ext-iter-updated");
        expect(metrics.overall_metrics.total_documents).toBe(1);
        expect(processedIterationDocument.choices[0].message.parsed).toEqual({ status: "ok" });
        expect(finalizedIteration.status).toBe("completed");

        const createEvalRequest = findRequest(mockClient.requests, "POST", "/evals/extract");
        expect(createEvalRequest.url).toBe("/evals/extract");
        expect(createEvalRequest.body?.json_schema).toEqual({
            type: "object",
            properties: {
                status: { type: "string" },
            },
        });

        const publishRequest = findRequest(mockClient.requests, "POST", "/evals/extract/eval-extract/publish");
        expect(publishRequest.url).toBe("/evals/extract/eval-extract/publish");
        expect(publishRequest.params).toEqual({ origin: "iter-extract" });

        const processRequest = findRequest(mockClient.requests, "POST", "/evals/extract/extract/eval-extract");
        expect(processRequest.url).toBe("/evals/extract/extract/eval-extract");
        expect(processRequest.body?.metadata).toBe(JSON.stringify({ source: "lifecycle" }));

        const addDocumentRequest = findRequest(
            mockClient.requests,
            "POST",
            "/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents"
        );
        expect(addDocumentRequest.url).toBe("/evals/extract/eval-extract/datasets/dataset-extract/dataset-documents");
        expect(addDocumentRequest.body?.project_id).toBe("eval-extract");
        expect(addDocumentRequest.body?.dataset_id).toBe("dataset-extract");

        const createIterationRequest = findRequest(
            mockClient.requests,
            "POST",
            "/evals/extract/eval-extract/datasets/dataset-extract/iterations"
        );
        expect(createIterationRequest.url).toBe("/evals/extract/eval-extract/datasets/dataset-extract/iterations");
        expect(createIterationRequest.body?.project_id).toBe("eval-extract");
        expect(createIterationRequest.body?.dataset_id).toBe("dataset-extract");

        const updateDraftRequest = findRequest(
            mockClient.requests,
            "PATCH",
            "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract"
        );
        expect(updateDraftRequest.body).toEqual({
            draft: {
                schema_overrides: {
                    descriptionsOverride: {
                        status: "Review status",
                    },
                },
            },
        });

        const getSchemaRequest = findRequest(
            mockClient.requests,
            "GET",
            "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/schema"
        );
        expect(getSchemaRequest.params).toEqual({ use_draft: true });

        const processDocumentsRequest = findRequest(
            mockClient.requests,
            "POST",
            "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/iteration-documents/processDocumentsFromDatasetId"
        );
        expect(processDocumentsRequest.body).toEqual({ dataset_document_id: "doc-extract" });

        const metricsRequest = findRequest(
            mockClient.requests,
            "GET",
            "/evals/extract/eval-extract/datasets/dataset-extract/iterations/iter-extract/metrics"
        );
        expect(metricsRequest.params).toEqual({ force_refresh: true });
    });
});
