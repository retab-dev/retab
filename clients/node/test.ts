import { Retab } from '@retab/node';
import { config } from 'dotenv';

config();

const client = new Retab();

const response = await client.documents.extract({
    documents: ["Invoice.pdf"],
    modality: "native",
    model: "gpt-4o-mini",
    json_schema: "Invoice_schema.json",
    temperature: 0,
});

console.log(JSON.stringify(response, null, 2));
