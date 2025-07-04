import { UiFormClient } from "@/index";
import fs from "fs";

// Bearer token authentication:
// UIFORM_BEARER_TOKEN=<token> or new UiFormClient({ auth: { bearer: "<token>" } })
// Master key authentication:
// UIFORM_MASTER_KEY=<key> or new UiFormClient({ auth: { masterKey: "<key>" } })
// Api key authentication:
// UIFORM_API_KEY=<key> or new UiFormClient({ auth: { apiKey: "<key>" } })
const client = new UiFormClient().v1;

(async () => {
  console.log(await client.secrets.apiKeys.get());

  let data = fs.readFileSync("test.pdf");
  let base64 = data.toString("base64");
  let gen = await client.documents.extractions.stream.post({
    json_schema: {
      type: "object",
      properties: {
        summary: {
          type: "string",
          description: "Summary of the document",
        },
      }
    },
    model: "gpt-4.1-mini",
    modality: "native",
    documents: [{
      filename: "test.pdf",
      url: "data:application/pdf;base64," + base64,
    }],
    stream: true,
  });
  console.log("Streaming...");
  for await (const chunk of gen) {
    console.log("Chunk: ", JSON.stringify(chunk, null, 2));
  }
  console.log("Done");
})();
