import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyConvertToEmailDataAndUploadFileV1AutomationsOutlookConvertToEmailDataAndUploadFilePost, EmailDataOutput } from "@/types";

export default class APIConvertToEmailDataAndUploadFile extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ returnInlineBody, ...body }: { returnInlineBody?: boolean } & BodyConvertToEmailDataAndUploadFileV1AutomationsOutlookConvertToEmailDataAndUploadFilePost): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/convert_to_email_data_and_upload_file`,
      method: "POST",
      params: { "return_inline_body": returnInlineBody },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
