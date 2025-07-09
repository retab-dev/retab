import fs from 'fs';
import { createReadStream, createWriteStream } from 'fs';
import { createInterface } from 'readline';
import { pipeline } from 'stream/promises';
import { Transform } from 'stream';

/**
 * JSONL (JSON Lines) utilities for handling line-delimited JSON data
 * Equivalent to Python's jsonlutils.py
 */

export interface JSONLOptions {
  encoding?: BufferEncoding;
  highWaterMark?: number;
}

/**
 * Read JSONL file and return array of parsed objects
 */
export async function readJSONL(filePath: string, options: JSONLOptions = {}): Promise<any[]> {
  const { encoding = 'utf8' } = options;
  const results: any[] = [];
  
  const fileStream = createReadStream(filePath, { encoding });
  const rl = createInterface({
    input: fileStream,
    crlfDelay: Infinity, // Handle Windows line endings
  });

  for await (const line of rl) {
    const trimmedLine = line.trim();
    if (trimmedLine) {
      try {
        results.push(JSON.parse(trimmedLine));
      } catch (error) {
        throw new Error(`Invalid JSON on line ${results.length + 1}: ${error instanceof Error ? error.message : String(error)}`);
      }
    }
  }

  return results;
}

/**
 * Write array of objects to JSONL file
 */
export async function writeJSONL(filePath: string, data: any[], options: JSONLOptions = {}): Promise<void> {
  const { encoding = 'utf8' } = options;
  
  const writeStream = createWriteStream(filePath, { encoding });
  
  for (const item of data) {
    const jsonLine = JSON.stringify(item) + '\n';
    writeStream.write(jsonLine);
  }
  
  writeStream.end();
  
  return new Promise((resolve, reject) => {
    writeStream.on('finish', resolve);
    writeStream.on('error', reject);
  });
}

/**
 * Append single object to JSONL file
 */
export async function appendJSONL(filePath: string, data: any, options: JSONLOptions = {}): Promise<void> {
  const { encoding = 'utf8' } = options;
  
  const jsonLine = JSON.stringify(data) + '\n';
  
  return new Promise((resolve, reject) => {
    fs.appendFile(filePath, jsonLine, { encoding }, (error) => {
      if (error) {
        reject(error);
      } else {
        resolve();
      }
    });
  });
}

/**
 * Stream JSONL file processing with transformation
 */
export async function streamJSONL<T, R>(
  inputPath: string,
  outputPath: string,
  transform: (item: T) => R | Promise<R>,
  options: JSONLOptions = {}
): Promise<void> {
  const { encoding = 'utf8' } = options;
  
  const inputStream = createReadStream(inputPath, { encoding });
  const outputStream = createWriteStream(outputPath, { encoding });
  
  const transformStream = new Transform({
    objectMode: true,
    async transform(chunk: Buffer, _encoding, callback) {
      try {
        const line = chunk.toString().trim();
        if (line) {
          const parsed = JSON.parse(line);
          const transformed = await transform(parsed);
          const jsonLine = JSON.stringify(transformed) + '\n';
          callback(null, jsonLine);
        } else {
          callback();
        }
      } catch (error: unknown) {
        callback(error instanceof Error ? error : new Error(String(error)));
      }
    },
  });
  
  const rl = createInterface({
    input: inputStream,
    crlfDelay: Infinity,
  });
  
  for await (const line of rl) {
    transformStream.write(line);
  }
  
  transformStream.end();
  
  await pipeline(transformStream, outputStream);
}

/**
 * Validate JSONL file format
 */
export async function validateJSONL(filePath: string): Promise<{ valid: boolean; errors: string[] }> {
  const errors: string[] = [];
  let lineNumber = 0;
  
  try {
    const fileStream = createReadStream(filePath, { encoding: 'utf8' });
    const rl = createInterface({
      input: fileStream,
      crlfDelay: Infinity,
    });

    for await (const line of rl) {
      lineNumber++;
      const trimmedLine = line.trim();
      
      if (trimmedLine) {
        try {
          JSON.parse(trimmedLine);
        } catch (error: unknown) {
          errors.push(`Line ${lineNumber}: Invalid JSON - ${error instanceof Error ? error.message : String(error)}`);
        }
      }
    }
  } catch (error: unknown) {
    errors.push(`File reading error: ${error instanceof Error ? error.message : String(error)}`);
  }
  
  return {
    valid: errors.length === 0,
    errors,
  };
}

