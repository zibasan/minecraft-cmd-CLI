/**
 * Type definitions for the enquirer module
 * These types provide proper TypeScript support for enquirer prompts
 */

export interface EnquirerChoice {
  name: string;
  value: string;
  enabled?: boolean;
}

export interface EnquirerBasePrompt {
  index?: number;
  cursor?: number;
  choices: EnquirerChoice[];
  symbols?: Record<string, string>;
  footer?: string;
  run: () => Promise<string | string[]>;
  render?: () => void;
}

export type EnquirerModule = {
  Input?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  MultiSelect?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  AutoComplete?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  Select?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  Toggle?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  Form?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  default?: {
    Input?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
    MultiSelect?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
    AutoComplete?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
    Select?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
    Toggle?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
    Form?: new (options: Record<string, unknown>) => EnquirerBasePrompt;
  };
};
