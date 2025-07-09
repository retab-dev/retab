import * as fs from 'fs';
import { config } from 'dotenv';
import { AsyncRetab } from '@retab/node';



async function main() {
  config();
  
  const reclient = new AsyncRetab();
  
  const processor = await reclient.processors.create({
    name: "Invoice Processor",
    json_schema: JSON.parse(fs.readFileSync('../../assets/code/invoice_schema.json', 'utf-8')),
    model: "gpt-4o-mini",
    modality: "native",
    temperature: 0.1,
    reasoning_effort: "medium",
    n_consensus: 3,  // Enable consensus with 3 parallel runs
    image_resolution_dpi: 150,
  });
  
  console.log(JSON.stringify(processor, null, 2));
  // completion = await reclient.processors.submit({ processor_id: processor.id, document: "../../assets/code/invoice.jpeg" });
  
}

main().catch(console.error);