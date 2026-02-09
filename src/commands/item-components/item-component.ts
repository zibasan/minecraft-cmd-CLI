import chalk from 'chalk';
import { createQuestion, selectFromList } from '../create.js';
import { warn, error } from '../../util/emojis.js';

export async function addItemComponentsQuestion(): Promise<string> {
  console.log(`${chalk.blue('Further target selector:')} ${chalk.green(`${chalk.bold('Yes')}`)}`);

  const addedComponents: string[] = [];
  const itemComponentsTypes = [
    'item_name',
    'custom_name',
    'lore',
    'damage',
    'enchantment_glint_override',
    'enchantments',
    'food',
    'max_damage',
    'max_stack_size',
    'rarity',
  ];

  let continueAdding = true;

  while (continueAdding) {
    const availableComponents = itemComponentsTypes.filter(
      (components) => !addedComponents.some((added) => added.split('=')[0] === components)
    );

    const componentsOptions = [
      ...availableComponents.map((s) => {
        const descriptions: { [key: string]: string } = {
          item_name: 'Item Name(Override the original name)',
          custom_name:
            'Item Name(looks like it was edited with an anvil, do not override the original name)',
          lore: 'Item Lore',
          damage: 'How much to reduce the durability',
          enchantment_glint_override: 'Whether show glint of enchantment(no enchantments)',
          enchantments: 'Item Enchantments',
          food: 'Setting edible items',
          max_damage: 'The maximum durability value of that item',
          max_stack_size: 'The maximum stack size value of that item',
          rarity: 'Item Rarity',
        };
        return `${s} - ${descriptions[s] || s}`;
      }),
      'OK',
    ];

    // Q3.55 Component Type
    const addItemComponents = (
      await selectFromList('Additional components: ', componentsOptions)
    ).split(' ')[0];

    console.log(
      chalk.blue(`Additional components:`),
      `${chalk.green(`${chalk.bold(addItemComponents)}`)}`
    );
    console.log('\n');

    switch (addItemComponents) {
      case 'item_name': {
        while (true) {
          const comp_itemName = await createQuestion(
            chalk.cyan('item_name(Override the original name). Type "back" to go back: ')
          );
          if (comp_itemName.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!comp_itemName.trim()) {
            console.log(error, chalk.red(' Please enter a item name(string).'));
            continue;
          }
          console.log(chalk.blue(`item_name: `), `${chalk.green.bold(`${comp_itemName}`)}`);
          console.log('\n');
          addedComponents.push(`item_name="${comp_itemName}"`);
          break;
        }
        break;
      }
      case 'custom_name': {
        while (true) {
          const comp_customName = await createQuestion(
            chalk.cyan(
              'custom_name(looks like it was edited with an anvil; ',
              chalk.italic('italic'),
              "). Type 'back' to go back: "
            )
          );
          if (comp_customName.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!comp_customName.trim()) {
            console.log(error, chalk.red(' Please enter a custom_name.'));
            continue;
          }
          console.log(chalk.blue(`custom_name: `), `${chalk.green.bold(comp_customName)}`);
          console.log('\n');

          addedComponents.push(`custom_name="${comp_customName}"`);
          break;
        }
        break;
      }
      case 'lore': {
        while (true) {
          const lore = await createQuestion(
            chalk.cyan(
              "lore(lore of the item, insert '<br>' to start a new line). Type 'back' to go back: "
            )
          );
          if (lore.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!lore.trim()) {
            console.log(error, chalk.red(' Please enter a lore.'));
            continue;
          }
          // <br>で分割して各行を処理
          const loreLines = lore
            .split(/<br>/)
            .map((line) => line.trim())
            .filter((line) => line);

          if (loreLines.length === 0) {
            console.log(error, chalk.red(' Please enter a valid lore.'));
            continue;
          }

          // 各行をコンソールに表示
          console.log(chalk.blue(`lore: `));
          loreLines.forEach((line) => {
            console.log(`  ${chalk.green.bold(line)}`);
          });
          console.log('\n');

          // 配列形式でaddedComponentsにpush
          const loreArray = JSON.stringify(loreLines);
          addedComponents.push(`lore=${loreArray}`);
          break;
        }
        break;
      }
      case 'damage': {
        while (true) {
          const comp_damage = await createQuestion(
            chalk.cyan(
              'damage(how much to reduce the durability (non-negative integer)). Type "back" to go back: '
            )
          );
          if (comp_damage.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!comp_damage.trim()) {
            console.log(
              error,
              chalk.red(' Please enter a valid damage value(non-negative integer).')
            );
            continue;
          }

          const damageNum = parseInt(comp_damage.trim(), 10);
          if (isNaN(damageNum) || damageNum < 0) {
            console.log(
              error,
              chalk.red(' Please enter a valid damage value (non-negative integer).')
            );
            continue;
          }

          console.log(chalk.blue(`damage: `), `${chalk.green.bold(damageNum.toString())}`);
          console.log('\n');
          addedComponents.push(`damage=${damageNum.toString()}`);
          break;
        }
        break;
      }

      case 'enchantment_glint_override': {
        while (true) {
          const comp_glintTF = await createQuestion(
            chalk.cyan(
              'enchantment_glint_override(whether to add the glow of the enchantment(no enchantment), boolean). Type "back" to go back: '
            )
          );
          if (comp_glintTF.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!comp_glintTF.trim()) {
            console.log(error, chalk.red('Please enter a boolean.'));
            continue;
          }
          console.log(
            chalk.blue(`enchantment_glint_override: `),
            `${chalk.green.bold(comp_glintTF)}`
          );
          console.log('\n');
          addedComponents.push(`enchantment_glint_override=${comp_glintTF}`);
          break;
        }
        break;
      }

      case 'enchantments': {
        const comp_enchantmentsList: string[] = [];
        let addMoreEnchantments = true;

        while (addMoreEnchantments) {
          // 1. エンチャント名を入力
          while (true) {
            const enchantmentName = await createQuestion(
              chalk.cyan('Enchantment name(e.g., sharpness, unbreaking). Type "back" to go back: ')
            );
            if (enchantmentName.trim().toLowerCase() === 'back') {
              console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
              console.log('\n');
              addMoreEnchantments = false;
              break;
            }
            if (!enchantmentName.trim()) {
              console.log(error, chalk.red(' Please enter an enchantment name.'));
              continue;
            }

            // 2. エンチャントレベルを入力（1~255）
            let validLevel = false;
            while (!validLevel) {
              const enchantmentLevel = await createQuestion(
                chalk.cyan('Enchantment level(1~255): ')
              );
              if (!enchantmentLevel.trim()) {
                console.log(error, chalk.red(' Please enter an enchantment level.'));
                continue;
              }
              const levelNum = parseInt(enchantmentLevel.trim(), 10);
              if (isNaN(levelNum) || levelNum < 1 || levelNum > 255) {
                console.log(error, chalk.red(' Please enter a valid level(1~255).'));
                continue;
              }

              // エンチャントを追加
              comp_enchantmentsList.push(`${enchantmentName}:${levelNum}`);
              console.log(
                chalk.blue(`Enchantment: `),
                `${chalk.green.bold(`${enchantmentName}:${levelNum}`)}`
              );
              console.log('\n');

              // 3. "他の"エンチャントを追加するかどうか選択
              const addMoreOptions = ['y - Add another enchantment', 'N - Finish'];
              const addMoreResult = await selectFromList(
                chalk.cyan('Add another enchantment?(y/N)'),
                addMoreOptions
              );

              if (addMoreResult.split(' ')[0].toLowerCase() === 'n') {
                addMoreEnchantments = false;
              } else {
                console.log(
                  `${chalk.blue('Add More Enchantments:')} ${chalk.green(`${chalk.bold('Yes')}`)}`
                );
              }

              validLevel = true;
              break;
            }

            if (!addMoreEnchantments) {
              break; // 外側の名前入力ループを抜ける
            }
          }
        }

        if (comp_enchantmentsList.length > 0) {
          const enchantmentsArray = '{' + comp_enchantmentsList.map((e) => `${e}`).join(', ') + '}';
          addedComponents.push(`enchantments=${enchantmentsArray}`);
          console.log(
            chalk.blue(`All enchantments: `),
            `${chalk.green(`${chalk.bold(enchantmentsArray)}`)}`
          );
          console.log('\n');
        }
        break;
      }

      case 'food': {
        const comp_food = await createQuestion(
          chalk.cyan(
            'food',
            '(format: <nutriton(int)>,<saturation(int)>,<can_always_eat(bool)>\n    nutrition: Amount of hunger level restored when eating\n    saturation: Amount of hidden hunger level restored when eating\n    can_always_eat: Whether can eat when the hunger level is MAX)',
            '. Type "back" to go back: '
          )
        );

        if (comp_food.toLowerCase() === 'back') {
          console.log(warn, chalk.yellow(' Cancelled. Back to selector selection.'));
          console.log('\n');
          break;
        }

        const [nutrition, saturation, can_always_eat] = comp_food.trim().split(',');
        const foodPrettied = `nutrition:${nutrition}, saturation:${saturation}, can_always_eat:${can_always_eat}`;

        console.log(chalk.blue(`food: `), `${chalk.green.bold(foodPrettied)}`);
        console.log('\n');
        addedComponents.push(`food={${foodPrettied}}`);
        break;
      }
      case 'max_damage': {
        while (true) {
          const comp_max_damage = await createQuestion(
            chalk.cyan(
              'max_damage(The maximum durability value of that item (non-negative integer)). Type "back" to go back: '
            )
          );
          if (comp_max_damage.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!comp_max_damage.trim()) {
            console.log(
              error,
              chalk.red(' Please enter a valid max damage(non-negative integer).')
            );
            continue;
          }

          const maxDamageNum = parseInt(comp_max_damage.trim(), 10);
          if (isNaN(maxDamageNum) || maxDamageNum < 0) {
            console.log(
              error,
              chalk.red(' Please enter a valid max damage(non-negative integer).')
            );
            continue;
          }

          console.log(chalk.blue(`max_damage: `), `${chalk.green.bold(maxDamageNum.toString())}`);
          console.log('\n');
          addedComponents.push(`max_damage=${maxDamageNum.toString()}`);
          break;
        }
        break;
      }
      case 'max_stack_size': {
        while (true) {
          const comp_max_stack_size = await createQuestion(
            chalk.cyan(
              'max_stack_size(The maximum stack size value of that item (int, 1-99)). Type "back" to go back: '
            )
          );
          if (comp_max_stack_size.trim().toLowerCase() === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }
          if (!comp_max_stack_size.trim()) {
            console.log(error, chalk.red(' Please enter a valid max damage(int, 1-99).'));
            continue;
          }

          const maxStackSizeNum = parseInt(comp_max_stack_size.trim(), 10);
          if (isNaN(maxStackSizeNum) || maxStackSizeNum < 0 || maxStackSizeNum > 99) {
            console.log(error, chalk.red(' Please enter a valid max damage(int, 1-99).'));
            continue;
          }

          console.log(
            chalk.blue(`max_stack_size: `),
            `${chalk.green.bold(maxStackSizeNum.toString())}`
          );
          console.log('\n');
          addedComponents.push(`max_damage=${maxStackSizeNum.toString()}`);
          break;
        }
        break;
      }
      case 'rarity': {
        const rarityList = [
          `common - Normal: ${chalk.white('white')}, Enchanted: ${chalk.hex('#55FFFF')('aqua')}`,
          `uncommon - Normal: ${chalk.hex('#FFFF55')('yellow')}, Enchanted: ${chalk.hex('#55FFFF')('aqua')}`,
          `rare - Normal: ${chalk.hex('#55FFFF')('aqua')}, Enchanted: ${chalk.hex('#FF55FF')('light_purple')}`,
          `epic - ${chalk.hex('#FF55FF')('light_purple')}`,
          'back - Go back to components selection',
        ];
        const comp_rarity = await selectFromList(chalk.cyan('rarity(item rarity): '), rarityList);
        const rarity = comp_rarity.split(' ')[0];

        if (rarity.toLowerCase() === 'back') {
          console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
          console.log('\n');
          break;
        }

        console.log(chalk.blue(`rarity:`), `${chalk.green(`${chalk.bold(rarity)}`)}`);
        console.log('\n');
        addedComponents.push(`rarity=${rarity}`);
        break;
      }
      case 'OK': {
        continueAdding = false;
        break;
      }
    }
  }
  const addComponents = addedComponents.join(',');

  console.log(
    chalk.blue(`All components:`),
    `${chalk.green(`${chalk.bold(addComponents || null)}`)}`
  );
  console.log('\n');

  return addComponents;
}
