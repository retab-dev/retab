import { AutomationConfig } from '../logs.js';
import { Modality } from '../modalities.js';
import { BrowserCanvas } from '../browser_canvas.js';

export interface CronSchedule {
  second?: number; // 0-59, defaults to 0
  minute: number; // 0-59
  hour: number; // 0-23
  day_of_month?: number; // 1-31, undefined means any day
  month?: number; // 1-12, undefined means every month
  day_of_week?: number; // 0-6, Sunday = 0, undefined means any day
}

export interface ScrappingConfig extends AutomationConfig {
  object: 'automation.scrapping_cron';
  id: string;
  link: string;
  schedule: CronSchedule;
  updated_at: string;
  webhook_url: string;
  webhook_headers: Record<string, string>;
  modality: Modality;
  image_resolution_dpi: number;
  browser_canvas: BrowserCanvas;
  model: string;
  json_schema: Record<string, any>;
  temperature: number;
  reasoning_effort: 'low' | 'medium' | 'high';
}