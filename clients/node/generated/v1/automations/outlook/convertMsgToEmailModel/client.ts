import { AbstractClient, CompositionClient, streamResponse } from '@/client';
import { BodyConvertMsgToEmailModelV1AutomationsOutlookConvertMsgToEmailModelPost, EmailDataOutput } from "@/types";

export default class APIConvertMsgToEmailModel extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ returnInlineBody, ...body }: { returnInlineBody?: boolean } & BodyConvertMsgToEmailModelV1AutomationsOutlookConvertMsgToEmailModelPost): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/automations/outlook/convert_msg_to_email_model`,
      method: "POST",
      params: { "return_inline_body": returnInlineBody },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return res.json();
    throw new Error("Bad content type");
  }
  
}
