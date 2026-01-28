#!/usr/bin/env node
import { Command } from 'commander';
import { helloCommand } from './commands/hello.js';
import { createCommand } from './commands/create.js';

const program = new Command();
program
  .name('mccmd')
  .version('0.0.0')
  .description('Genarate Minecraft Java Edition command on CLI.');
program.addCommand(helloCommand());
program.addCommand(createCommand());
program.parse(process.argv);

if (process.argv.length === 2) {
  program.help();
}
