import { describe, expect, test } from 'bun:test';

import { ZClassification, ZEdit, ZExtractionV2, ZParse, ZPartition, ZSplit } from '../src/types.js';

describe('processing resource schemas', () => {
  test('processing resource schemas accept canonical normalized consensus payloads', () => {
    const classification = ZClassification.parse({
      id: 'clss_123',
      file: {
        id: 'file_123',
        filename: 'doc.pdf',
        mime_type: 'application/pdf',
      },
      model: 'retab-small',
      categories: [{ name: 'invoice', description: 'Invoice' }],
      n_consensus: 1,
      output: { reasoning: '', category: 'invoice' },
      consensus: { choices: [], likelihood: null },
      updated_at: '2026-05-20T10:00:00Z',
    });

    const extraction = ZExtractionV2.parse({
      id: 'extr_123',
      file: {
        id: 'file_123',
        filename: 'doc.pdf',
        mime_type: 'application/pdf',
      },
      model: 'retab-small',
      json_schema: { type: 'object' },
      n_consensus: 1,
      image_resolution_dpi: 192,
      output: { invoice_number: 'INV-001' },
      consensus: { choices: [], likelihoods: null },
      metadata: {},
      origin: { type: 'project', id: 'proj_123' },
      updated_at: '2026-05-20T10:00:00Z',
    });

    const parse = ZParse.parse({
      id: 'parse_123',
      file: {
        id: 'file_123',
        filename: 'doc.pdf',
        mime_type: 'application/pdf',
      },
      model: 'retab-small',
      table_parsing_format: 'markdown',
      image_resolution_dpi: 192,
      output: { pages: ['hello'], text: 'hello' },
      updated_at: '2026-05-20T10:00:00Z',
    });

    const split = ZSplit.parse({
      id: 'splt_123',
      file: {
        id: 'file_123',
        filename: 'doc.pdf',
        mime_type: 'application/pdf',
      },
      model: 'retab-small',
      subdocuments: [{ name: 'invoice', description: 'Invoice' }],
      n_consensus: 1,
      output: [{ name: 'invoice', pages: [1] }],
      consensus: null,
      updated_at: '2026-05-20T10:00:00Z',
    });

    const partition = ZPartition.parse({
      id: 'prtn_123',
      file: {
        id: 'file_123',
        filename: 'doc.pdf',
        mime_type: 'application/pdf',
      },
      model: 'retab-small',
      key: 'invoice_number',
      instructions: 'Group by invoice number.',
      n_consensus: 1,
      output: [{ key: 'INV-001', pages: [1] }],
      consensus: { choices: [], likelihoods: null },
      origin: { type: 'project', id: 'proj_123' },
      updated_at: '2026-05-20T10:00:00Z',
    });

    const edit = ZEdit.parse({
      id: 'edt_123',
      file: {
        id: 'file_123',
        filename: 'doc.pdf',
        mime_type: 'application/pdf',
      },
      model: 'retab-small',
      instructions: 'Fill it.',
      config: { color: '#000080' },
      data: {
        form_data: [],
        filled_document: {
          filename: 'filled.pdf',
          url: 'data:application/pdf;base64,AA==',
        },
      },
      usage: { credits: 1 },
      updated_at: '2026-05-20T10:00:00Z',
    });

    expect(classification.consensus.choices).toEqual([]);
    expect('updated_at' in classification).toBe(false);
    expect(extraction.consensus.choices).toEqual([]);
    expect('origin' in extraction).toBe(false);
    expect('updated_at' in extraction).toBe(false);
    expect('updated_at' in parse).toBe(false);
    expect(split.consensus).toBeNull();
    expect('updated_at' in split).toBe(false);
    expect(partition.consensus.choices).toEqual([]);
    expect('origin' in partition).toBe(false);
    expect('updated_at' in partition).toBe(false);
    expect('updated_at' in edit).toBe(false);
  });
});
