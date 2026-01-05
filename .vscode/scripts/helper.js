import { spawn } from 'node:child_process';
import { existsSync, readdirSync } from 'node:fs';
import path from 'node:path';

const windowsCmd = ['npm', 'npx'];

function cmd(name) {
  if (process.platform === 'win32' && windowsCmd.includes(name)) {
    return `${name}.cmd`;
  }
  return name;
}

function run(name, args, options = {}) {
  console.log(`+ ${options.cwd || '.'} $ ${name} ${args.join(' ')}`);
  return new Promise((resolve, reject) => {
    const p = spawn(cmd(name), args, { stdio: 'inherit', ...options });
    p.on('close', (code) => (code === 0 ? resolve() : reject(code)));
    p.on('error', reject);
  });
}

async function ensureFrontendDeps() {
  if (!existsSync(path.join('frontend', 'node_modules'))) {
    console.log('Installing frontend dependencies…');
    await run('npm', ['ci'], { cwd: 'frontend' });
  }
}

function goServices() {
  return readdirSync('.').filter(
    (folder) => !folder.startsWith('.') && folder !== 'frontend' && existsSync(path.join(folder, 'go.mod'))
  );
}

const action = process.argv[2];

try {
  await ensureFrontendDeps();

  switch (action) {
    case 'fmt':
      await run('npm', ['run', 'format'], { cwd: 'frontend' });
      await run('golangci-lint', ['fmt']);
      break;
    case 'lint':
      await run('npm', ['run', 'lint:fix'], { cwd: 'frontend' });
      await run('golangci-lint', ['run', ...goServices().map((s) => `./${s}/...`)]);
      break;
    case 'tidy':
      for (const service of goServices()) {
        await run('go', ['mod', 'tidy'], { cwd: service });
      }
      await run('go', ['work', 'sync']);
      break;
    default:
      console.error(`Unknown action: ${action}`);
      process.exit(1);
  }
} catch (err) {
  console.error(err);
  process.exit(typeof err === 'number' ? err : 1);
}
