import { AbstractClient, CompositionClient } from '@/client';
import { EmailConversionRequest, EmailDataOutput } from "@/types";

export default class APIConvertEmlBytesToEmailModel extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: EmailConversionRequest): Promise<EmailDataOutput> {
    return this._fetch({
      url: `/v1/automations/outlook/convert_eml_bytes_to_email_model`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
  }
  
}
