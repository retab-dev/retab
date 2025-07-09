import fs from 'fs';
import path from 'path';
import { SyncAPIResource, AsyncAPIResource } from '../resource.js';
import { readJSONL, writeJSONL } from './jsonl.js';
import { displayMetrics, processDatasetAndComputeMetrics, DatasetMetrics } from './display.js';

/**
 * Advanced Dataset management utilities for ML training workflows
 * Equivalent to Python's jsonlUtils.py
 */

export interface FinetuningJSON {
  messages: Array<{
    role: 'system' | 'user' | 'assistant';
    content: string;
  }>;
}

export interface DocumentAnnotationPair {
  document: string | Buffer;
  annotation: Record<string, any>;
}

export interface BatchJSONLRequest {
  custom_id: string;
  method: 'POST';
  url: string;
  body: Record<string, any>;
}

export interface BatchJSONLResponse {
  id: string;
  custom_id: string;
  response: {
    status_code: number;
    request_id: string;
    body: Record<string, any>;
  };
  error?: {
    code: string;
    message: string;
  };
}

export interface AnnotationOptions {
  model?: string;
  temperature?: number;
  modality?: 'native' | 'text';
  maxConcurrency?: number;
  reasoning_effort?: 'low' | 'medium' | 'high';
  provider?: 'openai' | 'anthropic' | 'xai' | 'gemini';
  idempotencyKey?: string;
}

export interface SaveOptions {
  modality?: 'native' | 'text';
  imageResolutionDpi?: number;
  browserCanvas?: 'A3' | 'A4' | 'A5';
}

export class BaseDatasetsMixin {
  /**
   * Process dataset and compute comprehensive metrics
   */
  async pprint(
    datasetPath: string,
    inputTokenPrice: number = 0.00015,
    outputTokenPrice: number = 0.0006
  ): Promise<DatasetMetrics> {
    if (!fs.existsSync(datasetPath)) {
      throw new Error(`Dataset file not found: ${datasetPath}`);
    }

    const metrics = await processDatasetAndComputeMetrics(
      datasetPath,
      inputTokenPrice,
      outputTokenPrice
    );

    displayMetrics(metrics);
    return metrics;
  }

  /**
   * Save document-annotation pairs as JSONL training dataset
   */
  async save(
    jsonSchema: Record<string, any> | string,
    documentAnnotationPairsPaths: Array<{
      document: string;
      annotation: string;
    }>,
    datasetPath: string,
    options: SaveOptions = {}
  ): Promise<void> {
    const { modality = 'native' } = options;
    const finetuningData: FinetuningJSON[] = [];

    for (const { document: docPath, annotation: annPath } of documentAnnotationPairsPaths) {
      // Read document and annotation
      if (!fs.existsSync(docPath) || !fs.existsSync(annPath)) {
        throw new Error(`Document or annotation file not found: ${docPath}, ${annPath}`);
      }

      const annotation = JSON.parse(fs.readFileSync(annPath, 'utf-8'));
      
      // Create system message with schema
      const systemMessage = this.createSystemMessage(jsonSchema, modality);
      
      // Create user message with document
      const userMessage = await this.createUserMessage(docPath, modality, options);
      
      // Create assistant message with annotation
      const assistantMessage = {
        role: 'assistant' as const,
        content: JSON.stringify(annotation),
      };

      finetuningData.push({
        messages: [systemMessage, userMessage, assistantMessage],
      });
    }

    // Write to JSONL file
    await writeJSONL(datasetPath, finetuningData);
    console.log(`✅ Dataset saved to ${datasetPath} with ${finetuningData.length} examples`);
  }

