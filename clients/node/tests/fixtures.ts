import { z } from 'zod';

// Test data fixtures - mirror the Python test data
export const bookingConfirmationSchema = {
  type: 'object',
  title: 'BookingConfirmation',
  properties: {
    booking_number: {
      type: 'string',
      description: 'The booking confirmation number'
    },
    shipper_name: {
      type: 'string',
      description: 'Name of the shipping company'
    },
    container_number: {
      type: 'string',
      description: 'Container identification number'
    },
    vessel_name: {
      type: 'string',
      description: 'Name of the vessel/ship'
    },
    port_of_loading: {
      type: 'string',
      description: 'Port where cargo is loaded'
    },
    port_of_discharge: {
      type: 'string',
      description: 'Port where cargo is discharged'
    },
    estimated_departure: {
      type: 'string',
      format: 'date',
      description: 'Estimated departure date'
    },
    estimated_arrival: {
      type: 'string',
      format: 'date',
      description: 'Estimated arrival date'
    }
  },
  required: [
    'booking_number',
    'shipper_name',
    'container_number',
    'vessel_name',
    'port_of_loading',
    'port_of_discharge'
  ]
};

// Company schema fixture
export const companySchema = {
  type: 'object',
  title: 'Company',
  properties: {
    name: {
      type: 'string',
      description: 'Name of the identified company',
      'X-FieldPrompt': 'Look for the name of the company, or derive it from the logo'
    },
    type: {
      type: 'string',
      enum: ['school', 'investor', 'startup', 'corporate'],
      description: 'Type of the identified company',
      'X-FieldPrompt': 'Guess the type depending on slide context'
    },
    relationship: {
      type: 'string',
      enum: ['founderBackground', 'investor', 'competitor', 'client', 'partnership'],
      description: 'Relationship of the identified company with the startup from the deck',
      'X-FieldPrompt': 'Guess the relationship of the identified company with the startup from the deck'
    }
  },
  required: ['name', 'type', 'relationship']
};

// Zod schema fixtures for testing
export const PersonZodSchema = z.object({
  name: z.string().describe('Full name of the person'),
  age: z.number().min(0).max(150).describe('Age in years'),
  email: z.string().email().describe('Email address'),
  isActive: z.boolean().default(true).describe('Whether the person is active')
});

export const AddressZodSchema = z.object({
  street: z.string().describe('Street address'),
  city: z.string().describe('City name'),
  zipCode: z.string().describe('ZIP/postal code'),
  country: z.string().describe('Country name')
});

export const PersonWithAddressZodSchema = z.object({
  person: PersonZodSchema,
  address: AddressZodSchema.optional(),
  tags: z.array(z.string()).default([]).describe('List of tags')
});

// Mock pydantic model structure for testing
export const mockPydanticModel = {
  model_json_schema: () => companySchema,
  schema: companySchema
};

export const invalidPydanticModel = {
  // Missing model_json_schema method and schema property
  someOtherMethod: () => ({})
};

// Test constants
export const TEST_API_KEY = 'test-api-key-12345';
export const TEST_BASE_URL = 'https://api.test.retab.com';
export const INVALID_API_KEY = 'invalid-key';

// Helper functions for tests
export function generateRandomId(): string {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}

export function generateTestEmail(): string {
  return `test-${generateRandomId()}@example.com`;
}

export async function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}