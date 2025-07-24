// ---------------------------------------------
// Example: Prompt Optimization Usage
// ---------------------------------------------

import { optimizePrompt, calculateMetrics, testPromptVariations } from '@retab/node/utils';
import { config } from 'dotenv';

// Load environment variables
config();

async function demonstratePromptOptimization() {
  console.log('=== Prompt Optimization Demo ===\n');
  
  // Example 1: Basic prompt optimization
  const originalPrompt = `
    Please kindly analyze the following document and if you could provide a comprehensive summary that includes the main points, key findings, and important conclusions. I would really appreciate it if you could also highlight any recommendations or action items that are mentioned in the document. Thank you very much for your assistance with this task.
  `;
  
  console.log('1. Original prompt:');
  console.log(originalPrompt.trim());
  
  console.log('\n2. Optimizing for different goals...\n');
  
  // Optimize for different goals
  const optimizations = [
    { goal: 'accuracy', description: 'Enhanced clarity and structure' },
    { goal: 'speed', description: 'Reduced token count for faster processing' },
    { goal: 'cost', description: 'Minimized token usage to reduce costs' },
    { goal: 'balanced', description: 'Balanced optimization' },
  ];
  
  for (const { goal, description } of optimizations) {
    console.log(`--- ${goal.toUpperCase()} Optimization ---`);
    
    const result = optimizePrompt(originalPrompt.trim(), {
      optimizationGoal: goal,
      model: 'gpt-4o-mini',
    });
    
    console.log(`Optimized prompt:\n${result.optimizedPrompt}`);
    console.log(`\nImprovements: ${result.improvements.join(', ')}`);
    console.log(`Confidence: ${(result.confidence * 100).toFixed(1)}%`);
    
    console.log('\nMetrics comparison:');
    console.log(`  Token count: ${result.metrics.before.tokenCount} → ${result.metrics.after.tokenCount}`);
    console.log(`  Clarity: ${(result.metrics.before.clarity * 100).toFixed(1)}% → ${(result.metrics.after.clarity * 100).toFixed(1)}%`);
    console.log(`  Estimated cost: $${result.metrics.before.estimatedCost.toFixed(4)} → $${result.metrics.after.estimatedCost.toFixed(4)}`);
    console.log(`  Estimated latency: ${result.metrics.before.estimatedLatency}ms → ${result.metrics.after.estimatedLatency}ms`);
    
    console.log('\n' + '='.repeat(50) + '\n');
  }
}

async function demonstrateModelSpecificOptimization() {
  console.log('=== Model-Specific Optimization ===\n');
  
  const basePrompt = 'Analyze this data and provide insights with specific recommendations for improvement.';
  
  const models = ['gpt-4o', 'gpt-4o-mini', 'claude-3-opus', 'claude-3-haiku'];
  
  for (const model of models) {
    console.log(`--- ${model} Optimization ---`);
    
    const result = optimizePrompt(basePrompt, {
      model,
      optimizationGoal: 'balanced',
    });
    
    console.log(`Original: ${basePrompt}`);
    console.log(`Optimized: ${result.optimizedPrompt}`);
    
    const metrics = calculateMetrics(result.optimizedPrompt, model);
    console.log(`Metrics for ${model}:`);
    console.log(`  Tokens: ${metrics.tokenCount}`);
    console.log(`  Clarity: ${(metrics.clarity * 100).toFixed(1)}%`);
    console.log(`  Cost: $${metrics.estimatedCost.toFixed(4)}`);
    console.log(`  Latency: ${metrics.estimatedLatency}ms`);
    
    console.log('\n' + '-'.repeat(40) + '\n');
  }
}

async function demonstratePromptTesting() {
  console.log('=== Prompt Variation Testing ===\n');
  
  const basePrompt = 'Summarize this document';
  
  const variations = [
    'Provide a concise summary of this document',
    'Extract the key points from this document',
    'Create a brief overview of this document',
    'Summarize the main ideas in this document',
  ];
  
  // Mock test function that simulates performance scoring
  const mockTestFunction = async (prompt) => {
    // Simulate different performance scores based on prompt characteristics
    const metrics = calculateMetrics(prompt);
    
    // Mock scoring logic: balance between clarity and brevity
    const clarityScore = metrics.clarity * 40;
    const brevityScore = Math.max(0, 50 - metrics.tokenCount);
    const specificityScore = metrics.specificity * 10;
    
    return clarityScore + brevityScore + specificityScore;
  };
  
  console.log('Testing prompt variations...\n');
  
  try {
    const result = await testPromptVariations(
      basePrompt,
      variations,
      mockTestFunction,
      { model: 'gpt-4o-mini' }
    );
    
    console.log('Best performing prompt:');
    console.log(`"${result.optimizedPrompt}"`);
    console.log(`\nScore: ${result.score.toFixed(2)}`);
    console.log(`Confidence: ${(result.confidence * 100).toFixed(1)}%`);
    
    console.log('\nAll variations tested:');
    const allPrompts = [basePrompt, ...variations];
    for (const prompt of allPrompts) {
      const score = await mockTestFunction(prompt);
      console.log(`  "${prompt}" - Score: ${score.toFixed(2)}`);
    }
    
  } catch (error) {
    console.error('Error in prompt testing:', error);
  }
}

async function demonstrateMetricsCalculation() {
  console.log('\n=== Prompt Metrics Analysis ===\n');
  
  const testPrompts = [
    'Analyze data',
    'Please provide a comprehensive analysis of the data with detailed insights',
    'Analyze this data step-by-step: 1) identify patterns 2) extract insights 3) provide recommendations',
    'However, although the data appears complex, you must nevertheless provide a thorough analysis considering all variables',
  ];
  
  console.log('Analyzing metrics for different prompt styles:\n');
  
  testPrompts.forEach((prompt, index) => {
    console.log(`Prompt ${index + 1}: "${prompt}"`);
    
    const metrics = calculateMetrics(prompt);
    console.log('Metrics:');
    console.log(`  Token count: ${metrics.tokenCount}`);
    console.log(`  Clarity: ${(metrics.clarity * 100).toFixed(1)}%`);
    console.log(`  Specificity: ${(metrics.specificity * 100).toFixed(1)}%`);
    console.log(`  Complexity: ${(metrics.complexity * 100).toFixed(1)}%`);
    console.log(`  Estimated cost: $${metrics.estimatedCost.toFixed(4)}`);
    console.log(`  Estimated latency: ${metrics.estimatedLatency}ms`);
    
    console.log('\n' + '-'.repeat(40) + '\n');
  });
}

// Run the demos
async function runDemo() {
  await demonstratePromptOptimization();
  await demonstrateModelSpecificOptimization();
  await demonstratePromptTesting();
  await demonstrateMetricsCalculation();
  
  console.log('✅ Prompt optimization demo completed successfully!');
}

runDemo();