  /**
   * Change schema in existing dataset
   */
  async changeSchema(
    inputDatasetPath: string,
    jsonSchema: Record<string, any> | string,
    outputDatasetPath?: string,
    inplace: boolean = false
  ): Promise<void> {
    if (!fs.existsSync(inputDatasetPath)) {
      throw new Error(`Input dataset not found: ${inputDatasetPath}`);
    }

    const outputPath = inplace ? inputDatasetPath : (outputDatasetPath || inputDatasetPath);
    const tempPath = `${outputPath}.tmp`;

    try {
      const dataset = await readJSONL(inputDatasetPath);
      const newSystemMessage = this.createSystemMessage(jsonSchema, 'native');

      const updatedDataset = dataset.map((item: FinetuningJSON) => ({
        ...item,
        messages: [
          newSystemMessage,
          ...item.messages.slice(1), // Keep user and assistant messages
        ],
      }));

      await writeJSONL(tempPath, updatedDataset);
      
      // Atomic move
      fs.renameSync(tempPath, outputPath);
      console.log(`✅ Schema updated in ${outputPath}`);
    } catch (error) {
      // Cleanup temp file on error
      if (fs.existsSync(tempPath)) {
        fs.unlinkSync(tempPath);
      }
      throw error;
    }
  }

  /**
   * Stitch multiple documents and save as dataset
   */
  async stitchAndSave(
    jsonSchema: Record<string, any> | string,
    pairsPaths: Array<{
      documents: string[];
      annotation: string;
    }>,
    datasetPath: string,
    modality: 'native' | 'text' = 'native'
  ): Promise<void> {
    const finetuningData: FinetuningJSON[] = [];

    for (const { documents: docPaths, annotation: annPath } of pairsPaths) {
      if (!fs.existsSync(annPath)) {
        throw new Error(`Annotation file not found: ${annPath}`);
      }

      // Verify all document files exist
      for (const docPath of docPaths) {
        if (!fs.existsSync(docPath)) {
          throw new Error(`Document file not found: ${docPath}`);
        }
      }

      const annotation = JSON.parse(fs.readFileSync(annPath, 'utf-8'));
      
      const systemMessage = this.createSystemMessage(jsonSchema, modality);
      const userMessage = await this.createMultiDocumentUserMessage(docPaths, modality);
      
      const assistantMessage = {
        role: 'assistant' as const,
        content: JSON.stringify(annotation),
      };

      finetuningData.push({
        messages: [systemMessage, userMessage, assistantMessage],
      });
    }

    await writeJSONL(datasetPath, finetuningData);
    console.log(`✅ Stitched dataset saved to ${datasetPath} with ${finetuningData.length} examples`);
  }

  /**
   * Generate annotations for documents using AI models
   */
  async annotate(
    jsonSchema: Record<string, any> | string,
    documents: string[],
    datasetPath: string,
    options: AnnotationOptions = {}
  ): Promise<void> {
    const {
      model = 'gpt-4o-mini',
      temperature = 0.0,
      modality = 'native',
      maxConcurrency = 5,
      reasoning_effort = 'medium',
      provider = 'openai',
    } = options;

    console.log(`🚀 Starting annotation of ${documents.length} documents...`);
    
    const finetuningData: FinetuningJSON[] = [];
    const concurrencyLimit = Math.min(maxConcurrency, documents.length);
    
    // Process documents in batches
    for (let i = 0; i < documents.length; i += concurrencyLimit) {
      const batch = documents.slice(i, i + concurrencyLimit);
      const batchPromises = batch.map(async (docPath, index) => {
        const globalIndex = i + index;
        console.log(`📝 Processing document ${globalIndex + 1}/${documents.length}: ${path.basename(docPath)}`);
        
        try {
          const annotation = await this.generateAnnotation(
            jsonSchema,
            docPath,
            model,
            temperature,
            modality,
            reasoning_effort,
            provider
          );

          const systemMessage = this.createSystemMessage(jsonSchema, modality);
          const userMessage = await this.createUserMessage(docPath, modality);
          const assistantMessage = {
            role: 'assistant' as const,
            content: JSON.stringify(annotation),
          };

          return {
            messages: [systemMessage, userMessage, assistantMessage],
          };
        } catch (error) {
          console.error(`❌ Failed to process ${docPath}:`, error);
          return null;
        }
      });

      const batchResults = await Promise.all(batchPromises);
      finetuningData.push(...batchResults.filter(result => result !== null));
    }

    await writeJSONL(datasetPath, finetuningData);
    console.log(`✅ Annotation complete! Generated ${finetuningData.length}/${documents.length} annotations`);
  }

