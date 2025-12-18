# Spotlite - frontend

Frontend interface for Spotlite.

## Setup

- Requirements: Node 24+, npm
- Install dependencies:

```bash
# intall dependencies
cd frontend
npm ci

# run dev server
npm run start
```

## Scripts

- `npm run start` - dev server
- `npm run build` - production build
- `npm run lint` - run ESLint (fail on warnings)
- `npm run lint:fix` - auto-fix lintable issues
- `npm run format` - format source with Prettier
- `npm run format:check` - check formatting (CI)

## Notes

- Angular 21 + SCSS.
- Commit `package-lock.json` to ensure reproducible installs.
- Recommended VS Code extensions: Prettier, ESLint, Angular Language Service.