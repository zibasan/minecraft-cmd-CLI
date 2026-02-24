import { Command } from 'commander';
import chalk from 'chalk';
import clipboard from 'clipboardy';
// import { createInterface } from 'readline';
import ora from 'ora';
import Enquirer from 'enquirer';
// import figureSet from 'figures';

// path not required here

class CancelError extends Error {
  constructor() {
    super('CANCEL');
    this.name = 'CancelError';
  }
}

// 共通のエラーハンドリング関数
async function runPromptWithCancel(prompt: EnquirerBasePrompt, allowCancel: boolean) {
  try {
    return await prompt.run();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (e: any) {
    // 自分で投げたキャンセルエラーの場合
    if (e.message === 'CANCEL') {
      if (allowCancel) {
        // コンポーネント選択に戻す機能用（今はまだ使いませんが、準備しておきます）
        // 'back' という文字列が入力されたかのように振る舞う
        return 'back';
      } else {
        // メインの処理でのキャンセルの場合は終了
        console.log(
          chalk.bgYellow.black(' CANCELED '),
          chalk.yellow('Ctrl + C detected. This process will be closed...')
        );
        process.exit(0);
      }
    }
    throw e; // それ以外の予期せぬエラーはそのまま投げる
  }
}
import { sendNotify } from '../features/notifier.js';
import { addSelectorsQuestion } from './selectors/selectors.js';
import { addItemComponentsQuestion } from './item-components/item-component.js';
import { info, success, warn, error } from '../util/emojis.js';
import type { EnquirerBasePrompt, EnquirerModule } from '../types/enquirer.js';

export async function createQuestion(query: string, allowCancel: boolean = false): Promise<string> {
  const enquirerModule = Enquirer as EnquirerModule;
  // Enquirer の Input プロンプトを使用
  const Input = enquirerModule.Input || enquirerModule.default?.Input;

  if (!Input) {
    throw new Error('enquirer Input not available');
  }

  const prompt = new Input({
    message: chalk.cyan(query),
    onCancel: () => {
      throw new CancelError();
    },
  }) as EnquirerBasePrompt;

  return (await runPromptWithCancel(prompt, allowCancel)) as string;
}

/**
 *
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
    } catch {
      continue; // 次の拡張子を試す
    }
  }

  // ファイルが見つからなかった場合
  console.warn(
    warn,
    chalk.bgYellow.black(' WARN '),
    chalk.yellow(
      `${fileName} file not found. Autocomplete and validation for ${fileName} will be disabled.`
    )
  );
  return [];
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

export function suggestSimilar(input: string, pool: string[], max = 5): string[] {
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

/**
 * 選択肢から１つ選択させるプロンプト
 * @param message 質問のメッセージ
 * @param choices 選択肢の配列: string[]
 * @returns 選択された値: Promise<string>
 */
export async function selectFromList(
  message: string,
  choices: string[],
  allowCancel: boolean = false
): Promise<string> {
  const promptChoices = choices.map((c) => ({ name: c, value: c }));
  const enquirerModule = Enquirer as EnquirerModule;
  const Select = enquirerModule.Select || enquirerModule.default?.Select;
  if (!Select) {
    throw new Error('enquirer Select not available');
  }

  const prompt = new Select({
    name: 'selected',
    message,
    choices: promptChoices.map((p) => ({ name: p.name, value: p.value })),
    // show all choices
    limit: promptChoices.length,
    footer: chalk.gray.italic('Use arrow keys to navigate and press Enter to select.'),
    onCancel: () => {
      throw new CancelError();
    },
  }) as EnquirerBasePrompt;

  return (await runPromptWithCancel(prompt, allowCancel)) as unknown as string;
}

/** トグルプロンプトでYes/Noを選択させる
 * @param message 質問のメッセージ
 * @param trueLabel 'true'の時に表示するラベル
 * @param falseLabel 'false'の時に表示するラベル
 * @returns Promise<boolean> ユーザーの選択 (true/false)
 */

export async function toggleQuestion(
  message: string,
  trueLabel: string = 'Yes',
  falseLabel: string = 'No',
  allowCancel: boolean = false
): Promise<boolean> {
  const enquirerModule = Enquirer as EnquirerModule;
  const Toggle = enquirerModule.Toggle || enquirerModule.default?.Toggle;
  if (!Toggle) {
    throw new Error('enquirer Toggle not available');
  }

  const prompt = new Toggle({
    name: 'question',
    message: chalk.cyan(message),
    enabled: chalk.green(trueLabel),
    disabled: chalk.red(falseLabel),
    onCancel: () => {
      throw new CancelError();
    },
  }) as EnquirerBasePrompt;

  const answer = await runPromptWithCancel(prompt, allowCancel);
  if (answer === 'back') {
    return false;
  }
  return answer as unknown as boolean;
}

/**
 * フォームプロンプトを表示して複数の入力を受け取る
 * @param message 質問のメッセージ
 * @param choices 選択肢の配列
 * @param choices.name 名前
 * * @param choices.message メッセージ
 * * @param choices.initial 初期値（オプション）
 * @returns 選択された値: Promise<any> - オブジェクトで返す / キャンセル時には'__BACK__'を返す
 */
export async function fillOutForm(
  message: string,
  choices: {
    name: string;
    message: string;
    type: 'string' | 'number' | 'boolean';
    initial?: string;
  }[],
  allowCancel: boolean = false
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): Promise<any> {
  const enquirerModule = Enquirer as EnquirerModule;
  const Form = enquirerModule.Form || enquirerModule.default?.Form;
  if (!Form) {
    throw new Error('enquirer Form not available');
  }

  const prompt = new Form({
    name: 'form',
    message: chalk.cyan(message),
    choices,
    // show all choices
    stdout: process.stdout,
    stdin: process.stdin,
    onCancel: () => {
      throw new CancelError();
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    validate(value: any) {
      // 全項目が入力されているか簡易チェック
      for (const choice of choices) {
        const input = String(value[choice.name]).trim();
        if (!input) {
          return `${chalk.bgRed.white(' ERROR ')} ${chalk.red('All fields are required.')}`;
        }

        if (choice.type === 'number') {
          // 数字（整数・小数）以外を拒否
          if (isNaN(Number(input))) {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be a number.`)}`;
          }
        }

        if (choice.type === 'boolean') {
          // true / false 以外を拒否
          const lowered = input.toLowerCase();
          if (lowered !== 'true' && lowered !== 'false') {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be "true" or "false".`)}`;
          }
        }
      }
      return true;
    },
  }) as EnquirerBasePrompt;

  try {
    return await prompt.run();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    if (err.message === 'CANCEL') {
      if (allowCancel) {
        return '__BACK__';
      }
      console.log(
        chalk.bgYellow.black(' CANCELED '),
        chalk.yellow('Ctrl + C detected. This process will be closed...')
      );
      process.exit(0);
    }
    throw err;
  }
}

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.option('-c, --copy [boolean]', 'Whether to copy command to clipboard', true);
  cmd.option('-s, --silent', 'Whether to nofity when the command copied');
  cmd.option('--no-slash', 'Remove the leading slash("/") It is useful when using Command Block.');

  cmd.action(async (options) => {
    switch (options.copy) {
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

    if (options.slash === false) {
      console.log(
        `${chalk.bgBlue.black(' INFO ')} ${chalk.yellow.bold('Slash("/") will not be add to the command')}`
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
        // Q2: Target selector
        selector = await addSelectorsQuestion();
        console.log(chalk.blue(`Target selector:`), `${chalk.green(`${chalk.bold(selector)}`)}`);

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

        // Q3.5: Ask to add additional component
        const addComponentSelector = await toggleQuestion(chalk.cyan('Add item component(s)?: '));
        const shouldAdd = addComponentSelector === true;
        let addComponents: string = '';
        if (shouldAdd) {
          addComponents = await addItemComponentsQuestion();
        } else {
          console.log(`${chalk.blue('Add component(s):')} ${chalk.green(`${chalk.bold('No')}`)}`);
        }

        const addedComponentsTF: boolean = addComponents ? true : false;

        let item: string;

        if (addedComponentsTF) {
          item = `${itemName}[${addComponents}]`;
        } else {
          item = `${itemName}`;
        }

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

        generatedCommand = `/give ${selector} ${item} ${amount}`;
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
        const blocks = await loadDataLists('blocks', 'BLOCKS');
        do {
          sbPosition = await createQuestion(chalk.cyan('Position (e.g., 0 64 0): '));
          if (!sbPosition.trim()) {
            console.log(error, chalk.red('Please enter a position.'));
            continue;
          }

          // For block id, use enquirer AutoComplete for tab completion
          const enquirerModule = Enquirer as EnquirerModule;
          const AutoComplete = enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
          if (AutoComplete && blocks.length > 0) {
            const ac = new AutoComplete({
              name: 'block',
              message: 'Block (e.g., diamond_block): ',
              choices: blocks.map((b) => ({ name: `minecraft:${b}`, value: b })),
              limit: 10,
            }) as EnquirerBasePrompt;
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
        const fillBlocks = await loadDataLists('blocks', 'BLOCKS');
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
            const enquirerModule = Enquirer as EnquirerModule;
            const AutoComplete =
              enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
            if (AutoComplete) {
              const ac = new AutoComplete({
                name: 'fillBlock',
                message: 'Block (e.g., stone):',
                choices: fillBlocks.map((b) => ({ name: `minecraft:${b}`, value: b })),
                limit: 10,
              }) as EnquirerBasePrompt;
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

    if (options.slash === false && generatedCommand.startsWith('/')) {
      generatedCommand = generatedCommand.slice(1);
    }

    const spinner = ora({ text: 'Generating commands...', discardStdin: false }).start();
    await new Promise((r) => setTimeout(r, 600));
    spinner.stop();

    console.log(
      `${success} ${chalk.green('Generated! Command:')} ${chalk.blue(`${generatedCommand}`)}`
    );

    if (options.copy === 'false') {
      process.exit();
    } else {
      // readline インターフェースの代わりに Input プロンプトを使用
      const enquirerModule = Enquirer as EnquirerModule;
      // Enquirer の Input プロンプトを使用
      const Input = enquirerModule.Input || enquirerModule.default?.Input;

      if (!Input) {
        throw new Error('enquirer Input not available');
      }

      const confirmPrompt = new Input({
        message: `${info} ${chalk.cyan('Press Enter to copy to clipboard...')}`,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      }) as any;

      try {
        await confirmPrompt.run();
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
      }
    }
  });

  return cmd;
}