  /**
   * Update existing annotations with new model/schema
   */
  async updateAnnotations(
    jsonSchema: Record<string, any> | string,
    oldDatasetPath: string,
    newDatasetPath: string,
    options: AnnotationOptions = {}
  ): Promise<void> {
    if (!fs.existsSync(oldDatasetPath)) {
      throw new Error(`Old dataset not found: ${oldDatasetPath}`);
    }

    console.log(`🔄 Updating annotations from ${oldDatasetPath}...`);
    
    const oldDataset = await readJSONL(oldDatasetPath);
    const updatedDataset: FinetuningJSON[] = [];

    for (let i = 0; i < oldDataset.length; i++) {
      const item = oldDataset[i] as FinetuningJSON;
      console.log(`🔄 Updating annotation ${i + 1}/${oldDataset.length}`);

      try {
        // Extract document path from user message (this is simplified)
        const userContent = item.messages.find(m => m.role === 'user')?.content;
        if (!userContent) {
          console.warn(`⚠️  No user message found in item ${i + 1}, skipping`);
          continue;
        }

        // For this implementation, we assume the document is referenced in the user message
        // In practice, you'd need to store document paths or reconstruct them
        const newAnnotation = await this.generateAnnotationFromUserMessage(
          jsonSchema,
          userContent,
          options
        );

        const systemMessage = this.createSystemMessage(jsonSchema, options.modality || 'native');
        const assistantMessage = {
          role: 'assistant' as const,
          content: JSON.stringify(newAnnotation),
        };

        updatedDataset.push({
          messages: [
            systemMessage,
            item.messages.find(m => m.role === 'user')!,
            assistantMessage,
          ],
        });
      } catch (error) {
        console.error(`❌ Failed to update annotation ${i + 1}:`, error);
      }
    }

    await writeJSONL(newDatasetPath, updatedDataset);
    console.log(`✅ Updated ${updatedDataset.length}/${oldDataset.length} annotations`);
  }

  /**
   * Save batch annotation requests for OpenAI Batch API
   */
  async saveBatchAnnotateRequests(
    jsonSchema: Record<string, any> | string,
    documents: string[],
    batchRequestsPath: string,
    options: AnnotationOptions = {}
  ): Promise<void> {
    const {
      model = 'gpt-4o-mini',
      temperature = 0.0,
      modality = 'native',
    } = options;

    const batchRequests: BatchJSONLRequest[] = [];

    for (let i = 0; i < documents.length; i++) {
      const docPath = documents[i];
      const systemMessage = this.createSystemMessage(jsonSchema, modality);
      const userMessage = await this.createUserMessage(docPath, modality);

      batchRequests.push({
        custom_id: `doc_${i}_${path.basename(docPath, path.extname(docPath))}`,
        method: 'POST',
        url: '/v1/chat/completions',
        body: {
          model,
          messages: [systemMessage, userMessage],
          temperature,
          response_format: { type: 'json_object' },
        },
      });
    }

    await writeJSONL(batchRequestsPath, batchRequests);
    console.log(`✅ Saved ${batchRequests.length} batch requests to ${batchRequestsPath}`);
  }

  /**
   * Build dataset from batch API results
   */
  async buildDatasetFromBatchResults(
    jsonSchema: Record<string, any> | string,
    batchResultsPath: string,
    datasetPath: string,
    modality: 'native' | 'text' = 'native'
  ): Promise<void> {
    if (!fs.existsSync(batchResultsPath)) {
      throw new Error(`Batch results file not found: ${batchResultsPath}`);
    }

    const batchResults = await readJSONL(batchResultsPath);
    const finetuningData: FinetuningJSON[] = [];

    for (const result of batchResults as BatchJSONLResponse[]) {
      if (result.error) {
        console.warn(`⚠️  Skipping failed request ${result.custom_id}: ${result.error.message}`);
        continue;
      }

      const response = result.response.body;
      const content = response.choices?.[0]?.message?.content;
      
      if (!content) {
        console.warn(`⚠️  No content in response for ${result.custom_id}`);
        continue;
      }

      try {
        const annotation = JSON.parse(content);
        
        // Reconstruct messages (this is simplified)
        const systemMessage = this.createSystemMessage(jsonSchema, modality);
        
        // Extract user message from original request (would need to be stored)
        const userMessage = {
          role: 'user' as const,
          content: `Document content for ${result.custom_id}`,
        };
        
        const assistantMessage = {
          role: 'assistant' as const,
          content: JSON.stringify(annotation),
        };

        finetuningData.push({
          messages: [systemMessage, userMessage, assistantMessage],
        });
      } catch (error) {
        console.warn(`⚠️  Failed to parse annotation for ${result.custom_id}:`, error);
      }
    }

    await writeJSONL(datasetPath, finetuningData);
    console.log(`✅ Built dataset with ${finetuningData.length} examples from batch results`);
  }

