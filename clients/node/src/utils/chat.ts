/**
 * Chat message processing utilities
 * Equivalent to Python's utils/chat.py
 */

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'function' | 'tool';
  content: string | null;
  name?: string;
  function_call?: {
    name: string;
    arguments: string;
  };
  tool_calls?: Array<{
    id: string;
    type: 'function';
    function: {
      name: string;
      arguments: string;
    };
  }>;
  tool_call_id?: string;
}

export interface ChatCompletionRequest {
  model: string;
  messages: ChatMessage[];
  temperature?: number;
  max_tokens?: number;
  top_p?: number;
  frequency_penalty?: number;
  presence_penalty?: number;
  stop?: string | string[];
  stream?: boolean;
  functions?: Array<{
    name: string;
    description?: string;
    parameters: Record<string, any>;
  }>;
  function_call?: 'auto' | 'none' | { name: string };
  tools?: Array<{
    type: 'function';
    function: {
      name: string;
      description?: string;
      parameters: Record<string, any>;
    };
  }>;
  tool_choice?: 'auto' | 'none' | { type: 'function'; function: { name: string } };
}

export function formatMessagesForProvider(
  messages: ChatMessage[],
  provider: 'openai' | 'anthropic' | 'xai' | 'gemini'
): any[] {
  switch (provider) {
    case 'openai':
    case 'xai':
      return messages;
    
    case 'anthropic':
      return messages.map(msg => ({
        role: msg.role === 'system' ? 'user' : msg.role,
        content: msg.content,
      }));
    
    case 'gemini':
      return messages.map(msg => ({
        role: msg.role === 'assistant' ? 'model' : 'user',
        parts: [{ text: msg.content }],
      }));
    
    default:
      return messages;
  }
}

export function extractSystemPrompt(messages: ChatMessage[]): { system: string | null; filtered: ChatMessage[] } {
  const systemMessages = messages.filter(msg => msg.role === 'system');
  const nonSystemMessages = messages.filter(msg => msg.role !== 'system');
  
  const systemPrompt = systemMessages.length > 0 ? 
    systemMessages.map(msg => msg.content).join('\n') : null;
  
  return {
    system: systemPrompt,
    filtered: nonSystemMessages,
  };
}

export function validateMessages(messages: ChatMessage[]): string[] {
  const errors: string[] = [];
  
  if (!Array.isArray(messages) || messages.length === 0) {
    errors.push('Messages array is required and cannot be empty');
    return errors;
  }
  
  for (let i = 0; i < messages.length; i++) {
    const msg = messages[i];
    
    if (!msg.role) {
      errors.push(`Message at index ${i} is missing role`);
    }
    
    if (!['system', 'user', 'assistant', 'function', 'tool'].includes(msg.role)) {
      errors.push(`Message at index ${i} has invalid role: ${msg.role}`);
    }
    
    if (msg.content === null && !msg.function_call && !msg.tool_calls) {
      errors.push(`Message at index ${i} must have content, function_call, or tool_calls`);
    }
  }
  
  return errors;
}

export function countTokensInMessages(messages: ChatMessage[], _model: string = 'gpt-4o-mini'): number {
  // Simplified token counting - in production use tiktoken
  let totalTokens = 0;
  
  for (const message of messages) {
    // Role tokens
    totalTokens += 4; // Base tokens per message
    
    // Content tokens
    if (message.content) {
      totalTokens += Math.ceil(message.content.length / 4); // ~4 chars per token
    }
    
    // Function/tool call tokens
    if (message.function_call) {
      totalTokens += Math.ceil(JSON.stringify(message.function_call).length / 4);
    }
    
    if (message.tool_calls) {
      totalTokens += Math.ceil(JSON.stringify(message.tool_calls).length / 4);
    }
  }
  
  return totalTokens;
}

export default {
  formatMessagesForProvider,
  extractSystemPrompt,
  validateMessages,
  countTokensInMessages,
};