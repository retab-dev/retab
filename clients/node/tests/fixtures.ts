import * as fs from 'fs';
import * as path from 'path';

export interface EnvConfig {
    retabApiKey: string;
    retabApiBaseUrl: string;
}

export function getEnvConfig(): EnvConfig {
    const retabApiKey = process.env.RETAB_API_KEY;
    const retabApiBaseUrl = process.env.RETAB_API_BASE_URL;

    if (!retabApiKey) {
        throw new Error('RETAB_API_KEY must be set in environment');
    }
    if (!retabApiBaseUrl) {
        throw new Error('RETAB_API_BASE_URL must be set in environment');
    }

    return {
        retabApiKey,
        retabApiBaseUrl,
    };
}

export function getTestDataDir(): string {
    return path.join(__dirname, 'data');
}

export function getBookingConfirmationJsonSchema(): Record<string, any> {
    const testDataDir = getTestDataDir();
    const schemaPath = path.join(testDataDir, 'freight', 'booking_confirmation_schema_small.json');
    return JSON.parse(fs.readFileSync(schemaPath, 'utf-8'));
}

export function getBookingConfirmationFilePath1(): string {
    const testDataDir = getTestDataDir();
    return path.join(testDataDir, 'freight', 'booking_confirmation_1.jpg');
}

export function getBookingConfirmationFilePath2(): string {
    const testDataDir = getTestDataDir();
    return path.join(testDataDir, 'freight', 'booking_confirmation_2.jpg');
}

export function getBookingConfirmationData1(): Record<string, any> {
    const testDataDir = getTestDataDir();
    const dataPath = path.join(testDataDir, 'freight', 'booking_confirmation_1_data.json');
    return JSON.parse(fs.readFileSync(dataPath, 'utf-8'));
}

export function getBookingConfirmationData2(): Record<string, any> {
    const testDataDir = getTestDataDir();
    const dataPath = path.join(testDataDir, 'freight', 'booking_confirmation_2_data.json');
    return JSON.parse(fs.readFileSync(dataPath, 'utf-8'));
}

export function getPayslipFilePath(): string {
    const testDataDir = getTestDataDir();
    return path.join(testDataDir, 'payslip', 'payslip.pdf');
}

// Global test constants
export const TEST_MODEL = "gpt-4.1-nano";
export const TEST_MODALITY = "native_fast";
