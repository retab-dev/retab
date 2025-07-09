/**
 * Prompt optimization utilities for improving AI model performance
 * Equivalent to Python's prompt_optimization.py
 */

export interface PromptOptimizationConfig {
  maxTokens?: number;
  temperature?: number;
  model?: string;
  preserveStructure?: boolean;
  optimizationGoal?: 'accuracy' | 'speed' | 'cost' | 'balanced';
}

export interface PromptMetrics {
  tokenCount: number;
  clarity: number;
  specificity: number;
  complexity: number;
  estimatedCost: number;
  estimatedLatency: number;
}

export interface OptimizationResult {
  originalPrompt: string;
  optimizedPrompt: string;
  improvements: string[];
  metrics: {
    before: PromptMetrics;
    after: PromptMetrics;
  };
  confidence: number;
}

/**
 * Estimate token count for a prompt (rough approximation)
 */
export function estimateTokenCount(text: string): number {
  // Rough estimation: 4 characters per token on average
  return Math.ceil(text.length / 4);
}

/**
 * Calculate prompt clarity score based on readability metrics
 */
export function calculateClarity(prompt: string): number {
  const sentences = prompt.split(/[.!?]+/).filter(s => s.trim().length > 0);
  const words = prompt.split(/\s+/).filter(w => w.length > 0);
  
  if (sentences.length === 0 || words.length === 0) return 0;
  
  const avgWordsPerSentence = words.length / sentences.length;
  const avgWordLength = words.reduce((sum, word) => sum + word.length, 0) / words.length;
  
  // Flesch Reading Ease inspired calculation (simplified)
  const clarityScore = Math.max(0, Math.min(100, 
    206.835 - (1.015 * avgWordsPerSentence) - (84.6 * avgWordLength / 5)
  ));
  
  return clarityScore / 100;
}

/**
 * Calculate prompt specificity score
 */
export function calculateSpecificity(prompt: string): number {
  const specificityKeywords = [
    'specific', 'exactly', 'precise', 'detailed', 'step-by-step',
    'format', 'structure', 'example', 'template', 'must', 'should',
    'required', 'necessary', 'important', 'key', 'essential'
  ];
  
  const words = prompt.toLowerCase().split(/\s+/);
  const specificityWords = words.filter(word => 
    specificityKeywords.some(keyword => word.includes(keyword))
  );
  
  return Math.min(1, specificityWords.length / words.length * 10);
}

/**
 * Calculate prompt complexity score
 */
export function calculateComplexity(prompt: string): number {
  const complexityIndicators = [
    'however', 'although', 'whereas', 'furthermore', 'moreover',
    'consequently', 'therefore', 'nevertheless', 'nonetheless',
    'if', 'unless', 'provided', 'assuming', 'considering'
  ];
  
  const words = prompt.toLowerCase().split(/\s+/);
  const complexWords = words.filter(word => 
    complexityIndicators.some(indicator => word.includes(indicator))
  );
  
  const sentences = prompt.split(/[.!?]+/).filter(s => s.trim().length > 0);
  const avgSentenceLength = words.length / sentences.length;
  
  return Math.min(1, (complexWords.length / words.length * 5) + (avgSentenceLength / 50));
}

/**
 * Estimate cost based on token count and model
 */
export function estimateCost(tokenCount: number, model: string = 'gpt-4o-mini'): number {
  // Rough cost estimates per 1K tokens (as of 2024)
  const costPer1KTokens: Record<string, number> = {
    'gpt-4o': 0.03,
    'gpt-4o-mini': 0.00015,
    'gpt-4': 0.03,
    'gpt-3.5-turbo': 0.002,
    'claude-3-opus': 0.015,
    'claude-3-sonnet': 0.003,
    'claude-3-haiku': 0.00025,
  };
  
  const rate = costPer1KTokens[model] || 0.002;
  return (tokenCount / 1000) * rate;
}

/**
 * Estimate latency based on token count and model
 */
export function estimateLatency(tokenCount: number, model: string = 'gpt-4o-mini'): number {
  // Rough latency estimates in milliseconds
  const baseLatency: Record<string, number> = {
    'gpt-4o': 500,
    'gpt-4o-mini': 300,
    'gpt-4': 800,
    'gpt-3.5-turbo': 200,
    'claude-3-opus': 1000,
    'claude-3-sonnet': 600,
    'claude-3-haiku': 400,
  };
  
  const base = baseLatency[model] || 400;
  return base + (tokenCount * 2); // ~2ms per token
}

/**
 * Calculate comprehensive prompt metrics
 */
export function calculateMetrics(prompt: string, model: string = 'gpt-4o-mini'): PromptMetrics {
  const tokenCount = estimateTokenCount(prompt);
  
  return {
    tokenCount,
    clarity: calculateClarity(prompt),
    specificity: calculateSpecificity(prompt),
    complexity: calculateComplexity(prompt),
    estimatedCost: estimateCost(tokenCount, model),
    estimatedLatency: estimateLatency(tokenCount, model),
  };
}

/**
 * Optimize prompt for brevity while maintaining clarity
 */
export function optimizeForBrevity(prompt: string): string {
  return prompt
    // Remove redundant phrases
    .replace(/\b(please|kindly|if you would|if you could)\b/gi, '')
    // Simplify conjunctions
    .replace(/\bin order to\b/gi, 'to')
    .replace(/\bdue to the fact that\b/gi, 'because')
    .replace(/\bat this point in time\b/gi, 'now')
    // Remove excessive politeness
    .replace(/\bthank you\b/gi, '')
    .replace(/\bi would appreciate\b/gi, '')
    // Clean up extra spaces
    .replace(/\s+/g, ' ')
    .trim();
}

/**
 * Optimize prompt for clarity and structure
 */
