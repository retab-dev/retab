import { SyncAPIResource, AsyncAPIResource } from '../resource.js';
import OpenAI from 'openai';

// Valid OpenAI models for fine-tuning (using string array to avoid type conflicts)
const VALID_OPENAI_MODELS: string[] = [
  'gpt-4o-mini-2024-07-18',
  'gpt-4o-2024-08-06',
  'gpt-3.5-turbo-0125',
  'gpt-3.5-turbo-1106',
  'gpt-3.5-turbo-0613',
  'davinci-002',
  'babbage-002'
];

export class FineTuningJobs extends SyncAPIResource {
  /**
   * Create a new fine-tuning job.
   *
   * @param training_file - The ID of the uploaded training file
   * @param model - The model to fine-tune
   * @returns The fine-tuning job response
   */
  async create(params: {
    training_file: string;
    model: string;
    validation_file?: string;
    hyperparameters?: {
      batch_size?: number | 'auto';
      learning_rate_multiplier?: number | 'auto';
      n_epochs?: number | 'auto';
    };
    suffix?: string;
  }): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    // Validate model
    if (!VALID_OPENAI_MODELS.includes(params.model)) {
      throw new Error(
        `Model ${params.model} is not supported. Supported models are: ${VALID_OPENAI_MODELS.join(', ')}`
      );
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.create({
      training_file: params.training_file,
      model: params.model,
      validation_file: params.validation_file,
      hyperparameters: params.hyperparameters,
      suffix: params.suffix,
    });
  }

  /**
   * Retrieve the status of a fine-tuning job.
   *
   * @param fine_tuning_job_id - The ID of the fine-tuning job
   * @returns The fine-tuning job response
   */
  async retrieve(fine_tuning_job_id: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.retrieve(fine_tuning_job_id);
  }

  /**
   * List fine-tuning jobs.
   *
   * @param params - Optional parameters for listing
   * @returns List of fine-tuning jobs
   */
  async list(params: {
    after?: string;
    limit?: number;
  } = {}): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.list(params);
  }

  /**
   * Cancel a fine-tuning job.
   *
   * @param fine_tuning_job_id - The ID of the fine-tuning job to cancel
   * @returns The cancelled fine-tuning job response
   */
  async cancel(fine_tuning_job_id: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.cancel(fine_tuning_job_id);
  }

  /**
   * List events for a fine-tuning job.
   *
   * @param fine_tuning_job_id - The ID of the fine-tuning job
   * @param params - Optional parameters for listing events
   * @returns List of fine-tuning job events
   */
  async listEvents(fine_tuning_job_id: string, params: {
    after?: string;
    limit?: number;
  } = {}): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.listEvents(fine_tuning_job_id, params);
  }
}

export class AsyncFineTuningJobs extends AsyncAPIResource {
  /**
   * Create a new fine-tuning job asynchronously.
   */
  async create(params: {
    training_file: string;
    model: string;
    validation_file?: string;
    hyperparameters?: {
      batch_size?: number | 'auto';
      learning_rate_multiplier?: number | 'auto';
      n_epochs?: number | 'auto';
    };
    suffix?: string;
  }): Promise<any> {
    // Validate model
    if (!VALID_OPENAI_MODELS.includes(params.model)) {
      throw new Error(
        `Model ${params.model} is not supported. Supported models are: ${VALID_OPENAI_MODELS.join(', ')}`
      );
    }

    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.create({
      training_file: params.training_file,
      model: params.model,
      validation_file: params.validation_file,
      hyperparameters: params.hyperparameters,
      suffix: params.suffix,
    });
  }

  /**
   * Retrieve the status of a fine-tuning job asynchronously.
   */
  async retrieve(fine_tuning_job_id: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.retrieve(fine_tuning_job_id);
  }

  /**
   * List fine-tuning jobs asynchronously.
   */
  async list(params: {
    after?: string;
    limit?: number;
  } = {}): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.list(params);
  }

  /**
   * Cancel a fine-tuning job asynchronously.
   */
  async cancel(fine_tuning_job_id: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.cancel(fine_tuning_job_id);
  }

  /**
   * List events for a fine-tuning job asynchronously.
   */
  async listEvents(fine_tuning_job_id: string, params: {
    after?: string;
    limit?: number;
  } = {}): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.fineTuning.jobs.listEvents(fine_tuning_job_id, params);
  }
}

export class FineTuning extends SyncAPIResource {
  private _jobs: FineTuningJobs;

  constructor(client: any) {
    super(client);
    this._jobs = new FineTuningJobs(client);
  }

  get jobs(): FineTuningJobs {
    return this._jobs;
  }
}

export class AsyncFineTuning extends AsyncAPIResource {
  private _jobs: AsyncFineTuningJobs;

  constructor(client: any) {
    super(client);
    this._jobs = new AsyncFineTuningJobs(client);
  }

  get jobs(): AsyncFineTuningJobs {
    return this._jobs;
  }
}