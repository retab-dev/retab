const APIV1 = require("./api/client").default;
const { FetcherClient } = require("./client");

export * from "./types";
export * from "./client";

export interface ClientOptions {
  baseUrl?: string;
  apiKey?: string;
  bearer?: string;
  masterKey?: string;
}

export class Retab extends APIV1 {
  constructor(options?: ClientOptions) {
    super(new FetcherClient(options));
  }
}
