import * as z from 'zod';
import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  MIMEDataInput,
  ZMIMEData,
  WorkflowRun,
  ZWorkflowRun,
  PaginatedList,
  ZPaginatedList,
  CancelWorkflowResponse,
  ZCancelWorkflowResponse,
  WorkflowRunExportResponse,
  ZWorkflowRunExportResponse,
  WorkflowRunStatus,
  WorkflowRunTriggerType,
} from '../../../types.js';
import { ZFileRef, type FileRef } from '../../../generated_types.js';
import APIWorkflowRunSteps from './steps/client.js';

type WorkflowRunDocumentInput = MIMEDataInput | FileRef;

function isFileRefInput(document: WorkflowRunDocumentInput): document is FileRef {
  const parsed = ZFileRef.safeParse(document);
  return parsed.success;
}

function normalizeCsvParam(value?: string | string[]): string | undefined {
  if (value === undefined) {
    return undefined;
  }
  return Array.isArray(value) ? value.join(',') : value;
}

function normalizeDateParam(value?: string | Date): string | undefined {
  if (value === undefined) {
    return undefined;
  }
  if (value instanceof Date) {
    return value.toISOString().slice(0, 10);
  }
  return value;
}

/**
 * Workflow Runs API client for managing workflow executions.
 *
 * Sub-clients:
 * - steps: Step execution operations (get, list)
 */
export default class APIWorkflowRuns extends CompositionClient {
  public steps: APIWorkflowRunSteps;

  constructor(client: CompositionClient) {
    super(client);
    this.steps = new APIWorkflowRunSteps(this);
  }

  /**
   * Run a workflow with the provided inputs.
   *
   * This creates a workflow run and starts execution in the background.
   * The returned WorkflowRun will have lifecycle.status "running" or
   * "pending" - use get() to check for updates on the run lifecycle.
   *
   * @example
   * ```typescript
   * const run = await client.workflows.runs.create({
   *     workflowId: "wf_abc123",
   *     documents: { "start-block-1": "./invoice.pdf" },
   * });
   * ```
   */
  async create(
    {
      workflowId,
      documents,
      jsonInputs,
      version = 'production',
    }: {
      workflowId: string;
      documents?: Record<string, WorkflowRunDocumentInput>;
      jsonInputs?: Record<string, unknown>;
      version?: string;
    },
    options?: RequestOptions
  ): Promise<WorkflowRun> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {};

    if (documents) {
      const documentsPayload: Record<
        string,
        { id?: string; filename: string; content?: string; mime_type: string }
      > = {};

      for (const [blockId, document] of Object.entries(documents)) {
        if (isFileRefInput(document)) {
          documentsPayload[blockId] = {
            id: document.id,
            filename: document.filename,
            mime_type: document.mime_type,
          };
          continue;
        }
        const parsedDocument = await ZMIMEData.parseAsync(document);
        const content = parsedDocument.url.split(',')[1];
        const mimeType = parsedDocument.url.split(';')[0].split(':')[1];

        documentsPayload[blockId] = {
          filename: parsedDocument.filename,
          content: content,
          mime_type: mimeType,
        };
      }
      body.documents = documentsPayload;
    }

    if (jsonInputs) {
      body.json_inputs = jsonInputs;
    }
    body.version = version;

