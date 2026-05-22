import { CompositionClient, RequestOptions } from '../../../client.js';
import {
  DeclarativeApplyResponse,
  DeclarativeExportResponse,
  DeclarativePlanResponse,
  DeclarativeValidationResponse,
  ZDeclarativeApplyResponse,
  ZDeclarativeExportResponse,
  ZDeclarativePlanResponse,
  ZDeclarativeValidationResponse,
} from '../../../types.js';

/**
 * Declarative workflow spec operations.
 */
export default class APIWorkflowSpec extends CompositionClient {
  /**
   * Validate declarative workflow YAML without mutating workflow state.
   */
  async validate(
    yamlDefinition: string,
    options?: RequestOptions
  ): Promise<DeclarativeValidationResponse> {
    return this._fetchJson(ZDeclarativeValidationResponse, {
      url: '/workflows/spec/validate',
      method: 'POST',
      body: { yaml_definition: yamlDefinition, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Plan the changes needed to reconcile declarative YAML with the current draft.
   */
  async plan(yamlDefinition: string, options?: RequestOptions): Promise<DeclarativePlanResponse> {
    return this._fetchJson(ZDeclarativePlanResponse, {
      url: '/workflows/spec/plan',
      method: 'POST',
      body: { yaml_definition: yamlDefinition, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Apply declarative workflow YAML to draft workflow state.
   */
  async apply(yamlDefinition: string, options?: RequestOptions): Promise<DeclarativeApplyResponse> {
    return this._fetchJson(ZDeclarativeApplyResponse, {
      url: '/workflows/spec/apply',
      method: 'POST',
      body: { yaml_definition: yamlDefinition, ...(options?.body || {}) },
      params: options?.params,
      headers: options?.headers,
    });
  }

  /**
   * Export a workflow draft as canonical declarative YAML.
   */
  async export(workflowId: string, options?: RequestOptions): Promise<DeclarativeExportResponse> {
    return this._fetchJson(ZDeclarativeExportResponse, {
      url: `/workflows/${workflowId}/spec`,
      method: 'GET',
      params: options?.params,
      headers: options?.headers,
    });
  }
}
