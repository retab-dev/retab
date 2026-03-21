import { fileURLToPath } from "node:url";
import { readFileSync } from "node:fs";

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

function buildProject(body: Record<string, unknown> = {}) {
    return {
        id: "proj-123",
        name: typeof body.name === "string" ? body.name : "Project",
        updated_at: "2026-03-18T00:00:00Z",
        published_config: {
            inference_settings: {
                model: "retab-small",
                image_resolution_dpi: 192,
                n_consensus: 1,
            },
            json_schema: (body.json_schema as Record<string, unknown> | undefined) || {
                type: "object",
                properties: {
                    status: { type: "string" },
                },
            },
            origin: "manual",
        },
        draft_config: {
            inference_settings: {
                model: "retab-small",
                image_resolution_dpi: 192,
                n_consensus: 1,
            },
            json_schema: (body.json_schema as Record<string, unknown> | undefined) || {
                type: "object",
                properties: {
                    status: { type: "string" },
                },
            },
        },
        is_published: false,
        is_schema_generated: true,
    };
}

function buildDataset(body: Record<string, unknown> = {}) {
    return {
        id: "dataset-123",
        name: typeof body.name === "string" ? body.name : "Dataset",
        updated_at: "2026-03-18T00:00:00Z",
        base_json_schema: (body.base_json_schema as Record<string, unknown> | undefined) || {
            type: "object",
            properties: {
                status: { type: "string" },
            },
        },
        project_id: "proj-123",
    };
}

function buildDatasetDocument(body: Record<string, unknown> = {}) {
    return {
        id: "doc-123",
        updated_at: "2026-03-18T00:00:00Z",
        project_id: "proj-123",
        dataset_id: "dataset-123",
        mime_data: {
            id: "file-123",
            filename: "document.pdf",
            mime_type: "application/pdf",
        },
        prediction_data: (body.prediction_data as Record<string, unknown> | undefined) || {},
        extraction_id: typeof body.extraction_id === "string" ? body.extraction_id : null,
        validation_flags: (body.validation_flags as Record<string, unknown> | undefined) || {},
    };
}

function buildIteration(body: Record<string, unknown> = {}) {
    const draft = (body.draft as Record<string, unknown> | undefined) || {};

    return {
        id: "iter-123",
        updated_at: "2026-03-18T00:00:00Z",
        inference_settings: {
            model: "retab-small",
            reasoning_effort: "minimal",
            image_resolution_dpi: 192,
            n_consensus: 1,
        },
        schema_overrides: {},
        parent_id: typeof body.parent_id === "string" ? body.parent_id : null,
        project_id: "proj-123",
        dataset_id: "dataset-123",
        draft: {
            schema_overrides: (draft.schema_overrides as Record<string, unknown> | undefined) || {},
            inference_settings: (draft.inference_settings as Record<string, unknown> | undefined) || {
                model: "retab-small",
                reasoning_effort: "minimal",
                image_resolution_dpi: 192,
                n_consensus: 1,
            },
            updated_at: "2026-03-18T00:00:00Z",
        },
        status: typeof body.status === "string" ? body.status : "draft",
        finalized_at: typeof body.status === "string" && body.status === "completed" ? "2026-03-18T00:00:00Z" : null,
        last_finalize_error: null,
    };
}

function buildIterationDocument(body: Record<string, unknown> = {}) {
    return {
        id: "iter-doc-123",
        updated_at: "2026-03-18T00:00:00Z",
        project_id: "proj-123",
        iteration_id: "iter-123",
        dataset_id: "dataset-123",
        dataset_document_id: "doc-123",
        mime_data: {
            id: "file-123",
            filename: "document.pdf",
            mime_type: "application/pdf",
        },
        prediction_data: (body.prediction_data as Record<string, unknown> | undefined) || {},
        extraction_id: typeof body.extraction_id === "string" ? body.extraction_id : null,
    };
}

function buildParsedCompletion() {
    return {
        id: "chatcmpl-123",
        object: "chat.completion",
        created: 1710000000,
        model: "retab-small",
        choices: [
            {
                index: 0,
                finish_reason: "stop",
                message: {
                    role: "assistant",
                    content: "{}",
                    parsed: {},
                },
            },
        ],
        extraction_id: "extract-123",
    };
}