export function optimizeForClarity(prompt: string): string {
  // Add clear structure markers
  let optimized = prompt;
  
  // Add numbering to instructions if they appear to be steps
  if (prompt.includes('\n') && !prompt.match(/^\d+\./m)) {
    const lines = prompt.split('\n');
    const instructionLines = lines.filter(line => 
      line.trim().length > 0 && 
      (line.includes('please') || line.includes('should') || line.includes('must'))
    );
    
    if (instructionLines.length > 1) {
      optimized = lines.map((line) => {
        const trimmed = line.trim();
        if (trimmed.length > 0 && instructionLines.includes(line)) {
          return `${instructionLines.indexOf(line) + 1}. ${trimmed}`;
        }
        return line;
      }).join('\n');
    }
  }
  
  // Add clear formatting
  optimized = optimized
    .replace(/\b(important|key|essential|critical|note):\s*/gi, '**$1**: ')
    .replace(/\b(example|for example|e\.g\.)\s*/gi, '\n**Example**: ')
    .replace(/\b(format|structure|template)\s*/gi, '\n**$1**: ');
  
  return optimized;
}

/**
 * Optimize prompt for specific AI model characteristics
 */
export function optimizeForModel(prompt: string, model: string): string {
  let optimized = prompt;
  
  switch (model) {
    case 'gpt-4o':
    case 'gpt-4':
      // GPT-4 responds well to structured prompts
      optimized = optimizeForClarity(prompt);
      break;
      
    case 'gpt-4o-mini':
    case 'gpt-3.5-turbo':
      // Smaller models benefit from brevity
      optimized = optimizeForBrevity(prompt);
      break;
      
    case 'claude-3-opus':
    case 'claude-3-sonnet':
      // Claude models prefer more conversational tone
      optimized = prompt.replace(/\byou must\b/gi, 'please');
      break;
      
    case 'claude-3-haiku':
      // Haiku benefits from very concise prompts
      optimized = optimizeForBrevity(prompt);
      break;
  }
  
  return optimized;
}

/**
 * Main optimization function
 */
export function optimizePrompt(
  prompt: string,
  config: PromptOptimizationConfig = {}
): OptimizationResult {
  const {
    model = 'gpt-4o-mini',
    optimizationGoal = 'balanced',
    preserveStructure = false,
  } = config;
  
  const originalMetrics = calculateMetrics(prompt, model);
  let optimized = prompt;
  const improvements: string[] = [];
  
  // Apply optimizations based on goal
  switch (optimizationGoal) {
    case 'accuracy':
      optimized = optimizeForClarity(optimized);
      improvements.push('Enhanced clarity and structure');
      break;
      
    case 'speed':
      optimized = optimizeForBrevity(optimized);
      improvements.push('Reduced token count for faster processing');
      break;
      
    case 'cost':
      optimized = optimizeForBrevity(optimized);
      improvements.push('Minimized token usage to reduce costs');
      break;
      
    case 'balanced':
      optimized = optimizeForModel(optimized, model);
      improvements.push('Balanced optimization for model characteristics');
      break;
  }
  
  // Apply model-specific optimizations
  if (!preserveStructure) {
    optimized = optimizeForModel(optimized, model);
    improvements.push(`Optimized for ${model} characteristics`);
  }
  
  const optimizedMetrics = calculateMetrics(optimized, model);
  
  // Calculate confidence based on improvement metrics
  const tokenImprovement = (originalMetrics.tokenCount - optimizedMetrics.tokenCount) / originalMetrics.tokenCount;
  const clarityImprovement = optimizedMetrics.clarity - originalMetrics.clarity;
  const costImprovement = (originalMetrics.estimatedCost - optimizedMetrics.estimatedCost) / originalMetrics.estimatedCost;
  
  const confidence = Math.max(0, Math.min(1, 
    (tokenImprovement * 0.3) + (clarityImprovement * 0.4) + (costImprovement * 0.3)
  ));
  
  return {
    originalPrompt: prompt,
    optimizedPrompt: optimized,
    improvements,
    metrics: {
      before: originalMetrics,
      after: optimizedMetrics,
    },
    confidence,
  };
}

/**
 * Test prompt variations and return the best one
 */
export async function testPromptVariations(
  basePrompt: string,
  variations: string[],
  testFunction: (prompt: string) => Promise<number>, // Should return a score
  config: PromptOptimizationConfig = {}
): Promise<OptimizationResult & { score: number }> {
  const results: Array<{ prompt: string; score: number; metrics: PromptMetrics }> = [];
  
  // Test base prompt
  const baseScore = await testFunction(basePrompt);
  results.push({
    prompt: basePrompt,
    score: baseScore,
    metrics: calculateMetrics(basePrompt, config.model),
  });
  
  // Test variations
  for (const variation of variations) {
    const score = await testFunction(variation);
    results.push({
      prompt: variation,
      score,
      metrics: calculateMetrics(variation, config.model),
    });
  }
  
  // Find best performing variation
  const best = results.reduce((best, current) => 
    current.score > best.score ? current : best
  );
  
  const baseMetrics = calculateMetrics(basePrompt, config.model);
  
  return {
    originalPrompt: basePrompt,
    optimizedPrompt: best.prompt,
    improvements: ['Selected best performing variation from testing'],
    metrics: {
      before: baseMetrics,
      after: best.metrics,
    },
    confidence: 0.9, // High confidence since it's tested
    score: best.score,
  };
}

export default {
  estimateTokenCount,
  calculateClarity,
  calculateSpecificity,
  calculateComplexity,
  estimateCost,
  estimateLatency,
  calculateMetrics,
  optimizeForBrevity,
  optimizeForClarity,
  optimizeForModel,
  optimizePrompt,
  testPromptVariations,
};