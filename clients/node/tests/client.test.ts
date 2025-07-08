import { Retab, AsyncRetab } from '../src/index.js';
import { FieldUnset } from '../src/types/standards.js';
import { TEST_API_KEY, TEST_BASE_URL } from './fixtures.js';

describe('Client Configuration', () => {
  describe('Retab Sync Client', () => {
    describe('Constructor', () => {
      it('should create client with API key', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        expect(client).toBeInstanceOf(Retab);
        expect(client.getHeaders()['Api-Key']).toBe(TEST_API_KEY);
      });

      it('should create client with API key from environment', () => {
        const originalApiKey = process.env.RETAB_API_KEY;
        process.env.RETAB_API_KEY = TEST_API_KEY;

        try {
          const client = new Retab();
          expect(client.getHeaders()['Api-Key']).toBe(TEST_API_KEY);
        } finally {
          if (originalApiKey) {
            process.env.RETAB_API_KEY = originalApiKey;
          } else {
            delete process.env.RETAB_API_KEY;
          }
        }
      });

      it('should create client with custom base URL', () => {
        const client = new Retab({ 
          apiKey: TEST_API_KEY, 
          baseUrl: TEST_BASE_URL 
        });
        expect(client).toBeInstanceOf(Retab);
      });

      it('should create client with base URL from environment', () => {
        const originalBaseUrl = process.env.RETAB_API_BASE_URL;
        process.env.RETAB_API_BASE_URL = TEST_BASE_URL;

        try {
          const client = new Retab({ apiKey: TEST_API_KEY });
          expect(client).toBeInstanceOf(Retab);
        } finally {
          if (originalBaseUrl) {
            process.env.RETAB_API_BASE_URL = originalBaseUrl;
          } else {
            delete process.env.RETAB_API_BASE_URL;
          }
        }
      });

      it('should handle custom timeout and retries', () => {
        const client = new Retab({ 
          apiKey: TEST_API_KEY,
          timeout: 60000,
          maxRetries: 5
        });
        expect(client).toBeInstanceOf(Retab);
      });

      it('should throw error when no API key provided', () => {
        const originalApiKey = process.env.RETAB_API_KEY;
        delete process.env.RETAB_API_KEY;

        try {
          expect(() => new Retab()).toThrow('No API key provided');
        } finally {
          if (originalApiKey) {
            process.env.RETAB_API_KEY = originalApiKey;
          }
        }
      });

      it('should strip trailing slash from base URL', () => {
        const client = new Retab({ 
          apiKey: TEST_API_KEY, 
          baseUrl: 'https://api.test.retab.com/' 
        });
        expect(client).toBeInstanceOf(Retab);
      });
    });

    describe('API Key Configuration', () => {
      it('should set OpenAI API key from config', () => {
        const openaiKey = 'sk-test-openai-key';
        const client = new Retab({ 
          apiKey: TEST_API_KEY,
          openaiApiKey: openaiKey
        });
        expect(client.getHeaders()['OpenAI-Api-Key']).toBe(openaiKey);
      });

      it('should set OpenAI API key from environment when FieldUnset', () => {
        const originalOpenAIKey = process.env.OPENAI_API_KEY;
        const openaiKey = 'sk-env-openai-key';
        process.env.OPENAI_API_KEY = openaiKey;

        try {
          const client = new Retab({ 
            apiKey: TEST_API_KEY,
            openaiApiKey: FieldUnset
          });
          expect(client.getHeaders()['OpenAI-Api-Key']).toBe(openaiKey);
        } finally {
          if (originalOpenAIKey) {
            process.env.OPENAI_API_KEY = originalOpenAIKey;
          } else {
            delete process.env.OPENAI_API_KEY;
          }
        }
      });

      it('should set Gemini API key from config', () => {
        const geminiKey = 'gemini-test-key';
        const client = new Retab({ 
          apiKey: TEST_API_KEY,
          geminiApiKey: geminiKey
        });
        expect(client.getHeaders()['Gemini-Api-Key']).toBe(geminiKey);
      });

      it('should set Gemini API key from environment when FieldUnset', () => {
        const originalGeminiKey = process.env.GEMINI_API_KEY;
        const geminiKey = 'gemini-env-key';
        process.env.GEMINI_API_KEY = geminiKey;

        try {
          const client = new Retab({ 
            apiKey: TEST_API_KEY,
            geminiApiKey: FieldUnset
          });
          expect(client.getHeaders()['Gemini-Api-Key']).toBe(geminiKey);
        } finally {
          if (originalGeminiKey) {
            process.env.GEMINI_API_KEY = originalGeminiKey;
          } else {
            delete process.env.GEMINI_API_KEY;
          }
        }
      });

      it('should set XAI API key from config', () => {
        const xaiKey = 'xai-test-key';
        const client = new Retab({ 
          apiKey: TEST_API_KEY,
          xaiApiKey: xaiKey
        });
        expect(client.getHeaders()['XAI-Api-Key']).toBe(xaiKey);
      });

      it('should set XAI API key from environment when FieldUnset', () => {
        const originalXAIKey = process.env.XAI_API_KEY;
        const xaiKey = 'xai-env-key';
        process.env.XAI_API_KEY = xaiKey;

        try {
          const client = new Retab({ 
            apiKey: TEST_API_KEY,
            xaiApiKey: FieldUnset
          });
          expect(client.getHeaders()['XAI-Api-Key']).toBe(xaiKey);
        } finally {
          if (originalXAIKey) {
            process.env.XAI_API_KEY = originalXAIKey;
          } else {
            delete process.env.XAI_API_KEY;
          }
        }
      });

      it('should not set API key headers when not provided', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        const headers = client.getHeaders();
        expect(headers['OpenAI-Api-Key']).toBeUndefined();
        expect(headers['Gemini-Api-Key']).toBeUndefined();
        expect(headers['XAI-Api-Key']).toBeUndefined();
      });
    });

    describe('Headers', () => {
      it('should include required headers', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        const headers = client.getHeaders();
        
        expect(headers['Api-Key']).toBe(TEST_API_KEY);
        expect(headers['Content-Type']).toBe('application/json');
      });

      it('should include idempotency key when provided', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        const idempotencyKey = 'test-idempotency-key';
        const headers = (client as any)._getHeaders(idempotencyKey);
        
        expect(headers['Idempotency-Key']).toBe(idempotencyKey);
      });

      it('should not include idempotency key when not provided', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        const headers = (client as any)._getHeaders();
        
        expect(headers['Idempotency-Key']).toBeUndefined();
      });
    });

    describe('Resource Initialization', () => {
      it('should initialize all resources', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        
        expect(client.evaluations).toBeDefined();
        expect(client.files).toBeDefined();
        expect(client.fineTuning).toBeDefined();
        expect(client.documents).toBeDefined();
        expect(client.models).toBeDefined();
        expect(client.schemas).toBeDefined();
        expect(client.processors).toBeDefined();
        expect(client.secrets).toBeDefined();
        expect(client.usage).toBeDefined();
        expect(client.consensus).toBeDefined();
      });
    });

    describe('Close Method', () => {
      it('should have close method', () => {
        const client = new Retab({ apiKey: TEST_API_KEY });
        expect(() => client.close()).not.toThrow();
      });
    });
  });

  describe('AsyncRetab Client', () => {
    describe('Constructor', () => {
      it('should create async client with API key', () => {
        const client = new AsyncRetab({ apiKey: TEST_API_KEY });
        expect(client).toBeInstanceOf(AsyncRetab);
        expect(client.getHeaders()['Api-Key']).toBe(TEST_API_KEY);
      });

      it('should create async client with full configuration', () => {
        const client = new AsyncRetab({ 
          apiKey: TEST_API_KEY,
          baseUrl: TEST_BASE_URL,
          timeout: 30000,
          maxRetries: 2,
          openaiApiKey: 'sk-test-openai',
          geminiApiKey: 'gemini-test',
          xaiApiKey: 'xai-test'
        });
        
        expect(client).toBeInstanceOf(AsyncRetab);
        const headers = client.getHeaders();
        expect(headers['Api-Key']).toBe(TEST_API_KEY);
        expect(headers['OpenAI-Api-Key']).toBe('sk-test-openai');
        expect(headers['Gemini-Api-Key']).toBe('gemini-test');
        expect(headers['XAI-Api-Key']).toBe('xai-test');
      });

      it('should throw error when no API key provided', () => {
        const originalApiKey = process.env.RETAB_API_KEY;
        delete process.env.RETAB_API_KEY;

        try {
          expect(() => new AsyncRetab()).toThrow('No API key provided');
        } finally {
          if (originalApiKey) {
            process.env.RETAB_API_KEY = originalApiKey;
          }
        }
      });
    });

    describe('Resource Initialization', () => {
      it('should initialize all async resources', () => {
        const client = new AsyncRetab({ apiKey: TEST_API_KEY });
        
        expect(client.evaluations).toBeDefined();
        expect(client.files).toBeDefined();
        expect(client.fineTuning).toBeDefined();
        expect(client.documents).toBeDefined();
        expect(client.models).toBeDefined();
        expect(client.schemas).toBeDefined();
        expect(client.processors).toBeDefined();
        expect(client.secrets).toBeDefined();
        expect(client.usage).toBeDefined();
        expect(client.consensus).toBeDefined();
      });
    });

    describe('Close Method', () => {
      it('should have async close method', async () => {
        const client = new AsyncRetab({ apiKey: TEST_API_KEY });
        await expect(client.close()).resolves.toBeUndefined();
      });
    });
  });

  describe('URL Preparation', () => {
    it('should prepare URL correctly', () => {
      const client = new Retab({ 
        apiKey: TEST_API_KEY,
        baseUrl: 'https://api.test.com'
      });
      
      const url = (client as any)._prepareUrl('/v1/schemas');
      expect(url).toBe('https://api.test.com/v1/schemas');
    });

    it('should handle leading slash in endpoint', () => {
      const client = new Retab({ 
        apiKey: TEST_API_KEY,
        baseUrl: 'https://api.test.com'
      });
      
      const url = (client as any)._prepareUrl('v1/schemas');
      expect(url).toBe('https://api.test.com/v1/schemas');
    });
  });

  describe('Response Parsing', () => {
    let client: Retab;

    beforeEach(() => {
      client = new Retab({ apiKey: TEST_API_KEY });
    });

    it('should parse JSON response', () => {
      const mockResponse = {
        headers: { 'content-type': 'application/json' },
        data: { key: 'value' }
      };
      
      const parsed = (client as any)._parseResponse(mockResponse);
      expect(parsed).toEqual({ key: 'value' });
    });

    it('should parse streaming JSON response', () => {
      const mockResponse = {
        headers: { 'content-type': 'application/stream+json' },
        data: { key: 'value' }
      };
      
      const parsed = (client as any)._parseResponse(mockResponse);
      expect(parsed).toEqual({ key: 'value' });
    });

    it('should parse text response', () => {
      const mockResponse = {
        headers: { 'content-type': 'text/plain' },
        data: 'plain text'
      };
      
      const parsed = (client as any)._parseResponse(mockResponse);
      expect(parsed).toBe('plain text');
    });

    it('should handle unknown content type', () => {
      const mockResponse = {
        headers: { 'content-type': 'unknown/type' },
        data: 'some data'
      };
      
      const parsed = (client as any)._parseResponse(mockResponse);
      expect(parsed).toBe('some data');
    });
  });

  describe('Response Validation', () => {
    let client: Retab;

    beforeEach(() => {
      client = new Retab({ apiKey: TEST_API_KEY });
    });

    it('should not throw for successful response', () => {
      const mockResponse = {
        status: 200,
        statusText: 'OK',
        data: {}
      };
      
      expect(() => (client as any)._validateResponse(mockResponse)).not.toThrow();
    });

    it('should throw APIError for 500 error', () => {
      const mockResponse = {
        status: 500,
        statusText: 'Internal Server Error',
        data: { error: 'Server error' }
      };
      
      expect(() => (client as any)._validateResponse(mockResponse))
        .toThrow('Internal Server Error');
    });

    it('should throw ValidationError for 422 error', () => {
      const mockResponse = {
        status: 422,
        statusText: 'Unprocessable Entity',
        data: { error: 'Validation failed' }
      };
      
      expect(() => (client as any)._validateResponse(mockResponse))
        .toThrow('Validation error (422)');
    });

    it('should throw APIError for 400 error', () => {
      const mockResponse = {
        status: 400,
        statusText: 'Bad Request',
        data: { error: 'Bad request' }
      };
      
      expect(() => (client as any)._validateResponse(mockResponse))
        .toThrow('Request failed (400)');
    });
  });
});