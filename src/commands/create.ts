import { Command } from 'commander';
import chalk from 'chalk';
import clipboard from 'clipboardy';
import { createInterface } from 'readline';
import inquirer from 'inquirer';

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.action(async () => {
    const supportedTypes = ['give', 'teleport', 'setblock', 'fill', 'say', 'execute'];

    // Q1: Select command type
    const answer1 = await inquirer.prompt({
      type: 'list',
      name: 'commandType',
      message: 'Select a command type: ',
      choices: supportedTypes,
    });

    const commandType = answer1.commandType;
    console.log(chalk.blue(`Generate target: `), `${chalk.green(`${chalk.bold(commandType)}`)}`);

    let generatedCommand = '';

    switch (commandType) {
      case 'give':
        // Q2: Enter item name
        const answer2 = await inquirer.prompt([
          {
            type: 'input',
            name: 'itemName',
            message: 'Item name (e.g., diamond): ',
            validate: (input) => {
              if (!input.trim()) {
                return 'Please enter an item name.';
              }
              return true;
            },
          },
        ]);
        generatedCommand = `/give @p ${answer2.itemName}`;
        break;

      case 'teleport':
      case 'setblock':
      case 'fill':
      case 'say':
      case 'execute':

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
