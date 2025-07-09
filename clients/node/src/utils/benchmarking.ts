import { readJSONL, writeJSONL } from './jsonl.js';

/**
 * Benchmarking and evaluation utilities for model comparison
 * Equivalent to Python's benchmarking.py
 */

export interface EvaluationMetrics {
  accuracy: number;
  precision: number;
  recall: number;
  f1Score: number;
  exactMatch: number;
  levenshteinDistance: number;
  jaccardSimilarity: number;
  hammingDistance: number;
  fieldAccuracy: Record<string, number>;
  completeness: number;
  errorRate: number;
}

export interface SingleFileEvalResult {
  filename: string;
  metrics: EvaluationMetrics;
  predictions: any[];
  groundTruths: any[];
  differences: Record<string, any>[];
  executionTime: number;
}

export interface BenchmarkResult {
  model: string;
  overallMetrics: EvaluationMetrics;
  fileResults: SingleFileEvalResult[];
  aggregateStats: {
    meanAccuracy: number;
    stdDevAccuracy: number;
    meanF1: number;
    stdDevF1: number;
    totalFiles: number;
    totalPredictions: number;
  };
  executionTime: number;
}

export interface DictionaryDifference {
  field: string;
  predicted: any;
  groundTruth: any;
  differenceType: 'missing' | 'extra' | 'value_mismatch' | 'type_mismatch';
  path: string;
}

/**
 * Calculate Levenshtein distance between two strings
 */
export function levenshteinDistance(str1: string, str2: string): number {
  const matrix: number[][] = [];
  
  // Initialize matrix
  for (let i = 0; i <= str2.length; i++) {
    matrix[i] = [i];
  }
  for (let j = 0; j <= str1.length; j++) {
    matrix[0][j] = j;
  }
  
  // Fill matrix
  for (let i = 1; i <= str2.length; i++) {
    for (let j = 1; j <= str1.length; j++) {
      if (str2.charAt(i - 1) === str1.charAt(j - 1)) {
        matrix[i][j] = matrix[i - 1][j - 1];
      } else {
        matrix[i][j] = Math.min(
          matrix[i - 1][j - 1] + 1, // substitution
          matrix[i][j - 1] + 1,     // insertion
          matrix[i - 1][j] + 1      // deletion
        );
      }
    }
  }
  
  return matrix[str2.length][str1.length];
}

/**
 * Calculate Jaccard similarity between two sets
 */
export function jaccardSimilarity(set1: Set<any>, set2: Set<any>): number {
  const intersection = new Set([...set1].filter(x => set2.has(x)));
  const union = new Set([...set1, ...set2]);
  
  if (union.size === 0) return 1.0;
  return intersection.size / union.size;
}

/**
 * Calculate Hamming distance between two strings
 */
export function hammingDistance(str1: string, str2: string): number {
  if (str1.length !== str2.length) {
    throw new Error('Strings must be of equal length for Hamming distance');
  }
  
  let distance = 0;
  for (let i = 0; i < str1.length; i++) {
    if (str1[i] !== str2[i]) {
      distance++;
    }
  }
  return distance;
}

/**
 * Flatten nested object into dot-notation keys
 */
export function flattenObject(obj: any, prefix: string = ''): Record<string, any> {
  const flattened: Record<string, any> = {};
  
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {
      const newKey = prefix ? `${prefix}.${key}` : key;
      const value = obj[key];
      
      if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
        Object.assign(flattened, flattenObject(value, newKey));
      } else {
        flattened[newKey] = value;
      }
    }
  }
  
  return flattened;
}

/**
 * Compute detailed differences between two dictionaries
 */
export function computeDictDifference(predicted: any, groundTruth: any, path: string = ''): DictionaryDifference[] {
  const differences: DictionaryDifference[] = [];
  
  const flatPredicted = flattenObject(predicted);
  const flatGroundTruth = flattenObject(groundTruth);
  
  const allKeys = new Set([
    ...Object.keys(flatPredicted),
    ...Object.keys(flatGroundTruth)
  ]);
  
  for (const key of allKeys) {
    const fullPath = path ? `${path}.${key}` : key;
    const predValue = flatPredicted[key];
    const truthValue = flatGroundTruth[key];
    
    if (!(key in flatPredicted)) {
      differences.push({
        field: key,
        predicted: undefined,
        groundTruth: truthValue,
        differenceType: 'missing',
        path: fullPath,
      });
    } else if (!(key in flatGroundTruth)) {
      differences.push({
        field: key,
        predicted: predValue,
        groundTruth: undefined,
        differenceType: 'extra',
        path: fullPath,
      });
    } else if (predValue !== truthValue) {
      const diffType = typeof predValue !== typeof truthValue ? 'type_mismatch' : 'value_mismatch';
      differences.push({
        field: key,
        predicted: predValue,
        groundTruth: truthValue,
        differenceType: diffType,
        path: fullPath,
      });
    }
  }
  
  return differences;
}

