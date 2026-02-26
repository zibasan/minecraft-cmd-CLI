import Enquirer from 'enquirer';
import chalk from 'chalk';
import { runPromptWithCancel } from '../commands/create.js';
import type { EnquirerBasePrompt, EnquirerModule } from '../types/enquirer.js';

const { Input, Select, Toggle, Form } = Enquirer as unknown as EnquirerModule;

export async function createQuestion(query: string, allowCancel: boolean = false): Promise<string> {
  if (!Input) {
    throw new Error('enquirer Input not available');
  }

  const prompt = new Input({
    message: chalk.cyan(query),
    /*
      onCancel: () => {
      throw new CancelError();
    },
    */
  }) as EnquirerBasePrompt;

  const result = await runPromptWithCancel<string>(prompt, allowCancel);
  return result as string;
}

/**
 * 選択肢から１つ選択させるプロンプト
 * @param message 質問のメッセージ
 * @param choices 選択肢の配列: string[]
 * @returns 選択された値: Promise<string>
 */
export async function selectFromList(
  message: string,
  choices: string[],
  allowCancel: boolean = false
): Promise<string> {
  const promptChoices = choices.map((c) => ({ name: c, value: c }));
  if (!Select) {
    throw new Error('enquirer Select not available');
  }

  const prompt = new Select({
    name: 'selected',
    message,
    choices: promptChoices.map((p) => ({ name: p.name, value: p.value })),
    // show all choices
    limit: promptChoices.length,
    footer: chalk.gray.italic('Use arrow keys to navigate and press Enter to select.'),
    /*
    onCancel: () => {
      throw new CancelError();
    },
    */
  }) as EnquirerBasePrompt;

  return (await runPromptWithCancel<string>(prompt, allowCancel)) as string;
}

/** トグルプロンプトでYes/Noを選択させる
 * @param message 質問のメッセージ
 * @param trueLabel 'true'の時に表示するラベル
 * @param falseLabel 'false'の時に表示するラベル
 * @returns Promise<boolean> ユーザーの選択 (true/false)
 */

export async function toggleQuestion(
  message: string,
  trueLabel: string = 'Yes',
  falseLabel: string = 'No',
  allowCancel: boolean = false
): Promise<boolean> {
  if (!Toggle) {
    throw new Error('enquirer Toggle not available');
  }

  const prompt = new Toggle({
    name: 'question',
    message: chalk.cyan(message),
    enabled: chalk.green(trueLabel),
    disabled: chalk.red(falseLabel),
    /*
    onCancel: () => {
      throw new CancelError();
    },
    */
  }) as EnquirerBasePrompt;

  const answer = await runPromptWithCancel<boolean>(prompt, allowCancel);
  if (answer === 'back') {
    return false;
  }
  return answer;
}

/**
 * フォームプロンプトを表示して複数の入力を受け取る
 * @param message 質問のメッセージ
 * @param choices 選択肢の配列
 * * @param choices.name 名前
 * * @param choices.message メッセージ
 * * @param choices.initial 初期値（オプション）
 * @returns 選択された値: Promise<any> - オブジェクトで返す / キャンセル時には'__BACK__'を返す
 */
export async function fillOutForm(
  message: string,
  choices: {
    name: string;
    message: string;
    type: 'string' | 'number' | 'boolean';
    initial?: string;
  }[],
  allowCancel: boolean = false
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
): Promise<any> {
  if (!Form) {
    throw new Error('enquirer Form not available');
  }

  const prompt = new Form({
    name: 'form',
    message: chalk.cyan(message),
    choices,
    // show all choices
    stdout: process.stdout,
    stdin: process.stdin,
    /*
    onCancel: () => {
      throw new CancelError();
    },
    */
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    validate(value: any) {
      // 全項目が入力されているか簡易チェック
      for (const choice of choices) {
        const input = String(value[choice.name]).trim();
        if (!input) {
          return `${chalk.bgRed.white(' ERROR ')} ${chalk.red('All fields are required.')}`;
        }

        if (choice.type === 'number') {
          // 数字（整数・小数）以外を拒否
          if (isNaN(Number(input))) {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be a number.`)}`;
          }
        }

        if (choice.type === 'boolean') {
          // true / false 以外を拒否
          const lowered = input.toLowerCase();
          if (lowered !== 'true' && lowered !== 'false') {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be "true" or "false".`)}`;
          }
        }
      }
      return true;
    },
  }) as EnquirerBasePrompt;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const result = await runPromptWithCancel<any>(prompt, allowCancel);
  if (result === 'back') {
    return '__BACK__';
  }
  return result;
}
