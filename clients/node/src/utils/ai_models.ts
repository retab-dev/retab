import fs from 'fs';
import path from 'path';
import yaml from 'js-yaml';
import { fileURLToPath } from 'url';
import { 
  AIProvider, 
  ModelCard,
  ModelCardSchema
} from '../types/ai_models.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const MODEL_CARDS_DIR = path.join(__dirname, '_model_cards');

function mergeModelCards(base: Record<string, any>, override: Record<string, any>): Record<string, any> {
  const result = { ...base };
  for (const [key, value] of Object.entries(override)) {
    if (key === 'inherits') {
      continue;
    }
    if (typeof value === 'object' && value !== null && key in result && typeof result[key] === 'object') {
      result[key] = mergeModelCards(result[key], value);
    } else {
      result[key] = value;
    }
  }
  return result;
}

function loadModelCards(yamlFile: string): ModelCard[] {
  const yamlContent = fs.readFileSync(yamlFile, 'utf-8');
  const rawCards = yaml.load(yamlContent) as any[];
  const nameToCard: Record<string, any> = {};
  
  // First pass: collect base cards
  for (const card of rawCards) {
    if (!('inherits' in card)) {
      nameToCard[card.model] = card;
    }
  }

  const finalCards: ModelCard[] = [];
  for (const card of rawCards) {
    if ('inherits' in card) {
      const parent = nameToCard[card.inherits];
      const merged = mergeModelCards(parent, card);
      finalCards.push(ModelCardSchema.parse(merged));
    } else {
      finalCards.push(ModelCardSchema.parse(card));
    }
  }
  return finalCards;
}

// Create model cards directory structure if it doesn't exist
if (!fs.existsSync(MODEL_CARDS_DIR)) {
  fs.mkdirSync(MODEL_CARDS_DIR, { recursive: true });
  
  // Create basic model card files
  const openaiCards = [
    {
      model: 'gpt-4o',
      pricing: {
        text: { prompt: 2.5, completion: 10.0 },
      },
      capabilities: {
        modalities: ['text', 'image'],
        endpoints: ['chat_completions'],
        features: ['streaming', 'function_calling', 'structured_outputs'],
      },
    },
    {
      model: 'gpt-4o-mini',
      pricing: {
        text: { prompt: 0.15, completion: 0.6 },
      },
      capabilities: {
        modalities: ['text', 'image'],
        endpoints: ['chat_completions'],
        features: ['streaming', 'function_calling', 'structured_outputs'],
      },
    },
  ];

  fs.writeFileSync(path.join(MODEL_CARDS_DIR, 'openai.yaml'), yaml.dump(openaiCards));

  const anthropicCards = [
    {
      model: 'claude-3-5-sonnet-latest',
      pricing: {
        text: { prompt: 3.0, completion: 15.0 },
      },
      capabilities: {
        modalities: ['text', 'image'],
        endpoints: ['chat_completions'],
        features: ['streaming'],
      },
    },
  ];

  fs.writeFileSync(path.join(MODEL_CARDS_DIR, 'anthropic.yaml'), yaml.dump(anthropicCards));

  // Create empty files for other providers
  fs.writeFileSync(path.join(MODEL_CARDS_DIR, 'xai.yaml'), yaml.dump([]));
  fs.writeFileSync(path.join(MODEL_CARDS_DIR, 'gemini.yaml'), yaml.dump([]));
  fs.writeFileSync(path.join(MODEL_CARDS_DIR, 'auto.yaml'), yaml.dump([]));
}

// Load all model cards
let modelCards: ModelCard[] = [];
const modelCardsDict: Record<string, ModelCard> = {};

try {
  const cardFiles = ['openai.yaml', 'anthropic.yaml', 'xai.yaml', 'gemini.yaml', 'auto.yaml'];
  for (const file of cardFiles) {
    const filePath = path.join(MODEL_CARDS_DIR, file);
    if (fs.existsSync(filePath)) {
      modelCards = [...modelCards, ...loadModelCards(filePath)];
    }
  }
  
  for (const card of modelCards) {
    modelCardsDict[card.model as string] = card;
  }
} catch (error) {
  console.warn('Failed to load model cards:', error);
}

export function getModelFromModelId(modelId: string): string {
  if (modelId.startsWith('ft:')) {
    const parts = modelId.split(':');
    return parts[1];
  }
  return modelId;
}

export function getModelCard(model: string): ModelCard {
  const modelName = getModelFromModelId(model);
  if (modelName in modelCardsDict) {
    const modelCard = ModelCardSchema.parse({ ...modelCardsDict[modelName] });
    if (modelName !== model) {
      // Fine-tuned model -> Change the name
      modelCard.model = model;
      // Remove the fine-tuning feature (if exists)
      const features = modelCard.capabilities.features;
      const index = features.indexOf('fine_tuning');
      if (index > -1) {
        features.splice(index, 1);
      }
    }
    return modelCard;
  }

  throw new Error(`No model card found for model: ${modelName}`);
}

export function getProviderForModel(modelId: string): AIProvider {
  const modelName = getModelFromModelId(modelId);
  
  // Check OpenAI models
  if (modelName.startsWith('gpt-') || modelName.startsWith('o1') || modelName.startsWith('o3') || modelName.startsWith('o4')) {
    return 'OpenAI';
  }
  
  // Check Anthropic models
  if (modelName.startsWith('claude-')) {
    return 'Anthropic';
  }
  
  // Check xAI models
  if (modelName.startsWith('grok-')) {
    return 'xAI';
  }
  
  // Check Gemini models
  if (modelName.startsWith('gemini-')) {
    return 'Gemini';
  }
  
  // Check Retab models
  if (modelName.startsWith('auto-')) {
    return 'Retab';
  }
  
  throw new Error(`Unknown provider for model: ${modelName}`);
}

export function assertValidModelExtraction(model: string): void {
  if (!model || typeof model !== 'string') {
    throw new Error('Valid model must be provided for extraction');
  }
  
  // Additional validation logic can be added here
  try {
    getProviderForModel(model);
  } catch (error) {
    throw new Error(`Invalid model for extraction: ${model}`);
  }
}

export function assertValidModelSchemaGeneration(model: string): void {
  if (!model || typeof model !== 'string') {
    throw new Error('Valid model must be provided for schema generation');
  }
  
  const validModels = ['gpt-4o-2024-11-20', 'gpt-4o-mini', 'gpt-4o'];
  if (!validModels.includes(model)) {
    throw new Error(`Model ${model} not valid for schema generation`);
  }
}

export { modelCards, modelCardsDict };