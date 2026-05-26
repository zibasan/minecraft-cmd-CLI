import chalk from 'chalk';
import Enquirer from 'enquirer';
import type { EnquirerBasePrompt, EnquirerModule } from '../../types/enquirer.js';
import { createQuestion } from '../../util/questionsFunc.js';
import { error } from '../../util/symbols.js';
import { loadDataLists, suggestSimilar } from '../../util/utilsFunc.js';

const items = await loadDataLists('items', 'ITEMS');

export async function getItemName(): Promise<string> {
  let itemName = '';

  // --- 有効なアイテム名が決まるまで回るループ ---
  while (true) {
    const enquirerModule = Enquirer as EnquirerModule;
    const AutoComplete = enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
    if (AutoComplete && items.length > 0) {
      const ac = new AutoComplete({
        name: 'item',
        message: 'Item (Choose from the list or type to search e.g., diamond_sword): ',
        choices: items.map((i) => ({ name: i, value: i })),
        limit: 10,
      }) as EnquirerBasePrompt;
      try {
        const val = await ac.run();
        itemName = String(val).trim(); // value is normalized (no prefix)
      } catch (e) {
        if (e === '') {
          process.emit('SIGINT');
        } else {
          // fallback to plain input
          itemName = await createQuestion(chalk.cyan('Item (e.g., diamond_sword): '));
        }
      }
    } else {
      itemName = await createQuestion(chalk.cyan('Item (e.g., diamond_sword): '));
    }

    if (!itemName) {
      console.log(error, chalk.red('Please enter item name.'));
      continue;
    }

    // 入力値の正規化（minecraft:がなければ付与）
    const fullName = itemName.startsWith('minecraft:') ? itemName : `minecraft:${itemName}`;

    // 存在チェック (items内には既にminecraft:が含まれているため、fullNameと直接比較)
    if (!items.includes(fullName)) {
      const suggestions = suggestSimilar(fullName, items).filter((s) => s !== '__BACK__');

      console.log(chalk.red(`Item "${fullName}" not found.`));
      if (suggestions.length > 0) {
        console.log(chalk.yellow('Did you mean:'));
        for (const s of suggestions) {
          console.log(`  - ${s}`);
        }
      }
      console.log(error, chalk.cyan('Please enter a valid item name (try Tab to autocomplete).'));
      continue;
    }

    // チェック通過
    itemName = fullName;
    break;
  }
  return itemName;
}
