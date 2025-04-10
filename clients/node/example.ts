import { UiFormClient } from "@/index";
import req from "./req.json";

// Bearer token authentication:
// UIFORM_BEARER_TOKEN=<token> or new UiFormClient({ auth: { bearer: "<token>" } })
// Master key authentication:
// UIFORM_MASTER_KEY=<key> or new UiFormClient({ auth: { masterKey: "<key>" } })
// Api key authentication:
// UIFORM_API_KEY=<key> or new UiFormClient({ auth: { apiKey: "<key>" } })
const client = new UiFormClient().v1;

(async () => {
  console.log(await client.secrets.apiKeys.get());
  console.log(await client.iam.payments.subscriptionStatus.get());
  console.log(await client.db.files.get());

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
    model: "receipt",
    modality: "text",
    document: {
      filename: "test.pdf",
      url: "data:application/pdf;base64,...",
    },
  });
  console.log("Streaming...");
  for await (const chunk of gen) {
    console.log("Chunk: ", JSON.stringify(chunk, null, 2));
  }
  console.log("Done");
})();