  // Helper methods
  private createSystemMessage(
    jsonSchema: Record<string, any> | string,
    _modality: 'native' | 'text'
  ) {
    const schemaObj = typeof jsonSchema === 'string' ? JSON.parse(jsonSchema) : jsonSchema;
    const schemaStr = JSON.stringify(schemaObj, null, 2);
    
    return {
      role: 'system' as const,
      content: `You are an expert data extraction assistant. Extract information from the provided document according to the following JSON schema:\n\n${schemaStr}\n\nReturn only valid JSON that matches the schema exactly.`,
    };
  }

  private async createUserMessage(
    docPath: string,
    _modality: 'native' | 'text',
    _options: SaveOptions = {}
  ) {
    // This is a simplified implementation
    // In practice, you'd handle different file types, base64 encoding, etc.
    const content = fs.readFileSync(docPath, 'utf-8');
    
    return {
      role: 'user' as const,
      content: `Please extract data from this document:\n\n${content}`,
    };
  }

  private async createMultiDocumentUserMessage(
    docPaths: string[],
    _modality: 'native' | 'text'
  ) {
    const contents = docPaths.map((docPath, index) => {
      const content = fs.readFileSync(docPath, 'utf-8');
      return `Document ${index + 1} (${path.basename(docPath)}):\n${content}`;
    }).join('\n\n---\n\n');

    return {
      role: 'user' as const,
      content: `Please extract data from these documents:\n\n${contents}`,
    };
  }

  private async generateAnnotation(
    _jsonSchema: Record<string, any> | string,
    _docPath: string,
    _model: string,
    _temperature: number,
    _modality: 'native' | 'text',
    _reasoningEffort: 'low' | 'medium' | 'high',
    _provider: 'openai' | 'anthropic' | 'xai' | 'gemini'
  ): Promise<Record<string, any>> {
    // This would integrate with the actual AI providers
    // For now, return a placeholder implementation
    throw new Error('AI provider integration not implemented in this version');
  }

  private async generateAnnotationFromUserMessage(
    _jsonSchema: Record<string, any> | string,
    _userContent: string,
    _options: AnnotationOptions
  ): Promise<Record<string, any>> {
    // This would re-generate annotation from existing user message
    throw new Error('Annotation update not implemented in this version');
  }
}

export class Datasets extends SyncAPIResource {
  private mixin = new BaseDatasetsMixin();

  async pprint(
    datasetPath: string,
    inputTokenPrice?: number,
    outputTokenPrice?: number
  ): Promise<DatasetMetrics> {
    return this.mixin.pprint(datasetPath, inputTokenPrice, outputTokenPrice);
  }

  async save(
    jsonSchema: Record<string, any> | string,
    documentAnnotationPairsPaths: Array<{ document: string; annotation: string }>,
    datasetPath: string,
    options?: SaveOptions
  ): Promise<void> {
    return this.mixin.save(jsonSchema, documentAnnotationPairsPaths, datasetPath, options);
  }

  async changeSchema(
    inputDatasetPath: string,
    jsonSchema: Record<string, any> | string,
    outputDatasetPath?: string,
    inplace?: boolean
  ): Promise<void> {
    return this.mixin.changeSchema(inputDatasetPath, jsonSchema, outputDatasetPath, inplace);
  }