/**
 * Aggregate dictionary differences across multiple examples
 */
export function aggregateDictDifferences(differences: DictionaryDifference[][]): Record<string, {
  count: number;
  percentage: number;
  examples: DictionaryDifference[];
}> {
  const aggregated: Record<string, DictionaryDifference[]> = {};
  
  // Group differences by field path
  for (const diffList of differences) {
    for (const diff of diffList) {
      if (!aggregated[diff.path]) {
        aggregated[diff.path] = [];
      }
      aggregated[diff.path].push(diff);
    }
  }
  
  const totalExamples = differences.length;
  const result: Record<string, { count: number; percentage: number; examples: DictionaryDifference[] }> = {};
  
  for (const [path, diffs] of Object.entries(aggregated)) {
    result[path] = {
      count: diffs.length,
      percentage: (diffs.length / totalExamples) * 100,
      examples: diffs.slice(0, 5), // Keep first 5 examples
    };
  }
  
  return result;
}

/**
 * Calculate comprehensive evaluation metrics
 */
export function calculateMetrics(predictions: any[], groundTruths: any[]): EvaluationMetrics {
  if (predictions.length !== groundTruths.length) {
    throw new Error('Predictions and ground truths must have the same length');
  }
  
  const n = predictions.length;
  let exactMatches = 0;
  let totalLevenshtein = 0;
  let totalJaccard = 0;
  let totalHamming = 0;
  let validHamming = 0;
  
  const fieldAccuracy: Record<string, { correct: number; total: number }> = {};
  const differences: DictionaryDifference[][] = [];
  
  for (let i = 0; i < n; i++) {
    const pred = predictions[i];
    const truth = groundTruths[i];
    
    // Exact match
    if (JSON.stringify(pred) === JSON.stringify(truth)) {
      exactMatches++;
    }
    
    // String representations for text-based metrics
    const predStr = JSON.stringify(pred);
    const truthStr = JSON.stringify(truth);
    
    // Levenshtein distance
    totalLevenshtein += levenshteinDistance(predStr, truthStr);
    
    // Jaccard similarity (using character sets)
    const predSet = new Set(predStr.split(''));
    const truthSet = new Set(truthStr.split(''));
    totalJaccard += jaccardSimilarity(predSet, truthSet);
    
    // Hamming distance (only for same-length strings)
    if (predStr.length === truthStr.length) {
      totalHamming += hammingDistance(predStr, truthStr);
      validHamming++;
    }
    
    // Field-level accuracy
    const diff = computeDictDifference(pred, truth);
    differences.push(diff);
    
    const flatPred = flattenObject(pred);
    const flatTruth = flattenObject(truth);
    
    for (const key of Object.keys(flatTruth)) {
      if (!fieldAccuracy[key]) {
        fieldAccuracy[key] = { correct: 0, total: 0 };
      }
      fieldAccuracy[key].total++;
      if (flatPred[key] === flatTruth[key]) {
        fieldAccuracy[key].correct++;
      }
    }
  }
  
  // Calculate field accuracy percentages
  const fieldAccuracyPercentages: Record<string, number> = {};
  for (const [field, stats] of Object.entries(fieldAccuracy)) {
    fieldAccuracyPercentages[field] = (stats.correct / stats.total) * 100;
  }
  
  // Calculate aggregate differences
  const aggregatedDiffs = aggregateDictDifferences(differences);
  const completeness = 100 - (Object.keys(aggregatedDiffs).length / Object.keys(fieldAccuracy).length) * 100;
  
  return {
    accuracy: (exactMatches / n) * 100,
    precision: (exactMatches / n) * 100, // Simplified for exact match scenario
    recall: (exactMatches / n) * 100,    // Simplified for exact match scenario
    f1Score: (exactMatches / n) * 100,   // Simplified for exact match scenario
    exactMatch: (exactMatches / n) * 100,
    levenshteinDistance: totalLevenshtein / n,
    jaccardSimilarity: (totalJaccard / n) * 100,
    hammingDistance: validHamming > 0 ? totalHamming / validHamming : 0,
    fieldAccuracy: fieldAccuracyPercentages,
    completeness: Math.max(0, completeness),
    errorRate: ((n - exactMatches) / n) * 100,
  };
}

/**
 * Single file evaluation class
 */
export class SingleFileEval {
  private filename: string;
  private predictions: any[];
  private groundTruths: any[];
  
  constructor(filename: string, predictions: any[], groundTruths: any[]) {
    this.filename = filename;
    this.predictions = predictions;
    this.groundTruths = groundTruths;
  }
  
  async evaluate(): Promise<SingleFileEvalResult> {
    const startTime = Date.now();
    
    const metrics = calculateMetrics(this.predictions, this.groundTruths);
    const differences: Record<string, any>[] = [];
    
    for (let i = 0; i < this.predictions.length; i++) {
      const diff = computeDictDifference(this.predictions[i], this.groundTruths[i]);
      if (diff.length > 0) {
        differences.push({
          index: i,
          differences: diff,
        });
      }
    }
    
    const executionTime = Date.now() - startTime;
    
    return {
      filename: this.filename,
      metrics,
      predictions: this.predictions,
      groundTruths: this.groundTruths,
      differences,
      executionTime,
    };
  }
}

