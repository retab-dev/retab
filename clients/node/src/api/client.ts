import { AbstractClient, CompositionClient } from '../client.js';
import APISchemas from './schemas/client';
import APIExtractions from './extractions/client';
import APIWorkflows from './workflows/client';
import APIFiles from './files/client';
import APIJobs from './jobs/client';
import APIParses from './parses/client';
import APIClassifications from './classifications/client';
import APISplits from './splits/client';
import APIPartitions from './partitions/client';
import APIEdits from './edits/client';

export default class APIV1 extends CompositionClient {
  constructor(client: AbstractClient) {
    super(client);
  }

  files = new APIFiles(this);
  schemas = new APISchemas(this);
  extractions = new APIExtractions(this);
  workflows = new APIWorkflows(this);
  jobs = new APIJobs(this);
  parses = new APIParses(this);
  classifications = new APIClassifications(this);
  splits = new APISplits(this);
  partitions = new APIPartitions(this);
  edits = new APIEdits(this);
}
