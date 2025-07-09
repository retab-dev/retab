import { createHash, createHmac } from 'crypto';

/**
 * Cryptographic hashing utilities
 * Equivalent to Python's utils/hashing.py
 */

export function md5(data: string | Buffer): string {
  return createHash('md5').update(data).digest('hex');
}

export function sha256(data: string | Buffer): string {
  return createHash('sha256').update(data).digest('hex');
}

export function sha512(data: string | Buffer): string {
  return createHash('sha512').update(data).digest('hex');
}

export function hmacSha256(data: string | Buffer, secret: string): string {
  return createHmac('sha256', secret).update(data).digest('hex');
}

export function contentHash(content: any): string {
  const normalized = typeof content === 'string' ? content : JSON.stringify(content);
  return sha256(normalized);
}

export default {
  md5,
  sha256,
  sha512,
  hmacSha256,
  contentHash,
};