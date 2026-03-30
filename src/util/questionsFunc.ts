/* eslint-disable @typescript-eslint/no-explicit-any */

import chalk from 'chalk';
import Enquirer from 'enquirer';
import type { EnquirerBasePrompt, EnquirerModule } from '../types/enquirer.js';
import { warn } from './symbols.js';

const { Input, Select, Toggle, Form } = Enquirer as unknown as EnquirerModule;

// --- 追加: 最小限のプロンプト実行関数 ---
export async function runPrompt(prompt: EnquirerBasePrompt): Promise<unknown> {
  try {
    return await prompt.run();
  } catch (e) {
    // EnquirerのCtrl+Cキャンセル(空文字)、または閉じられたあとのエラーの場合は握りつぶす
    if (e === '') {
      process.emit('SIGINT'); // Ctrl+Cのシグナルを送ることで、index.tsで設定した終了処理を呼び出す

      // Promiseを解決させない(永遠に待機)ことで、ここで処理の進行を完全にストップさせ、
      // index.ts 側で設定した Ctrl+C の終了処理が実行されるのを静かに待ちます。
      return new Promise(() => {
        /* 永遠に待機 */
      });
    }
    if (typeof e === 'object' && e !== null && 'name' in e && e.name === 'FOOD_CANCELED') {
      return;
    }
    throw e; // それ以外の予期せぬエラーは投げる
  }
}
// ----------------------------------------

// --- 追加: 最小限のプロンプト実行関数(fillOutForm専用) ---
export async function runForm(
  prompt: EnquirerBasePrompt,
  allowCancel: boolean = true
): Promise<unknown> {
  try {
    return await prompt.run();
  } catch (e: unknown) {
    // EnquirerのCtrl+Cキャンセル(空文字)、または閉じられたあとのエラーの場合は握りつぶす
    if (e === '' && allowCancel === false) {
      process.emit('SIGINT'); // Ctrl+Cのシグナルを送ることで、index.tsで設定した終了処理を呼び出す

      // Promiseを解決させない(永遠に待機)ことで、ここで処理の進行を完全にストップさせ、
      // index.ts 側で設定した Ctrl+C の終了処理が実行されるのを静かに待ちます。
      return new Promise(() => {
        /* 永遠に待機 */
      });
    }
    if (e === '' && allowCancel === true) {
      console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
      console.log('\n');

      return '__BACK__'; // フォームがキャンセルされたことを示す特別な値
    }
    throw e; // それ以外の予期せぬエラーは投げる
  }
}
// ----------------------------------------

export async function createQuestion(query: string, initial?: number): Promise<string> {
  if (!Input) {
    throw new Error('enquirer Input not available');
  }

  const prompt = new Input({
    message: chalk.cyan(query),
    initial: initial ? String(initial) : undefined,
  }) as EnquirerBasePrompt;

  return (await runPrompt(prompt)) as string;
}

/**
 * 選択肢から１つ選択させるプロンプト
 * @param message 質問のメッセージ
 * @param choices 選択肢の配列: string[]
 * @returns 選択された値: Promise<string>
 */
export async function selectFromList(message: string, choices: string[]): Promise<string> {
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
  }) as EnquirerBasePrompt;

  return (await runPrompt(prompt)) as string;
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
  falseLabel: string = 'No'
): Promise<boolean> {
  if (!Toggle) {
    throw new Error('enquirer Toggle not available');
  }

  const prompt = new Toggle({
    name: 'question',
    message: chalk.cyan(message),
    enabled: chalk.green(trueLabel),
    disabled: chalk.red(falseLabel),
  }) as EnquirerBasePrompt;

  return (await runPrompt(prompt)) as boolean;
}

/**
 * 選択肢を検索できるSelectプロンプト
 * @param message 質問のメッセージ
 * @param fallbackMessage AutoCompleteが利用できない場合の質問のメッセージ
 * @param choices 選択肢の配列: string[]
 * @param allowBack backオプションを追加するかどうか。デフォルト値：true
 * @returns 選択された値: string
 */
