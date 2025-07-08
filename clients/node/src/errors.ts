export class MaxRetriesExceeded extends Error {
  public attempts: number;
  public cause?: Error;
  
  constructor(attempts: number, cause?: Error) {
    const message = cause 
      ? `Max tries exceeded after ${attempts} attempts. ${cause.message}`
      : `Max tries exceeded after ${attempts} attempts.`;
    super(message);
    this.name = 'MaxRetriesExceeded';
    this.attempts = attempts;
    if (cause) {
      this.cause = cause;
    }
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      attempts: this.attempts,
      stack: this.stack
    };
  }
}

export class APIError extends Error {
  public status: number;
  public response?: any;
  
  constructor(status: number, message: string, response?: any) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.response = response;
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      status: this.status,
      response: this.response,
      stack: this.stack
    };
  }
}

export class ValidationError extends Error {
  public details?: any;
  
  constructor(message: string, details?: any) {
    super(message);
    this.name = 'ValidationError';
    this.details = details;
  }

  toJSON() {
    return {
      name: this.name,
      message: this.message,
      details: this.details,
      stack: this.stack
    };
  }
}