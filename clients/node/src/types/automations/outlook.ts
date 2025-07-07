import { AutomationConfig, UpdateAutomationRequest } from '../logs.js';
import { ListMetadata } from '../pagination.js';

export interface AutomationLevel {
  distance_threshold: number;
  score_threshold: number;
}

export interface MatchParams {
  endpoint: string;
  headers: Record<string, string>;
  path: string;
}

export interface FetchParams {
  endpoint: string;
  headers: Record<string, string>;
  name: string;
}

export interface Outlook extends AutomationConfig {
  object: 'automation.outlook';
  id: string;
  authorized_domains: string[];
  authorized_emails: string[];
  layout_schema?: Record<string, any>;
  match_params: MatchParams[];
  fetch_params: FetchParams[];
}

export interface ListOutlooks {
  data: Outlook[];
  list_metadata: ListMetadata;
}

export interface UpdateOutlookRequest extends UpdateAutomationRequest {
  authorized_domains?: string[];
  authorized_emails?: string[];
  match_params?: MatchParams[];
  fetch_params?: FetchParams[];
  layout_schema?: Record<string, any>;
}