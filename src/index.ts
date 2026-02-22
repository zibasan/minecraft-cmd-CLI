#!/usr/bin/env node
import { Command } from 'commander';
import { createCommand } from './commands/create.js';
import { blockCommand } from './commands/block.js';
import { colorCodeCommand } from './commands/colorcode.js';

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