/**
 * Plot metrics with uncertainty (text-based visualization)
 */
export function plotMetricsWithUncertainty(results: BenchmarkResult[]): void {
  console.log('\n📊 Model Performance Comparison');
  console.log('═'.repeat(60));
  
  const maxModelNameLength = Math.max(...results.map(r => r.model.length));
  
  console.log(`\n${'Model'.padEnd(maxModelNameLength)} | Accuracy | F1 Score | Exec Time`);
  console.log('─'.repeat(maxModelNameLength + 35));
  
  for (const result of results) {
    const accuracy = result.overallMetrics.accuracy.toFixed(1);
    const f1 = result.overallMetrics.f1Score.toFixed(1);
    const execTime = `${(result.executionTime / 1000).toFixed(1)}s`;
    
    console.log(
      `${result.model.padEnd(maxModelNameLength)} | ${accuracy.padStart(6)}% | ${f1.padStart(6)}% | ${execTime.padStart(8)}`
    );
  }
  
  // Show best performing model
  const bestModel = results.reduce((best, current) => 
    current.overallMetrics.accuracy > best.overallMetrics.accuracy ? current : best
  );
  
  console.log(`\n🏆 Best performing model: ${bestModel.model} (${bestModel.overallMetrics.accuracy.toFixed(1)}% accuracy)`);
  console.log('═'.repeat(60));
}

/**
 * Benchmark multiple models
 */
export async function benchmark(
  models: string[],
  testDataPath: string,
  groundTruthPath: string,
  evaluationFunction: (model: string, testData: any[]) => Promise<any[]>
): Promise<BenchmarkResult[]> {
  console.log(`🚀 Starting benchmark of ${models.length} models...`);
  
  // Load test data and ground truth
  const testData = await readJSONL(testDataPath);
  const groundTruth = await readJSONL(groundTruthPath);
  
  if (testData.length !== groundTruth.length) {
    throw new Error('Test data and ground truth must have the same length');
  }
  
  const results: BenchmarkResult[] = [];
  
  for (let i = 0; i < models.length; i++) {
    const model = models[i];
    console.log(`\n📊 Evaluating model: ${model} (${i + 1}/${models.length})`);
    
    const startTime = Date.now();
    
    try {
      // Get predictions from model
      const predictions = await evaluationFunction(model, testData);
      
      if (predictions.length !== groundTruth.length) {
        throw new Error(`Model ${model} returned ${predictions.length} predictions, expected ${groundTruth.length}`);
      }
      
      // Calculate metrics
      const metrics = calculateMetrics(predictions, groundTruth);
      
      // Create single file evaluation
      const fileEval = new SingleFileEval(testDataPath, predictions, groundTruth);
      const fileResult = await fileEval.evaluate();
      
      const executionTime = Date.now() - startTime;
      
      results.push({
        model,
        overallMetrics: metrics,
        fileResults: [fileResult],
        aggregateStats: {
          meanAccuracy: metrics.accuracy,
          stdDevAccuracy: 0, // Would need multiple runs to calculate
          meanF1: metrics.f1Score,
          stdDevF1: 0,
          totalFiles: 1,
          totalPredictions: predictions.length,
        },
        executionTime,
      });
      
      console.log(`   ✅ Accuracy: ${metrics.accuracy.toFixed(1)}%`);
      console.log(`   ⏱️  Time: ${(executionTime / 1000).toFixed(1)}s`);
      
    } catch (error) {
      console.error(`   ❌ Failed to evaluate ${model}:`, error);
    }
  }
  
  // Display final comparison
  plotMetricsWithUncertainty(results);
  
  return results;
}

/**
 * Save benchmark results to file
 */
export async function saveBenchmarkResults(
  results: BenchmarkResult[],
  outputPath: string
): Promise<void> {
  const summary = {
    timestamp: new Date().toISOString(),
    totalModels: results.length,
    results: results.map(r => ({
      model: r.model,
      accuracy: r.overallMetrics.accuracy,
      f1Score: r.overallMetrics.f1Score,
      executionTime: r.executionTime,
      totalPredictions: r.aggregateStats.totalPredictions,
    })),
    detailed: results,
  };
  
  await writeJSONL(outputPath, [summary]);
  console.log(`📄 Benchmark results saved to ${outputPath}`);
}

export default {
  SingleFileEval,
  calculateMetrics,
  computeDictDifference,
  aggregateDictDifferences,
  levenshteinDistance,
  jaccardSimilarity,
  hammingDistance,
  flattenObject,
  plotMetricsWithUncertainty,
  benchmark,
  saveBenchmarkResults,
};