import { Retab } from '@retab/node';
import { config } from 'dotenv';



async function main() {
  // Load environment variables
  config();
  
  const client = new Retab();
  
  // Multiple extraction results to reconcile
  const extractions = [
    {
      title: 'Quantum Algorithms in Interstellar Navigation',
      authors: ['Dr. Stella Voyager', 'Dr. Nova Star', 'Dr. Lyra Hunter'],
      year: 2025,
      keywords: ['quantum computing', 'space navigation', 'algorithms'],
    },
    {
      title: 'Quantum Algorithms for Interstellar Navigation',
      authors: ['Dr. S. Voyager', 'Dr. N. Star', 'Dr. L. Hunter'],
      year: 2025,
      keywords: ['quantum algorithms', 'interstellar navigation', 'space travel'],
    },
    {
      title: 'Application of Quantum Algorithms in Space Navigation',
      authors: ['Stella Voyager', 'Nova Star', 'Lyra Hunter'],
      year: 2025,
      keywords: ['quantum computing', 'navigation', 'space exploration'],
    },
  ];
  
  // Reconcile the different extraction results into a consensus
  const response = await client.consensus.reconcile({ 
    list_dicts: extractions, 
    mode: 'aligned' 
  });
  
  const consensusResult = response.consensus_dict;
  const consensusConfidence = response.likelihoods;
  
  console.log(`Consensus: ${JSON.stringify(consensusResult, null, 2)}`);
  console.log(`Confidence scores: ${JSON.stringify(consensusConfidence, null, 2)}`);
  
}

main().catch(console.error);