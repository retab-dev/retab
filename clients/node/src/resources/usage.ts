import { SyncAPIResource, AsyncAPIResource } from '../resource.js';
import { PreparedRequest } from '../types/standards.js';
import { Amount, MonthlyUsageResponse } from '../types/ai_models.js';
import { AutomationLog, LogCompletionRequest } from '../types/logs.js';

// Global variable to track total cost
let totalCost = 0.0;

export class UsageMixin {
  prepareMonthlyCreditsUsage(): PreparedRequest {
    return {
      method: 'GET',
      url: '/v1/usage/monthly_credits',
    };
  }

  prepareTotal(params: {
    start_date?: Date;
    end_date?: Date;
  } = {}): PreparedRequest {
    const queryParams: Record<string, string> = {};
    if (params.start_date) {
      queryParams.start_date = params.start_date.toISOString();
    }
    if (params.end_date) {
      queryParams.end_date = params.end_date.toISOString();
    }

    return {
      method: 'GET',
      url: '/v1/usage/total',
      params: queryParams,
    };
  }

  prepareMailbox(params: {
    email: string;
    start_date?: Date;
    end_date?: Date;
  }): PreparedRequest {
    const queryParams: Record<string, string> = {};
    if (params.start_date) {
      queryParams.start_date = params.start_date.toISOString();
    }
    if (params.end_date) {
      queryParams.end_date = params.end_date.toISOString();
    }

    return {
      method: 'GET',
      url: `/v1/processors/automations/mailboxes/${params.email}/usage`,
      params: queryParams,
    };
  }

  prepareLink(params: {
    link_id: string;
    start_date?: Date;
    end_date?: Date;
  }): PreparedRequest {
    const queryParams: Record<string, string> = {};
    if (params.start_date) {
      queryParams.start_date = params.start_date.toISOString();
    }
    if (params.end_date) {
      queryParams.end_date = params.end_date.toISOString();
    }

    return {
      method: 'GET',
      url: `/v1/processors/automations/links/${params.link_id}/usage`,
      params: queryParams,
    };
  }

  prepareSchema(params: {
    schema_id: string;
    start_date?: Date;
    end_date?: Date;
  }): PreparedRequest {
    const queryParams: Record<string, string> = {};
    if (params.start_date) {
      queryParams.start_date = params.start_date.toISOString();
    }
    if (params.end_date) {
      queryParams.end_date = params.end_date.toISOString();
    }

    return {
      method: 'GET',
      url: `/v1/schemas/${params.schema_id}/usage`,
      params: queryParams,
    };
  }

  prepareSchemaData(params: {
    schema_data_id: string;
    start_date?: Date;
    end_date?: Date;
  }): PreparedRequest {
    const queryParams: Record<string, string> = {};
    if (params.start_date) {
      queryParams.start_date = params.start_date.toISOString();
    }
    if (params.end_date) {
      queryParams.end_date = params.end_date.toISOString();
    }

    return {
      method: 'GET',
      url: `/v1/schemas/${params.schema_data_id}/usage_data`,
      params: queryParams,
    };
  }

  prepareLog(params: {
    response_format: any; // ResponseFormat type
    completion: any; // ChatCompletion type
  }): PreparedRequest {
    let logCompletionRequest: LogCompletionRequest;

    if (params.response_format && typeof params.response_format === 'object') {
      if ('json_schema' in params.response_format) {
        const jsonSchema = params.response_format.json_schema as any;
        if ('schema' in jsonSchema) {
          logCompletionRequest = {
            json_schema: jsonSchema.schema,
            completion: params.completion,
          };
        } else {
          throw new Error('Invalid response format');
        }
      } else {
        throw new Error('Invalid response format');
      }
    } else {
      throw new Error('Invalid response format');
    }

    return {
      method: 'POST',
      url: '/v1/usage/log',
      data: logCompletionRequest,
    };
  }
}

export class Usage extends SyncAPIResource {
  private mixin = new UsageMixin();

  /**
   * Get monthly credits usage information.
   * Credits are calculated dynamically based on MIME type and consumption.
   *
   * @returns Monthly usage data including credits consumed and limits
   * @throws RetabAPIError if the API request fails
   */
  monthlyCreditsUsage(): Promise<MonthlyUsageResponse> {
    const request = this.mixin.prepareMonthlyCreditsUsage();
    const response = this._client._preparedRequest(request);
    return response as Promise<MonthlyUsageResponse>;
  }

