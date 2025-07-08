import { AbstractClient, CompositionClient, streamResponse, DateOrISO } from '@/client';
import * as z from 'zod';
import { ZEvent, Event } from "@/types";

export default class APIEventId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(eventId: string): Promise<Event> {
    let res = await this._fetch({
      url: `/v1/events/${eventId}`,
      method: "GET",
      auth: ["HTTPBearer", "Master Key", "API Key", "Outlook Auth"],
    });
    if (res.headers.get("Content-Type") === "application/json") return ZEvent.parse(await res.json());
    throw new Error("Bad content type");
  }
  
}
