export interface WebhookRequest {
  completion: any; // RetabParsedChatCompletion
  user?: string;
  file_payload: any; // MIMEData
  metadata?: Record<string, any>;
}

export interface BaseWebhookRequest {
  completion: any; // RetabParsedChatCompletion
  user?: string;
  file_payload: any; // BaseMIMEData
  metadata?: Record<string, any>;
}