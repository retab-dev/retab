import { Retab } from '@retab/node';
import { config } from 'dotenv';

config();

const client = new Retab();

const result = await client.documents.parse({
    document: "document.pdf",
    model: "gemini-2.5.flash",
    table_parsing_format: "html",
    image_resolution_dpi: 72,
    browser_canvas: "A4"
});

// Access parsed content
result.pages.forEach((pageContent, index) => {
    console.log(`Page ${index + 1}: ${pageContent}`);
});

console.log(`Total pages: ${result.usage.page_count}`);
console.log(`Credits used: ${result.usage.credits}`);
