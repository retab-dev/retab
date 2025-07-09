import { createHash, createHmac, timingSafeEqual } from 'crypto';

/**
 * Webhook secret management utilities for secure webhook verification
 * Equivalent to Python's webhook_secrets.py
 */

export interface WebhookConfig {
  secret: string;
  tolerance?: number; // seconds
  encoding?: 'hex' | 'base64';
}

export interface WebhookHeaders {
  signature?: string;
  timestamp?: string;
  [key: string]: string | undefined;
}

/**
 * Generate webhook signature for payload
 */
export function generateSignature(
  payload: string | Buffer,
  secret: string,
  timestamp?: number,
  encoding: 'hex' | 'base64' = 'hex'
): string {
  const ts = timestamp || Math.floor(Date.now() / 1000);
  const payloadString = Buffer.isBuffer(payload) ? payload.toString('utf8') : payload;
  const signedPayload = `${ts}.${payloadString}`;
  
  const hmac = createHmac('sha256', secret);
  hmac.update(signedPayload);
  
  return hmac.digest(encoding);
}

/**
 * Verify webhook signature
 */
export function verifySignature(
  payload: string | Buffer,
  signature: string,
  secret: string,
  timestamp?: number,
  config: Partial<WebhookConfig> = {}
): boolean {
  const { tolerance = 300, encoding = 'hex' } = config; // 5 minute default tolerance
  
  try {
    const ts = timestamp || Math.floor(Date.now() / 1000);
    const expectedSignature = generateSignature(payload, secret, ts, encoding);
    
    // Use timing-safe comparison to prevent timing attacks
    const sigBuffer = Buffer.from(signature, encoding);
    const expectedBuffer = Buffer.from(expectedSignature, encoding);
    
    if (sigBuffer.length !== expectedBuffer.length) {
      return false;
    }
    
    const isSignatureValid = timingSafeEqual(sigBuffer, expectedBuffer);
    
    // Check timestamp tolerance if provided
    if (timestamp && tolerance > 0) {
      const now = Math.floor(Date.now() / 1000);
      const isTimestampValid = Math.abs(now - timestamp) <= tolerance;
      return isSignatureValid && isTimestampValid;
    }
    
    return isSignatureValid;
  } catch (error) {
    return false;
  }
}

/**
 * Parse webhook headers for signature and timestamp
 */
export function parseWebhookHeaders(headers: WebhookHeaders): {
  signature?: string;
  timestamp?: number;
} {
  const signature = headers['x-webhook-signature'] || 
                   headers['x-retab-signature'] || 
                   headers['signature'];
  
  const timestampHeader = headers['x-webhook-timestamp'] || 
                         headers['x-retab-timestamp'] || 
                         headers['timestamp'];
  
  const timestamp = timestampHeader ? parseInt(timestampHeader, 10) : undefined;
  
  return { signature, timestamp };
}

/**
 * Verify webhook request
 */
export function verifyWebhookRequest(
  payload: string | Buffer,
  headers: WebhookHeaders,
  secret: string,
  config: Partial<WebhookConfig> = {}
): boolean {
  const { signature, timestamp } = parseWebhookHeaders(headers);
  
  if (!signature) {
    return false;
  }
  
  return verifySignature(payload, signature, secret, timestamp, config);
}

/**
 * Generate webhook secret
 */
export function generateWebhookSecret(length: number = 32): string {
  return createHash('sha256')
    .update(Math.random().toString())
    .update(Date.now().toString())
    .digest('hex')
    .slice(0, length);
}

/**
 * Create webhook signature header
 */
export function createSignatureHeader(
  payload: string | Buffer,
  secret: string,
  encoding: 'hex' | 'base64' = 'hex'
): { signature: string; timestamp: number } {
  const timestamp = Math.floor(Date.now() / 1000);
  const signature = generateSignature(payload, secret, timestamp, encoding);
  
  return { signature, timestamp };
}

/**
 * Webhook middleware for Express-like frameworks
 */
export function webhookMiddleware(secret: string, config: Partial<WebhookConfig> = {}) {
  return (req: any, res: any, next: any) => {
    try {
      const payload = req.rawBody || req.body;
      const headers = req.headers;
      
      if (!verifyWebhookRequest(payload, headers, secret, config)) {
        return res.status(401).json({ error: 'Invalid webhook signature' });
      }
      
      next();
    } catch (error) {
      return res.status(400).json({ error: 'Webhook verification failed' });
    }
  };
}

export default {
  generateSignature,
  verifySignature,
  parseWebhookHeaders,
  verifyWebhookRequest,
  generateWebhookSecret,
  createSignatureHeader,
  webhookMiddleware,
};