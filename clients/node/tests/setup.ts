import * as dotenv from 'dotenv';
import * as path from 'path';

// Load environment based on NODE_ENV or test flags
const env = process.env.NODE_ENV || 'test';

let envFile: string;
if (process.env.RETAB_ENV_FILE) {
    envFile = process.env.RETAB_ENV_FILE;
} else if (env === 'production') {
    envFile = path.join(__dirname, '../../../.env.production');
} else if (env === 'local') {
    envFile = path.join(__dirname, '../../../.env.local');
} else if (env === 'staging') {
    envFile = path.join(__dirname, '../../../.env.staging');
} else {
    // Default to local for tests
    envFile = path.join(__dirname, '../../../.env.local');
}

console.log(`Loading environment from: ${envFile}`);
dotenv.config({ path: envFile, override: true });

// Bun test configuration is handled differently - no need for explicit timeout setup
