import { Command } from 'commander';
import chalk from 'chalk';
import clipboard from 'clipboardy';
import ora from 'ora';
import Enquirer from 'enquirer';
import { sendNotify } from '../features/notifier.js';
import { addSelectorsQuestion } from './selectors/selectors.js';
import { addItemComponentsQuestion } from './item-components/item-component.js';
import { info, success, warn, error } from '../util/symbols.js';
import type { EnquirerBasePrompt, EnquirerModule } from '../types/enquirer.js';
import { createQuestion, selectFromList, toggleQuestion } from '../util/questionsFunc.js';
import { loadDataLists, suggestSimilar, isValidPosition } from '../util/utilsFunc.js';
// import figureSet from 'figures';

const { Input } = Enquirer as unknown as EnquirerModule;

// 共通のエラーハンドリング関数
export async function runPromptWithCancel<T>(
  prompt: EnquirerBasePrompt,
  allowCancel: boolean
): Promise<T | 'back'> {
  try {
    const result = await prompt.run();
    return result as T;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (e: any) {
    // 自分で投げたキャンセルエラーの場合
    if (e === '' || e.code === 'ERR_USE_AFTER_CLOSE' || e.message === 'CANCEL') {
      if (allowCancel) {
        // コンポーネント選択に戻す機能用（今はまだ使いませんが、準備しておきます）
        // 'back' という文字列が入力されたかのように振る舞う
        return 'back';
      } else {
        // カーソルを再表示してから終了する (重要)
        process.stdout.write('\x1B[?25h');
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

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.option('-c, --copy [boolean]', 'Whether to copy command to clipboard', true);
  cmd.option('-s, --silent', 'Whether to nofity when the command copied');
  cmd.option('--no-slash', 'Remove the leading slash("/") It is useful when using Command Block.');

  cmd.action(async (options) => {
    switch (options.copy) {
      case 'false': {
        console.log(`${info} ${chalk.green.bold('The command will not be copied to clipboard')}`);
        break;
      }
      default: {
        console.log(`${info} ${chalk.green.bold('The command will be copied to clipboard')}`);
        break;
      }
    }

    if (options.silent) {
      console.log(
        `${warn} ${chalk.yellow.bold('Notification will not be sent when the command copied')}`
      );
    }

    if (options.slash === false) {
      console.log(`${info} ${chalk.yellow.bold('Slash("/") will not be add to the command')}`);
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
      if (!Input) {
        throw new Error('enquirer Input not available');
      }

      const confirmPrompt = new Input({
        message: `${info} ${chalk.cyan('Press Enter to copy to clipboard...')}`,
        /*
        onCancel: () => {
          throw new CancelError();
        },
        */
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      }) as any;

      try {
        await runPromptWithCancel(confirmPrompt, false);
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
