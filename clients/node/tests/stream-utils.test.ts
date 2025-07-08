import {
  StreamContextManager,
  createStreamContextManager,
  withStreamContext
} from '../src/utils/stream.js';

// Helper function to create test async generators
async function* createTestGenerator(items: any[]): AsyncGenerator<any, void, unknown> {
  for (const item of items) {
    yield item;
  }
}

// Helper function to create an async generator that throws an error
async function* createErrorGenerator(items: any[], errorIndex: number): AsyncGenerator<any, void, unknown> {
  for (let i = 0; i < items.length; i++) {
    if (i === errorIndex) {
      throw new Error(`Test error at index ${i}`);
    }
    yield items[i];
  }
}

// Helper function to create an infinite async generator for testing cleanup
async function* createInfiniteGenerator(): AsyncGenerator<number, void, unknown> {
  let i = 0;
  while (true) {
    yield i++;
    await new Promise(resolve => setTimeout(resolve, 1)); // Small delay
  }
}

describe('Stream Utilities Tests', () => {
  describe('StreamContextManager', () => {
    describe('Constructor and Basic Iteration', () => {
      it('should create StreamContextManager with async generator', () => {
        const generator = createTestGenerator([1, 2, 3]);
        const stream = new StreamContextManager(generator);
        
        expect(stream).toBeInstanceOf(StreamContextManager);
      });

      it('should iterate through simple values', async () => {
        const testData = [1, 2, 3, 4, 5];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        
        expect(results).toEqual(testData);
      });

      it('should iterate through complex objects', async () => {
        const testData = [
          { id: 1, name: 'Alice' },
          { id: 2, name: 'Bob' },
          { id: 3, name: 'Charlie' }
        ];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        
        expect(results).toEqual(testData);
      });

      it('should handle empty generator', async () => {
        const generator = createTestGenerator([]);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        
        expect(results).toEqual([]);
      });

      it('should handle generator with mixed types', async () => {
        const testData = [1, 'string', { key: 'value' }, [1, 2, 3], null, undefined];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        
        expect(results).toEqual(testData);
      });
    });

    describe('collect() method', () => {
      it('should collect all items into an array', async () => {
        const testData = ['a', 'b', 'c', 'd'];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const results = await stream.collect();
        
        expect(results).toEqual(testData);
      });

      it('should collect empty stream into empty array', async () => {
        const generator = createTestGenerator([]);
        const stream = new StreamContextManager(generator);
        
        const results = await stream.collect();
        
        expect(results).toEqual([]);
      });

      it('should collect large datasets efficiently', async () => {
        const largeData = Array.from({ length: 1000 }, (_, i) => i);
        const generator = createTestGenerator(largeData);
        const stream = new StreamContextManager(generator);
        
        const results = await stream.collect();
        
        expect(results).toEqual(largeData);
        expect(results.length).toBe(1000);
      });

      it('should handle nested objects in collection', async () => {
        const nestedData = [
          { items: [1, 2, 3] },
          { items: ['a', 'b', 'c'] },
          { items: [{ nested: true }] }
        ];
        const generator = createTestGenerator(nestedData);
        const stream = new StreamContextManager(generator);
        
        const results = await stream.collect();
        
        expect(results).toEqual(nestedData);
        expect(results[2].items[0].nested).toBe(true);
      });
    });

    describe('first() method', () => {
      it('should return first item from stream', async () => {
        const testData = [10, 20, 30];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const first = await stream.first();
        
        expect(first).toBe(10);
      });

      it('should return null for empty stream', async () => {
        const generator = createTestGenerator([]);
        const stream = new StreamContextManager(generator);
        
        const first = await stream.first();
        
        expect(first).toBeNull();
      });

      it('should return first item even if it is falsy', async () => {
        const testData = [0, 1, 2];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const first = await stream.first();
        
        expect(first).toBe(0);
      });

      it('should handle first item being null or undefined', async () => {
        const nullData = [null, 'second', 'third'];
        const nullGenerator = createTestGenerator(nullData);
        const nullStream = new StreamContextManager(nullGenerator);
        
        const undefinedData = [undefined, 'second', 'third'];
        const undefinedGenerator = createTestGenerator(undefinedData);
        const undefinedStream = new StreamContextManager(undefinedGenerator);
        
        const firstNull = await nullStream.first();
        const firstUndefined = await undefinedStream.first();
        
        expect(firstNull).toBeNull();
        expect(firstUndefined).toBeUndefined();
      });

      it('should stop iteration after finding first item', async () => {
        const generator = createInfiniteGenerator();
        const stream = new StreamContextManager(generator);
        
        const first = await stream.first();
        
        expect(first).toBe(0);
        // Ensure the stream is properly closed after getting first item
        await stream.close();
      });
    });

    describe('last() method', () => {
      it('should return last item from stream', async () => {
        const testData = [10, 20, 30];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const last = await stream.last();
        
        expect(last).toBe(30);
      });

      it('should return null for empty stream', async () => {
        const generator = createTestGenerator([]);
        const stream = new StreamContextManager(generator);
        
        const last = await stream.last();
        
        expect(last).toBeNull();
      });

      it('should return single item for single-item stream', async () => {
        const testData = ['only-item'];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const last = await stream.last();
        
        expect(last).toBe('only-item');
      });

      it('should handle last item being falsy', async () => {
        const testData = [1, 2, 0];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const last = await stream.last();
        
        expect(last).toBe(0);
      });

      it('should iterate through entire stream to find last item', async () => {
        const largeData = Array.from({ length: 100 }, (_, i) => i);
        const generator = createTestGenerator(largeData);
        const stream = new StreamContextManager(generator);
        
        const last = await stream.last();
        
        expect(last).toBe(99);
      });
    });

    describe('close() method', () => {
      it('should close stream properly', async () => {
        const generator = createTestGenerator([1, 2, 3]);
        const stream = new StreamContextManager(generator);
        
        await stream.close();
        
        // After closing, iterating should yield no items
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        expect(results).toEqual([]);
      });

      it('should be idempotent (safe to call multiple times)', async () => {
        const generator = createTestGenerator([1, 2, 3]);
        const stream = new StreamContextManager(generator);
        
        await stream.close();
        await stream.close();
        await stream.close();
        
        // Should not throw or cause issues
        expect(true).toBe(true);
      });

      it('should handle close during iteration gracefully', async () => {
        const generator = createTestGenerator([1, 2, 3, 4, 5]);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
          if (item === 2) {
            await stream.close();
            break;
          }
        }
        
        expect(results).toEqual([1, 2]);
      });

      it('should handle generator return errors gracefully', async () => {
        // Create a mock generator that throws on return
        const mockGenerator = {
          next: jest.fn()
            .mockResolvedValueOnce({ value: 1, done: false })
            .mockResolvedValueOnce({ value: undefined, done: true }),
          return: jest.fn().mockRejectedValue(new Error('Return error')),
          throw: jest.fn(),
          [Symbol.asyncIterator]: function() { return this; }
        };
        
        const stream = new StreamContextManager(mockGenerator as any);
        
        // Should not throw even if generator.return() throws
        await expect(stream.close()).resolves.not.toThrow();
      });
    });

    describe('Error Handling', () => {
      it('should handle generator errors during iteration', async () => {
        const generator = createErrorGenerator([1, 2, 3], 1);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        
        await expect(async () => {
          for await (const item of stream) {
            results.push(item);
          }
        }).rejects.toThrow('Test error at index 1');
        
        expect(results).toEqual([1]);
      });

      it('should handle errors in collect method', async () => {
        const generator = createErrorGenerator([1, 2, 3], 2);
        const stream = new StreamContextManager(generator);
        
        await expect(stream.collect()).rejects.toThrow('Test error at index 2');
      });

      it('should handle errors in first method', async () => {
        const generator = createErrorGenerator([1, 2, 3], 0);
        const stream = new StreamContextManager(generator);
        
        await expect(stream.first()).rejects.toThrow('Test error at index 0');
      });

      it('should handle errors in last method', async () => {
        const generator = createErrorGenerator([1, 2, 3], 1);
        const stream = new StreamContextManager(generator);
        
        await expect(stream.last()).rejects.toThrow('Test error at index 1');
      });

      it('should ensure cleanup happens even with errors', async () => {
        const generator = createErrorGenerator([1, 2, 3], 1);
        const stream = new StreamContextManager(generator);
        
        try {
          await stream.collect();
        } catch (error) {
          // Expected error
        }
        
        // Stream should be marked as finished after error
        const moreResults: any[] = [];
        for await (const item of stream) {
          moreResults.push(item);
        }
        expect(moreResults).toEqual([]);
      });
    });

    describe('Symbol.asyncIterator Implementation', () => {
      it('should be properly iterable with for-await-of', async () => {
        const testData = ['async', 'iterator', 'test'];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        
        expect(results).toEqual(testData);
      });

      it('should handle multiple concurrent iterations correctly', async () => {
        const testData = [1, 2, 3];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        // First iteration should consume the stream
        const results1: any[] = [];
        for await (const item of stream) {
          results1.push(item);
        }
        
        // Second iteration should get no items (stream is consumed)
        const results2: any[] = [];
        for await (const item of stream) {
          results2.push(item);
        }
        
        expect(results1).toEqual(testData);
        expect(results2).toEqual([]);
      });

      it('should cleanup automatically at end of iteration', async () => {
        const testData = [1, 2, 3];
        const generator = createTestGenerator(testData);
        const stream = new StreamContextManager(generator);
        
        // Complete iteration
        for await (const _item of stream) {
          // Just iterate
        }
        
        // Stream should be finished and closed
        const results: any[] = [];
        for await (const item of stream) {
          results.push(item);
        }
        expect(results).toEqual([]);
      });
    });
  });

  describe('createStreamContextManager', () => {
    it('should create StreamContextManager instance', () => {
      const generator = createTestGenerator([1, 2, 3]);
      const stream = createStreamContextManager(generator);
      
      expect(stream).toBeInstanceOf(StreamContextManager);
    });

    it('should create functional stream manager', async () => {
      const testData = ['factory', 'test'];
      const generator = createTestGenerator(testData);
      const stream = createStreamContextManager(generator);
      
      const results = await stream.collect();
      
      expect(results).toEqual(testData);
    });

    it('should create independent stream managers', async () => {
      const data1 = [1, 2];
      const data2 = [3, 4];
      
      const stream1 = createStreamContextManager(createTestGenerator(data1));
      const stream2 = createStreamContextManager(createTestGenerator(data2));
      
      const results1 = await stream1.collect();
      const results2 = await stream2.collect();
      
      expect(results1).toEqual(data1);
      expect(results2).toEqual(data2);
    });
  });

  describe('withStreamContext', () => {
    it('should execute handler with stream context', async () => {
      const testData = [10, 20, 30];
      const generator = createTestGenerator(testData);
      
      const result = await withStreamContext(generator, async (stream) => {
        return await stream.collect();
      });
      
      expect(result).toEqual(testData);
    });

    it('should automatically close stream after handler execution', async () => {
      const testData = [1, 2, 3];
      const generator = createTestGenerator(testData);
      
      let streamRef: StreamContextManager;
      
      await withStreamContext(generator, async (stream) => {
        streamRef = stream;
        return await stream.first();
      });
      
      // Try to use the stream after withStreamContext completes
      const results: any[] = [];
      for await (const item of streamRef!) {
        results.push(item);
      }
      
      expect(results).toEqual([]); // Should be empty because stream was closed
    });

    it('should close stream even if handler throws error', async () => {
      const testData = [1, 2, 3];
      const generator = createTestGenerator(testData);
      
      let streamRef: StreamContextManager;
      
      await expect(
        withStreamContext(generator, async (stream) => {
          streamRef = stream;
          throw new Error('Handler error');
        })
      ).rejects.toThrow('Handler error');
      
      // Stream should still be closed after error
      const results: any[] = [];
      for await (const item of streamRef!) {
        results.push(item);
      }
      
      expect(results).toEqual([]);
    });

    it('should return handler result', async () => {
      const testData = ['a', 'b', 'c'];
      const generator = createTestGenerator(testData);
      
      const result = await withStreamContext(generator, async (stream) => {
        const first = await stream.first();
        return `First item: ${first}`;
      });
      
      expect(result).toBe('First item: a');
    });

    it('should handle complex handler operations', async () => {
      const testData = Array.from({ length: 10 }, (_, i) => i);
      const generator = createTestGenerator(testData);
      
      const result = await withStreamContext(generator, async (stream) => {
        const all = await stream.collect();
        const sum = all.reduce((acc, val) => acc + val, 0);
        const avg = sum / all.length;
        return { count: all.length, sum, average: avg };
      });
      
      expect(result.count).toBe(10);
      expect(result.sum).toBe(45);
      expect(result.average).toBe(4.5);
    });

    it('should handle async handler with delays', async () => {
      const testData = [1, 2, 3];
      const generator = createTestGenerator(testData);
      
      const startTime = Date.now();
      
      const result = await withStreamContext(generator, async (stream) => {
        await new Promise(resolve => setTimeout(resolve, 10));
        return await stream.collect();
      });
      
      const elapsed = Date.now() - startTime;
      
      expect(result).toEqual(testData);
      expect(elapsed).toBeGreaterThanOrEqual(10);
    });
  });

  describe('Integration Tests', () => {
    it('should work together for complete streaming workflow', async () => {
      const data = [
        { id: 1, value: 'first' },
        { id: 2, value: 'second' },
        { id: 3, value: 'third' },
        { id: 4, value: 'fourth' },
        { id: 5, value: 'fifth' }
      ];
      
      const generator = createTestGenerator(data);
      
      const results = await withStreamContext(generator, async (stream) => {
        const first = await stream.first();
        
        // Create a new stream for getting last (since first consumes the stream)
        const newGenerator = createTestGenerator(data);
        const newStream = createStreamContextManager(newGenerator);
        const last = await newStream.last();
        
        // Create another stream for collecting all
        const allGenerator = createTestGenerator(data);
        const allStream = createStreamContextManager(allGenerator);
        const all = await allStream.collect();
        
        return {
          first,
          last,
          all,
          count: all.length
        };
      });
      
      expect(results.first).toEqual(data[0]);
      expect(results.last).toEqual(data[4]);
      expect(results.all).toEqual(data);
      expect(results.count).toBe(5);
    });

    it('should handle streaming of different data types', async () => {
      async function* mixedGenerator() {
        yield 42;
        yield 'string';
        yield { object: true };
        yield [1, 2, 3];
        yield null;
        yield undefined;
        yield new Date('2024-01-01');
      }
      
      const stream = createStreamContextManager(mixedGenerator());
      const results = await stream.collect();
      
      expect(results).toHaveLength(7);
      expect(typeof results[0]).toBe('number');
      expect(typeof results[1]).toBe('string');
      expect(typeof results[2]).toBe('object');
      expect(Array.isArray(results[3])).toBe(true);
      expect(results[4]).toBeNull();
      expect(results[5]).toBeUndefined();
      expect(results[6]).toBeInstanceOf(Date);
    });
  });

  describe('Performance and Edge Cases', () => {
    it('should handle very large streams efficiently', async () => {
      const largeSize = 10000;
      async function* largeGenerator() {
        for (let i = 0; i < largeSize; i++) {
          yield i;
        }
      }
      
      const startTime = Date.now();
      const stream = createStreamContextManager(largeGenerator());
      const results = await stream.collect();
      const elapsed = Date.now() - startTime;
      
      expect(results).toHaveLength(largeSize);
      expect(results[0]).toBe(0);
      expect(results[largeSize - 1]).toBe(largeSize - 1);
      expect(elapsed).toBeLessThan(1000); // Should complete within 1 second
    });

    it('should handle streams with async operations', async () => {
      async function* asyncGenerator() {
        for (let i = 0; i < 5; i++) {
          await new Promise(resolve => setTimeout(resolve, 1));
          yield `async-${i}`;
        }
      }
      
      const stream = createStreamContextManager(asyncGenerator());
      const results = await stream.collect();
      
      expect(results).toEqual(['async-0', 'async-1', 'async-2', 'async-3', 'async-4']);
    });

    it('should handle empty async generator', async () => {
      async function* emptyGenerator() {
        return;
      }
      
      const stream = createStreamContextManager(emptyGenerator());
      
      const first = await stream.first();
      const last = await stream.last();
      const all = await stream.collect();
      
      expect(first).toBeNull();
      expect(last).toBeNull();
      expect(all).toEqual([]);
    });

    it('should handle generator with complex object types', async () => {
      async function* complexGenerator() {
        yield { type: 'object', value: 1 };
        yield new Map([['key', 'value']]);
        yield new Set([1, 2, 3]);
        yield /regex/g;
        yield new Date('2024-01-01');
      }
      
      const stream = createStreamContextManager(complexGenerator());
      const results = await stream.collect();
      
      expect(results).toHaveLength(5);
      expect(results[0]).toEqual({ type: 'object', value: 1 });
      expect(results[1]).toBeInstanceOf(Map);
      expect(results[2]).toBeInstanceOf(Set);
      expect(results[3]).toBeInstanceOf(RegExp);
      expect(results[4]).toBeInstanceOf(Date);
    });
  });
});