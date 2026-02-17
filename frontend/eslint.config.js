// Modern ESM flat-config for ESLint (v9+).
// - Uses FlatCompat to reuse legacy shareable configs (TypeScript & Angular).
// - ESM style (import/export) follows the ESLint docs and avoids CommonJS interop issues.
// - Template rules are scoped to HTML files only to prevent template rules
//   from running on TypeScript files.
//
// Usage:
//   Run `npm ci` then `npm run lint` in the frontend folder.
import { defineConfig, globalIgnores } from 'eslint/config';
import js from '@eslint/js';
import { FlatCompat } from '@eslint/eslintrc';
import tsParser from '@typescript-eslint/parser';
import templateParser from '@angular-eslint/template-parser';

const compat = new FlatCompat({ recommendedConfig: js.configs.recommended });

// Use FlatCompat to create TS-only and template-only entries so that
// TypeScript rules don't run against HTML files and vice-versa.
const tsExtends = compat.extends(
  'plugin:@typescript-eslint/recommended',
  'plugin:@angular-eslint/recommended'
);
const tsConfigs = tsExtends.map((cfg) => ({ ...cfg, files: ['**/*.ts'] }));

const templateExtends = compat.extends('plugin:@angular-eslint/template/recommended');
const templateConfigs = templateExtends.map((cfg) => ({ ...cfg, files: ['**/*.html'] }));

export default defineConfig([
  // global ignores (clear and explicit)
  globalIgnores(['dist/**', 'node_modules/**', 'e2e/**']),

  // base JS rules
  js.configs.recommended,

  // include Prettier to disable formatting rules globally
  ...compat.extends('prettier'),

  // include TypeScript/Angular recommended configs restricted to TS files
  ...tsConfigs,

  // include template recommended configs restricted to HTML files
  ...templateConfigs,

  // TypeScript files: set parser and parserOptions.project
  {
    files: ['**/*.ts'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        project: ['./tsconfig.app.json', './tsconfig.json', './tsconfig.spec.json'],
        ecmaVersion: 2020,
        sourceType: 'module',
      },
    },
  },

  // HTML templates: set the template parser (rules come from templateConfigs)
  {
    files: ['**/*.html'],
    languageOptions: {
      parser: templateParser,
    },
  },

  // Custom rules overrides
  {
    files: ['**/*.ts'],
    rules: {
      '@angular-eslint/no-input-rename': 'off',
    },
  },
]);
