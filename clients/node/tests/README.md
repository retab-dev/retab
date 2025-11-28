# Retab Node SDK Test Suite

This is a complete test suite for the Retab Node SDK, translated from the Python test suite. It helps identify what functionality is missing in the Node SDK compared to the Python implementation.

## Setup

1. Install dependencies:
```bash
npm install
```

2. Set up environment variables by creating the appropriate `.env` file in the project root:
   - `.env.local` for local testing
   - `.env.staging` for staging testing  
   - `.env.production` for production testing

The `.env` file should contain:
```
RETAB_API_KEY=your_api_key_here
RETAB_API_BASE_URL=https://api.retab.com  # or your local/staging URL
```

## Running Tests

### Against local server (http://localhost:4000):
```bash
npm run test:local
```

### Against staging server:
```bash
NODE_ENV=staging npm test
```

### Against production server (https://api.retab.com):
```bash
NODE_ENV=production npm test
```

### Run all tests:
```bash
npm test
```

### Run tests with coverage:
```bash
npm run test:coverage
```

### Run tests in watch mode:
```bash
npm run test:watch
```

### Run a specific test:
```bash
npm test -- --testNamePattern="test_evaluation_crud_basic"
```

## Test Structure

The test suite covers the following functionality:

1. **Basic Project CRUD Operations**
   - Create, read, update, delete projects
   - List projects

2. **Document Management**
   - Add documents to projects
   - Update document annotations
   - List and delete documents

3. **Iteration Management**
   - Create iterations with inference settings
   - List and delete iterations
   - Check iteration status *(may not be implemented yet)*

4. **Document Processing**
   - Bulk processing via `process` method *(may not be implemented yet)*
   - Individual document processing via `process_document` method *(may not be implemented yet)*

5. **Complete Workflows**
   - End-to-end testing with multiple documents and iterations
   - Selective processing scenarios

## Missing Functionality

The tests use `(client.projects.iterations as any)` for the following methods that may not be implemented in the Node SDK yet:

- `status(projectId, iterationId)` - Check status of documents in an iteration
- `process(projectId, iterationId, options)` - Bulk process documents in an iteration  
- `process_document(projectId, iterationId, documentId)` - Process a single document

When these methods are missing, the tests will fail and clearly indicate what needs to be implemented.

## Test Data

The tests use the same test data as the Python test suite:
- `booking_confirmation_schema_small.json` - JSON schema for extraction
- `booking_confirmation_1.jpg` and `booking_confirmation_2.jpg` - Test images
- `booking_confirmation_1_data.json` and `booking_confirmation_2_data.json` - Expected annotations

## Notes

- Tests have a 30-second timeout to accommodate API processing times
- Tests clean up created resources automatically
- Each test uses unique names with nanoid to avoid conflicts
- Tests are designed to run independently and can be run in parallel

# Example of running a specific test:
```bash
cd /workspaces/retab/open-source/sdk/clients/node && bun test tests/documents/extract.test.ts -t "test_extract_with_multipart" 2>&1
```