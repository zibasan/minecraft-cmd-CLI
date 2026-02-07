import { Command } from 'commander';
import chalk from 'chalk';
import clipboard from 'clipboardy';
import { createInterface } from 'readline';
import ora from 'ora';
import { fileURLToPath } from 'url';
// path not required here

import { sendNotify } from '../features/notifier.js';
import { addtionalSelectorsQuestion } from './selectors/selectors.js';

import { info, success, warn, error } from '../util/emojis.js';

// Type definitions for enquirer
interface EnquirerChoice {
  name: string;
  value: string;
  enabled?: boolean;
}

interface EnquirerPrompt {
  index?: number;
  cursor?: number;
  choices: EnquirerChoice[];
  run: () => Promise<string | string[]>;
  render?: () => void;
}

export function createQuestion(query: string): Promise<string> {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(query, (answer) => {
      rl.close();
      resolve(answer);
    });
  });
}

async function loadBlocksList(): Promise<string[]> {
  // Try to import the generated JS module (when running built dist)
  try {
    const mod = await import(fileURLToPath(new URL('../data/blocks.js', import.meta.url)));
    const list = (mod?.BLOCKS || mod?.default || []) as string[];
    return list;
  } catch {
    // Try to import TS directly (when running with ts-node)
    try {
      const mod = await import(fileURLToPath(new URL('../data/blocks.ts', import.meta.url)));
      const list = (mod?.BLOCKS || mod?.default || []) as string[];
      return list;
    } catch {
      // If blocks file is not available, log a warning and return empty array
      console.warn(
        chalk.yellow(
          'Warning: blocks.ts/blocks.js file not found. Block autocomplete and validation will be disabled.'
        )
      );
      return [];
    }
  }
}

