export class MaxRetriesExceeded extends Error {
  public cause?: Error;
  
  constructor(tries: number, cause?: Error) {
    super(`Max tries exceeded after ${tries} tries.`);
    this.name = 'MaxRetriesExceeded';
    if (cause) {
      this.cause = cause;
    }
  }
}

export class APIError extends Error {
  constructor(
    public statusCode: number,
    public message: string,
    public response?: any
  ) {
    super(message);
    this.name = 'APIError';
  }
}

export class ValidationError extends APIError {
  constructor(message: string, response?: any) {
    super(422, message, response);
    this.name = 'ValidationError';
  }
}