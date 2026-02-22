import { Command } from 'commander';
import chalk from 'chalk';
import Table from 'cli-table3';
import zalgo from 'zalgo-js';

export let symbol: string;
export let sampleText: string;

export function colorCodeCommand(): Command {
  const cmd = new Command('colorcode');
  cmd.description('Show Minecraft Colorcode');
  cmd.option('--symbol <type>', "Sets which symbol to display; '&' or '§'", 'section');
  cmd.addHelpText(
    'after',
    'e.g., mc-cmd colorcode --symbol amp / mc-cmd colorcode --symbol section'
  );
  cmd.option(
    '-t, --sampletxt <text>',
    'Text which display to sample text',
    'This is a sample text.'
  );

  cmd.action(async (options) => {
    const symbolType = options.symbol;

    switch (symbolType) {
      case 'amp': {
        symbol = '&';
        break;
      }
      case 'section': {
        symbol = '§';
        break;
      }
      default: {
        symbol = '§';
      }
    }
    sampleText = options.sampletxt || 'This is a sample text.';
    console.log(
      `${chalk.bgBlue.white(' INFO ')} ${chalk.green.bold('The color code will be displayed with the symbol: ', chalk.blue.bold(symbol))}\n`
    );
    console.log(
      `${chalk.bgBlue.white(' INFO ')} ${chalk.green.bold('The sample text will be displayed in this text: ', chalk.blue.bold(sampleText))}`
    );

    const table = new Table({
      head: ['Code', 'Color/Decoration Name', '', 'Sample'],
      colWidths: [7, 20, 7, 25],
      wordWrap: true,
      colAligns: ['center', 'left', 'center', 'left'],
    });

    table.push(
      [`${symbol}0`, 'black', '-->', `${chalk.hex('#000000').bgWhite(sampleText)}`],
      [`${symbol}1`, 'dark_blue', '-->', `${chalk.hex('#0000AA').bgBlack(sampleText)}`],
      [`${symbol}2`, 'dark_green', '-->', `${chalk.hex('#00AA00').bgBlack(sampleText)}`],
      [`${symbol}3`, 'dark_aqua', '-->', `${chalk.hex('#00AAAA').bgBlack(sampleText)}`],
      [`${symbol}4`, 'dark_red', '-->', `${chalk.hex('#AA0000').bgBlack(sampleText)}`],
      [`${symbol}5`, 'dark_purple', '-->', `${chalk.hex('#AA00AA').bgBlack(sampleText)}`],
      [`${symbol}6`, 'gold', '-->', `${chalk.hex('#FFAA00').bgBlack(sampleText)}`],
      [`${symbol}7`, 'gray', '-->', `${chalk.hex('#AAAAAA').bgWhite(sampleText)}`],
      [`${symbol}8`, 'dark_gray', '-->', `${chalk.hex('#555555').bgWhite(sampleText)}`],
      [`${symbol}9`, 'blue', '-->', `${chalk.hex('#5555FF').bgBlack(sampleText)}`],
      [`${symbol}a`, 'green', '-->', `${chalk.hex('#55FF55').bgBlack(sampleText)}`],
      [`${symbol}b`, 'aqua', '-->', `${chalk.hex('#55FFFF').bgBlack(sampleText)}`],
      [`${symbol}c`, 'red', '-->', `${chalk.hex('#FF5555').bgBlack(sampleText)}`],
      [`${symbol}d`, 'light_purple', '-->', `${chalk.hex('#FF55FF').bgBlack(sampleText)}`],
      [`${symbol}e`, 'yellow', '-->', `${chalk.hex('#FFFF55').bgBlack(sampleText)}`],
      [`${symbol}f`, 'white', '-->', `${chalk.hex('#FFFFFF').bgBlack(sampleText)}`],
      [``, '', '', ``],
      [`${symbol}k`, 'Obfuscated', '-->', `${zalgo.default(sampleText)}(Garbled Characters)`],
      [`${symbol}l`, 'Bold', '-->', `${chalk.bold(sampleText)}`],
      [`${symbol}m`, 'StrikeThrough', '-->', `${chalk.strikethrough(sampleText)}`],
      [`${symbol}n`, 'Underline', '-->', `${chalk.underline(sampleText)}`],
      [`${symbol}o`, 'Italic', '-->', `${chalk.italic(sampleText)}`],
      [`${symbol}r`, 'Reset', '-->', sampleText]
    );

    console.log(`\n\n${table.toString()}\n`);
  });
  return cmd;
}
