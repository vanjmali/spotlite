# Spotlite — frontend

Essential setup

- Requirements: Node 24+, npm
- Install dependencies:

```powershell
cd frontend
npm ci
```

- Start development server

```powershell
npm run start
```

Key scripts

- npm run start      — dev server
- npm run build      — production build
- npm run lint       — run ESLint (fail on warnings)
- npm run lint:fix   — auto-fix lintable issues
- npm run format     — format source with Prettier
- npm run format:check — check formatting (CI)

Notes

- The project uses Angular 21 and SCSS. 
- Commit `package-lock.json` to ensure reproducible installs.
- Recommended VS Code extensions: Prettier, ESLint, Angular Language Service.

