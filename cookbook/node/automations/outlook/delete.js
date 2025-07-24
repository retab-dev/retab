import { config } from 'dotenv';
import { Retab } from '@retab/node';

config({ path: '../../../.env.production' });

const reclient = new Retab();

const automation = await reclient.processors.automations.outlook.delete({
  outlook_id: "outlook_DDw72RgWgIfKFXGaWXcGu",
});

console.log(JSON.stringify(automation, null, 2));