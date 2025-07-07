import { AutomationConfig, UpdateAutomationRequest } from '../logs.js';
import { ListMetadata } from '../pagination.js';

export interface Mailbox extends AutomationConfig {
  object: 'automation.mailbox';
  id: string;
  email: string;
  authorized_domains: string[];
  authorized_emails: string[];
}

export interface ListMailboxes {
  data: Mailbox[];
  list_metadata: ListMetadata;
}

export interface UpdateMailboxRequest extends UpdateAutomationRequest {
  authorized_domains?: string[];
  authorized_emails?: string[];
}