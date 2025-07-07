import { AutomationConfig, UpdateAutomationRequest } from '../logs.js';
import { ListMetadata } from '../pagination.js';

export interface Link extends AutomationConfig {
  object: 'automation.link';
  id: string;
  password?: string;
}

export interface ListLinks {
  data: Link[];
  list_metadata: ListMetadata;
}

export interface UpdateLinkRequest extends UpdateAutomationRequest {
  password?: string;
}