function buildResponse(params: RecordedRequest): unknown {
    if (params.url === "/projects" && params.method === "POST") {
        return buildProject(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123" && params.method === "GET") {
        return buildProject();
    }
    if (params.url === "/projects" && params.method === "GET") {
        return {
            data: [buildProject()],
            list_metadata: {},
        };
    }
    if (params.url === "/projects/proj-123" && params.method === "PATCH") {
        return buildProject(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123" && params.method === "DELETE") {
        return { success: true, id: "proj-123" };
    }
    if (params.url === "/projects/proj-123/publish" && params.method === "POST") {
        return { ...buildProject(), is_published: true };
    }
    if (params.url === "/projects/proj-123/datasets" && params.method === "POST") {
        return buildDataset(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123" && params.method === "GET") {
        return buildDataset();
    }
    if (params.url === "/projects/proj-123/datasets" && params.method === "GET") {
        return {
            data: [buildDataset()],
            list_metadata: {},
        };
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123" && params.method === "PATCH") {
        return buildDataset(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123" && params.method === "DELETE") {
        return { success: true, id: "dataset-123" };
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/duplicate" && params.method === "POST") {
        return buildDataset(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/dataset-documents" && params.method === "POST") {
        return buildDatasetDocument(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/dataset-documents/doc-123" && params.method === "GET") {
        return buildDatasetDocument();
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/dataset-documents" && params.method === "GET") {
        return [buildDatasetDocument()];
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/dataset-documents/doc-123" && params.method === "PATCH") {
        return buildDatasetDocument(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/dataset-documents/doc-123" && params.method === "DELETE") {
        return { success: true, id: "doc-123" };
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations" && params.method === "POST") {
        return buildIteration(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123" && params.method === "GET") {
        return buildIteration();
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations" && params.method === "GET") {
        return {
            data: [buildIteration()],
            list_metadata: {},
        };
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123" && params.method === "PATCH") {
        return buildIteration(params.body as Record<string, unknown>);
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123" && params.method === "DELETE") {
        return { success: true, id: "iter-123" };
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/finalize" && params.method === "POST") {
        return buildIteration({ status: "completed" });
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/schema" && params.method === "GET") {
        return {
            json_schema: {
                type: "object",
                properties: {
                    status: { type: "string" },
                },
            },
        };
    }
    if (
        params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/iteration-documents/processDocumentsFromDatasetId" &&
        params.method === "POST"
    ) {
        return { success: true, id: "iter-doc-123" };
    }
    if (
        params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/iteration-documents/iter-doc-123" &&
        params.method === "GET"
    ) {
        return buildIterationDocument();
    }
    if (
        params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/iteration-documents" &&
        params.method === "GET"
    ) {
        return {
            data: [buildIterationDocument()],
        };
    }
    if (
        params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/iteration-documents/iter-doc-123" &&
        params.method === "PATCH"
    ) {
        return buildIterationDocument(params.body as Record<string, unknown>);
    }
    if (
        params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/iteration-documents/iter-doc-123" &&
        params.method === "DELETE"
    ) {
        return { success: true, id: "iter-doc-123" };
    }
    if (params.url === "/projects/proj-123/datasets/dataset-123/iterations/iter-123/metrics" && params.method === "GET") {
        return {
            exact_match: 0.95,
        };
    }
    if (params.url === "/projects/extract/proj-123" && params.method === "POST") {
        return buildParsedCompletion();
    }
    if (params.url === "/projects/split/proj-123" && params.method === "POST") {
        return buildParsedCompletion();
    }

    throw new Error(`Unhandled mock request: ${params.method} ${params.url}`);
}

describe("projects resource parity", () => {
    test("projects expose datasets and nested iterations", () => {
        const api = new APIV1(new MockClient());

        expect(api.projects.datasets).toBeDefined();
        expect(api.projects.datasets.iterations).toBeDefined();
    });

    test("project CRUD and nested dataset lifecycle methods mirror the python client", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);
        const schemaPath = fileURLToPath(new URL("./data/freight/booking_confirmation_schema_small.json", import.meta.url));
        const documentPath = fileURLToPath(new URL("./data/freight/booking_confirmation_1.jpg", import.meta.url));
        const expectedSchema = JSON.parse(readFileSync(schemaPath, "utf-8")) as Record<string, unknown>;

        const createdProject = await api.projects.create({
            name: "Project A",
            json_schema: schemaPath,
        });
        const fetchedProject = await api.projects.get("proj-123");
        const listedProjects = await api.projects.list();
        const preparedProjectUpdate = await api.projects.prepare_update("proj-123", {
            name: "Project B",
            json_schema: schemaPath,
        });
        const publishedProject = await api.projects.publish("proj-123");
        const deletedProject = await api.projects.delete("proj-123");

        const createdDataset = await api.projects.datasets.create("proj-123", {
            name: "Dataset A",
            base_json_schema: schemaPath,
        });
        const fetchedDataset = await api.projects.datasets.get("proj-123", "dataset-123");
        const listedDatasets = await api.projects.datasets.list("proj-123");
        const updatedDataset = await api.projects.datasets.update("proj-123", "dataset-123", {
            name: "Dataset B",
        });
        const duplicatedDataset = await api.projects.datasets.duplicate("proj-123", "dataset-123", {
            name: "Dataset Copy",
        });
        const createdDatasetDocument = await api.projects.datasets.addDocument("proj-123", "dataset-123", {
            mime_data: documentPath,
            prediction_data: { status: "pending" },
        });
        const fetchedDatasetDocument = await api.projects.datasets.getDocument("proj-123", "dataset-123", "doc-123");
        const listedDatasetDocuments = await api.projects.datasets.listDocuments("proj-123", "dataset-123");
        const updatedDatasetDocument = await api.projects.datasets.updateDocument("proj-123", "dataset-123", "doc-123", {
            validation_flags: { reviewed: true },
            extraction_id: "extract-456",
        });
        const deletedDatasetDocument = await api.projects.datasets.deleteDocument("proj-123", "dataset-123", "doc-123");
        const deletedDataset = await api.projects.datasets.delete("proj-123", "dataset-123");

        expect(createdProject.name).toBe("Project A");
        expect(fetchedProject.id).toBe("proj-123");
        expect(listedProjects).toHaveLength(1);
        expect(preparedProjectUpdate).toMatchObject({
            url: "/projects/proj-123",
            method: "PATCH",
        });
        expect((preparedProjectUpdate.body as Record<string, unknown>).name).toBe("Project B");
        expect((preparedProjectUpdate.body as Record<string, unknown>).json_schema).toEqual(expectedSchema);
        expect(publishedProject.is_published).toBe(true);
        expect(deletedProject.success).toBe(true);

        expect(createdDataset.name).toBe("Dataset A");
        expect(fetchedDataset.id).toBe("dataset-123");
        expect(listedDatasets).toHaveLength(1);
        expect(updatedDataset.name).toBe("Dataset B");
        expect(duplicatedDataset.name).toBe("Dataset Copy");
        expect(createdDatasetDocument.id).toBe("doc-123");
        expect(fetchedDatasetDocument.id).toBe("doc-123");
        expect(listedDatasetDocuments).toHaveLength(1);
        expect(updatedDatasetDocument.validation_flags.reviewed).toBe(true);
        expect(deletedDatasetDocument.success).toBe(true);
        expect(deletedDataset.success).toBe(true);

        const createProjectRequest = findRequest(mockClient.requests, "POST", "/projects");
        expect(createProjectRequest.body?.json_schema).toEqual(expectedSchema);

        const createDatasetRequest = findRequest(mockClient.requests, "POST", "/projects/proj-123/datasets");
        expect(createDatasetRequest.body?.base_json_schema).toEqual(createProjectRequest.body?.json_schema);

        const addDocumentRequest = findRequest(
            mockClient.requests,
            "POST",
            "/projects/proj-123/datasets/dataset-123/dataset-documents"
        );
        expect(addDocumentRequest.body?.project_id).toBe("proj-123");
        expect(addDocumentRequest.body?.dataset_id).toBe("dataset-123");
    });

    test("project iterations support draft updates, schema previews, metrics, and document operations", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);

        const createdIteration = await api.projects.datasets.iterations.create("proj-123", "dataset-123");
        const fetchedIteration = await api.projects.datasets.iterations.get("proj-123", "dataset-123", "iter-123");
        const listedIterations = await api.projects.datasets.iterations.list("proj-123", "dataset-123");
        const updatedIteration = await api.projects.datasets.iterations.updateDraft("proj-123", "dataset-123", "iter-123", {
            schema_overrides: {
                descriptionsOverride: {
                    status: "Review status",
                },
            },
        });
        const schema = await api.projects.datasets.iterations.getSchema("proj-123", "dataset-123", "iter-123", {
            useDraft: true,
        });
        const processDocumentsResult = await api.projects.datasets.iterations.processDocuments("proj-123", "dataset-123", "iter-123", "doc-123");
        const fetchedIterationDocument = await api.projects.datasets.iterations.getDocument("proj-123", "dataset-123", "iter-123", "iter-doc-123");
        const listedIterationDocuments = await api.projects.datasets.iterations.listDocuments("proj-123", "dataset-123", "iter-123");
        const updatedIterationDocument = await api.projects.datasets.iterations.updateDocument("proj-123", "dataset-123", "iter-123", "iter-doc-123", {
            extraction_id: "extract-789",
        });
        const deletedIterationDocument = await api.projects.datasets.iterations.deleteDocument("proj-123", "dataset-123", "iter-123", "iter-doc-123");
        const metrics = await api.projects.datasets.iterations.getMetrics("proj-123", "dataset-123", "iter-123", {
            forceRefresh: true,
        });
        const finalizedIteration = await api.projects.datasets.iterations.finalize("proj-123", "dataset-123", "iter-123");
        const deletedIteration = await api.projects.datasets.iterations.delete("proj-123", "dataset-123", "iter-123");

        expect(createdIteration.id).toBe("iter-123");
        expect(fetchedIteration.id).toBe("iter-123");
        expect(listedIterations).toHaveLength(1);
        expect(updatedIteration.draft.schema_overrides.descriptionsOverride).toEqual({ status: "Review status" });
        expect(schema.json_schema).toBeDefined();
        expect(processDocumentsResult.success).toBe(true);
        expect(fetchedIterationDocument.id).toBe("iter-doc-123");
        expect(listedIterationDocuments).toHaveLength(1);
        expect(updatedIterationDocument.extraction_id).toBe("extract-789");
        expect(deletedIterationDocument.success).toBe(true);
        expect(metrics.exact_match).toBe(0.95);
        expect(finalizedIteration.status).toBe("completed");
        expect(deletedIteration.success).toBe(true);

        const updateDraftRequest = findRequest(
            mockClient.requests,
            "PATCH",
            "/projects/proj-123/datasets/dataset-123/iterations/iter-123"
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
            "/projects/proj-123/datasets/dataset-123/iterations/iter-123/schema"
        );
        expect(getSchemaRequest.params).toEqual({ use_draft: true });

        const getMetricsRequest = findRequest(
            mockClient.requests,
            "GET",
            "/projects/proj-123/datasets/dataset-123/iterations/iter-123/metrics"
        );
        expect(getMetricsRequest.params).toEqual({ force_refresh: true });
    });

    test("projects.extract supports the python proxy semantics for single and multiple documents", async () => {
        const mockClient = new MockClient();
        const api = new APIV1(mockClient);
        const documentPath1 = fileURLToPath(new URL("./data/freight/booking_confirmation_1.jpg", import.meta.url));
        const documentPath2 = fileURLToPath(new URL("./data/freight/booking_confirmation_2.jpg", import.meta.url));

        const response = await api.projects.extract({
            project_id: "proj-123",
            documents: [documentPath1, documentPath2],
        });

        expect(response.choices).toHaveLength(1);

        const extractRequest = findRequest(mockClient.requests, "POST", "/projects/extract/proj-123");
        expect(extractRequest.bodyMime).toBe("multipart/form-data");
        expect(Array.isArray(extractRequest.body?.documents)).toBe(true);
        expect((extractRequest.body?.documents as unknown[])).toHaveLength(2);
    });
});