export async function autoComplete(
  message: string,
  fallbackMessage: string,
  choices: string[],
  allowBack: boolean = true
): Promise<string> {
  const enquirerModule = Enquirer as EnquirerModule;
  const AutoComplete = enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;

  // Backを許可する(true)ならbackオプションを追加、許可しない(false)ならchoicesのみ
  const finalChoices = allowBack
    ? [
        { name: 'mc_cmd_gen_cli:back', value: '__BACK__' },
        ...choices.map((c) => ({ name: c, value: c })),
      ]
    : [...choices.map((c) => ({ name: c, value: c }))];

  let input = '';
  if (AutoComplete && choices.length > 0) {
    const ac = new AutoComplete({
      name: 'choices',
      message:
        message +
        chalk.gray.italic(' (Type to search, or select "mc_cmd_gen_cli:back" to go back)'),
      choices: finalChoices,
      limit: 10,
    }) as EnquirerBasePrompt;

    try {
      const val = await runPrompt(ac);
      input = String(val).trim();
    } catch {
      input = await createQuestion(chalk.cyan(fallbackMessage));
    }
  } else {
    input = await createQuestion(chalk.cyan(fallbackMessage));
  }

  input = input.trim();

  return input;
}

/**
 * フォームプロンプトを表示して複数の入力を受け取る
 * @param message 質問のメッセージ
 * @param choices 選択肢の配列
 * * @param choices.name 入力項目の名前（後で結果をオブジェクトで受け取るときのキーになる）
 * * @param choices.message メッセージ
 * * @param choices.initial 初期値（オプション）
 * @param choices.type 入力の種類（'string' | 'number' | 'boolean' | 'string | number' | 'number | ~' | 'number | ~ | ^'）
 * @param allowCancel フォームのキャンセルを許可するかどうか（デフォルト: true）。trueの場合、ユーザーがフォームをキャンセルしたときに'__BACK__'を返す。
 * @returns 選択された値: Promise<any> - オブジェクトで返す / キャンセル時には'__BACK__'を返す
 */
export async function fillOutForm(
  message: string,
  choices: {
    name: string;
    message: string;
    type: 'string' | 'number' | 'boolean' | 'string | number' | 'number | ~' | 'number | ~ | ^';
    initial?: string;
  }[],
  allowCancel: boolean = true
  // biome-ignore lint/suspicious/noExplicitAny: どの値を返すか不明なため
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
    // biome-ignore lint/suspicious/noExplicitAny: どの値を返すか不明なため
    validate(value: any) {
      // 全項目が入力されているか簡易チェック
      for (const choice of choices) {
        const input = String(value[choice.name]).trim();
        if (!input) {
          return `${chalk.bgRed.white(' ERROR ')} ${chalk.red('All fields are required.')}`;
        }

        if (choice.type === 'number') {
          // 数字（整数・小数）以外を拒否
          if (Number.isNaN(Number(input))) {
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

        if (choice.type === 'string | number') {
          // 数字（整数・小数）または文字列以外を拒否
          if (Number.isNaN(Number(input)) && typeof input !== 'string') {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be a number or a string.`)}`;
          }
        }

        if (choice.type === 'number | ~ | ^') {
          // 数字（整数・小数）、または ~ / ^ から始まる数字を許可する正規表現
          // 例: 5, -3, 0.5, ~, ~5, ~-2.5, ^, ^2
          const isValidLocation = /^(~|\^)?-?\d*(\.\d+)?$/.test(input);
          if (!isValidLocation) {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be a number, or start with "~" or "^".`)}`;
          }
        }

        if (choice.type === 'number | ~') {
          // 数字（整数・小数）、または ~ / ^ から始まる数字を許可する正規表現
          // 例: 5, -3, 0.5, ~, ~5, ~-2.5, ^, ^2
          const isValidLocation = /^(~|-?\d+(\.\d+)?|~-?\d+(\.\d+)?)$/.test(input);
          if (!isValidLocation) {
            return `${chalk.bgRed.white(' ERROR ')} ${chalk.red(`${choice.message} must be a number, or start with "~" or "^".`)}`;
          }
        }
      }

      return true;
    },
  }) as EnquirerBasePrompt;

  return await runForm(prompt, allowCancel);
}
