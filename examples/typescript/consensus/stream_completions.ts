import { AsyncRetab } from '@retab/node';
import { config } from 'dotenv';


async function main() {
  // ---------------------------------------------
  // Example: Streaming Consensus Completions
  // ---------------------------------------------
  
  
  // Load environment variables
  config();
  
  const client = new AsyncRetab();
  
  async function streamConsensusCompletions() {
    console.log('Starting consensus streaming completions example...\n');
    
    try {
      // Create a streaming consensus completion
      const stream = client.consensus.completions.stream({
        model: 'gpt-4o-mini',
        n_consensus: 3,
        temperature: 0.1,
        reasoning_effort: 'medium',
        messages: [
          {
            role: 'system',
            content: 'You are a helpful assistant that provides structured analysis.',
          },
          {
            role: 'user',
            content: 'Analyze the pros and cons of remote work and provide a structured summary.',
          },
        ],
        response_format: {
          type: 'json_schema',
          json_schema: {
            name: 'RemoteWorkAnalysis',
            schema: {
              type: 'object',
              properties: {
                pros: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'List of advantages of remote work',
                },
                cons: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'List of disadvantages of remote work',
                },
                overall_assessment: {
                  type: 'string',
                  description: 'Overall assessment of remote work',
                },
              },
              required: ['pros', 'cons', 'overall_assessment'],
            },
            strict: true,
          },
        },
      });
  
      console.log('Streaming consensus results:');
      console.log('==========================\n');
  
      // Process streaming results
      for await (const chunk of stream) {
        if (chunk.choices && chunk.choices[0]) {
          const content = chunk.choices[0].message?.content;
          if (content) {
            console.log('Chunk received:', content);
          }
        }
        
        // Show consensus metadata if available
        if (chunk.consensus_metadata) {
          console.log('Consensus metadata:', {
            n_consensus: chunk.consensus_metadata.n_consensus,
            agreement_score: chunk.consensus_metadata.agreement_score,
          });
        }
      }
  
      console.log('\n✅ Consensus streaming completed successfully!');
      
    } catch (error) {
      console.error('Error in consensus streaming:', error);
    }
  }
  
  async function demonstrateConsensusFeatures() {
    console.log('=== Consensus Streaming Features Demo ===\n');
    
    // Example 1: Stream completions
    await streamConsensusCompletions();
    
    console.log('\n' + '='.repeat(50) + '\n');
    
    // Example 2: Non-streaming consensus with parse
    try {
      console.log('Non-streaming consensus parse example:');
      
      const result = await client.consensus.completions.parse({
        model: 'gpt-4o-mini',
        n_consensus: 2,
        temperature: 0.0,
        messages: [
          {
            role: 'user',
            content: 'What are the main benefits of TypeScript over JavaScript?',
          },
        ],
        response_format: {
          type: 'json_schema',
          json_schema: {
            name: 'TypeScriptBenefits',
            schema: {
              type: 'object',
              properties: {
                benefits: {
                  type: 'array',
                  items: { type: 'string' },
                },
                summary: { type: 'string' },
              },
              required: ['benefits', 'summary'],
            },
          },
        },
      });
      
      console.log('Parsed result:', JSON.stringify(result, null, 2));
      
    } catch (error) {
      console.error('Error in consensus parse:', error);
    }
  }
  
  // Run the demo
  demonstrateConsensusFeatures();
  
}

main().catch(console.error);