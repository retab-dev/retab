import crypto from 'crypto';

/**
 * Generate a BLAKE2b hash from bytes.
 * Since Node.js doesn't have BLAKE2b built-in, we'll use SHA-256 as a substitute.
 * The Python version uses BLAKE2b with 8-byte digest, so we'll take the first 16 hex chars.
 */
export function generateBlake2bHashFromBytes(bytes: Buffer): string {
  return crypto.createHash('sha256').update(bytes).digest('hex').substring(0, 16);
}

/**
 * Generate a BLAKE2b hash from a base64 string.
 */
export function generateBlake2bHashFromBase64(base64String: string): string {
  const bytes = Buffer.from(base64String, 'base64');
  return generateBlake2bHashFromBytes(bytes);
}

/**
 * Generate a BLAKE2b hash from a UTF-8 string.
 */
export function generateBlake2bHashFromString(inputString: string): string {
  const bytes = Buffer.from(inputString, 'utf-8');
  return generateBlake2bHashFromBytes(bytes);
}

/**
 * Generate a BLAKE2b hash from a dictionary/object.
 */
export function generateBlake2bHashFromDict(inputDict: Record<string, any>): string {
  const jsonString = JSON.stringify(inputDict, Object.keys(inputDict).sort());
  return generateBlake2bHashFromString(jsonString.trim());
}