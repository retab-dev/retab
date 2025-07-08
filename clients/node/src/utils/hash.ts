import * as blake2 from 'blake2';

/**
 * Generate a BLAKE2b hash from bytes.
 * Uses the blake2 package to match Python's hashlib.blake2b with 8-byte digest.
 */
export function generateBlake2bHashFromBytes(bytes: Buffer): string {
  const hash = blake2.createHash('blake2b', { digestLength: 8 });
  hash.update(bytes);
  return hash.digest('hex');
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