    return this._fetchJson(ZWorkflowRun, {
      url: `/workflows/${workflowId}/run`,
      method: 'POST',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get a workflow run by ID.
   */
  async get(runId: string, options?: RequestOptions): Promise<WorkflowRun> {
    return this._fetchJson(ZWorkflowRun, {
      url: `/workflows/runs/${runId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * List workflow runs with filtering and pagination.
   */
  async list(
    {
      workflowId,
      status,
      statuses,
      excludeStatus,
      triggerType,
      triggerTypes,
      fromDate,
      toDate,
      minCost,
      maxCost,
      minDuration,
      maxDuration,
      search,
      sortBy,
      fields,
      before,
      after,
      limit = 20,
      order = 'desc',
    }: {
      workflowId?: string;
      status?: WorkflowRunStatus;
      statuses?: WorkflowRunStatus[] | string;
      excludeStatus?: WorkflowRunStatus;
      triggerType?: WorkflowRunTriggerType;
      triggerTypes?: WorkflowRunTriggerType[] | string;
      fromDate?: string | Date;
      toDate?: string | Date;
      minCost?: number;
      maxCost?: number;
      minDuration?: number;
      maxDuration?: number;
      search?: string;
      sortBy?: string;
      fields?: string[] | string;
      before?: string;
      after?: string;
      limit?: number;
      order?: 'asc' | 'desc';
    } = {},
    options?: RequestOptions
  ): Promise<PaginatedList> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const params: Record<string, any> = {
      workflow_id: workflowId,
      status,
      statuses: normalizeCsvParam(statuses),
      exclude_status: excludeStatus,
      trigger_type: triggerType,
      trigger_types: normalizeCsvParam(triggerTypes),
      from_date: normalizeDateParam(fromDate),
      to_date: normalizeDateParam(toDate),
      min_cost: minCost,
      max_cost: maxCost,
      min_duration: minDuration,
      max_duration: maxDuration,
      search,
      sort_by: sortBy,
      fields: normalizeCsvParam(fields),
      before,
      after,
      limit,
      order,
    };

    const cleanParams = Object.fromEntries(
      Object.entries(params).filter(([_, v]) => v !== undefined)
    );

    return this._fetchJson(ZPaginatedList, {
      url: '/workflows/runs',
      method: 'GET',
      params: { ...cleanParams, ...(options?.params || {}) },
      headers: options?.headers,
    });
  }

  /**
   * Delete a workflow run and its associated step data.
   */
  async delete(runId: string, options?: RequestOptions): Promise<void> {
    return this._fetchJson({
      url: `/workflows/runs/${runId}`,
      method: 'DELETE',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Cancel a running or pending workflow run.
   */
  async cancel(
    runId: string,
    { commandId }: { commandId?: string } = {},
    options?: RequestOptions
  ): Promise<CancelWorkflowResponse> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {};
    if (commandId !== undefined) body.command_id = commandId;

    return this._fetchJson(ZCancelWorkflowResponse, {
      url: `/workflows/runs/${runId}/cancel`,
      method: 'POST',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Restart a completed or failed workflow run with the same inputs.
   */
  async restart(
    runId: string,
    { commandId }: { commandId?: string } = {},
    options?: RequestOptions
  ): Promise<WorkflowRun> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {};
    if (commandId !== undefined) body.command_id = commandId;

    return this._fetchJson(ZWorkflowRun, {
      url: `/workflows/runs/${runId}/restart`,
      method: 'POST',
      body: { ...body, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get the configuration snapshot used for a run.
   */
  async getConfig(runId: string, options?: RequestOptions): Promise<Record<string, unknown>> {
    return this._fetchJson(z.record(z.any()), {
      url: `/workflows/runs/${runId}/config`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get the DAG-ordered execution order for a run.
   */
  async executionOrder(runId: string, options?: RequestOptions): Promise<Record<string, unknown>> {
    return this._fetchJson(z.record(z.any()), {
      url: `/workflows/runs/${runId}/execution-order`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Get a signed URL for downloading a document from a run step.
   */
  async getDocumentUrl(
    runId: string,
    blockId: string,
    options?: RequestOptions
  ): Promise<Record<string, unknown>> {
    return this._fetchJson(z.record(z.any()), {
      url: `/workflows/runs/${runId}/documents/${blockId}`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Export run results as structured CSV data.
   *
   * @returns Object with `csv_data` (string), `rows` (number), `columns` (number)
   */
  async export(
    {
      workflowId,
      blockId,
      exportSource = 'outputs',
      selectedRunIds,
      status,
      excludeStatus,
      fromDate,
      toDate,
      triggerTypes,
      preferredColumns,
    }: {
      workflowId: string;
      blockId: string;
      exportSource?: 'outputs' | 'inputs';
      selectedRunIds?: string[];
      status?: string;
      excludeStatus?: string;
      fromDate?: string | Date;
      toDate?: string | Date;
      triggerTypes?: WorkflowRunTriggerType[];
      preferredColumns?: string[];
    },
    options?: RequestOptions
  ): Promise<WorkflowRunExportResponse> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const body: Record<string, any> = {
      workflow_id: workflowId,
      block_id: blockId,
      export_source: exportSource,
      preferred_columns: preferredColumns || [],
    };
    if (selectedRunIds !== undefined) body.selected_run_ids = selectedRunIds;
    if (status !== undefined) body.status = status;
    if (excludeStatus !== undefined) body.exclude_status = excludeStatus;
    if (fromDate !== undefined) body.from_date = normalizeDateParam(fromDate);
    if (toDate !== undefined) body.to_date = normalizeDateParam(toDate);
    if (triggerTypes !== undefined) body.trigger_types = triggerTypes;

    return this._fetchJson(ZWorkflowRunExportResponse, {
      url: '/workflows/runs/export_payload',
      method: 'POST',
      body: { ...body, ...((options?.body as Record<string, unknown>) || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }
}
