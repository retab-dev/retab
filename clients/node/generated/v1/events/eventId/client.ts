import { AbstractClient, CompositionClient } from '@/client';
import { Event } from "@/types";

export default class APIEventId extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }


  async get(eventId: string): Promise<Event> {
    return this._fetch({
      url: `/v1/events/${eventId}`,
      method: "GET",
      params: {  },
      headers: {  },
    });
  }
  
}
