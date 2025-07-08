import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZBodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost, BodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost, ZEmailDataOutput, EmailDataOutput } from "@/types";

export default class APIConvertMsgToEmailModel extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ returnInlineBody, ...body }: { returnInlineBody?: boolean } & BodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/convert_msg_to_email_model`,
      method: "POST",
      params: { "return_inline_body": returnInlineBody },
      body: body,
      bodyMime: "multipart/form-data",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEmailDataOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
