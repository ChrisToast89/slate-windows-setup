package patch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wassermanproductions/slate-installer/internal/logx"
)

// ApplyBrainWindowsPatch applies the verified Windows CLI resolution fix to brain.ts
// if upstream has not already merged an equivalent (IS_WIN marker).
func ApplyBrainWindowsPatch(sourceRoot string) (string, error) {
	brainPath := filepath.Join(sourceRoot, "src", "main", "brain.ts")
	raw, err := os.ReadFile(brainPath)
	if err != nil {
		return "", fmt.Errorf("read brain.ts: %w", err)
	}
	src := string(raw)
	if strings.Contains(src, "IS_WIN") || strings.Contains(src, "winShimTarget") {
		logx.Log("brain.ts already has Windows CLI fix — skipping patch")
		return "Windows brain fix already present (no change needed).", nil
	}

	// 1) import dirname
	if !strings.Contains(src, "dirname") {
		src = strings.Replace(src,
			`import { join, extname } from 'path'`,
			`import { join, extname, dirname } from 'path'`,
			1)
	}

	// 2) insert IS_WIN + WIN_CLI_DIRS + winShimTarget before resolveCli
	if !strings.Contains(src, "const IS_WIN") {
		marker := "// The ChatGPT desktop app bundles"
		insert := `const IS_WIN = process.platform === 'win32'

// Windows equivalents — where npm puts global CLI shims.
const WIN_CLI_DIRS = [
  join(homedir(), 'AppData', 'Roaming', 'npm'),
  join(homedir(), '.local', 'bin'),
  join(homedir(), 'bin')
]

// npm's global installer puts a <name>.cmd shim on Windows. Node can't spawn .cmd
// safely with shell:true for AI prompts — resolve the real .exe instead.
function winShimTarget(cmdPath: string): string | null {
  try {
    const content = readFileSync(cmdPath, 'utf8')
    const match = content.match(/"%dp0%\\([^"]+\.exe)"/i)
    if (!match) return null
    const exe = join(dirname(cmdPath), match[1])
    return existsSync(exe) ? exe : null
  } catch {
    return null
  }
}

` + marker
		if !strings.Contains(src, marker) {
			// insert after CLI_DIRS block end — before resolveCli
			marker = "function resolveCli(name: string): string {"
			insert = `const IS_WIN = process.platform === 'win32'

const WIN_CLI_DIRS = [
  join(homedir(), 'AppData', 'Roaming', 'npm'),
  join(homedir(), '.local', 'bin'),
  join(homedir(), 'bin')
]

function winShimTarget(cmdPath: string): string | null {
  try {
    const content = readFileSync(cmdPath, 'utf8')
    const match = content.match(/"%dp0%\\([^"]+\.exe)"/i)
    if (!match) return null
    const exe = join(dirname(cmdPath), match[1])
    return existsSync(exe) ? exe : null
  } catch {
    return null
  }
}

` + marker
		}
		if !strings.Contains(src, marker) {
			return "", fmt.Errorf("could not locate insertion point in brain.ts — upstream may have changed")
		}
		src = strings.Replace(src, marker, insert, 1)
	}

	// 3) patch resolveCli body start
	oldResolve := `function resolveCli(name: string): string {
  if (name === 'codex' && existsSync(CODEX_BUNDLED)) return CODEX_BUNDLED
  for (const dir of CLI_DIRS) {`
	newResolve := `function resolveCli(name: string): string {
  if (name === 'codex' && existsSync(CODEX_BUNDLED)) return CODEX_BUNDLED
  if (IS_WIN) {
    for (const dir of WIN_CLI_DIRS) {
      const shim = join(dir, ` + "`${name}.cmd`" + `)
      if (existsSync(shim)) {
        const exe = winShimTarget(shim)
        if (exe) return exe
      }
    }
    return name
  }
  for (const dir of CLI_DIRS) {`
	if strings.Contains(src, oldResolve) {
		src = strings.Replace(src, oldResolve, newResolve, 1)
	} else if !strings.Contains(src, "if (IS_WIN)") {
		// looser: after CODEX_BUNDLED line
		needle := "if (name === 'codex' && existsSync(CODEX_BUNDLED)) return CODEX_BUNDLED\n"
		if strings.Contains(src, needle) {
			src = strings.Replace(src, needle, needle+`  if (IS_WIN) {
    for (const dir of WIN_CLI_DIRS) {
      const shim = join(dir, `+"`${name}.cmd`"+`)
      if (existsSync(shim)) {
        const exe = winShimTarget(shim)
        if (exe) return exe
      }
    }
    return name
  }
`, 1)
		} else {
			return "", fmt.Errorf("could not patch resolveCli in brain.ts")
		}
	}

	// 4) brainEnv PATH separator
	oldEnv := `function brainEnv(): NodeJS.ProcessEnv {
  const extra = CLI_DIRS.join(':')
  return { ...process.env, PATH: ` + "`${process.env.PATH ?? ''}:${extra}`" + ` }
}`
	newEnv := `function brainEnv(): NodeJS.ProcessEnv {
  const dirs = IS_WIN ? WIN_CLI_DIRS : CLI_DIRS
  const sep = IS_WIN ? ';' : ':'
  return { ...process.env, PATH: ` + "`${process.env.PATH ?? ''}${sep}${dirs.join(sep)}`" + ` }
}`
	if strings.Contains(src, oldEnv) {
		src = strings.Replace(src, oldEnv, newEnv, 1)
	} else if strings.Contains(src, "CLI_DIRS.join(':')") {
		src = strings.Replace(src,
			`const extra = CLI_DIRS.join(':')
  return { ...process.env, PATH: `+"`${process.env.PATH ?? ''}:${extra}`"+` }`,
			`const dirs = IS_WIN ? WIN_CLI_DIRS : CLI_DIRS
  const sep = IS_WIN ? ';' : ':'
  return { ...process.env, PATH: `+"`${process.env.PATH ?? ''}${sep}${dirs.join(sep)}`"+` }`,
			1)
	}

	if err := os.WriteFile(brainPath, []byte(src), 0o644); err != nil {
		return "", err
	}
	logx.Log("Applied Windows brain.ts patch")
	return "Applied Windows Claude/Codex detection fix to brain.ts.", nil
}
