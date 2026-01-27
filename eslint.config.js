import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import json from '@eslint/json';

export default [
  {
    ignores: ['dist/**'],
  },

  {
    files: ['**/*.{js,mjs,cjs}'],
    languageOptions: {
      globals: globals.node,
    },
    ...js.configs.recommended,
  },

  {
    files: ['**/*.{ts,mts,cts}'],
    languageOptions: {
      globals: globals.node,
      parser: tseslint.parser,
    },
    plugins: {
      '@typescript-eslint': tseslint.plugin,
    },
    rules: {
      ...tseslint.configs.recommended.rules,
    },
  },

  {
    files: ['**/*.json'],
    language: 'json/json',
    plugins: { json },
    ...json.configs.recommended,
  },
];
