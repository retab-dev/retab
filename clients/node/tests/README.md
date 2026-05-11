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

## Test Structure

The test suite covers the following functionality:

1. **Document APIs**
   - Parse, extract, classify, split, partition, and edit documents
   - Upload and retrieve files

2. **Workflow APIs**
   - Create, update, publish, run, and inspect workflows
   - Manage workflow blocks, edges, tests, runs, and run steps

3. **Operational APIs**
   - List jobs
   - Validate response contracts and error handling

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
cd /workspaces/retab/open-source/sdk/clients/node && bun test tests/extractions/extraction.test.ts -t "test_extract_with_multipart" 2>&1
```