  async stitchAndSave(
    jsonSchema: Record<string, any> | string,
    pairsPaths: Array<{ documents: string[]; annotation: string }>,
    datasetPath: string,
    modality?: 'native' | 'text'
  ): Promise<void> {
    return this.mixin.stitchAndSave(jsonSchema, pairsPaths, datasetPath, modality);
  }

  async annotate(
    jsonSchema: Record<string, any> | string,
    documents: string[],
    datasetPath: string,
    options?: AnnotationOptions
  ): Promise<void> {
    return this.mixin.annotate(jsonSchema, documents, datasetPath, options);
  }

  async updateAnnotations(
    jsonSchema: Record<string, any> | string,
    oldDatasetPath: string,
    newDatasetPath: string,
    options?: AnnotationOptions
  ): Promise<void> {
    return this.mixin.updateAnnotations(jsonSchema, oldDatasetPath, newDatasetPath, options);
  }

  async saveBatchAnnotateRequests(
    jsonSchema: Record<string, any> | string,
    documents: string[],
    batchRequestsPath: string,
    options?: AnnotationOptions
  ): Promise<void> {
    return this.mixin.saveBatchAnnotateRequests(jsonSchema, documents, batchRequestsPath, options);
  }

  async buildDatasetFromBatchResults(
    jsonSchema: Record<string, any> | string,
    batchResultsPath: string,
    datasetPath: string,
    modality?: 'native' | 'text'
  ): Promise<void> {
    return this.mixin.buildDatasetFromBatchResults(jsonSchema, batchResultsPath, datasetPath, modality);
  }
}

export class AsyncDatasets extends AsyncAPIResource {
  private mixin = new BaseDatasetsMixin();

  async pprint(
    datasetPath: string,
    inputTokenPrice?: number,
    outputTokenPrice?: number
  ): Promise<DatasetMetrics> {
    return this.mixin.pprint(datasetPath, inputTokenPrice, outputTokenPrice);
  }

  async save(
    jsonSchema: Record<string, any> | string,
    documentAnnotationPairsPaths: Array<{ document: string; annotation: string }>,
    datasetPath: string,
    options?: SaveOptions
  ): Promise<void> {
    return this.mixin.save(jsonSchema, documentAnnotationPairsPaths, datasetPath, options);
  }

  async changeSchema(
    inputDatasetPath: string,
    jsonSchema: Record<string, any> | string,
    outputDatasetPath?: string,
    inplace?: boolean
  ): Promise<void> {
    return this.mixin.changeSchema(inputDatasetPath, jsonSchema, outputDatasetPath, inplace);
  }

  async stitchAndSave(
    jsonSchema: Record<string, any> | string,
    pairsPaths: Array<{ documents: string[]; annotation: string }>,
    datasetPath: string,
    modality?: 'native' | 'text'
  ): Promise<void> {
    return this.mixin.stitchAndSave(jsonSchema, pairsPaths, datasetPath, modality);
  }

  async annotate(
    jsonSchema: Record<string, any> | string,
    documents: string[],
    datasetPath: string,
    options?: AnnotationOptions
  ): Promise<void> {
    return this.mixin.annotate(jsonSchema, documents, datasetPath, options);
  }

  async updateAnnotations(
    jsonSchema: Record<string, any> | string,
    oldDatasetPath: string,
    newDatasetPath: string,
    options?: AnnotationOptions
  ): Promise<void> {
    return this.mixin.updateAnnotations(jsonSchema, oldDatasetPath, newDatasetPath, options);
  }

  async saveBatchAnnotateRequests(
    jsonSchema: Record<string, any> | string,
    documents: string[],
    batchRequestsPath: string,
    options?: AnnotationOptions
  ): Promise<void> {
    return this.mixin.saveBatchAnnotateRequests(jsonSchema, documents, batchRequestsPath, options);
  }

  async buildDatasetFromBatchResults(
    jsonSchema: Record<string, any> | string,
    batchResultsPath: string,
    datasetPath: string,
    modality?: 'native' | 'text'
  ): Promise<void> {
    return this.mixin.buildDatasetFromBatchResults(jsonSchema, batchResultsPath, datasetPath, modality);
  }
}

export default {
  Datasets,
  AsyncDatasets,
  BaseDatasetsMixin,
};