/**
 * Count lines in JSONL file
 */
export async function countJSONL(filePath: string): Promise<number> {
  let count = 0;
  
  const fileStream = createReadStream(filePath, { encoding: 'utf8' });
  const rl = createInterface({
    input: fileStream,
    crlfDelay: Infinity,
  });

  for await (const line of rl) {
    if (line.trim()) {
      count++;
    }
  }
  
  return count;
}

/**
 * Split JSONL file into chunks
 */
export async function splitJSONL(
  inputPath: string,
  outputDir: string,
  chunkSize: number,
  options: JSONLOptions = {}
): Promise<string[]> {
  const { encoding = 'utf8' } = options;
  const outputFiles: string[] = [];
  let currentChunk = 0;
  let currentCount = 0;
  let currentStream: fs.WriteStream | null = null;
  
  const getOutputPath = (chunk: number): string => {
    const outputPath = `${outputDir}/chunk_${chunk.toString().padStart(3, '0')}.jsonl`;
    outputFiles.push(outputPath);
    return outputPath;
  };
  
  const fileStream = createReadStream(inputPath, { encoding });
  const rl = createInterface({
    input: fileStream,
    crlfDelay: Infinity,
  });

  for await (const line of rl) {
    const trimmedLine = line.trim();
    if (trimmedLine) {
      if (currentCount === 0) {
        if (currentStream) {
          currentStream.end();
        }
        currentStream = createWriteStream(getOutputPath(currentChunk), { encoding });
        currentChunk++;
      }
      
      currentStream!.write(trimmedLine + '\n');
      currentCount++;
      
      if (currentCount >= chunkSize) {
        currentCount = 0;
      }
    }
  }
  
  if (currentStream) {
    currentStream.end();
  }
  
  return outputFiles;
}

/**
 * Merge multiple JSONL files
 */
export async function mergeJSONL(
  inputPaths: string[],
  outputPath: string,
  options: JSONLOptions = {}
): Promise<void> {
  const { encoding = 'utf8' } = options;
  
  const outputStream = createWriteStream(outputPath, { encoding });
  
  for (const inputPath of inputPaths) {
    const fileStream = createReadStream(inputPath, { encoding });
    const rl = createInterface({
      input: fileStream,
      crlfDelay: Infinity,
    });

    for await (const line of rl) {
      const trimmedLine = line.trim();
      if (trimmedLine) {
        outputStream.write(trimmedLine + '\n');
      }
    }
  }
  
  outputStream.end();
  
  return new Promise((resolve, reject) => {
    outputStream.on('finish', resolve);
    outputStream.on('error', reject);
  });
}

/**
 * Filter JSONL file based on predicate
 */
export async function filterJSONL<T>(
  inputPath: string,
  outputPath: string,
  predicate: (item: T) => boolean | Promise<boolean>,
  options: JSONLOptions = {}
): Promise<number> {
  const { encoding = 'utf8' } = options;
  let filteredCount = 0;
  
  const outputStream = createWriteStream(outputPath, { encoding });
  const fileStream = createReadStream(inputPath, { encoding });
  const rl = createInterface({
    input: fileStream,
    crlfDelay: Infinity,
  });

  for await (const line of rl) {
    const trimmedLine = line.trim();
    if (trimmedLine) {
      try {
        const parsed = JSON.parse(trimmedLine);
        if (await predicate(parsed)) {
          outputStream.write(trimmedLine + '\n');
          filteredCount++;
        }
      } catch (error) {
        // Skip invalid JSON lines
      }
    }
  }
  
  outputStream.end();
  
  return new Promise((resolve, reject) => {
    outputStream.on('finish', () => resolve(filteredCount));
    outputStream.on('error', reject);
  });
}

export default {
  readJSONL,
  writeJSONL,
  appendJSONL,
  streamJSONL,
  validateJSONL,
  countJSONL,
  splitJSONL,
  mergeJSONL,
  filterJSONL,
};