import logSymbols from 'log-symbols';
import chalk from 'chalk';

export const info = `${chalk.bgBlue.white(` ${logSymbols.info} INFO `)}`;
export const success = `${chalk.bgGreen.white(` ${logSymbols.success} SUCCESS `)}`;
export const warn = `${chalk.bgYellow.black(` ${logSymbols.warning} WARN `)}`;
export const error = `${chalk.bgRed.white(` ${logSymbols.error} ERROR `)}`;
