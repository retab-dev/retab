import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZEmailConversionRequest, EmailConversionRequest, ZEmailDataOutput, EmailDataOutput } from "@/types";

export default class APIConvertEmlBytesToEmailModel extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async post({ ...body }: EmailConversionRequest): Promise<EmailDataOutput> {
    let res = await this._fetch({
      url: `/v1/processors/automations/outlook/convert_eml_bytes_to_email_model`,
      method: "POST",
      body: body,
      bodyMime: "application/json",
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEmailDataOutput.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