function levenshtein(a: string, b: string): number {
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

function suggestSimilar(input: string, pool: string[], max = 5): string[] {
  const scores = pool.map((p) => ({ p, d: levenshtein(input, p) }));
  scores.sort((a, b) => a.d - b.d);
  return scores.slice(0, max).map((s) => s.p);
}

function isValidPositionToken(token: string): boolean {
  // allow: 5, -3, ~, ~5
  return /^(?:~-?\d+|~|-?\d+)$/.test(token);
}

function isValidPosition(pos: string): boolean {
  const tokens = pos.trim().split(/\s+/);
  if (tokens.length !== 3) return false;
  return tokens.every(isValidPositionToken);
}

export async function selectFromList(message: string, choices: string[]): Promise<string> {
  const promptChoices = choices.map((c) => ({ name: c, value: c }));
  const enquirerModule = (await import('enquirer')) as {
    MultiSelect?: new (options: Record<string, unknown>) => EnquirerPrompt;
    default?: { MultiSelect?: new (options: Record<string, unknown>) => EnquirerPrompt };
  };
  const MultiSelect = enquirerModule.MultiSelect || enquirerModule.default?.MultiSelect;
  if (!MultiSelect) {
    throw new Error('enquirer MultiSelect not available');
  }

  const prompt = new MultiSelect({
    name: 'selected',
    message,
    choices: promptChoices.map((p) => ({ name: p.name, value: p.value })),
    // show all choices
    limit: promptChoices.length,
  }) as EnquirerPrompt;

  const stdin = process.stdin;
  const onData = (chunk: Buffer | string) => {
    const key = typeof chunk === 'string' ? chunk : chunk.toString('utf8');
    if (key === ' ') {
      // When space pressed, enforce single selection: mark only the focused choice as enabled
      try {
        // try to read current index from prompt
        const idx = typeof prompt.index === 'number' ? prompt.index : (prompt.cursor ?? 0);
        prompt.choices.forEach((c: EnquirerChoice, i: number) => {
          c.enabled = i === idx;
        });
        try {
          // re-render to update visual checkboxes
          prompt.render?.();
        } catch {
          void 0;
        }
      } catch {
        // ignore
      }
    }
  };

  stdin.resume();
  stdin.on('data', onData);

  try {
    const result = await prompt.run();
    // MultiSelect returns an array — but we enforce single selection above
    if (Array.isArray(result)) {
      if (result.length > 0) {
        return result[0];
      }
      // If nothing was checked (user pressed Enter), use focused index
      const idx = typeof prompt.index === 'number' ? prompt.index : (prompt.cursor ?? 0);
      const choice = prompt.choices[idx];
      return choice && choice.value ? choice.value : '';
    }
    return result;
  } finally {
    stdin.removeListener('data', onData);
    stdin.pause();
  }
}

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.option('-c, --copy [boolean]', 'Whether copy command to clipboard', true);
  cmd.option('-s, --silent', 'Whether nofity when the command copied');

  cmd.action(async (options) => {
    switch (options.copy) {
      case 'true': {
        console.log(
          `${chalk.bgBlue.white(' INFO ')} ${chalk.green.bold('The command will be copied to clipboard')}`
        );
        break;
      }
      case 'false': {
        console.log(
          `${chalk.bgBlue.white(' INFO ')} ${chalk.green.bold('The command will not be copied to clipboard')}`
        );
        break;
      }
      default: {
        console.log(
          `${chalk.bgBlue.white(' INFO ')} ${chalk.green.bold('The command will be copied to clipboard')}`
        );
        break;
      }
    }

    if (options.silent) {
      console.log(
        `${chalk.bgYellow.black(' WARN ')} ${chalk.yellow.bold('Notification will not be sent when the command copied')}`
      );
    }

    const supportedTypes = ['give', 'teleport', 'setblock', 'fill', 'say', 'execute'];

    // Q1: Select command type
    const commandType = await selectFromList('Select a command type:', supportedTypes);

    console.log(chalk.blue(`Generate target:`), `${chalk.green(`${chalk.bold(commandType)}`)}`);

    let generatedCommand = '';
    let selector = '';

    console.log('\n');

    switch (commandType) {
      case 'give': {
        // Q2: Choose Target
        const target = [
          '@p - Near Player',
          '@a - All Player',
          '@s - Myself',
          '@r - Random Player',
          "@n - A Nearest Player (1.21+, same '@p[sort=nearest, limit=1]')",
        ];
        const targetType = (await selectFromList('Select a target selector type:', target)).split(
          ' '
        )[0];

        console.log(chalk.blue(`Target:`), `${chalk.green(`${chalk.bold(targetType)}`)}`);
        console.log('\n');

        // Q2.5: Ask to refine target selector
        const refineSelector = await createQuestion(
          chalk.cyan('Want to further refine your target selector? (y/N): ')
        );
        const shouldRefine = refineSelector.toLowerCase() === 'y';
        let addSelectors: string = '';
        if (shouldRefine) {
          addSelectors = await addtionalSelectorsQuestion();
        } else {
          console.log(
            `${chalk.blue('Further target selector:')} ${chalk.green(`${chalk.bold('No')}`)}`
          );
        }

        const addedSelectorsTF: boolean = addSelectors ? true : false;

        if (addedSelectorsTF) {
          selector = `${targetType}[${addSelectors}]`;
        } else {
          selector = `${targetType}`;
        }

        console.log('\n');

        // Q3: Enter item name (repeat until valid)
        let itemName = '';
        do {
          itemName = await createQuestion(chalk.cyan('Item name (e.g., diamond): '));
          if (!itemName.trim()) {
            console.log(error, chalk.red('Please enter an item name.'));
          }
        } while (!itemName.trim());
        console.log(chalk.blue(`Item name:`), `${chalk.green(`${chalk.bold(itemName)}`)}`);
        console.log('\n');

        // Q4: Amount
        let amount = '';
        do {
          amount = await createQuestion(
            chalk.cyan("Item amount(How many? If empty, it'll set 1.): ")
          );
          if (!amount.trim()) {
            amount = '1';
            break;
          }
          if (!/^[0-9]+$/.test(amount.trim())) {
            console.log(error, chalk.red('Amount must be a positive integer.'));
            amount = '';
          }
        } while (!amount.trim());
        console.log(chalk.blue(`Item amount:`), `${chalk.green(`${chalk.bold(amount)}`)}`);
        console.log('\n');

        generatedCommand = `/give ${selector} ${itemName} ${amount}`;
        break;
      }
      case 'teleport': {
        let destination = '';
        do {
          destination = await createQuestion(
            chalk.cyan('Destination player/entity or coordinates (e.g., @p or 0 64 0): ')
          );
          if (!destination.trim()) {
            console.log(error, chalk.red('Please enter a destination.'));
            continue;
          }
          // allow selector like @p or coordinates
          const isSelector = destination.trim().startsWith('@');
          const isCoords = isValidPosition(destination.trim());
          if (!isSelector && !isCoords) {
            console.log(
              error,
              chalk.red(
                'Destination must be a selector (e.g., @p) or three coordinates (e.g., 0 64 0).'
              )
            );
            destination = '';
          }
        } while (!destination.trim());
        generatedCommand = `/teleport ${destination}`;
        break;
      }
      case 'setblock': {
        let sbPosition = '';
        let sbBlock = '';
        // Load block list for validation / autocomplete
        const blocks = await loadBlocksList();
        do {
          sbPosition = await createQuestion(chalk.cyan('Position (e.g., 0 64 0): '));
          if (!sbPosition.trim()) {
            console.log(error, chalk.red('Please enter a position.'));
            continue;
          }

          // For block id, use enquirer AutoComplete for tab completion
          const enquirerModule = (await import('enquirer')) as {
            AutoComplete?: new (options: Record<string, unknown>) => EnquirerPrompt;
            default?: { AutoComplete?: new (options: Record<string, unknown>) => EnquirerPrompt };
          };
          const AutoComplete = enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
          if (AutoComplete && blocks.length > 0) {
            const ac = new AutoComplete({
              name: 'block',
              message: 'Block (e.g., diamond_block): ',
              choices: blocks.map((b) => ({ name: `minecraft:${b}`, value: b })),
              limit: 10,
            }) as EnquirerPrompt;
            try {
              const val = await ac.run();
              sbBlock = String(val).trim(); // value is normalized (no prefix)
            } catch {
              // fallback to plain input
              sbBlock = await createQuestion(chalk.cyan('Block (e.g., diamond_block): '));
            }
          } else {
            sbBlock = await createQuestion(chalk.cyan('Block (e.g., diamond_block): '));
          }

          // Normalize block id: allow with or without minecraft: prefix
          const normalized = sbBlock.startsWith('minecraft:') ? sbBlock.slice(10) : sbBlock;
          const exists = blocks.includes(normalized);
          if (!sbPosition.trim() || !sbBlock.trim()) {
            console.log(error, chalk.red('Please enter position and block.'));
            continue;
          }
          if (!exists) {
            const suggestions = suggestSimilar(normalized, blocks).map((s) => `minecraft:${s}`);
            console.log(chalk.red(`Block ID "${sbBlock}" not found.`));
            if (suggestions.length > 0) {
              console.log(chalk.yellow('Did you mean:'));
              suggestions.forEach((s) => console.log(`  - ${s}`));
            }
            console.log(
              error,
              chalk.cyan('Please enter a valid block ID (try Tab to autocomplete).')
            );
            sbBlock = '';
            continue;
          }
        } while (!sbPosition.trim() || !sbBlock.trim());
        generatedCommand = `/setblock ${sbPosition} minecraft:${sbBlock.startsWith('minecraft:') ? sbBlock.slice(10) : sbBlock}`;
        break;
      }
      case 'fill': {
        let fillFrom = '';
        let fillTo = '';
        let fillBlock = '';
        // load blocks for autocomplete
        const fillBlocks = await loadBlocksList();
        do {
          fillFrom = await createQuestion(chalk.cyan('From position (e.g., 0 64 0): '));
          if (!isValidPosition(fillFrom)) {
            console.log(
              error,
              chalk.red('From position must be three coordinates (e.g., 0 64 0) or use ~ notation.')
            );
            fillFrom = '';
            continue;
          }
          fillTo = await createQuestion(chalk.cyan('To position (e.g., 10 64 10): '));
          if (!isValidPosition(fillTo)) {
            console.log(
              error,
              chalk.red('To position must be three coordinates (e.g., 10 64 10) or use ~ notation.')
            );
            fillTo = '';
            continue;
          }
          if (fillBlocks.length > 0) {
            const enquirerModule = (await import('enquirer')) as {
              AutoComplete?: new (options: Record<string, unknown>) => EnquirerPrompt;
              default?: { AutoComplete?: new (options: Record<string, unknown>) => EnquirerPrompt };
            };
            const AutoComplete =
              enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
            if (AutoComplete) {
              const ac = new AutoComplete({
                name: 'fillBlock',
                message: 'Block (e.g., stone):',
                choices: fillBlocks.map((b) => ({ name: `minecraft:${b}`, value: b })),
                limit: 10,
              }) as EnquirerPrompt;
              try {
                const val = await ac.run();
                fillBlock = String(val).trim();
              } catch {
                fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
              }
            } else {
              fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
            }
          } else {
            fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
          }

          // normalize and validate
          const normalizedFill = fillBlock.startsWith('minecraft:')
            ? fillBlock.slice(10)
            : fillBlock;
          if (!fillBlock.trim()) {
            console.log(error, chalk.red('Please enter a block ID.'));
            fillBlock = '';
            continue;
          }
          if (fillBlocks.length > 0 && !fillBlocks.includes(normalizedFill)) {
            const suggestions = suggestSimilar(normalizedFill, fillBlocks).map(
              (s) => `minecraft:${s}`
            );
            console.log(error, chalk.red(`Block ID "${fillBlock}" not found.`));
            if (suggestions.length) {
              console.log(chalk.yellow('Did you mean:'));
              suggestions.forEach((s) => console.log(`  - ${s}`));
            }
            fillBlock = '';
            continue;
          }
        } while (!fillFrom.trim() || !fillTo.trim() || !fillBlock.trim());
        // ensure final output includes minecraft: prefix
        const outBlock = fillBlock.startsWith('minecraft:') ? fillBlock : `minecraft:${fillBlock}`;
        generatedCommand = `/fill ${fillFrom} ${fillTo} ${outBlock}`;
        break;
      }
      case 'say': {
        let message = '';
        do {
          message = await createQuestion(chalk.cyan('Message: '));
          if (!message.trim()) console.log(error, chalk.red('Please enter a message.'));
        } while (!message.trim());
        generatedCommand = `/say ${message}`;
        break;
      }
      case 'execute': {
        let execTarget = '';
        let execCommand = '';
        do {
          execTarget = await createQuestion(chalk.cyan('Target selector (e.g., @a): '));
          execCommand = await createQuestion(chalk.cyan('Command to execute: '));
          if (!execTarget.trim() || !execCommand.trim())
            console.log(error, chalk.red('Please enter target and command.'));
        } while (!execTarget.trim() || !execCommand.trim());
        generatedCommand = `/execute as ${execTarget} at @s run ${execCommand}`;
        break;
      }

      default: {
        console.log(error, chalk.red(`Unknown command type: ${commandType}`));
        console.log(warn, chalk.yellow(`"${commandType}" is not yet supported. Sorry!`));
        process.exit(1);
      }
    }

    const spinner = ora({ text: 'Generating commands...', discardStdin: false }).start();
    await new Promise((r) => setTimeout(r, 600));
    spinner.stop();

    console.log(
      `${success} ${chalk.green('Generated! Command:')} ${chalk.blue(`${generatedCommand}`)}`
    );

    // Wait for Enter key
    const rl = createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    if (options.copy === 'true') {
      rl.question(`${info} chalk.cyan('Press Enter to copy to clipboard...'`, async () => {
        try {
          await clipboard.write(generatedCommand);
          console.log(success, chalk.green('Command copied to clipboard!'));
          if (!options.silent) {
            sendNotify('Minecraft-Command-Gen-CLI', '✅️ The command was copied successfully.');
          }
        } catch {
          console.log(error, chalk.red('Failed to copy command to clipboard'));
          if (!options.silent) {
            sendNotify('Minecraft-Command-Gen-CLI', '❌️ Failed to copy command.');
          }
        } finally {
          rl.close();
        }
      });
    } else if (options.copy === 'false') {
      process.exit();
    } else {
      rl.question(`${info} ${chalk.cyan('Press Enter to copy to clipboard...')}`, async () => {
        try {
          await clipboard.write(generatedCommand);
          console.log(success, chalk.green('Command copied to clipboard!'));
          if (!options.silent) {
            sendNotify('Minecraft-Command-Gen-CLI', '✅️ The command was copied successfully.');
          }
        } catch {
          console.log(error, chalk.red('Failed to copy command to clipboard'));
          if (!options.silent) {
            sendNotify('Minecraft-Command-Gen-CLI', '❌️ Failed to copy command.');
          }
        } finally {
          rl.close();
        }
      });
    }
  });

  return cmd;
}
