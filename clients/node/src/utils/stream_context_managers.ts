/**
 * Stream context management utilities for proper resource cleanup
 * Equivalent to Python's stream_context_managers.py
 */

export interface AsyncDisposable {
  [Symbol.asyncDispose](): Promise<void>;
}

export interface Disposable {
  [Symbol.dispose](): void;
}

/**
 * AsyncGenerator context manager for proper cleanup
 */
export class AsyncGeneratorContextManager<T> implements AsyncDisposable {
  private generator: AsyncGenerator<T, void, unknown>;
  private isFinalized: boolean = false;

  constructor(generator: AsyncGenerator<T, void, unknown>) {
    this.generator = generator;
  }

  async *[Symbol.asyncIterator](): AsyncGenerator<T, void, unknown> {
    try {
      yield* this.generator;
    } finally {
      await this.cleanup();
    }
  }

  async next(): Promise<IteratorResult<T, void>> {
    try {
      return await this.generator.next();
    } catch (error) {
      await this.cleanup();
      throw error;
    }
  }

  async return(value?: void): Promise<IteratorResult<T, void>> {
    try {
      return await this.generator.return(value);
    } finally {
      await this.cleanup();
    }
  }

  async throw(error?: any): Promise<IteratorResult<T, void>> {
    try {
      return await this.generator.throw(error);
    } finally {
      await this.cleanup();
    }
  }

  private async cleanup(): Promise<void> {
    if (!this.isFinalized) {
      this.isFinalized = true;
      try {
        // Ensure generator is properly closed
        if (typeof this.generator.return === 'function') {
          await this.generator.return();
        }
      } catch (error) {
        // Ignore cleanup errors
      }
    }
  }

  async [Symbol.asyncDispose](): Promise<void> {
    await this.cleanup();
  }
}

/**
 * Synchronous generator context manager
 */
export class GeneratorContextManager<T> implements Disposable {
  private generator: Generator<T, void, unknown>;
  private isFinalized: boolean = false;

  constructor(generator: Generator<T, void, unknown>) {
    this.generator = generator;
  }

  *[Symbol.iterator](): Generator<T, void, unknown> {
    try {
      yield* this.generator;
    } finally {
      this.cleanup();
    }
  }

  next(): IteratorResult<T, void> {
    try {
      return this.generator.next();
    } catch (error) {
      this.cleanup();
      throw error;
    }
  }

  return(value?: void): IteratorResult<T, void> {
    try {
      return this.generator.return(value);
    } finally {
      this.cleanup();
    }
  }

  throw(error?: any): IteratorResult<T, void> {
    try {
      return this.generator.throw(error);
    } finally {
      this.cleanup();
    }
  }

  private cleanup(): void {
    if (!this.isFinalized) {
      this.isFinalized = true;
      try {
        // Ensure generator is properly closed
        if (typeof this.generator.return === 'function') {
          this.generator.return();
        }
      } catch (error) {
        // Ignore cleanup errors
      }
    }
  }

  [Symbol.dispose](): void {
    this.cleanup();
  }
}

/**
 * Decorator to wrap async generator functions with context management
 */
export function asAsyncContextManager<T extends any[], R>(
  generatorFunction: (...args: T) => AsyncGenerator<R, void, unknown>
) {
  return (...args: T): AsyncGeneratorContextManager<R> => {
    const generator = generatorFunction(...args);
    return new AsyncGeneratorContextManager(generator);
  };
}

/**
 * Decorator to wrap sync generator functions with context management
 */
export function asContextManager<T extends any[], R>(
  generatorFunction: (...args: T) => Generator<R, void, unknown>
) {
  return (...args: T): GeneratorContextManager<R> => {
    const generator = generatorFunction(...args);
    return new GeneratorContextManager(generator);
  };
}

/**
 * Stream wrapper that provides automatic cleanup and error handling
 */
export class StreamWrapper<T> implements AsyncDisposable {
  protected stream: AsyncGenerator<T, void, unknown>;
  protected isActive: boolean = true;
  private cleanup: (() => Promise<void>) | null = null;

  constructor(
    stream: AsyncGenerator<T, void, unknown>,
    cleanup?: () => Promise<void>
  ) {
    this.stream = stream;
    this.cleanup = cleanup || null;
  }

  async *[Symbol.asyncIterator](): AsyncGenerator<T, void, unknown> {
    try {
      while (this.isActive) {
        const result = await this.stream.next();
        if (result.done) {
          break;
        }
        yield result.value;
      }
    } finally {
      await this.close();
    }
  }

  async close(): Promise<void> {
    if (this.isActive) {
      this.isActive = false;
      
      try {
        // Close the stream
        if (typeof this.stream.return === 'function') {
          await this.stream.return();
        }
      } catch (error) {
        // Ignore stream closure errors
      }

      // Run custom cleanup
      if (this.cleanup) {
        try {
          await this.cleanup();
        } catch (error) {
          console.warn('Stream cleanup failed:', error);
        }
      }
    }
  }

  async [Symbol.asyncDispose](): Promise<void> {
    await this.close();
  }
}

/**
 * Timeout wrapper for streams
 */
export class StreamTimeoutWrapper<T> extends StreamWrapper<T> {
  private timeout: number;
  private lastActivity: number;
  private timeoutId: NodeJS.Timeout | null = null;

  constructor(
    stream: AsyncGenerator<T, void, unknown>,
    timeoutMs: number,
    cleanup?: () => Promise<void>
  ) {
    super(stream, cleanup);
    this.timeout = timeoutMs;
    this.lastActivity = Date.now();
  }

