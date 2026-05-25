import chalk from 'chalk';
import { createQuestion, toggleQuestion } from '../../util/questionsFunc.js';
import { error } from '../../util/symbols.js';
import { getItemName } from './item.js';
import { addItemComponentsQuestion } from './item-component.js';

export async function selectItem(): Promise<string> {
  const itemName = await getItemName();

  // Ask to add additional component
  const addComponentSelector = await toggleQuestion(chalk.cyan('Add item component(s)?: '));
  const shouldAdd = addComponentSelector === true;
  let addComponents: string = '';
  if (shouldAdd) {
    addComponents = await addItemComponentsQuestion();
  } else {
    console.log(`${chalk.blue('Add component(s):')} ${chalk.green(`${chalk.bold('No')}`)}`);
  }

  const addedComponentsTF: boolean = !!addComponents;

  let item: string;

  if (addedComponentsTF) {
    item = `${itemName}[${addComponents}]`;
  } else {
    item = `${itemName}`;
  }

  // Amount
  let amount = '';
  do {
    amount = await createQuestion(chalk.cyan("Item amount(How many? If empty, it'll set 1.): "));
    if (!amount.trim()) {
      amount = '1';
      break;
    }
    if (!/^[0-9]+$/.test(amount.trim())) {
      console.log(error, chalk.red('Amount must be a positive integer.'));
      amount = '';
    }
  } while (!amount.trim());
  console.log(chalk.blue(`Item amount:`), `${chalk.green(`${chalk.bold(amount)}`)}`);
  console.log('\n');

  const result = `${item} ${amount}`;
  return result;
}
