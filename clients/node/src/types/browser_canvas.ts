import { z } from 'zod';

export const BrowserCanvasSchema = z.enum(['A3', 'A4', 'A5']);
export type BrowserCanvas = z.infer<typeof BrowserCanvasSchema>;