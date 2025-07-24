// ---------------------------------------------
// Quick example: Extract structured data using Retab's all-in-one `.parse()` method.
// ---------------------------------------------

import { Retab } from '@retab/node';
import { config } from 'dotenv';

// Load environment variables
config();

const retabApiKey = process.env.RETAB_API_KEY;
if (!retabApiKey) {
  throw new Error('Missing RETAB_API_KEY');
}

// Retab Setup
const reclient = new Retab({ api_key: retabApiKey });

// Document Extraction via Retab API
const response = await reclient.documents.extract({
  document: '../../assets/code/invoice.jpeg',
  model: 'gpt-4.1',
  json_schema: '../../assets/code/invoice_schema.json',
  modality: 'native',
  image_resolution_dpi: 96,
  browser_canvas: 'A4',
  temperature: 0.1,
  n_consensus: 4, // Number of parallel extraction runs for consensus voting. Values > 1 enable consensus mode.
});

// Output
console.log('CONSENSUS EXTRACTION RESULTS');

for (let i = 0; i < response.choices.length; i++) {
  console.log(`\nConsensus Result #${i + 1}:`);
  console.log('-'.repeat(40));
  try {
    const content = response.choices[i].message.content;
    if (content && (content.trim().startsWith('{') || content.trim().startsWith('['))) {
      const parsed = JSON.parse(content);
      console.log(JSON.stringify(parsed, null, 2));
    } else {
      console.log(content || 'No content');
    }
  } catch (error) {
    const content = response.choices[i].message.content;
    console.log(content || 'No content');
  }
}

// Display likelihoods with better formatting
if (response.likelihoods) {
  console.log('\nCONSENSUS LIKELIHOODS:');
  console.log('-'.repeat(40));

  // Handle both array and object formats for likelihoods
  if (Array.isArray(response.likelihoods)) {
    response.likelihoods.forEach((likelihood, i) => {
      console.log(`Result #${i + 1}: ${likelihood.toFixed(4)}`);
    });

    // Show the most confident result
    if (response.likelihoods.length > 0) {
      const maxLikelihood = Math.max(...response.likelihoods);
      const bestIdx = response.likelihoods.indexOf(maxLikelihood);
      console.log(`\n🎯 Most confident result: #${bestIdx + 1} (likelihood: ${maxLikelihood.toFixed(4)})`);
    }
  } else {
    // Handle object format
    for (const [key, likelihood] of Object.entries(response.likelihoods)) {
      const formattedLikelihood = typeof likelihood === 'number' ? likelihood.toFixed(4) : likelihood;
      console.log(`${key}: ${formattedLikelihood}`);
    }
  }
}