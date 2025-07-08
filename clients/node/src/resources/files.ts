import { SyncAPIResource, AsyncAPIResource } from '../resource.js';
import OpenAI from 'openai';

type FilePurpose = 'assistants' | 'batch' | 'fine-tune' | 'vision';

export class Files extends SyncAPIResource {
  /**
   * Create a file using OpenAI's API.
   * This integrates with the OpenAI SDK to upload files.
   *
   * @param file - The file to upload
   * @param purpose - The purpose of the file upload
   * @returns Promise<any> - The OpenAI file response
   */
  async create(file: any, purpose: FilePurpose): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.create({ file, purpose });
  }

  /**
   * List files using OpenAI's API.
   *
   * @param purpose - Optional purpose filter
   * @returns Promise<any> - The OpenAI files list response
   */
  async list(purpose?: FilePurpose): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.list({ purpose });
  }

  /**
   * Retrieve a file using OpenAI's API.
   *
   * @param fileId - The ID of the file to retrieve
   * @returns Promise<any> - The OpenAI file response
   */
  async retrieve(fileId: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.retrieve(fileId);
  }

  /**
   * Delete a file using OpenAI's API.
   *
   * @param fileId - The ID of the file to delete
   * @returns Promise<any> - The OpenAI delete response
   */
  async delete(fileId: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.delete(fileId);
  }

  /**
   * Retrieve file content using OpenAI's API.
   *
   * @param fileId - The ID of the file to retrieve content for
   * @returns Promise<any> - The file content
   */
  async content(fileId: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.content(fileId);
  }
}

export class AsyncFiles extends AsyncAPIResource {
  /**
   * Create a file using OpenAI's API asynchronously.
   * This integrates with the OpenAI SDK to upload files.
   *
   * @param file - The file to upload
   * @param purpose - The purpose of the file upload
   * @returns Promise<any> - The OpenAI file response
   */
  async create(file: any, purpose: FilePurpose): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.create({ file, purpose });
  }

  /**
   * List files using OpenAI's API asynchronously.
   *
   * @param purpose - Optional purpose filter
   * @returns Promise<any> - The OpenAI files list response
   */
  async list(purpose?: FilePurpose): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.list({ purpose });
  }

  /**
   * Retrieve a file using OpenAI's API asynchronously.
   *
   * @param fileId - The ID of the file to retrieve
   * @returns Promise<any> - The OpenAI file response
   */
  async retrieve(fileId: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.retrieve(fileId);
  }

  /**
   * Delete a file using OpenAI's API asynchronously.
   *
   * @param fileId - The ID of the file to delete
   * @returns Promise<any> - The OpenAI delete response
   */
  async delete(fileId: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.delete(fileId);
  }

  /**
   * Retrieve file content using OpenAI's API asynchronously.
   *
   * @param fileId - The ID of the file to retrieve content for
   * @returns Promise<any> - The file content
   */
  async content(fileId: string): Promise<any> {
    const openaiApiKey = this._client.getHeaders()['OpenAI-Api-Key'];
    if (!openaiApiKey) {
      throw new Error('OpenAI API key not found. Please provide it in the client configuration.');
    }

    const openaiClient = new OpenAI({ apiKey: openaiApiKey });
    return await openaiClient.files.content(fileId);
  }
}