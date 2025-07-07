export interface ReconciliationRequest {
  list_dicts: Record<string, any>[];
  reference_schema?: Record<string, any>;
  mode?: 'direct' | 'aligned';
}

export interface ReconciliationResponse {
  consensus_dict: Record<string, any>;
  likelihoods: Record<string, any>;
}