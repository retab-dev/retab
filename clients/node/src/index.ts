import APIV1 from "./api/client";
import { ClientOptions, FetcherClient } from "./client";

export * from "./types";
export * from "./client";



export class Retab extends APIV1 {
  constructor(options?: ClientOptions) {
    super(new FetcherClient(options));
  }
}
