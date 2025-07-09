/**
 * Usage tracking and analytics utilities
 * Equivalent to Python's utils/usage/usage.py
 */

export interface UsageMetrics {
  timestamp: Date;
  organization_id: string;
  user_id?: string;
  operation_type: 'extraction' | 'generation' | 'analysis' | 'batch' | 'streaming';
  model_used: string;
  provider: string;
  tokens_used: {
    input: number;
    output: number;
    total: number;
  };
  cost_usd: number;
  latency_ms: number;
  success: boolean;
  error_type?: string;
  request_id: string;
  metadata?: Record<string, any>;
}

export interface UsageAggregation {
  period: 'hour' | 'day' | 'week' | 'month';
  start_time: Date;
  end_time: Date;
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  total_tokens: number;
  total_cost_usd: number;
  average_latency_ms: number;
  operations_by_type: Record<string, number>;
  models_by_usage: Record<string, {
    requests: number;
    tokens: number;
    cost: number;
  }>;
  providers_by_usage: Record<string, {
    requests: number;
    tokens: number;
    cost: number;
  }>;
  error_breakdown: Record<string, number>;
}

export class UsageTracker {
  private metrics: UsageMetrics[] = [];

  recordUsage(metrics: Omit<UsageMetrics, 'timestamp'>): void {
    const usage: UsageMetrics = {
      ...metrics,
      timestamp: new Date(),
    };
    
    this.metrics.push(usage);
  }

  getUsageAggregation(
    organizationId: string,
    period: 'hour' | 'day' | 'week' | 'month',
    startTime?: Date,
    endTime?: Date
  ): UsageAggregation {
    const now = new Date();
    const start = startTime || this.getPeriodStart(period, now);
    const end = endTime || now;

    const relevantMetrics = this.metrics.filter(m => 
      m.organization_id === organizationId &&
      m.timestamp >= start &&
      m.timestamp <= end
    );

    const totalRequests = relevantMetrics.length;
    const successfulRequests = relevantMetrics.filter(m => m.success).length;
    const failedRequests = totalRequests - successfulRequests;

    const totalTokens = relevantMetrics.reduce((sum, m) => sum + m.tokens_used.total, 0);
    const totalCost = relevantMetrics.reduce((sum, m) => sum + m.cost_usd, 0);
    const averageLatency = totalRequests > 0 ? 
      relevantMetrics.reduce((sum, m) => sum + m.latency_ms, 0) / totalRequests : 0;

    const operationsByType: Record<string, number> = {};
    relevantMetrics.forEach(m => {
      operationsByType[m.operation_type] = (operationsByType[m.operation_type] || 0) + 1;
    });

    const modelsByUsage: Record<string, { requests: number; tokens: number; cost: number }> = {};
    relevantMetrics.forEach(m => {
      if (!modelsByUsage[m.model_used]) {
        modelsByUsage[m.model_used] = { requests: 0, tokens: 0, cost: 0 };
      }
      modelsByUsage[m.model_used].requests++;
      modelsByUsage[m.model_used].tokens += m.tokens_used.total;
      modelsByUsage[m.model_used].cost += m.cost_usd;
    });

    const providersByUsage: Record<string, { requests: number; tokens: number; cost: number }> = {};
    relevantMetrics.forEach(m => {
      if (!providersByUsage[m.provider]) {
        providersByUsage[m.provider] = { requests: 0, tokens: 0, cost: 0 };
      }
      providersByUsage[m.provider].requests++;
      providersByUsage[m.provider].tokens += m.tokens_used.total;
      providersByUsage[m.provider].cost += m.cost_usd;
    });

    const errorBreakdown: Record<string, number> = {};
    relevantMetrics.filter(m => !m.success).forEach(m => {
      const errorType = m.error_type || 'unknown';
      errorBreakdown[errorType] = (errorBreakdown[errorType] || 0) + 1;
    });

    return {
      period,
      start_time: start,
      end_time: end,
      total_requests: totalRequests,
      successful_requests: successfulRequests,
      failed_requests: failedRequests,
      total_tokens: totalTokens,
      total_cost_usd: totalCost,
      average_latency_ms: averageLatency,
      operations_by_type: operationsByType,
      models_by_usage: modelsByUsage,
      providers_by_usage: providersByUsage,
      error_breakdown: errorBreakdown,
    };
  }

  private getPeriodStart(period: 'hour' | 'day' | 'week' | 'month', from: Date): Date {
    const start = new Date(from);
    
    switch (period) {
      case 'hour':
        start.setMinutes(0, 0, 0);
        break;
      case 'day':
        start.setHours(0, 0, 0, 0);
        break;
      case 'week':
        start.setDate(start.getDate() - start.getDay());
        start.setHours(0, 0, 0, 0);
        break;
      case 'month':
        start.setDate(1);
        start.setHours(0, 0, 0, 0);
        break;
    }
    
    return start;
  }
}

// Export class but not types in default export
export default {
  UsageTracker,
};