  async *[Symbol.asyncIterator](): AsyncGenerator<T, void, unknown> {
    try {
      this.startTimeout();
      
      while (this.isActive) {
        const result = await this.stream.next();
        this.updateActivity();
        
        if (result.done) {
          break;
        }
        
        yield result.value;
      }
    } finally {
      this.clearTimeout();
      await this.close();
    }
  }

  private startTimeout(): void {
    this.timeoutId = setTimeout(() => {
      if (Date.now() - this.lastActivity > this.timeout) {
        this.close().catch(console.error);
      }
    }, this.timeout);
  }

  private clearTimeout(): void {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
      this.timeoutId = null;
    }
  }

  private updateActivity(): void {
    this.lastActivity = Date.now();
  }

  async close(): Promise<void> {
    this.clearTimeout();
    await super.close();
  }
}

/**
 * Batch processing wrapper for streams
 */
export class StreamBatchProcessor<T> implements AsyncDisposable {
  private stream: AsyncGenerator<T, void, unknown>;
  private batchSize: number;
  private currentBatch: T[] = [];
  private isActive: boolean = true;
  private cleanup: (() => Promise<void>) | null = null;

  constructor(
    stream: AsyncGenerator<T, void, unknown>,
    batchSize: number,
    cleanup?: () => Promise<void>
  ) {
    this.stream = stream;
    this.batchSize = batchSize;
    this.cleanup = cleanup || null;
  }

  async *[Symbol.asyncIterator](): AsyncGenerator<T[], void, unknown> {
    try {
      while (this.isActive) {
        const result = await this.stream.next();
        
        if (result.done) {
          // Yield remaining items in batch
          if (this.currentBatch.length > 0) {
            yield [...this.currentBatch];
            this.currentBatch = [];
          }
          break;
        }
        
        this.currentBatch.push(result.value);
        
        if (this.currentBatch.length >= this.batchSize) {
          yield [...this.currentBatch];
          this.currentBatch = [];
        }
      }
    } finally {
      await this.close();
    }
  }

  async close(): Promise<void> {
    if (this.isActive) {
      this.isActive = false;
      
      try {
        if (typeof this.stream.return === 'function') {
          await this.stream.return();
        }
      } catch (error) {
        // Ignore stream closure errors
      }

      if (this.cleanup) {
        try {
          await this.cleanup();
        } catch (error) {
          console.warn('Stream cleanup failed:', error);
        }
      }
    }
  }

  async [Symbol.asyncDispose](): Promise<void> {
    await this.close();
  }
}

/**
 * Error recovery wrapper for streams
 */
export class StreamErrorRecovery<T> extends StreamWrapper<T> {
  private maxRetries: number;
  private retryCount: number = 0;
  private errorHandler: ((error: Error) => boolean) | null = null;

  constructor(
    stream: AsyncGenerator<T, void, unknown>,
    maxRetries: number = 3,
    errorHandler?: (error: Error) => boolean,
    cleanup?: () => Promise<void>
  ) {
    super(stream, cleanup);
    this.maxRetries = maxRetries;
    this.errorHandler = errorHandler || null;
  }

  async *[Symbol.asyncIterator](): AsyncGenerator<T, void, unknown> {
    try {
      while (this.isActive && this.retryCount <= this.maxRetries) {
        try {
          const result = await this.stream.next();
          
          if (result.done) {
            break;
          }
          
          // Reset retry count on successful read
          this.retryCount = 0;
          yield result.value;
          
        } catch (error) {
          this.retryCount++;
          
          const shouldRetry = this.errorHandler ? 
            this.errorHandler(error as Error) : 
            this.retryCount <= this.maxRetries;
          
          if (!shouldRetry) {
            throw error;
          }
          
          // Wait before retry
          await new Promise(resolve => setTimeout(resolve, Math.pow(2, this.retryCount) * 1000));
        }
      }
    } finally {
      await this.close();
    }
  }
}

/**
 * Utility functions for creating context-managed streams
 */
export const streamUtils = {
  /**
   * Wrap an async generator with automatic cleanup
   */
  wrap: <T>(
    stream: AsyncGenerator<T, void, unknown>,
    cleanup?: () => Promise<void>
  ): StreamWrapper<T> => {
    return new StreamWrapper(stream, cleanup);
  },

  /**
   * Add timeout handling to a stream
   */
  withTimeout: <T>(
    stream: AsyncGenerator<T, void, unknown>,
    timeoutMs: number,
    cleanup?: () => Promise<void>
  ): StreamTimeoutWrapper<T> => {
    return new StreamTimeoutWrapper(stream, timeoutMs, cleanup);
  },

  /**
   * Add batch processing to a stream
   */
  withBatching: <T>(
    stream: AsyncGenerator<T, void, unknown>,
    batchSize: number,
    cleanup?: () => Promise<void>
  ): StreamBatchProcessor<T> => {
    return new StreamBatchProcessor(stream, batchSize, cleanup);
  },

  /**
   * Add error recovery to a stream
   */
  withErrorRecovery: <T>(
    stream: AsyncGenerator<T, void, unknown>,
    maxRetries: number = 3,
    errorHandler?: (error: Error) => boolean,
    cleanup?: () => Promise<void>
  ): StreamErrorRecovery<T> => {
    return new StreamErrorRecovery(stream, maxRetries, errorHandler, cleanup);
  },
};

export default {
  AsyncGeneratorContextManager,
  GeneratorContextManager,
  StreamWrapper,
  StreamTimeoutWrapper,
  StreamBatchProcessor,
  StreamErrorRecovery,
  asAsyncContextManager,
  asContextManager,
  streamUtils,
};