/**
 * Comprehensive parity tests comparing Python and Node.js SDK behavior
 * 
 * These tests verify that both SDKs behave identically by:
 * 1. Running the same operations on both SDKs
 * 2. Comparing outputs, error messages, and behavior
 * 3. Validating API request structures match
 */

import { spawn } from 'child_process';
import { join } from 'path';
import { Retab, AsyncRetab } from '../src/index.js';
import { TEST_API_KEY } from './fixtures.js';

// Helper function to run Python code and get results
async function runPythonCode(code: string): Promise<any> {
  return new Promise((resolve, reject) => {
    const pythonPath = join(process.cwd(), '../python');
    const pythonProcess = spawn('python3', ['-c', code], {
      cwd: pythonPath,
      env: { ...process.env, PYTHONPATH: pythonPath }
    });
    
    let stdout = '';
    let stderr = '';
    
    pythonProcess.stdout.on('data', (data: any) => {
      stdout += data.toString();
    });
    
    pythonProcess.stderr.on('data', (data: any) => {
      stderr += data.toString();
    });
    
    pythonProcess.on('close', (code: any) => {
      if (code === 0) {
        try {
          // Try to parse JSON output
          const result = JSON.parse(stdout.trim());
          resolve(result);
        } catch {
          // If not JSON, return as text
          resolve(stdout.trim());
        }
      } else {
        reject(new Error(`Python process failed with code ${code}: ${stderr}`));
      }
    });
  });
}

// Helper function to create comparable test data
const createTestSchema = () => ({
  type: 'object',
  properties: {
    name: { type: 'string' },
    age: { type: 'number' },
    email: { type: 'string', format: 'email' }
  },
  required: ['name', 'age']
});

