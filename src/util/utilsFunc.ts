// eslint-disable-next-line @typescript-eslint/no-unused-vars

import chalk from 'chalk';
import { warn } from './symbols.js';

/**
 * @param fileName ロードさせたいファイル名 e.g., blocks, sounds etc.
 * @param constantName そのファイルでexportしている変数名 e.g., BLOCKS, SOUNDS etc.
 * @returns ロードしたいデータのリスト
 */

export async function loadDataLists(fileName: string, constantName: string): Promise<string[]> {
  const fileExtensions = ['js', 'ts'];

  for (const ext of fileExtensions) {
    try {
      const url = new URL(`../data/${fileName}.${ext}`, import.meta.url).href;
      const mod = await import(url);

      // 動的に定数名でアクセス (mod[constantName])
      const list = (mod?.[constantName] || mod?.default || []) as string[];

      return list;
    } catch {}
  }

  // ファイルが見つからなかった場合
  console.warn(
    warn,
    chalk.yellow(
      `Referenced file was not found. Autocomplete and validation for this command will be disabled.`
    )
  );

  return [];
}

export function levenshtein(a: string, b: string): number {
  const dp = Array.from({ length: a.length + 1 }, () => new Array(b.length + 1).fill(0));
  for (let i = 0; i <= a.length; i++) {
    dp[i][0] = i;
  }
  for (let j = 0; j <= b.length; j++) {
    dp[0][j] = j;
  }
  for (let i = 1; i <= a.length; i++) {
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      dp[i][j] = Math.min(dp[i - 1][j] + 1, dp[i][j - 1] + 1, dp[i - 1][j - 1] + cost);
    }
  }

  return dp[a.length][b.length];
}

export function suggestSimilar(input: string, pool: string[], max = 5): string[] {
  const scores = pool.map((p) => ({ p, d: levenshtein(input, p) }));
  scores.sort((a, b) => a.d - b.d);

  return scores.slice(0, max).map((s) => s.p);
}

export function isValidPositionToken(token: string): boolean {
  // allow: 5, -3, ~, ~5
  return /^(?:~-?\d+|~|-?\d+)$/.test(token);
}

export function isValidPosition(pos: string): boolean {
  const tokens = pos.trim().split(/\s+/);
  if (tokens.length !== 3) return false;

  return tokens.every(isValidPositionToken);
}
