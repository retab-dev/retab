import { AbstractClient, CompositionClient } from '@/client';
import { BodyConvertToEmailDataAndUploadFileV1AutomationsOutlookConvertToEmailDataAndUploadFilePost, EmailDataOutput } from "@/types";

export default class APIConvertToEmailDataAndUploadFile extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ returnInlineBody, ...body }: { returnInlineBody?: boolean } & BodyConvertToEmailDataAndUploadFileV1AutomationsOutlookConvertToEmailDataAndUploadFilePost): Promise<EmailDataOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/convert_to_email_data_and_upload_file`,
      method: "POST",
      params: { "return_inline_body": returnInlineBody },
      headers: {  },
      body: body,
      bodyMime: "multipart/form-data",
    });
  }
  
}