describe('Python vs Node.js SDK Parity Tests', () => {
  let nodeClient: Retab;
  let nodeAsyncClient: AsyncRetab;

  beforeAll(async () => {
    nodeClient = new Retab({ 
      apiKey: TEST_API_KEY,
      openaiApiKey: 'test-openai-key'
    });
    nodeAsyncClient = new AsyncRetab({ 
      apiKey: TEST_API_KEY,
      openaiApiKey: 'test-openai-key'
    });
  });

  describe('Client Configuration Parity', () => {
    it('should have identical client initialization behavior', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}', openai_api_key='test-openai-key')
    result = {
        'success': True,
        'api_key': client.api_key,
        'base_url': client.base_url,
        'timeout': client.timeout,
        'max_retries': client.max_retries,
        'has_openai_key': 'OpenAI-Api-Key' in client.headers
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      const nodeResult = {
        success: true,
        api_key: (nodeClient as any).apiKey,
        base_url: (nodeClient as any).baseUrl,
        timeout: (nodeClient as any).timeout,
        max_retries: (nodeClient as any).maxRetries,
        has_openai_key: !!nodeClient.getHeaders()['OpenAI-Api-Key']
      };

      expect(pythonResult.success).toBe(nodeResult.success);
      expect(pythonResult.api_key).toBe(nodeResult.api_key);
      expect(pythonResult.base_url).toBe(nodeResult.base_url);
      expect(pythonResult.has_openai_key).toBe(nodeResult.has_openai_key);
    });

    it('should handle missing API key identically', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab()  # No API key provided
    print(json.dumps({'success': True}))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);

      let nodeResult;
      try {
        new Retab();
        nodeResult = { success: true };
      } catch (error: any) {
        nodeResult = { success: false, error: error.message };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (!pythonResult.success && !nodeResult.success) {
        // Both should fail with similar error messages about missing API key
        expect(pythonResult.error).toContain('API key');
        expect(nodeResult.error).toContain('API key');
      }
    });

    it('should handle custom base URL identically', async () => {
      const customUrl = 'https://custom.api.url';
      
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}', base_url='${customUrl}')
    result = {
        'success': True,
        'base_url': client.base_url
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      const nodeClient = new Retab({ 
        apiKey: TEST_API_KEY,
        baseUrl: customUrl
      });
      
      const nodeResult = {
        success: true,
        base_url: (nodeClient as any).baseUrl
      };

      expect(pythonResult.success).toBe(nodeResult.success);
      expect(pythonResult.base_url).toBe(nodeResult.base_url);
    });
  });

  describe('Schema Operations Parity', () => {
    it('should create schemas with identical structure', async () => {
      const testSchema = createTestSchema();
      
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    schema = client.schemas.load(json_schema=${JSON.stringify(testSchema)})
    result = {
        'success': True,
        'json_schema': schema.json_schema,
        'has_save_method': hasattr(schema, 'save'),
        'schema_type': type(schema).__name__
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      let nodeResult;
      try {
        const schema = nodeClient.schemas.load(testSchema);
        nodeResult = {
          success: true,
          json_schema: schema.json_schema,
          has_save_method: typeof schema.save === 'function',
          schema_type: schema.constructor.name
        };
      } catch (error: any) {
        nodeResult = { success: false, error: error.message };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        expect(pythonResult.json_schema).toEqual(nodeResult.json_schema);
        expect(pythonResult.has_save_method).toBe(nodeResult.has_save_method);
        expect(pythonResult.schema_type).toBe(nodeResult.schema_type);
      }
    });

    it('should validate pydantic model handling consistently', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
from pydantic import BaseModel
import json

class TestModel(BaseModel):
    name: str
    age: int

try:
    client = Retab(api_key='${TEST_API_KEY}')
    schema = client.schemas.load(pydantic_model=TestModel)
    result = {
        'success': True,
        'has_json_schema': hasattr(schema, 'json_schema') and schema.json_schema is not None,
        'schema_properties': list(schema.json_schema.get('properties', {}).keys()) if hasattr(schema, 'json_schema') else []
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      let nodeResult;
      try {
        // In Node.js, we simulate pydantic model with pre-serialized schema
        const mockPydanticModel = {
          model_json_schema: () => ({
            type: 'object',
            properties: {
              name: { type: 'string' },
              age: { type: 'integer' }
            },
            required: ['name', 'age']
          })
        };
        
        const schema = nodeClient.schemas.load(null, mockPydanticModel);
        nodeResult = {
          success: true,
          has_json_schema: schema.json_schema !== null,
          schema_properties: Object.keys(schema.json_schema?.properties || {})
        };
      } catch (error: any) {
        nodeResult = { success: false, error: error.message };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        expect(pythonResult.has_json_schema).toBe(nodeResult.has_json_schema);
        expect(pythonResult.schema_properties.sort()).toEqual(nodeResult.schema_properties?.sort() || []);
      }
    });

    it('should prepare schema requests with identical structure', async () => {
      
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    prepared = client.schemas.prepare_generate(
        documents=['test.pdf'],
        instructions='Test instructions',
        model='gpt-4o-2024-11-20',
        temperature=0.5
    )
    result = {
        'success': True,
        'method': prepared.method,
        'url': prepared.url,
        'has_data': prepared.data is not None,
        'data_keys': list(prepared.data.keys()) if prepared.data else []
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      let nodeResult;
      try {
        const prepared = (nodeClient.schemas as any).mixin.prepareGenerate(
          ['test.pdf'],
          'Test instructions',
          'gpt-4o-2024-11-20',
          0.5
        );
        nodeResult = {
          success: true,
          method: prepared.method,
          url: prepared.url,
          has_data: prepared.data !== null,
          data_keys: Object.keys(prepared.data || {})
        };
      } catch (error: any) {
        nodeResult = { success: false, error: error.message };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        expect(pythonResult.method).toBe(nodeResult.method);
        expect(pythonResult.url).toBe(nodeResult.url);
        expect(pythonResult.has_data).toBe(nodeResult.has_data);
        expect(pythonResult.data_keys?.sort() || []).toEqual(nodeResult.data_keys?.sort() || []);
      }
    });
  });

  describe('Error Handling Parity', () => {
    it('should handle invalid model names identically', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    # Test with invalid model that should be rejected by validation
    prepared = client.schemas.prepare_generate(
        documents=['test.pdf'],
        model='invalid-model-name'
    )
    print(json.dumps({'success': True}))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e), 'error_type': type(e).__name__}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      let nodeResult;
      try {
        (nodeClient.schemas as any).mixin.prepareGenerate(
          ['test.pdf'],
          null,
          'invalid-model-name'
        );
        nodeResult = { success: true };
      } catch (error: any) {
        nodeResult = { 
          success: false, 
          error: error.message,
          error_type: error.constructor.name
        };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (!pythonResult.success && !nodeResult.success) {
        // Both should fail with validation-related errors
        expect(pythonResult.error).toBeTruthy();
        expect(nodeResult.error).toBeTruthy();
      }
    });

    it('should handle missing required parameters identically', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    # Missing required documents parameter
    schema = client.schemas.load()  # No parameters
    print(json.dumps({'success': True}))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e), 'error_type': type(e).__name__}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      let nodeResult;
      try {
        nodeClient.schemas.load();  // No parameters
        nodeResult = { success: true };
      } catch (error: any) {
        nodeResult = { 
          success: false, 
          error: error.message,
          error_type: error.constructor.name
        };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (!pythonResult.success && !nodeResult.success) {
        // Both should fail with parameter validation errors
        expect(pythonResult.error).toBeTruthy();
        expect(nodeResult.error).toBeTruthy();
        // Error messages should be similar
        expect(pythonResult.error.toLowerCase()).toContain('provided');
        expect(nodeResult.error.toLowerCase()).toContain('provided');
      }
    });
  });

  describe('Resource Structure Parity', () => {
    it('should have identical resource availability', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    result = {
        'success': True,
        'has_schemas': hasattr(client, 'schemas'),
        'has_documents': hasattr(client, 'documents'),
        'has_files': hasattr(client, 'files'),
        'has_fine_tuning': hasattr(client, 'fine_tuning'),
        'has_consensus': hasattr(client, 'consensus'),
        'has_processors': hasattr(client, 'processors'),
        'schemas_methods': dir(client.schemas),
        'documents_methods': dir(client.documents)
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      const nodeResult = {
        success: true,
        has_schemas: !!nodeClient.schemas,
        has_documents: !!nodeClient.documents,
        has_files: !!nodeClient.files,
        has_fine_tuning: !!nodeClient.fineTuning,
        has_consensus: !!nodeClient.consensus,
        has_processors: !!nodeClient.processors,
        schemas_methods: Object.getOwnPropertyNames(Object.getPrototypeOf(nodeClient.schemas)),
        documents_methods: Object.getOwnPropertyNames(Object.getPrototypeOf(nodeClient.documents))
      };

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        expect(pythonResult.has_schemas).toBe(nodeResult.has_schemas);
        expect(pythonResult.has_documents).toBe(nodeResult.has_documents);
        expect(pythonResult.has_files).toBe(nodeResult.has_files);
        expect(pythonResult.has_fine_tuning).toBe(nodeResult.has_fine_tuning);
        expect(pythonResult.has_consensus).toBe(nodeResult.has_consensus);
        expect(pythonResult.has_processors).toBe(nodeResult.has_processors);
        
        // Check that both have core methods
        const coreSchemasMethods = ['load', 'generate', 'evaluate', 'enhance'];
        coreSchemasMethods.forEach(method => {
          expect(pythonResult.schemas_methods).toContain(method);
          expect(nodeResult.schemas_methods).toContain(method);
        });
      }
    });

    it('should have identical async client structure', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import AsyncRetab
import json

try:
    client = AsyncRetab(api_key='${TEST_API_KEY}')
    result = {
        'success': True,
        'has_schemas': hasattr(client, 'schemas'),
        'has_documents': hasattr(client, 'documents'),
        'schemas_type': type(client.schemas).__name__,
        'client_type': type(client).__name__
    }
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      const nodeResult = {
        success: true,
        has_schemas: !!nodeAsyncClient.schemas,
        has_documents: !!nodeAsyncClient.documents,
        schemas_type: nodeAsyncClient.schemas.constructor.name,
        client_type: nodeAsyncClient.constructor.name
      };

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        expect(pythonResult.has_schemas).toBe(nodeResult.has_schemas);
        expect(pythonResult.has_documents).toBe(nodeResult.has_documents);
        expect(pythonResult.client_type).toBe(nodeResult.client_type);
        // Schema types should follow similar naming patterns
        expect(pythonResult.schemas_type).toContain('Schema');
        expect(nodeResult.schemas_type).toContain('Schema');
      }
    });
  });

  describe('API Request Structure Parity', () => {
    it('should generate identical document extraction requests', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    prepared = client.documents.extractions.prepare_extract(
        documents=['test.pdf'],
        json_schema={'type': 'object', 'properties': {'name': {'type': 'string'}}},
        model='gpt-4o-mini',
        modality='native'
    )
    result = {
        'success': True,
        'method': prepared.method,
        'url': prepared.url,
        'has_data': prepared.data is not None
    }
    if prepared.data:
        result['data_keys'] = list(prepared.data.keys())
        result['model'] = prepared.data.get('model')
        result['modality'] = prepared.data.get('modality')
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      let nodeResult;
      try {
        const prepared = (nodeClient.documents.extractions as any).mixin.prepareExtraction(
          { type: 'object', properties: { name: { type: 'string' } } },
          null,
          ['test.pdf'],
          undefined,
          undefined,
          'gpt-4o-mini',
          undefined,
          'native'
        );
        nodeResult = {
          success: true,
          method: prepared.method,
          url: prepared.url,
          has_data: prepared.data !== null,
          data_keys: prepared.data ? Object.keys(prepared.data) : [],
          model: prepared.data ? prepared.data.model : undefined,
          modality: prepared.data ? prepared.data.modality : undefined
        };
      } catch (error: any) {
        nodeResult = { success: false, error: error.message };
      }

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        expect(pythonResult.method).toBe(nodeResult.method);
        expect(pythonResult.url).toBe(nodeResult.url);
        expect(pythonResult.has_data).toBe(nodeResult.has_data);
        if (pythonResult.has_data && nodeResult.has_data) {
          expect(pythonResult.model).toBe((nodeResult as any).model);
          expect(pythonResult.modality).toBe((nodeResult as any).modality);
          expect(pythonResult.data_keys.sort()).toEqual((nodeResult as any).data_keys.sort());
        }
      }
    });
  });

  describe('Type System Parity', () => {
    it('should handle type validation consistently', async () => {
      const pythonCode = `
import sys
sys.path.append('.')
from retab import Retab
import json

try:
    client = Retab(api_key='${TEST_API_KEY}')
    
    # Test with various parameter types
    test_cases = []
    
    # Test string documents
    try:
        prepared = client.schemas.prepare_generate(documents=['test.pdf'])
        test_cases.append({'case': 'string_documents', 'success': True})
    except Exception as e:
        test_cases.append({'case': 'string_documents', 'success': False, 'error': str(e)})
    
    # Test invalid documents type
    try:
        prepared = client.schemas.prepare_generate(documents=123)  # Invalid type
        test_cases.append({'case': 'invalid_documents', 'success': True})
    except Exception as e:
        test_cases.append({'case': 'invalid_documents', 'success': False, 'error': str(e)})
    
    result = {'success': True, 'test_cases': test_cases}
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
      `;

      const pythonResult = await runPythonCode(pythonCode);
      
      const nodeTestCases = [];
      
      // Test string documents
      try {
        (nodeClient.schemas as any).mixin.prepareGenerate(['test.pdf']);
        nodeTestCases.push({ case: 'string_documents', success: true });
      } catch (error: any) {
        nodeTestCases.push({ case: 'string_documents', success: false, error: error.message });
      }
      
      // Test invalid documents type
      try {
        (nodeClient.schemas as any).prepareGenerate({ documents: 123 });  // Invalid type
        nodeTestCases.push({ case: 'invalid_documents', success: true });
      } catch (error: any) {
        nodeTestCases.push({ case: 'invalid_documents', success: false, error: error.message });
      }
      
      const nodeResult = { success: true, test_cases: nodeTestCases };

      expect(pythonResult.success).toBe(nodeResult.success);
      if (pythonResult.success && nodeResult.success) {
        // Compare test case results
        pythonResult.test_cases.forEach((pythonCase: any) => {
          const nodeCase = nodeResult.test_cases.find((nc: any) => nc.case === pythonCase.case);
          expect(nodeCase).toBeDefined();
          expect(pythonCase.success).toBe(nodeCase!.success);
        });
      }
    });
  });
});