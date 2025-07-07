import { config } from 'dotenv';
import { Retab } from '@retab/node';

// Load environment variables
config({ path: '../../../.env.production' });

const reclient = new Retab();

const automation = await reclient.processors.automations.endpoints.delete(
  "endp_DDw72RgWgIfKFXGaWXcGu"
);

console.log('Endpoint deleted successfully');