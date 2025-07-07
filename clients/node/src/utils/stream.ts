export class StreamContextManager {
  private generator: AsyncGenerator<any, void, unknown>;
  private isFinished: boolean = false;

  constructor(generator: AsyncGenerator<any, void, unknown>) {
    this.generator = generator;
  }

  async *[Symbol.asyncIterator]() {
    try {
      while (!this.isFinished) {
        const { value, done } = await this.generator.next();
        if (done) {
          this.isFinished = true;
          break;
        }
        yield value;
      }
    } finally {
      await this.close();
    }
  }

  async close(): Promise<void> {
    if (!this.isFinished) {
      try {
        await this.generator.return(undefined);
      } catch (error) {
        // Ignore errors during cleanup
      }
      this.isFinished = true;
    }
  }

  async collect(): Promise<any[]> {
    const results: any[] = [];
    for await (const item of this) {
      results.push(item);
    }
    return results;
  }

  async first(): Promise<any | null> {
    for await (const item of this) {
      return item;
    }
    return null;
  }

  async last(): Promise<any | null> {
    let lastItem: any = null;
    for await (const item of this) {
      lastItem = item;
    }
    return lastItem;
  }
}

export function createStreamContextManager(generator: AsyncGenerator<any, void, unknown>): StreamContextManager {
  return new StreamContextManager(generator);
}

export async function withStreamContext<T>(
  generator: AsyncGenerator<any, void, unknown>,
  handler: (stream: StreamContextManager) => Promise<T>
): Promise<T> {
  const stream = new StreamContextManager(generator);
  try {
    return await handler(stream);
  } finally {
    await stream.close();
  }
}