import fs from 'fs';
import { readJSONL } from './jsonl.js';

/**
 * Rich display and visualization utilities for datasets and metrics
 * Equivalent to Python's display.py
 */

export interface DatasetMetrics {
  totalExamples: number;
  inputTokens: {
    total: number;
    min: number;
    max: number;
    mean: number;
    median: number;
    p95: number;
    p99: number;
  };
  outputTokens: {
    total: number;
    min: number;
    max: number;
    mean: number;
    median: number;
    p95: number;
    p99: number;
  };
  totalTokens: {
    total: number;
    min: number;
    max: number;
    mean: number;
    median: number;
  };
  estimatedCost: {
    input: number;
    output: number;
    total: number;
  };
  messageStats: {
    systemMessages: number;
    userMessages: number;
    assistantMessages: number;
    avgMessagesPerExample: number;
  };
  contentAnalysis: {
    avgSystemLength: number;
    avgUserLength: number;
    avgAssistantLength: number;
    hasImages: boolean;
    imageCount: number;
  };
}

export interface TokenCountResult {
  textTokens: number;
  imageTokens: number;
  totalTokens: number;
}

/**
 * Count tokens in text using a simple approximation
 * In production, you'd want to use tiktoken equivalent for JavaScript
 */
export function countTokens(text: string, _model: string = 'gpt-4o-mini'): number {
  // Simple approximation: ~4 characters per token for English text
  // This is a rough estimate; for production use tiktoken-js or similar
  const avgCharsPerToken = 4;
  return Math.ceil(text.length / avgCharsPerToken);
}

/**
 * Count tokens in content (text + images)
 */
export function countContentTokens(content: string, _model: string = 'gpt-4o-mini'): TokenCountResult {
  let textTokens = 0;
  let imageTokens = 0;

  // Check for image references (simplified detection)
  const imagePatterns = [
    /data:image\/[^;]+;base64,/g,
    /!\[.*?\]\(.*?\)/g, // Markdown images
    /<img[^>]*>/g, // HTML images
  ];

  let textContent = content;

  // Count and remove image references
  for (const pattern of imagePatterns) {
    const matches = content.match(pattern);
    if (matches) {
      // OpenAI vision pricing: roughly 85 tokens per image for low detail
      imageTokens += matches.length * 85;
      textContent = textContent.replace(pattern, '[IMAGE]');
    }
  }

  // Count text tokens
  textTokens = countTokens(textContent, _model);

  return {
    textTokens,
    imageTokens,
    totalTokens: textTokens + imageTokens,
  };
}

/**
 * Calculate statistical metrics for an array of numbers
 */
export function calculateStats(values: number[]): {
  min: number;
  max: number;
  mean: number;
  median: number;
  p95: number;
  p99: number;
  total: number;
} {
  if (values.length === 0) {
    return { min: 0, max: 0, mean: 0, median: 0, p95: 0, p99: 0, total: 0 };
  }

  const sorted = [...values].sort((a, b) => a - b);
  const total = values.reduce((sum, val) => sum + val, 0);
  const mean = total / values.length;

  const getPercentile = (p: number) => {
    const index = Math.ceil((p / 100) * sorted.length) - 1;
    return sorted[Math.max(0, index)];
  };

  return {
    min: sorted[0],
    max: sorted[sorted.length - 1],
    mean: Math.round(mean * 100) / 100,
    median: sorted[Math.floor(sorted.length / 2)],
    p95: getPercentile(95),
    p99: getPercentile(99),
    total,
  };
}

/**
 * Process dataset and compute comprehensive metrics
 */
