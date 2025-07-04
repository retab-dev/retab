import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyConvertToEmailDataAndUploadFileV1ProcessorsAutomationsOutlookConvertToEmailDataAndUploadFilePost, EmailDataOutput } from "@/types";

export default class APIConvertToEmailDataAndUploadFile extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ returnInlineBody, ...body }: { returnInlineBody?: boolean } & BodyConvertToEmailDataAndUploadFileV1ProcessorsAutomationsOutlookConvertToEmailDataAndUploadFilePost): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/convert_to_email_data_and_upload_file`,
      method: "POST",
      params: { "return_inline_body": returnInlineBody },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json() as any;
    throw new Error("Bad content type");
  }
  
}
