import { UiFormClient } from "@/index";

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
})();
