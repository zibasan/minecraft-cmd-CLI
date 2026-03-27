import chalk from 'chalk';
import {
  autoComplete,
  createQuestion,
  fillOutForm,
  selectFromList,
  toggleQuestion,
} from '../../util/questionsFunc.js';
import { error, warn } from '../../util/symbols.js';
import { COMPONENT_DESCRIPTIONS } from '../../util/utils.js';
import { loadDataLists, suggestSimilar } from '../../util/utilsFunc.js';

/**
 * 追加するコンポーネントを選択、さらにその内容を指定させる
 * @return 追加されたコンポーネントを整形した文字列を返す
 */
export async function addItemComponentsQuestion(): Promise<string> {
  console.log(`${chalk.blue('Add component(s):')} ${chalk.green(`${chalk.bold('Yes')}`)}`);

  const addedComponents: string[] = [];
  const itemComponentsTypes = Object.keys(COMPONENT_DESCRIPTIONS);

  let continueAdding = true;

  while (continueAdding) {
    const availableComponents = itemComponentsTypes.filter(
      (components) => !addedComponents.some((added) => added.split('=')[0] === components)
    );

    const componentsOptions = [
      ...availableComponents.map((s) => `${s} - ${COMPONENT_DESCRIPTIONS[s] || s}`),
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
              'custom_name(do not override the original name)',
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
          if (Number.isNaN(damageNum) || damageNum < 0) {
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
          const comp_glintTF = await selectFromList(
            chalk.cyan(
              'enchantment_glint_override(whether to add the glow of the enchantment(no enchantment), boolean). Type "back" to go back: '
            ),
            ['true', 'false', 'back']
          );
          if (comp_glintTF === 'back') {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
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

        const enchantments = await loadDataLists('enchantments', 'ENCHANTMENTS');

        // 複数のエンチャントを追加するための大枠のループ
        while (addMoreEnchantments) {
          let enchantments_name = '';
          let nameBacked = false;

          // 1. エンチャント名を入力するループ
          while (true) {
            enchantments_name = await autoComplete(
              chalk.cyan(
                'Enchantment name (e.g., minecraft:sharpness). Choose "back" to go back: '
              ),
              chalk.cyan('Enchantment name (e.g., minecraft:sharpness). Type "back" to go back: '),
              enchantments
            );

            // キャンセル判定
            if (
              enchantments_name.trim() === '__BACK__' ||
              enchantments_name.trim().toLowerCase() === 'back'
            ) {
              nameBacked = true;
              break; // 名前入力ループを抜ける
            }

            // Normalize enchantments id: allow with or without minecraft: prefix
            const normalized = enchantments_name.startsWith('minecraft:')
              ? enchantments_name.slice(10)
              : enchantments_name;
            const exists = enchantments.includes(normalized);

            if (!enchantments_name.trim()) {
              console.log(error, chalk.red('Please enter enchantments name.'));
              continue;
            }
            if (!exists) {
              const suggestions = suggestSimilar(normalized, enchantments)
                .filter((s) => s !== '__BACK__')
                .map((s) => `minecraft:${s}`);

              console.log(chalk.red(`Enchantments "${enchantments_name}" not found.`));
              if (suggestions.length > 0) {
                console.log(chalk.yellow('Did you mean:'));
                for (const s of suggestions) {
                  console.log(`  - ${s}`);
                }
              }
              console.log(
                error,
                chalk.cyan('Please enter a valid enchantments name (try Tab to autocomplete).')
              );
              continue;
            }

            // 正しいエンチャントが入力されたら名前入力ループを抜ける
            break;
          }

          // ★ 名前入力で「戻る」が選ばれた場合、エンチャント追加ループ自体を終了し、コンポーネント選択へ戻る
          if (nameBacked) {
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            console.log('\n');
            break;
          }

          // 2. エンチャントレベルを入力するループ（1~255）
          let levelBacked = false;
          while (true) {
            const enchantmentLevel = await createQuestion(
              chalk.cyan(`Enchantment(${enchantments_name}) level(1~255) Type "back" to go back: `)
            );

            // キャンセル判定
            if (
              enchantmentLevel.trim() === '__BACK__' ||
              enchantmentLevel.trim().toLowerCase() === 'back'
            ) {
              levelBacked = true;
              break; // レベル入力ループを抜ける
            }

            if (!enchantmentLevel.trim()) {
              console.log(error, chalk.red(' Please enter an enchantment level.'));
              continue;
            }
            const levelNum = parseInt(enchantmentLevel.trim(), 10);
            if (Number.isNaN(levelNum) || levelNum < 1 || levelNum > 255) {
              console.log(error, chalk.red(' Please enter a valid enchantment level(1~255).'));
              continue;
            }

            // 正しいレベルが入力されたら、リストに追加してループを抜ける
            comp_enchantmentsList.push(`${enchantments_name}:${levelNum}`);
            console.log(
              chalk.blue(`Enchantment: `),
              `${chalk.green.bold(`${enchantments_name}:${levelNum}`)}`
            );
            console.log('\n');
            break;
          }

          // ★ レベル入力で「戻る」が選ばれた場合、名前入力（ループの先頭）に戻る
          if (levelBacked) {
            console.log(warn, chalk.yellow(' Cancelled. Back to enchantments selection.'));
            console.log('\n');
            continue; // while (addMoreEnchantments) の先頭へ戻る
          }

          // 3. "他の"エンチャントを追加するかどうか選択
          const addMoreResult = await toggleQuestion(chalk.cyan('Add another enchantment?'));

          if (addMoreResult === false) {
            addMoreEnchantments = false;
          } else {
            console.log(
              `${chalk.blue('Add More Enchantments:')} ${chalk.green(`${chalk.bold('Yes')}`)}`
            );
          }
        }

        // 全ての入力が終わり、1つでもエンチャントが追加されていればコンポーネントを生成
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
        const comp_food = await fillOutForm(
          chalk.cyan(
            'Enter food properties (nutrition, saturation, can_always_eat). Press Ctrl+C to cancel and go back.'
          ),
          [
            { name: 'nutrition', message: 'Nutrition (int)', initial: '5', type: 'number' },
            { name: 'saturation', message: 'Saturation (float)', initial: '0.3', type: 'number' },
            {
              name: 'can_always_eat',
              message: 'Can always eat (bool)',
              initial: 'false',
              type: 'boolean',
            },
          ]
        );

        if (comp_food === '__BACK__') {
          break;
        }

        const { nutrition, saturation, can_always_eat } = comp_food;
        const foodPrettied = `nutrition:${nutrition}, saturation:${saturation}, can_always_eat:${can_always_eat}`;

        console.log(chalk.blue(`food: `), `${chalk.green.bold(foodPrettied)}`);
        console.log('\n');
        addedComponents.push(`food={${foodPrettied}}`);
        break;
      }

      case 'break_sound': {
        const sounds = await loadDataLists('sounds', 'SOUNDS');
        let final_soundName = '';
        let backed = false;

        // --- 有効なサウンド名が決まるまで回るループ ---
        while (true) {
          const input = await autoComplete(
            chalk.cyan('break_sound: Select a sound...'),
            chalk.cyan('break_sound (e.g. block.stone.break etc.)'),
            sounds
          );

          // 戻る処理
          if (input === '__BACK__') {
            backed = true;
            break;
          }

          if (!input) {
            console.log(error, chalk.red('Please enter sound name.'));
            continue;
          }

          // 入力値の正規化（minecraft:がなければ付与）
          const fullName = input.startsWith('minecraft:') ? input : `minecraft:${input}`;

          // 存在チェック (sounds内には既にminecraft:が含まれているため、fullNameと直接比較)
          if (!sounds.includes(fullName)) {
            const suggestions = suggestSimilar(fullName, sounds).filter((s) => s !== '__BACK__');

            console.log(chalk.red(`Sound "${fullName}" not found.`));
            if (suggestions.length > 0) {
              console.log(chalk.yellow('Did you mean:'));
              for (const s of suggestions) {
                console.log(`  - ${s}`);
              }
            }
            console.log(
              error,
              chalk.cyan('Please enter a valid sound name (try Tab to autocomplete).')
            );
            continue;
          }

          // チェック通過
          final_soundName = fullName;
          break;
        }

        // 「back」が選択された場合
        if (backed) {
          console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
          console.log('\n');
          break;
        }

        // コンポーネントに追加
        if (final_soundName) {
          console.log(chalk.blue(`break_sound: `), `${chalk.green.bold(final_soundName)}`);
          addedComponents.push(`break_sound="${final_soundName}"`);
          console.log('\n');
        }
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
          if (Number.isNaN(maxDamageNum) || maxDamageNum < 0) {
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
          if (Number.isNaN(maxStackSizeNum) || maxStackSizeNum < 0 || maxStackSizeNum > 99) {
            console.log(error, chalk.red(' Please enter a valid max damage(int, 1-99).'));
            continue;
          }

          console.log(
            chalk.blue(`max_stack_size: `),
            `${chalk.green.bold(maxStackSizeNum.toString())}`
          );
          console.log('\n');
          addedComponents.push(`max_stack_size=${maxStackSizeNum.toString()}`);
          break;
        }
        break;
      }

      case 'can_break': {
        const comp_blocksList: string[] = [];
        let addMoreBlocks = true;
        let backed = false;
        const blocks = await loadDataLists('blocks', 'BLOCKS');

        while (addMoreBlocks) {
          let current_name = ''; // このターンで入力する名前をリセット

          // --- 1. 有効なブロック名が決まるまで回るループ ---
          while (true) {
            const input = await autoComplete(
              chalk.cyan(`can_break (Current: ${comp_blocksList.length}) - Select a block... `),
              chalk.cyan('can_break (e.g. stone). Type "back" to go back: '),
              blocks
            );
            if (input === '__BACK__') {
              backed = true;
              break; // 内側の入力ループを抜ける
            }
            if (!input) {
              console.log(error, chalk.red('Please enter block name.'));
              continue;
            }

            // プレフィックスの正規化（minecraft:を抜いた純粋なIDにする）
            const normalized = input.startsWith('minecraft:') ? input.slice(10) : input;
            const fullName = `minecraft:${normalized}`;

            // 重複チェック
            if (comp_blocksList.includes(fullName)) {
              console.log(
                chalk.bgRed.white(' DUPLICATE '),
                chalk.red(`"${fullName}" is already added.\n\n`)
              );
              continue; // 入力待ち（このwhileループの先頭）に戻る
            }

            // 存在チェック
            if (!blocks.includes(normalized)) {
              const suggestions = suggestSimilar(normalized, blocks)
                .filter((s) => s !== '__BACK__')
                .map((b) => `minecraft:${b}`);

              console.log(chalk.red(`Block "${fullName}" not found.`));
              if (suggestions.length > 0) {
                console.log(chalk.yellow('Did you mean:'));
                for (const s of suggestions) {
                  console.log(`  - ${s}`);
                }
              }
              continue; // 入力待ちに戻る
            }

            // 全てのチェックを通過
            current_name = fullName;
            break; // 「有効な名前が決まるまで回るループ」を抜ける
          }

          if (backed && comp_blocksList.length === 0) {
            addMoreBlocks = false;
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            break;
          }

          if (backed && comp_blocksList.length > 0) {
            addMoreBlocks = false;
            console.log(warn, chalk.yellow(' Cancelled. Finish adding blocks to "can_break".'));
            break;
          }

          // リストに追加
          comp_blocksList.push(current_name);
          console.log(chalk.blue(`Added Block: `), `${chalk.green.bold(current_name)}`);

          // --- 2. 他のブロックを追加するかどうか選択 ---
          const addMoreResult = await toggleQuestion(
            chalk.cyan('Add another block to "can_break"?')
          );
          console.log('\n');

          if (addMoreResult === false) {
            addMoreBlocks = false;
          }
        }

        // 最終的なコンポーネントの書き出し
        if (comp_blocksList.length > 0) {
          const blocksArray = comp_blocksList.map((b) => `"${b}"`).join(', ');
          addedComponents.push(`can_break={blocks:[${blocksArray}]}`);
          console.log(
            chalk.blue(`All blocks added to "can_break": `),
            chalk.green.bold(`[${blocksArray}]`)
          );
          console.log('\n');
        }

        if (comp_blocksList.length === 0) {
          break; // can_breakコンポーネントは追加せず、空文字を返す
        }

        break;
      }

      case 'can_place_on': {
        const comp_blocksList: string[] = [];
        let addMoreBlocks = true;
        let backed = false;
        const blocks = await loadDataLists('blocks', 'BLOCKS');

        while (addMoreBlocks) {
          let current_name = ''; // このターンで入力する名前をリセット

          // --- 1. 有効なブロック名が決まるまで回るループ ---
          while (true) {
            const input = await autoComplete(
              `can_place_on (Current: ${comp_blocksList.length}) - Select a block...`,
              'Block ID (e.g., stone): ',
              blocks
            );

            if (input === '__BACK__') {
              backed = true;
              break; // 内側の入力ループを抜ける
            }
            if (!input) {
              console.log(error, chalk.red('Please enter block name.'));
              continue;
            }

            // プレフィックスの正規化（minecraft:を抜いた純粋なIDにする）
            const normalized = input.startsWith('minecraft:') ? input.slice(10) : input;
            const fullName = `minecraft:${normalized}`;

            // 重複チェック
            if (comp_blocksList.includes(fullName)) {
              console.log(
                chalk.bgRed.white(' DUPLICATE '),
                chalk.red(`"${fullName}" is already added.\n\n`)
              );
              continue; // 入力待ち（このwhileループの先頭）に戻る
            }

            // 存在チェック
            if (!blocks.includes(normalized)) {
              const suggestions = suggestSimilar(normalized, blocks)
                .filter((s) => s !== '__BACK__')
                .map((b) => `minecraft:${b}`);

              console.log(chalk.red(`Block "${fullName}" not found.`));
              if (suggestions.length > 0) {
                console.log(chalk.yellow('Did you mean:'));
                for (const s of suggestions) {
                  console.log(`  - ${s}`);
                }
              }
              continue; // 入力待ちに戻る
            }

            // 全てのチェックを通過
            current_name = fullName;
            break; // 「有効な名前が決まるまで回るループ」を抜ける
          }

          if (backed && comp_blocksList.length === 0) {
            addMoreBlocks = false;
            console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
            break;
          }

          if (backed && comp_blocksList.length > 0) {
            addMoreBlocks = false;
            console.log(warn, chalk.yellow(' Cancelled. Finish adding blocks to "can_place_on".'));
            break;
          }

          // リストに追加
          comp_blocksList.push(current_name);
          console.log(chalk.blue(`Added Block: `), `${chalk.green.bold(current_name)}`);

          // --- 2. 他のブロックを追加するかどうか選択 ---
          const addMoreResult = await toggleQuestion(
            chalk.cyan('Add another block to "can_place_on"?')
          );
          console.log('\n');

          if (addMoreResult === false) {
            addMoreBlocks = false;
          }
        }

        // 最終的なコンポーネントの書き出し
        if (comp_blocksList.length > 0) {
          const blocksArray = comp_blocksList.map((b) => `"${b}"`).join(', ');
          addedComponents.push(`can_place_on={blocks:[${blocksArray}]}`);
          console.log(
            chalk.blue(`All blocks added to "can_place_on": `),
            chalk.green.bold(`[${blocksArray}]`)
          );
          console.log('\n');
        }

        if (comp_blocksList.length === 0) {
          break; // can_place_onコンポーネントは追加せず、空文字を返す
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