  /**
   * Get the total usage cost within an optional date range.
   *
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  total(_params: {
    start_date?: Date;
    end_date?: Date;
  } = {}): Amount {
    return { value: totalCost, currency: 'USD' };
  }

  /**
   * Get the total usage cost for a mailbox within an optional date range.
   *
   * @param email - The email address of the mailbox
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  mailbox(params: {
    email: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareMailbox(params);
    const response = this._client._preparedRequest(request);
    return response as Promise<Amount>;
  }

  /**
   * Get the total usage cost for a link within an optional date range.
   *
   * @param link_id - The ID of the link
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  link(params: {
    link_id: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareLink(params);
    const response = this._client._preparedRequest(request);
    return response as Promise<Amount>;
  }

  /**
   * Get the total usage cost for a schema within an optional date range.
   *
   * @param schema_id - The ID of the schema
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  schema(params: {
    schema_id: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareSchema(params);
    const response = this._client._preparedRequest(request);
    return response as Promise<Amount>;
  }

  /**
   * Get the total usage cost for a schema data within an optional date range.
   *
   * @param schema_data_id - The ID of the schema data
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  schemaData(params: {
    schema_data_id: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareSchemaData(params);
    const response = this._client._preparedRequest(request);
    return response as Promise<Amount>;
  }

  /**
   * Logs an OpenAI request completion as an automation log to make usage calculation possible.
   *
   * Example:
   * ```typescript
   * const client = new OpenAI();
   * const completion = await client.beta.chat.completions.parse({
   *   model: "gpt-4o-2024-08-06",
   *   messages: [
   *     {"role": "developer", "content": "Extract the event information."},
   *     {"role": "user", "content": "Alice and Bob are going to a science fair on Friday."},
   *   ],
   *   response_format: CalendarEvent,
   * });
   * reclient.usage.log({
   *   response_format: CalendarEvent,
   *   completion: completion
   * });
   * ```
   *
   * @param response_format - The response format of the OpenAI request
   * @param completion - The completion of the OpenAI request
   * @returns The automation log
   */
  log(params: {
    response_format: any;
    completion: any;
  }): Promise<AutomationLog> {
    const request = this.mixin.prepareLog(params);
    const response = this._client._preparedRequest(request);
    return response as Promise<AutomationLog>;
  }
}

export class AsyncUsage extends AsyncAPIResource {
  private mixin = new UsageMixin();

  /**
   * Get monthly credits usage information asynchronously.
   * Credits are calculated dynamically based on MIME type and consumption.
   *
   * @returns Monthly usage data including credits consumed and limits
   * @throws RetabAPIError if the API request fails
   */
  async monthlyCreditsUsage(): Promise<MonthlyUsageResponse> {
    const request = this.mixin.prepareMonthlyCreditsUsage();
    const response = await this._client._preparedRequest(request);
    return response as MonthlyUsageResponse;
  }

  /**
   * Get the total usage cost within an optional date range.
   *
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  async total(_params: {
    start_date?: Date;
    end_date?: Date;
  } = {}): Promise<Amount> {
    return { value: totalCost, currency: 'USD' };
  }

  /**
   * Get the total usage cost for a mailbox within an optional date range.
   *
   * @param email - The email address of the mailbox
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  async mailbox(params: {
    email: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareMailbox(params);
    const response = await this._client._preparedRequest(request);
    return response as Amount;
  }

  /**
   * Get the total usage cost for a link within an optional date range.
   *
   * @param link_id - The ID of the link
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  async link(params: {
    link_id: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareLink(params);
    const response = await this._client._preparedRequest(request);
    return response as Amount;
  }

  /**
   * Get the total usage cost for a schema within an optional date range.
   *
   * @param schema_id - The ID of the schema
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  async schema(params: {
    schema_id: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareSchema(params);
    const response = await this._client._preparedRequest(request);
    return response as Amount;
  }

  /**
   * Get the total usage cost for a schema data within an optional date range.
   *
   * @param schema_data_id - The ID of the schema data
   * @param start_date - Start date for usage calculation
   * @param end_date - End date for usage calculation
   * @returns The total usage cost
   */
  async schemaData(params: {
    schema_data_id: string;
    start_date?: Date;
    end_date?: Date;
  }): Promise<Amount> {
    const request = this.mixin.prepareSchemaData(params);
    const response = await this._client._preparedRequest(request);
    return response as Amount;
  }

  /**
   * Logs an OpenAI request completion as an automation log to make usage calculation possible.
   *
   * @param response_format - The response format of the OpenAI request
   * @param completion - The completion of the OpenAI request
   * @returns The automation log
   */
  async log(params: {
    response_format: any;
    completion: any;
  }): Promise<AutomationLog> {
    const request = this.mixin.prepareLog(params);
    const response = await this._client._preparedRequest(request);
    return response as AutomationLog;
  }
}