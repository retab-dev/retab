import { APIError, ValidationError, MaxRetriesExceeded } from '../src/errors.js';

describe('Error Classes', () => {
  describe('APIError', () => {
    it('should create APIError with status and message', () => {
      const error = new APIError(400, 'Bad Request');
      
      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(APIError);
      expect(error.status).toBe(400);
      expect(error.message).toBe('Bad Request');
      expect(error.name).toBe('APIError');
      expect(error.response).toBeUndefined();
    });

    it('should create APIError with status, message, and response', () => {
      const responseData = { error: 'Invalid request', code: 'BAD_REQUEST' };
      const error = new APIError(400, 'Bad Request', responseData);
      
      expect(error.status).toBe(400);
      expect(error.message).toBe('Bad Request');
      expect(error.response).toEqual(responseData);
    });

    it('should have proper error name', () => {
      const error = new APIError(500, 'Internal Server Error');
      expect(error.name).toBe('APIError');
    });

    it('should be catchable as Error', () => {
      const error = new APIError(404, 'Not Found');
      
      expect(() => {
        throw error;
      }).toThrow(Error);
    });

    it('should preserve stack trace', () => {
      const error = new APIError(500, 'Server Error');
      expect(error.stack).toBeDefined();
    });
  });

  describe('ValidationError', () => {
    it('should create ValidationError with message', () => {
      const error = new ValidationError('Validation failed');
      
      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(ValidationError);
      expect(error.message).toBe('Validation failed');
      expect(error.name).toBe('ValidationError');
      expect(error.details).toBeUndefined();
    });

    it('should create ValidationError with message and details', () => {
      const details = {
        field_errors: [
          { field: 'name', message: 'Name is required' },
          { field: 'email', message: 'Invalid email format' }
        ]
      };
      const error = new ValidationError('Validation failed', details);
      
      expect(error.message).toBe('Validation failed');
      expect(error.details).toEqual(details);
    });

    it('should have proper error name', () => {
      const error = new ValidationError('Invalid input');
      expect(error.name).toBe('ValidationError');
    });

    it('should be catchable as Error', () => {
      const error = new ValidationError('Validation error');
      
      expect(() => {
        throw error;
      }).toThrow(Error);
    });

    it('should preserve stack trace', () => {
      const error = new ValidationError('Validation failed');
      expect(error.stack).toBeDefined();
    });
  });

  describe('MaxRetriesExceeded', () => {
    it('should create MaxRetriesExceeded with attempts count', () => {
      const error = new MaxRetriesExceeded(3);
      
      expect(error).toBeInstanceOf(Error);
      expect(error).toBeInstanceOf(MaxRetriesExceeded);
      expect(error.attempts).toBe(3);
      expect(error.message).toContain('3 attempts');
      expect(error.name).toBe('MaxRetriesExceeded');
      expect(error.cause).toBeUndefined();
    });

    it('should create MaxRetriesExceeded with attempts and cause', () => {
      const cause = new Error('Network timeout');
      const error = new MaxRetriesExceeded(5, cause);
      
      expect(error.attempts).toBe(5);
      expect(error.message).toContain('5 attempts');
      expect(error.cause).toBe(cause);
    });

    it('should have proper error name', () => {
      const error = new MaxRetriesExceeded(2);
      expect(error.name).toBe('MaxRetriesExceeded');
    });

    it('should be catchable as Error', () => {
      const error = new MaxRetriesExceeded(4);
      
      expect(() => {
        throw error;
      }).toThrow(Error);
    });

    it('should preserve stack trace', () => {
      const error = new MaxRetriesExceeded(1);
      expect(error.stack).toBeDefined();
    });

    it('should include cause in message when provided', () => {
      const cause = new Error('Connection refused');
      const error = new MaxRetriesExceeded(3, cause);
      
      expect(error.message).toContain('Connection refused');
    });
  });

  describe('Error Inheritance', () => {
    it('should maintain proper inheritance chain for APIError', () => {
      const error = new APIError(400, 'Bad Request');
      
      expect(error instanceof Error).toBe(true);
      expect(error instanceof APIError).toBe(true);
      expect(Object.getPrototypeOf(error)).toBe(APIError.prototype);
    });

    it('should maintain proper inheritance chain for ValidationError', () => {
      const error = new ValidationError('Validation failed');
      
      expect(error instanceof Error).toBe(true);
      expect(error instanceof ValidationError).toBe(true);
      expect(Object.getPrototypeOf(error)).toBe(ValidationError.prototype);
    });

    it('should maintain proper inheritance chain for MaxRetriesExceeded', () => {
      const error = new MaxRetriesExceeded(3);
      
      expect(error instanceof Error).toBe(true);
      expect(error instanceof MaxRetriesExceeded).toBe(true);
      expect(Object.getPrototypeOf(error)).toBe(MaxRetriesExceeded.prototype);
    });
  });

  describe('Error Serialization', () => {
    it('should serialize APIError properly', () => {
      const responseData = { error: 'Bad request', details: ['Invalid field'] };
      const error = new APIError(400, 'Bad Request', responseData);
      
      const serialized = JSON.stringify(error);
      const parsed = JSON.parse(serialized);
      
      expect(parsed.name).toBe('APIError');
      expect(parsed.message).toBe('Bad Request');
      expect(parsed.status).toBe(400);
      expect(parsed.response).toEqual(responseData);
    });

    it('should serialize ValidationError properly', () => {
      const details = { field: 'email', error: 'Invalid format' };
      const error = new ValidationError('Validation failed', details);
      
      const serialized = JSON.stringify(error);
      const parsed = JSON.parse(serialized);
      
      expect(parsed.name).toBe('ValidationError');
      expect(parsed.message).toBe('Validation failed');
      expect(parsed.details).toEqual(details);
    });

    it('should serialize MaxRetriesExceeded properly', () => {
      const error = new MaxRetriesExceeded(5);
      
      const serialized = JSON.stringify(error);
      const parsed = JSON.parse(serialized);
      
      expect(parsed.name).toBe('MaxRetriesExceeded');
      expect(parsed.attempts).toBe(5);
      expect(parsed.message).toContain('5 attempts');
    });
  });

  describe('Error Creation Patterns', () => {
    it('should handle different status codes for APIError', () => {
      const statuses = [400, 401, 403, 404, 422, 429, 500, 502, 503, 504];
      
      statuses.forEach(status => {
        const error = new APIError(status, `HTTP ${status}`);
        expect(error.status).toBe(status);
        expect(error.message).toBe(`HTTP ${status}`);
      });
    });

    it('should handle complex validation details', () => {
      const complexDetails = {
        errors: [
          {
            field: 'user.profile.email',
            code: 'INVALID_EMAIL',
            message: 'Email format is invalid'
          },
          {
            field: 'user.profile.age',
            code: 'OUT_OF_RANGE',
            message: 'Age must be between 0 and 150'
          }
        ],
        request_id: 'req_123456',
        timestamp: '2024-01-01T12:00:00Z'
      };
      
      const error = new ValidationError('Multiple validation errors', complexDetails);
      expect(error.details).toEqual(complexDetails);
      expect(error.details.errors).toHaveLength(2);
    });

    it('should handle retry scenarios', () => {
      const networkError = new Error('ECONNREFUSED');
      const retryError = new MaxRetriesExceeded(3, networkError);
      
      expect(retryError.attempts).toBe(3);
      expect(retryError.cause).toBe(networkError);
      expect(retryError.message).toContain('ECONNREFUSED');
    });
  });

  describe('Error Comparison', () => {
    it('should differentiate between error types', () => {
      const apiError = new APIError(500, 'Server Error');
      const validationError = new ValidationError('Validation Error');
      const retryError = new MaxRetriesExceeded(3);
      
      expect(apiError).not.toEqual(validationError);
      expect(validationError).not.toEqual(retryError);
      expect(retryError).not.toEqual(apiError);
    });

    it('should compare errors of same type', () => {
      const error1 = new APIError(400, 'Bad Request');
      const error2 = new APIError(400, 'Bad Request');
      
      expect(error1.status).toBe(error2.status);
      expect(error1.message).toBe(error2.message);
      expect(error1.name).toBe(error2.name);
    });
  });
});