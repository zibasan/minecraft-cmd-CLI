import { Command } from 'commander';
import chalk from 'chalk';

export function helloCommand(): Command {
  const cmd = new Command('hello');
  cmd.description('Say hello and demonstrate separated command files');
  cmd.action(() => {
    console.log(chalk.green('Hello from your scaffolded CLI!'));
  });
  return cmd;
}
