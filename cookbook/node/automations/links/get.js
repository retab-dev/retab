import { config } from 'dotenv';
import { Retab } from '@retab/node';

config({ path: '../../../.env.production' });

const reclient = new Retab();

const automation = await reclient.processors.automations.links.get({
  link_id: "lnk_Xf15nXFpo7mwfGT1aSYo4",
});

console.log(JSON.stringify(automation, null, 2));