import { config } from 'dotenv';
import { Retab } from '@retab/node';

config({ path: '../../../.env.production' });

const reclient = new Retab();

const automation = await reclient.processors.automations.mailboxes.delete({
  mailbox_id: "mb_FRf6FX5fYkenZ_JJlL5GD",
});

console.log(JSON.stringify(automation, null, 2));