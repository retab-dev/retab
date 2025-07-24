import { AsyncRetab } from '@retab/node';
import { config } from 'dotenv';


async function main() {
  // ---------------------------------------------
  // Example: Streaming Consensus Responses
  // ---------------------------------------------
  
  
  // Load environment variables
  config();
  
  const client = new AsyncRetab();
  
  async function streamConsensusResponses() {
    console.log('Starting consensus streaming responses example...\n');
    
    try {
      // Create a streaming consensus response
      const stream = client.consensus.responses.stream({
        model: 'gpt-4o-mini',
        n_consensus: 3,
        temperature: 0.2,
        reasoning_effort: 'high',
        messages: [
          {
            role: 'system',
            content: 'You are an expert data analyst providing insights.',
          },
          {
            role: 'user',
            content: 'Analyze the trend of AI adoption in different industries and provide key insights.',
          },
        ],
        response_format: {
          type: 'json_schema',
          json_schema: {
            name: 'AIAdoptionAnalysis',
            schema: {
              type: 'object',
              properties: {
                industries: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      name: { type: 'string' },
                      adoption_level: { type: 'string', enum: ['low', 'medium', 'high'] },
                      key_applications: { type: 'array', items: { type: 'string' } },
                    },
                    required: ['name', 'adoption_level', 'key_applications'],
                  },
                },
                trends: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'Key trends in AI adoption',
                },
                predictions: {
                  type: 'string',
                  description: 'Future predictions for AI adoption',
                },
              },
              required: ['industries', 'trends', 'predictions'],
            },
            strict: true,
          },
        },
      });
  
      console.log('Streaming consensus responses:');
      console.log('=============================\n');
  
      let completeResponse = '';
      
      // Process streaming results
      for await (const chunk of stream) {
        if (chunk.choices && chunk.choices[0]) {
          const content = chunk.choices[0].message?.content;
          if (content) {
            completeResponse += content;
            console.log('Streaming chunk:', content);
          }
        }
        
        // Show consensus metadata if available
        if (chunk.consensus_metadata) {
          console.log('Consensus metadata:', {
            n_consensus: chunk.consensus_metadata.n_consensus,
            agreement_score: chunk.consensus_metadata.agreement_score,
            reconciliation_method: chunk.consensus_metadata.reconciliation_method,
          });
        }
      }
  
      console.log('\n✅ Final response:', completeResponse);
      console.log('\n✅ Consensus streaming completed successfully!');
      
    } catch (error) {
      console.error('Error in consensus streaming:', error);
    }
  }
  
  async function demonstrateStreamParse() {
    console.log('\n=== Stream Parse Example ===\n');
    
    try {
      // Stream with automatic parsing
      const stream = client.consensus.responses.streamParse({
        model: 'gpt-4o-mini',
        n_consensus: 2,
        temperature: 0.1,
        messages: [
          {
            role: 'user',
            content: 'List 5 popular programming languages and their main use cases.',
          },
        ],
        response_format: {
          type: 'json_schema',
          json_schema: {
            name: 'ProgrammingLanguages',
            schema: {
              type: 'object',
              properties: {
                languages: {
                  type: 'array',
                  items: {
                    type: 'object',
                    properties: {
                      name: { type: 'string' },
                      use_cases: { type: 'array', items: { type: 'string' } },
                    },
                    required: ['name', 'use_cases'],
                  },
                },
              },
              required: ['languages'],
            },
          },
        },
      });
  
      console.log('Streaming with parse:');
      console.log('====================\n');
  
      for await (const chunk of stream) {
        if (chunk.choices && chunk.choices[0]) {
          console.log('Parsed chunk:', chunk.choices[0].message?.content);
        }
      }
  
      console.log('\n✅ Stream parse completed successfully!');
      
    } catch (error) {
      console.error('Error in stream parse:', error);
    }
  }
  
  async function demonstrateConsensusResponses() {
    console.log('=== Consensus Responses Features Demo ===\n');
    
    // Example 1: Stream responses
    await streamConsensusResponses();
    
    console.log('\n' + '='.repeat(50) + '\n');
    
    // Example 2: Stream parse
    await demonstrateStreamParse();
    
    console.log('\n' + '='.repeat(50) + '\n');
    
    // Example 3: Non-streaming consensus response
    try {
      console.log('Non-streaming consensus response example:');
      
      const result = await client.consensus.responses.create({
        model: 'gpt-4o-mini',
        n_consensus: 3,
        temperature: 0.0,
        messages: [
          {
            role: 'user',
            content: 'What are the key principles of good software design?',
          },
        ],
      });
      
      console.log('Response result:', JSON.stringify(result, null, 2));
      
    } catch (error) {
      console.error('Error in consensus response:', error);
    }
  }
  
  // Run the demo
  demonstrateConsensusResponses();
  
}

main().catch(console.error);