export async function processDatasetAndComputeMetrics(
  datasetPath: string,
  inputTokenPrice: number = 0.00015,
  outputTokenPrice: number = 0.0006,
  model: string = 'gpt-4o-mini'
): Promise<DatasetMetrics> {
  if (!fs.existsSync(datasetPath)) {
    throw new Error(`Dataset file not found: ${datasetPath}`);
  }

  const dataset = await readJSONL(datasetPath);
  
  const inputTokenCounts: number[] = [];
  const outputTokenCounts: number[] = [];
  const totalTokenCounts: number[] = [];
  
  let systemMessages = 0;
  let userMessages = 0;
  let assistantMessages = 0;
  let totalMessages = 0;
  
  let systemLengths: number[] = [];
  let userLengths: number[] = [];
  let assistantLengths: number[] = [];
  
  let imageCount = 0;
  let hasImages = false;

  for (const example of dataset) {
    if (!example.messages || !Array.isArray(example.messages)) {
      continue;
    }

    let exampleInputTokens = 0;
    let exampleOutputTokens = 0;

    for (const message of example.messages) {
      totalMessages++;
      const content = message.content || '';
      const tokenCount = countContentTokens(content, model);
      
      // Track content lengths
      const contentLength = content.length;

      switch (message.role) {
        case 'system':
          systemMessages++;
          exampleInputTokens += tokenCount.totalTokens;
          systemLengths.push(contentLength);
          break;
        case 'user':
          userMessages++;
          exampleInputTokens += tokenCount.totalTokens;
          userLengths.push(contentLength);
          break;
        case 'assistant':
          assistantMessages++;
          exampleOutputTokens += tokenCount.totalTokens;
          assistantLengths.push(contentLength);
          break;
      }

      // Check for images
      if (tokenCount.imageTokens > 0) {
        hasImages = true;
        imageCount += tokenCount.imageTokens / 85; // Rough estimate
      }
    }

    inputTokenCounts.push(exampleInputTokens);
    outputTokenCounts.push(exampleOutputTokens);
    totalTokenCounts.push(exampleInputTokens + exampleOutputTokens);
  }

  const inputStats = calculateStats(inputTokenCounts);
  const outputStats = calculateStats(outputTokenCounts);
  const totalStats = calculateStats(totalTokenCounts);

  return {
    totalExamples: dataset.length,
    inputTokens: inputStats,
    outputTokens: outputStats,
    totalTokens: totalStats,
    estimatedCost: {
      input: (inputStats.total * inputTokenPrice) / 1000,
      output: (outputStats.total * outputTokenPrice) / 1000,
      total: ((inputStats.total * inputTokenPrice) + (outputStats.total * outputTokenPrice)) / 1000,
    },
    messageStats: {
      systemMessages,
      userMessages,
      assistantMessages,
      avgMessagesPerExample: Math.round((totalMessages / dataset.length) * 100) / 100,
    },
    contentAnalysis: {
      avgSystemLength: systemLengths.length > 0 ? Math.round((systemLengths.reduce((a, b) => a + b, 0) / systemLengths.length) * 100) / 100 : 0,
      avgUserLength: userLengths.length > 0 ? Math.round((userLengths.reduce((a, b) => a + b, 0) / userLengths.length) * 100) / 100 : 0,
      avgAssistantLength: assistantLengths.length > 0 ? Math.round((assistantLengths.reduce((a, b) => a + b, 0) / assistantLengths.length) * 100) / 100 : 0,
      hasImages,
      imageCount: Math.round(imageCount),
    },
  };
}

/**
 * Display metrics in a formatted table
 */
