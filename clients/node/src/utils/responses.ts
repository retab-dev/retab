export interface ChatMessage {
  role: 'system' | 'user' | 'assistant' | 'developer';
  content: string | ChatContentPart[];
}

export interface ChatContentPart {
  type: 'text' | 'image_url' | 'input_audio';
  text?: string;
  image_url?: {
    url: string;
    detail?: 'low' | 'high' | 'auto';
  };
  input_audio?: {
    data: string;
    format: string;
  };
}

export interface ParsedChatCompletion {
  id: string;
  object: 'chat.completion';
  created: number;
  model: string;
  choices: Array<{
    index: number;
    message: {
      role: 'assistant';
      content: string;
    };
    finish_reason: string;
  }>;
  usage?: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
  likelihoods?: Record<string, any>;
}

export function convertToOpenAIFormat(messages: ChatMessage[]): any[] {
  return messages.map(message => ({
    role: message.role,
    content: message.content,
  }));
}

export function convertFromOpenAIFormat(messages: any[]): ChatMessage[] {
  return messages.map(message => ({
    role: message.role,
    content: message.content,
  }));
}

export function separateMessages(messages: ChatMessage[]): {
  systemMessage?: ChatMessage;
  userMessages: ChatMessage[];
  assistantMessages: ChatMessage[];
} {
  let systemMessage: ChatMessage | undefined;
  const userMessages: ChatMessage[] = [];
  const assistantMessages: ChatMessage[] = [];

  for (const message of messages) {
    if (message.role === 'system' || message.role === 'developer') {
      systemMessage = message;
    } else if (message.role === 'user') {
      userMessages.push(message);
    } else if (message.role === 'assistant') {
      assistantMessages.push(message);
    }
  }

  return { systemMessage, userMessages, assistantMessages };
}

export function stringifyMessages(messages: ChatMessage[], maxLength: number = 100): string {
  const truncate = (text: string, maxLen: number): string => {
    return text.length <= maxLen ? text : `${text.substring(0, maxLen)}...`;
  };

  const serialized = messages.map(message => {
    const { role, content } = message;

    if (typeof content === 'string') {
      return { role, content: truncate(content, maxLength) };
    } else if (Array.isArray(content)) {
      const truncatedContent = content.map(part => {
        if (part.type === 'text' && part.text) {
          return { ...part, text: truncate(part.text, maxLength) };
        } else if (part.type === 'image_url' && part.image_url) {
          return {
            ...part,
            image_url: {
              ...part.image_url,
              url: truncate(part.image_url.url, maxLength)
            }
          };
        }
        return part;
      });
      return { role, content: truncatedContent };
    }

    return { role, content };
  });

  return JSON.stringify(serialized, null, 2);
}