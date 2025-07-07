import { AutomationConfig, UpdateAutomationRequest } from '../logs.js';
import { ListMetadata } from '../pagination.js';

export interface Endpoint extends AutomationConfig {
  object: 'automation.endpoint';
  id: string;
}

export interface ListEndpoints {
  data: Endpoint[];
  list_metadata: ListMetadata;
}

export interface UpdateEndpointRequest extends UpdateAutomationRequest {
  // Inherits all properties from UpdateAutomationRequest
}