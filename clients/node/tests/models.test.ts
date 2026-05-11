import { describe, expect, test } from 'bun:test';

import { ZClassification, ZExtractionV2, ZPartition, ZSplit } from '../src/types.js';

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
    });

    expect(classification.consensus.choices).toEqual([]);
    expect(extraction.consensus.choices).toEqual([]);
    expect(split.consensus).toBeNull();
    expect(partition.consensus.choices).toEqual([]);
  });
});
