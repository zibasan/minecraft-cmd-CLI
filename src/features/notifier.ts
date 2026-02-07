import notifier from 'node-notifier';
import path from 'path';
import { fileURLToPath } from 'url';
import ora from 'ora';
import chalk from 'chalk';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export const sendNotify = (title: string, msg: string) => {
  // spinnerを起動
  const spinner = ora('Waiting to confirm notification... (Confirm to exit.)').start();

  notifier.notify(
    {
      title: title,
      message: msg,
      icon: path.join(__dirname, '/imgs/MC_CMD_CLI.jpg'),
      sound: true,
    },
    (err, _response, _metadata) => {
      // 2. 通知に対して何かアクションがあった（閉じた、クリックした、タイムアウトした）時に実行される
      if (err) {
        spinner.fail('An error has occurred.');
        console.error(err);
      } else {
        // スピナーを「完了」状態にして止める
        spinner.succeed(chalk.green('DONE'));
      }
    }
  );
};
