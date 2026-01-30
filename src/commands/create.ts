import { Command } from 'commander';
import chalk from 'chalk';
import clipboard from 'clipboardy';
import { createInterface } from 'readline';

import { addtionalSelectorsQuestion } from './selectors/selectors.js';

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

export async function selectFromList(message: string, choices: string[]): Promise<string> {
  console.log(chalk.cyan(message));
  choices.forEach((choice, index) => {
    console.log(`  ${chalk.yellow(index + 1)}. ${choice}`);
  });

  let selection: number = 0;
  let validInput = false;

  while (!validInput) {
    const input = (await createQuestion(chalk.cyan('Choose: '))).trim();

    // Ensure the input is a strict integer string (digits only) before parsing.
    if (!/^[0-9]+$/.test(input)) {
      console.log(chalk.red(`Invalid selection. Please choose 1-${choices.length}`));
      continue;
    }
    selection = parseInt(input, 10);

    if (selection >= 1 && selection <= choices.length) {
      validInput = true;
    } else {
      console.log(chalk.red(`Invalid selection. Please choose 1-${choices.length}`));
    }
  }

  return choices[selection - 1];
}

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.action(async () => {
    const supportedTypes = ['give', 'teleport', 'setblock', 'fill', 'say', 'execute'];

    // Q1: Select command type
    const commandType = await selectFromList('Select a command type:', supportedTypes);

    console.log(chalk.blue(`Generate target:`), `${chalk.green(`${chalk.bold(commandType)}`)}`);

    let generatedCommand = '';
    let selector = '';

    console.log('\n');

    switch (commandType) {
      case 'give':
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

        // Q3: Enter item name
        const itemName = await createQuestion(chalk.cyan('Item name (e.g., diamond): '));
        if (!itemName.trim()) {
          console.log(chalk.red('Please enter an item name.'));
          process.exit(1);
        }
        console.log(chalk.blue(`Item name:`), `${chalk.green(`${chalk.bold(itemName)}`)}`);
        console.log('\n');

        // Q4: Amount
        let amount = await createQuestion(
          chalk.cyan("Item amount(How many? If empty, it'll set 1.): ")
        );
        if (!amount.trim()) {
          amount = '1';
        }
        console.log(chalk.blue(`Item amount:`), `${chalk.green(`${chalk.bold(amount)}`)}`);
        console.log('\n');

        generatedCommand = `/give ${selector} ${itemName} ${amount}`;
        break;

      case 'teleport':
        const destination = await createQuestion(
          chalk.cyan('Destination player/entity or coordinates (e.g., @p or 0 64 0): ')
        );
        if (!destination.trim()) {
          console.log(chalk.red('Please enter a destination.'));
          process.exit(1);
        }
        generatedCommand = `/teleport ${destination}`;
        break;

      case 'setblock':
        const sbPosition = await createQuestion(chalk.cyan('Position (e.g., 0 64 0): '));
        const sbBlock = await createQuestion(chalk.cyan('Block (e.g., diamond_block): '));
        if (!sbPosition.trim() || !sbBlock.trim()) {
          console.log(chalk.red('Please enter position and block.'));
          process.exit(1);
        }
        generatedCommand = `/setblock ${sbPosition} ${sbBlock}`;
        break;

      case 'fill':
        const fillFrom = await createQuestion(chalk.cyan('From position (e.g., 0 64 0): '));
        const fillTo = await createQuestion(chalk.cyan('To position (e.g., 10 64 10): '));
        const fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
        if (!fillFrom.trim() || !fillTo.trim() || !fillBlock.trim()) {
          console.log(chalk.red('Please enter all positions and block.'));
          process.exit(1);
        }
        generatedCommand = `/fill ${fillFrom} ${fillTo} ${fillBlock}`;
        break;

      case 'say':
        const message = await createQuestion(chalk.cyan('Message: '));
        if (!message.trim()) {
          console.log(chalk.red('Please enter a message.'));
          process.exit(1);
        }
        generatedCommand = `/say ${message}`;
        break;

      case 'execute':
        const execTarget = await createQuestion(chalk.cyan('Target selector (e.g., @a): '));
        const execCommand = await createQuestion(chalk.cyan('Command to execute: '));
        if (!execTarget.trim() || !execCommand.trim()) {
          console.log(chalk.red('Please enter target and command.'));
          process.exit(1);
        }
        generatedCommand = `/execute as ${execTarget} at @s run ${execCommand}`;
        break;
      default:
        console.log(chalk.red(`✗ Unknown command type: ${commandType}`));
        console.log(chalk.red(`✗ "${commandType}" is not yet supported. Sorry!`));
        process.exit(1);
    }

    console.log(`${chalk.green('Generated! Command:')} ${chalk.blue(`${generatedCommand}`)}`);

    // Wait for Enter key
    const rl = createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    rl.question(chalk.cyan('Press Enter to copy to clipboard...'), async () => {
      try {
        await clipboard.write(generatedCommand);
        console.log(chalk.green('✓ Command copied to clipboard!'));
      } catch (error) {
        console.log(chalk.red('✗ Failed to copy command to clipboard'));
      } finally {
        rl.close();
      }
    });
  });

  return cmd;
}