export function displayMetrics(metrics: DatasetMetrics): void {
  console.log('\n📊 Dataset Analysis Report');
  console.log('═'.repeat(50));
  
  // Basic Stats
  console.log(`\n📈 Basic Statistics:`);
  console.log(`   Total Examples: ${metrics.totalExamples.toLocaleString()}`);
  console.log(`   Avg Messages/Example: ${metrics.messageStats.avgMessagesPerExample}`);
  
  // Message Distribution
  console.log(`\n💬 Message Distribution:`);
  console.log(`   System Messages: ${metrics.messageStats.systemMessages.toLocaleString()}`);
  console.log(`   User Messages: ${metrics.messageStats.userMessages.toLocaleString()}`);
  console.log(`   Assistant Messages: ${metrics.messageStats.assistantMessages.toLocaleString()}`);

  // Token Statistics
  console.log(`\n🔢 Token Statistics:`);
  console.log(`   Input Tokens:`);
  console.log(`     Total: ${metrics.inputTokens.total.toLocaleString()}`);
  console.log(`     Mean: ${metrics.inputTokens.mean.toLocaleString()}`);
  console.log(`     Median: ${metrics.inputTokens.median.toLocaleString()}`);
  console.log(`     Min: ${metrics.inputTokens.min.toLocaleString()}`);
  console.log(`     Max: ${metrics.inputTokens.max.toLocaleString()}`);
  console.log(`     95th percentile: ${metrics.inputTokens.p95.toLocaleString()}`);
  console.log(`     99th percentile: ${metrics.inputTokens.p99.toLocaleString()}`);

  console.log(`\n   Output Tokens:`);
  console.log(`     Total: ${metrics.outputTokens.total.toLocaleString()}`);
  console.log(`     Mean: ${metrics.outputTokens.mean.toLocaleString()}`);
  console.log(`     Median: ${metrics.outputTokens.median.toLocaleString()}`);
  console.log(`     Min: ${metrics.outputTokens.min.toLocaleString()}`);
  console.log(`     Max: ${metrics.outputTokens.max.toLocaleString()}`);
  console.log(`     95th percentile: ${metrics.outputTokens.p95.toLocaleString()}`);
  console.log(`     99th percentile: ${metrics.outputTokens.p99.toLocaleString()}`);

  // Cost Estimation
  console.log(`\n💰 Cost Estimation:`);
  console.log(`   Input Cost: $${metrics.estimatedCost.input.toFixed(4)}`);
  console.log(`   Output Cost: $${metrics.estimatedCost.output.toFixed(4)}`);
  console.log(`   Total Cost: $${metrics.estimatedCost.total.toFixed(4)}`);

  // Content Analysis
  console.log(`\n📝 Content Analysis:`);
  console.log(`   Avg System Message Length: ${metrics.contentAnalysis.avgSystemLength.toLocaleString()} chars`);
  console.log(`   Avg User Message Length: ${metrics.contentAnalysis.avgUserLength.toLocaleString()} chars`);
  console.log(`   Avg Assistant Message Length: ${metrics.contentAnalysis.avgAssistantLength.toLocaleString()} chars`);
  
  if (metrics.contentAnalysis.hasImages) {
    console.log(`   Images Detected: ${metrics.contentAnalysis.imageCount.toLocaleString()}`);
  }

  console.log('\n' + '═'.repeat(50));
}

/**
 * Format large numbers with appropriate units
 */
export function formatNumber(num: number): string {
  if (num >= 1_000_000) {
    return `${(num / 1_000_000).toFixed(1)}M`;
  } else if (num >= 1_000) {
    return `${(num / 1_000).toFixed(1)}K`;
  }
  return num.toLocaleString();
}

/**
 * Create a simple ASCII progress bar
 */
export function createProgressBar(current: number, total: number, width: number = 40): string {
  const percentage = Math.min(current / total, 1);
  const filled = Math.floor(percentage * width);
  const empty = width - filled;
  
  return `[${'█'.repeat(filled)}${' '.repeat(empty)}] ${(percentage * 100).toFixed(1)}% (${current}/${total})`;
}

/**
 * Display progress with a progress bar
 */
export function displayProgress(current: number, total: number, message?: string): void {
  const progressBar = createProgressBar(current, total);
  const output = message ? `${message} ${progressBar}` : progressBar;
  
  // Clear line and write progress (works in most terminals)
  process.stdout.write(`\r${output}`);
  
  if (current >= total) {
    process.stdout.write('\n');
  }
}

export default {
  processDatasetAndComputeMetrics,
  displayMetrics,
  countTokens,
  countContentTokens,
  calculateStats,
  formatNumber,
  createProgressBar,
  displayProgress,
};