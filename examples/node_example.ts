import { readMIMEDataFromFile } from 'mimelib/io';
import { UiForm } from 'uiform';
import { RoadBookingConfirmationData } from 'cubeblock/dataclasses/road/booking/structures';

const UIFORM_API_KEY = "your_UIFORM_API_KEY";
const client = new UiForm({
    apiKey: UIFORM_API_KEY
});

const schema = RoadBookingConfirmationData.modelJsonSchema();

const pdfDocument = await readMIMEDataFromFile("path/to/your/pdf/file.pdf");

const documents = [pdfDocument];

const regexInstructions = [{
    name: "vat_number",
    pattern: "\\b[A-Z]{2}\\d{9}\\b",
    description: "VAT number in the format XX999999999"
}];

const TextOperations = {
    regexInstructions: regexInstructions
};

try {
    const response = await client.processDocument({
        schema,
        documents,
        model: "gpt-4o-mini",
        temperature: 0,
        TextOperations
    });
    console.log(response);
} catch (error) {
    console.error(error);
}

