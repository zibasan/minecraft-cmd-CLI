#!/usr/bin/env node
import chalk from 'chalk';
import { Command } from 'commander';
import { createCommand } from './commands/create.js';
import { blockCommand } from './commands/block.js';
import { colorCodeCommand } from './commands/colorcode.js';

process.on('SIGINT', () => {
  console.log(
    chalk.bgYellow.black(' CANCELED '),
    chalk.yellow('Ctrl + C was detected. This process will be closed...')
  );
  process.stdout.write('\x1B[?25h');
  process.exit(0);
});

const program = new Command();
program
  .name('mccmd')
  .version('0.0.0')
  .description('Generate Minecraft Java Edition command on CLI.');
program.addCommand(createCommand());
program.addCommand(blockCommand());
program.addCommand(colorCodeCommand());
program.parse(process.argv);

if (process.argv.length === 2) {
  program.help();
}
