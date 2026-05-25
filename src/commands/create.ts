import chalk from 'chalk';
import clipboard from 'clipboardy';
import { Command } from 'commander';
import Enquirer from 'enquirer';
import ora from 'ora';
import type { enchantmentData } from '../data/enchantments.js';
import { selectItem } from '../features/items/item-select.js';
import { sendNotify } from '../features/notifier.js';
import { addSelectorsQuestion } from '../features/selectors/selector.js';
import { getSlot } from '../features/slots/slot.js';
import type { EnquirerBasePrompt, EnquirerModule } from '../types/enquirer.js';
import {
  autoComplete,
  createQuestion,
  fillOutForm,
  runPrompt,
  selectFromList,
  toggleQuestion,
} from '../util/questionsFunc.js';
import { error, info, success, warn } from '../util/symbols.js';
import {
  ITEM_COMMANDS_DESCRIPTIONS,
  SETBLOCK_OPTIONS_DESCRIPTIONS,
  TP_COMMAND_DESCRIPTIONS,
} from '../util/utils.js';
import { loadDataLists, suggestSimilar } from '../util/utilsFunc.js';

// import figureSet from 'figures';

const { Input } = Enquirer as unknown as EnquirerModule;

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.option('-c, --copy [boolean]', 'Whether to copy command to clipboard', true);
  cmd.option('-s, --silent', 'Whether to nofity when the command copied');
  cmd.option('--no-slash', 'Remove the leading slash("/") It is useful when using Command Block.');

  cmd.action(async (options) => {
    switch (options.copy) {
      case 'false': {
        console.log(`${info} ${chalk.green.bold('The command will not be copied to clipboard')}`);
        break;
      }
      default: {
        console.log(`${info} ${chalk.green.bold('The command will be copied to clipboard')}`);
        break;
      }
    }

    if (options.silent) {
      console.log(
        `${warn} ${chalk.yellow.bold('Notification will not be sent when the command copied')}`
      );
    }

    if (options.slash === false) {
      console.log(`${info} ${chalk.yellow.bold('Slash("/") will not be add to the command')}`);
    }

    const supportedTypes = [
      'give',
      'teleport',
      'setblock',
      'fill',
      'say',
      'execute',
      'item',
      'effect',
      'summon',
      'enchant',
      'xp',
    ];

    // Q1: Select command type
    const commandType = await selectFromList('Select a command type:', supportedTypes);

    console.log(chalk.blue(`Generate target:`), `${chalk.green(`${chalk.bold(commandType)}`)}`);

    let generatedCommand = '';
    let selector = '';

    console.log('\n');

    switch (commandType) {
      case 'give': {
        // Q2: Target selector
        selector = await addSelectorsQuestion();
        console.log(chalk.blue(`Target selector:`), `${chalk.green(`${chalk.bold(selector)}`)}`);

        const item = await selectItem();

        generatedCommand = `/give ${selector} ${item}`;
        break;
      }
      case 'teleport': {
        // 1. 形式の選択
        const tpOptions = TP_COMMAND_DESCRIPTIONS.map((d) => `${d.cmd} - ${d.description}`);
        // 座標限界：29999983, Y, 29999983
        const selectedTpOption = await selectFromList(
          chalk.cyan('Teleport Command type:'),
          tpOptions
        );

        // 番号を取り出す
        const tpTypeStr = selectedTpOption.split('.')[0];
        console.log(chalk.blue('Selected teleport type:'), chalk.green(selectedTpOption));

        // 使う変数を初期化
        let targets = '';
        let destination = '';
        let location = '';
        let rotation = '';
        let facingLocation = '';
        let facingEntity = '';
        let facingAnchor = '';

        // 必要な要素を順番に聞く

        if (['2', '4', '5', '6', '7'].includes(tpTypeStr)) {
          console.log(chalk.gray.italic('<targets> - Specify the entity to teleport'));
          targets = await addSelectorsQuestion();
          console.log(chalk.blue(`<targets>:`), `${chalk.green(`${chalk.bold(targets)}`)}\n`);
        }

        if (['1', '2'].includes(tpTypeStr)) {
          console.log(chalk.gray.italic('<destination> - Specifies the entity to teleport to'));
          destination = await addSelectorsQuestion();
          console.log(
            chalk.blue(`<destination>:`),
            `${chalk.green(`${chalk.bold(destination)}`)}\n`
          );
        }

        if (['3', '4', '5', '6', '7'].includes(tpTypeStr)) {
          console.log(
            chalk.gray.italic(
              '<location> - Specifies the location(e.g., ~ ~ ~, 29 63 -48) to teleport to'
            )
          );
          const locResult = await fillOutForm(
            chalk.cyan('Enter location coordinates (e.g., 0 64 0 or ~ ~ ~).'),
            [
              {
                name: 'X',
                message: 'X coordinate',
                initial: '~',
                type: 'number | ~ | ^',
              },
              {
                name: 'Y',
                message: 'Y coordinate',
                initial: '~',
                type: 'number | ~ | ^',
              },
              {
                name: 'Z',
                message: 'Z coordinate',
                initial: '~',
                type: 'number | ~ | ^',
              },
            ],
            false
          );
          location = `${locResult.X} ${locResult.Y} ${locResult.Z}`;
          console.log(chalk.blue(`<location>:`), `${chalk.green(`${chalk.bold(location)}`)}\n`);
        }

        if (tpTypeStr === '5') {
          console.log(
            chalk.gray.italic(
              '<rotation> - Specify the direction(e.g., ~ ~, 90 0) after teleportation'
            )
          );
          const rotResult = await fillOutForm(
            chalk.cyan(
              'Enter the direction... (Horizontal rotation (yaw) and vertical rotation (pitch))'
            ),
            [
              {
                name: 'yaw',
                message: 'horizontal rotation',
                initial: '90',
                type: 'number | ~',
              },
              {
                name: 'pitch',
                message: 'vertical rotation',
                initial: '0',
                type: 'number | ~',
              },
            ],
            false
          );
          rotation = `${rotResult.yaw} ${rotResult.pitch}`;
          console.log(chalk.blue(`<rotation>:`), `${chalk.green(`${chalk.bold(rotation)}`)}\n`);
        }

        if (tpTypeStr === '6') {
          console.log(
            chalk.gray.italic(
              '<facingLocation> - Specifies the coordinates(e.g., ~ ~ ~, 29 63 -48) the entity will face after teleporting'
            )
          );
          const faceLocResult = await fillOutForm(
            chalk.cyan('Enter facing location coordinates...'),
            [
              {
                name: 'X',
                message: 'X coordinate',
                initial: '~',
                type: 'number | ~ | ^',
              },
              {
                name: 'Y',
                message: 'Y coordinate',
                initial: '~',
                type: 'number | ~ | ^',
              },
              {
                name: 'Z',
                message: 'Z coordinate',
                initial: '~',
                type: 'number | ~ | ^',
              },
            ],
            false
          );
          facingLocation = `${faceLocResult.X} ${faceLocResult.Y} ${faceLocResult.Z}`;
          console.log(
            chalk.blue(`<facingLocation>:`),
            `${chalk.green(`${chalk.bold(facingLocation)}`)}\n`
          );
        }

        if (tpTypeStr === '7') {
          console.log(
            chalk.gray.italic(
              '<facingEntity> - Specifies the entity the entity will face after teleporting'
            )
          );
          facingEntity = await addSelectorsQuestion();
          console.log(
            chalk.blue(`<facingEntity>:`),
            `${chalk.green(`${chalk.bold(facingEntity)}`)}\n`
          );

          const anchorChoice = await selectFromList(chalk.cyan('Facing Anchor (optional):'), [
            'eyes',
            'feet',
            'skip (do not specify anchor)',
          ]);

          if (!anchorChoice.startsWith('skip')) {
            facingAnchor = anchorChoice;
            console.log(
              chalk.blue(`Facing Anchor:`),
              `${chalk.green(`${chalk.bold(facingAnchor)}`)}\n`
            );
          }
        }

        // --- 3. 最後に選ばれたパターンに合わせてコマンドを組み立てる ---
        switch (tpTypeStr) {
          case '1':
            generatedCommand = `/teleport ${destination}`;
            break;
          case '2':
            generatedCommand = `/teleport ${targets} ${destination}`;
            break;
          case '3':
            generatedCommand = `/teleport ${location}`;
            break;
          case '4':
            generatedCommand = `/teleport ${targets} ${location}`;
            break;
          case '5':
            generatedCommand = `/teleport ${targets} ${location} ${rotation}`;
            break;
          case '6':
            generatedCommand = `/teleport ${targets} ${location} facing ${facingLocation}`;
            break;
          case '7':
            generatedCommand = `/teleport ${targets} ${location} facing entity ${facingEntity}${facingAnchor ? ` ${facingAnchor}` : ''}`;
            break;
        }
        break;
      }
      case 'setblock': {
        let sbPosition = '';
        let sbBlock = '';
        let selectedSbOption = '';
        let sbOption = '';
        // Load block list for validation / autocomplete
        const blocks = await loadDataLists('blocks', 'BLOCKS');

        const locResult = await fillOutForm(
          chalk.cyan('Enter location coordinates (e.g., 0 64 0 or ~ ~ ~).'),
          [
            {
              name: 'X',
              message: 'X coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
            {
              name: 'Y',
              message: 'Y coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
            {
              name: 'Z',
              message: 'Z coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
          ],
          false
        );
        sbPosition = `${locResult.X} ${locResult.Y} ${locResult.Z}`;
        console.log(chalk.blue('Position:'), chalk.green(sbPosition), '\n');

        do {
          // For block id, use enquirer AutoComplete for tab completion
          const enquirerModule = Enquirer as EnquirerModule;
          const AutoComplete = enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
          if (AutoComplete && blocks.length > 0) {
            const ac = new AutoComplete({
              name: 'block',
              message: 'Block (e.g., diamond_block): ',
              choices: blocks.map((b) => ({
                name: `minecraft:${b}`,
                value: b,
              })),
              limit: 10,
            }) as EnquirerBasePrompt;
            try {
              const val = await ac.run();
              sbBlock = String(val).trim(); // value is normalized (no prefix)
            } catch {
              // fallback to plain input
              sbBlock = await createQuestion(chalk.cyan('Block (e.g., diamond_block): '));
            }
          } else {
            sbBlock = await createQuestion(chalk.cyan('Block (e.g., diamond_block): '));
          }

          // Normalize block id: allow with or without minecraft: prefix
          const normalized = sbBlock.startsWith('minecraft:') ? sbBlock.slice(10) : sbBlock;
          const exists = blocks.includes(normalized);
          if (!sbPosition.trim() || !sbBlock.trim()) {
            console.log(error, chalk.red('Please enter position and block.'));
            continue;
          }
          if (!exists) {
            const suggestions = suggestSimilar(normalized, blocks).map((s) => `minecraft:${s}`);
            console.log(chalk.red(`Block ID "${sbBlock}" not found.`));
            if (suggestions.length > 0) {
              console.log(chalk.yellow('Did you mean:'));
              for (const s of suggestions) {
                console.log(`  - ${s}`);
              }
            }
            console.log(
              error,
              chalk.cyan('Please enter a valid block ID (try Tab to autocomplete).')
            );
            sbBlock = '';
          }
        } while (!sbPosition.trim() || !sbBlock.trim());

        const sbOptions = SETBLOCK_OPTIONS_DESCRIPTIONS.map(
          (d) => `${d.options} - ${d.description}`
        );
        // 座標限界：29999983, Y, 29999983
        while (true) {
          selectedSbOption = await selectFromList(chalk.cyan('Setblock Option:'), sbOptions);
          const optionKey = selectedSbOption.split(' - ')[0];
          if (optionKey === 'Skip') {
            selectedSbOption = 'Skiped';
            sbOption = '';
            break;
          } else if (optionKey) {
            selectedSbOption = optionKey;
            sbOption = selectedSbOption;
            break;
          }
        }

        console.log('\n', chalk.blue('Selected option:'), chalk.green(selectedSbOption));

        generatedCommand = `/setblock ${sbPosition} minecraft:${sbBlock.startsWith('minecraft:') ? sbBlock.slice(10) : sbBlock} ${sbOption}`;
        break;
      }
      case 'fill': {
        let fillFrom = '';
        let fillTo = '';
        let fillBlock = '';
        // load blocks for autocomplete
        const fillBlocks = await loadDataLists('blocks', 'BLOCKS');
        const locResult1 = await fillOutForm(
          chalk.cyan(
            'Enter the block coordinates(e.g., 0 64 0 or ~ ~ ~) to specify as the starting point...'
          ),
          [
            {
              name: 'X',
              message: 'X coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
            {
              name: 'Y',
              message: 'Y coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
            {
              name: 'Z',
              message: 'Z coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
          ],
          false
        );
        fillFrom = `${locResult1.X} ${locResult1.Y} ${locResult1.Z}`;
        console.log(chalk.blue('Position:'), chalk.green(fillFrom), '\n');

        const locResult2 = await fillOutForm(
          chalk.cyan(
            'Enter the block coordinates(e.g., 0 64 0 or ~ ~ ~) to specify as the ending point...'
          ),
          [
            {
              name: 'X',
              message: 'X coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
            {
              name: 'Y',
              message: 'Y coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
            {
              name: 'Z',
              message: 'Z coordinate',
              initial: '~',
              type: 'number | ~ | ^',
            },
          ],
          false
        );
        fillTo = `${locResult2.X} ${locResult2.Y} ${locResult2.Z}`;
        console.log(chalk.blue('Position:'), chalk.green(fillTo), '\n');

        do {
          if (fillBlocks.length > 0) {
            const enquirerModule = Enquirer as EnquirerModule;
            const AutoComplete =
              enquirerModule.AutoComplete || enquirerModule.default?.AutoComplete;
            if (AutoComplete) {
              const ac = new AutoComplete({
                name: 'fillBlock',
                message: 'Block (e.g., stone):',
                choices: fillBlocks.map((b) => ({
                  name: `minecraft:${b}`,
                  value: b,
                })),
                limit: 10,
              }) as EnquirerBasePrompt;
              try {
                const val = await ac.run();
                fillBlock = String(val).trim();
              } catch {
                fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
              }
            } else {
              fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
            }
          } else {
            fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
          }

          // normalize and validate
          const normalizedFill = fillBlock.startsWith('minecraft:')
            ? fillBlock.slice(10)
            : fillBlock;
          if (!fillBlock.trim()) {
            console.log(error, chalk.red('Please enter a block ID.'));
            fillBlock = '';
            continue;
          }
          if (fillBlocks.length > 0 && !fillBlocks.includes(normalizedFill)) {
            const suggestions = suggestSimilar(normalizedFill, fillBlocks).map(
              (s) => `minecraft:${s}`
            );
            console.log(error, chalk.red(`Block ID "${fillBlock}" not found.`));
            if (suggestions.length) {
              console.log(chalk.yellow('Did you mean:'));
              for (const s of suggestions) {
                console.log(`  - ${s}`);
              }
            }
            fillBlock = '';
          }
        } while (!fillBlock.trim());
        // ensure final output includes minecraft: prefix
        const outBlock = fillBlock.startsWith('minecraft:') ? fillBlock : `minecraft:${fillBlock}`;
        generatedCommand = `/fill ${fillFrom} ${fillTo} ${outBlock}`;
        break;
      }
      case 'say': {
        let message = '';
        do {
          message = await createQuestion(chalk.cyan('Message: '));
          if (!message.trim()) console.log(error, chalk.red('Please enter a message.'));
        } while (!message.trim());
        generatedCommand = `/say ${message}`;
        break;
      }
      // TODO: executeのコマンドをより細かく指定できるようにする
      case 'execute': {
        let execTarget = '';
        let execCommand = '';
        do {
          execTarget = await createQuestion(chalk.cyan('Target selector (e.g., @a): '));
          execCommand = await createQuestion(chalk.cyan('Command to execute: '));
          if (!execTarget.trim() || !execCommand.trim())
            console.log(error, chalk.red('Please enter target and command.'));
        } while (!execTarget.trim() || !execCommand.trim());
        generatedCommand = `/execute as ${execTarget} at @s run ${execCommand}`;
        break;
      }

      case 'item': {
        console.log(
          warn,
          chalk.yellow(
            'The "item" command generator is only supporting "replace" subcommand. Sorry!'
          )
        );
        const targets = ['block', 'entity'];
        const replaceTarget = await selectFromList(chalk.cyan('Replace target:'), targets);

        let targetOptionResult = '';

        switch (replaceTarget) {
          case 'block': {
            let blockPos = '';
            const posResult = await fillOutForm(
              chalk.cyan(
                'Enter the block coordinates(e.g., 0 64 0 or ~ ~ ~) to specify target block...'
              ),
              [
                {
                  name: 'X',
                  message: 'X coordinate',
                  initial: '~',
                  type: 'number | ~ | ^',
                },
                {
                  name: 'Y',
                  message: 'Y coordinate',
                  initial: '~',
                  type: 'number | ~ | ^',
                },
                {
                  name: 'Z',
                  message: 'Z coordinate',
                  initial: '~',
                  type: 'number | ~ | ^',
                },
              ],
              false
            );
            blockPos = `${posResult.X} ${posResult.Y} ${posResult.Z}`;
            console.log(chalk.blue('Position:'), chalk.green.bold(blockPos), '\n');
            targetOptionResult = `block ${blockPos}`;
            break;
          }

          case 'entity': {
            const targetEntity = await addSelectorsQuestion();
            targetOptionResult = `entity ${targetEntity}`;
            break;
          }
        }

        const slot = await getSlot();
        console.log(chalk.blue(`Selected Slot:`), chalk.green.bold(slot), '\n');

        const itemCommandOption = ITEM_COMMANDS_DESCRIPTIONS.map(
          (d) => `${d.options} - ${d.description}`
        );
        const selectedItemOption = await selectFromList(
          chalk.cyan('Item Command Option:'),
          itemCommandOption
        );
        console.log(chalk.blue('Selected option:'), chalk.green(selectedItemOption), '\n');
        const itemOptionKey = selectedItemOption.split(' - ')[0];

        if (itemOptionKey === 'with') {
          const item = await selectItem();
          generatedCommand = `/item replace ${targetOptionResult} ${slot} with ${item}`;
        }
        if (itemOptionKey === 'from') {
          const sourceTarget = await selectFromList(chalk.cyan('Source target:'), targets);

          let sourceTargetOptionResult = '';

          switch (sourceTarget) {
            case 'block': {
              let blockPos = '';
              const posResult = await fillOutForm(
                chalk.cyan(
                  'Enter the block coordinates(e.g., 0 64 0 or ~ ~ ~) to specify target block...'
                ),
                [
                  {
                    name: 'X',
                    message: 'X coordinate',
                    initial: '~',
                    type: 'number | ~ | ^',
                  },
                  {
                    name: 'Y',
                    message: 'Y coordinate',
                    initial: '~',
                    type: 'number | ~ | ^',
                  },
                  {
                    name: 'Z',
                    message: 'Z coordinate',
                    initial: '~',
                    type: 'number | ~ | ^',
                  },
                ],
                false
              );
              blockPos = `${posResult.X} ${posResult.Y} ${posResult.Z}`;
              sourceTargetOptionResult = `block ${blockPos}`;
              break;
            }

            case 'entity': {
              const targetEntity = await addSelectorsQuestion();
              sourceTargetOptionResult = `entity ${targetEntity}`;
              break;
            }
          }
          const sourceTargetSlot = await getSlot();

          generatedCommand = `/item replace ${targetOptionResult} ${slot} from ${sourceTargetOptionResult} ${sourceTargetSlot}`;
        }

        break;
      }

      case 'effect': {
        const effects = await loadDataLists('effects', 'EFFECTS');
        const effectTypes = [
          'give - give an effect to entit(ies)',
          'clear - take one or all effect(s) from entit(ies)',
        ];
        let selectedEffectType = await selectFromList(chalk.cyan('Effect type:'), effectTypes);
        selectedEffectType = selectedEffectType.split(' - ')[0];

        let effectTarget = '';
        let effectName = '';
        let duration = '';
        let amplifier = '';
        let effectOption: boolean = false;

        switch (selectedEffectType) {
          case 'give': {
            effectTarget = await addSelectorsQuestion();
            while (true) {
              const input = await autoComplete(
                chalk.cyan('Select a effect name...'),
                chalk.cyan('Enter a effect name(e.g., night_vision etc.)...'),
                effects,
                false
              );
              if (!input) {
                console.log(error, chalk.red('Please enter effect name.'));
                continue;
              }
              // 入力値の正規化（minecraft:がなければ付与）
              const fullName = input.startsWith('minecraft:') ? input : `minecraft:${input}`;

              // 存在チェック (effects内には既にminecraft:が含まれているため、fullNameと直接比較)
              if (!effects.includes(fullName)) {
                const suggestions = suggestSimilar(fullName, effects).filter(
                  (s) => s !== '__BACK__'
                );

                console.log(chalk.red(`Effect "${fullName}" not found.`));
                if (suggestions.length > 0) {
                  console.log(chalk.yellow('Did you mean:'));
                  for (const s of suggestions) {
                    console.log(`  - ${s}`);
                  }
                }
                console.log(
                  error,
                  chalk.cyan('Please enter a valid effect name (try Tab to autocomplete).')
                );
                continue;
              }

              // チェック通過
              effectName = fullName;
              break;
            }

            const isInfiniteLists = [
              'infinite - Set the effect duration to infinite',
              'seconds - Set the effect duration in seconds',
            ];

            let isInfiniteAnswer = await selectFromList(
              'Which option do you want to use to specify the duration of the effect?',
              isInfiniteLists
            );

            isInfiniteAnswer = isInfiniteAnswer.split(' - ')[0];

            let isInfinite: boolean = false;
            switch (isInfiniteAnswer) {
              case 'infinite': {
                isInfinite = true;
                break;
              }
              case 'seconds': {
                isInfinite = false;
                break;
              }
            }

            if (isInfinite === true) {
              duration = 'infinite';
            }

            if (isInfinite === false) {
              while (true) {
                duration = await createQuestion(chalk.cyan('Duration in seconds: '));
                break;
              }
            }

            amplifier = await createQuestion(
              chalk.cyan('Amplifier level (0 for level 1, 0-255): ')
            );

            effectOption = await toggleQuestion('Do you want to hide the effect particles?');

            generatedCommand = `/effect give ${effectTarget} ${effectName} ${duration} ${amplifier} ${effectOption}`;
            break;
          }
          case 'clear': {
            effectTarget = await addSelectorsQuestion();
            const clearOption = await toggleQuestion(
              'Do you want to clear all effects or only a specific effect?',
              'All',
              'Specific'
            );

            const isAll = clearOption;

            switch (isAll) {
              case true: {
                generatedCommand = `/effect clear ${effectTarget}`;
                break;
              }
              case false: {
                while (true) {
                  const input = await autoComplete(
                    chalk.cyan('Select a effect name...'),
                    chalk.cyan('Enter a effect name(e.g., night_vision etc.)...'),
                    effects,
                    false
                  );
                  if (!input) {
                    console.log(error, chalk.red('Please enter effect name.'));
                    continue;
                  }
                  // 入力値の正規化（minecraft:がなければ付与）
                  const fullName = input.startsWith('minecraft:') ? input : `minecraft:${input}`;

                  // 存在チェック (effects内には既にminecraft:が含まれているため、fullNameと直接比較)
                  if (!effects.includes(fullName)) {
                    const suggestions = suggestSimilar(fullName, effects).filter(
                      (s) => s !== '__BACK__'
                    );

                    console.log(chalk.red(`Effect "${fullName}" not found.`));
                    if (suggestions.length > 0) {
                      console.log(chalk.yellow('Did you mean:'));
                      for (const s of suggestions) {
                        console.log(`  - ${s}`);
                      }
                    }
                    console.log(
                      error,
                      chalk.cyan('Please enter a valid effect name (try Tab to autocomplete).')
                    );
                    continue;
                  }

                  // チェック通過
                  effectName = fullName;
                  break;
                }
                generatedCommand = `/effect clear ${effectTarget} ${effectName}`;
                break;
              }
            }
          }
        }
        break;
      }
      case 'enchant': {
        const enchantmentsData = await loadDataLists<enchantmentData>(
          'enchantments',
          'ENCHANTMENTS'
        );
        const enchantments = enchantmentsData.map((e) => e.name);
        const target = await addSelectorsQuestion();

        let enchantment = '';
        let normalizedEnch = '';
        while (true) {
          enchantment = await autoComplete(
            chalk.cyan('Enchantment name (e.g., minecraft:sharpness):'),
            chalk.cyan('Enter an enchantment name (e.g., minecraft:sharpness)...'),
            enchantments,
            false
          );

          // Normalize enchantments id: allow with or without minecraft: prefix
          normalizedEnch = enchantment.startsWith('minecraft:')
            ? enchantment.slice(10)
            : enchantment;
          const exists = enchantments.includes(normalizedEnch);

          if (!enchantment.trim()) {
            console.log(error, chalk.red('Please enter enchantments name.'));
            continue;
          }
          if (!exists) {
            const suggestions = suggestSimilar(normalizedEnch, enchantments).map(
              (s) => `minecraft:${s}`
            );

            console.log(chalk.red(`Enchantments "${enchantment}" not found.`));
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

          break;
        }
        const selectedEnchData = enchantmentsData.find((e) => e.name === normalizedEnch);
        const maxLevel = selectedEnchData ? selectedEnchData.maxLevel : 1;

        const lvString = maxLevel === 1 ? '1' : `1~${maxLevel}`;

        let enchantmentLevel = '';
        if (maxLevel === 1) {
          enchantmentLevel = '1';
        } else {
          while (true) {
            enchantmentLevel = await createQuestion(
              chalk.cyan(`Enchantment(${enchantment}) level(${lvString}): `),
              1
            );

            if (!enchantmentLevel.trim()) {
              console.log(error, chalk.red(' Please enter an enchantment level.'));
              continue;
            }
            const levelNum = parseInt(enchantmentLevel.trim(), 10);
            if (Number.isNaN(levelNum) || levelNum < 1 || levelNum > maxLevel) {
              console.log(
                error,
                chalk.red(` Please enter a valid enchantment level(${lvString}).`)
              );
              continue;
            }

            // 正しいレベルが入力されたら、リストに追加してループを抜ける
            console.log(
              chalk.blue(`Enchantment: `),
              `${chalk.green.bold(`${enchantment} - ${levelNum}`)}`
            );
            console.log('\n');
            break;
          }
        }

        generatedCommand = `/enchant ${target} ${enchantment} ${enchantmentLevel}`;
        break;
      }
      // TODO: summonコマンド生成の処理実装
      // TODO: xpコマンド生成の処理実装
      // TODO: 他コマンドの生成処理も実装する

      default: {
        console.log(error, chalk.red(`Unknown command type: ${commandType}`));
        console.log(warn, chalk.yellow(`"${commandType}" is not yet supported. Sorry!`));
        process.exit(1);
      }
    }

    if (options.slash === false && generatedCommand.startsWith('/')) {
      generatedCommand = generatedCommand.slice(1);
    }

    const spinner = ora({
      text: 'Generating commands...',
      discardStdin: false,
    }).start();
    await new Promise((r) => setTimeout(r, 600));
    spinner.stop();

    console.log(
      `${success} ${chalk.green('Generated! Command:')} ${chalk.blue(`${generatedCommand}`)}\n`
    );

    if (options.copy === 'false') {
      process.exit();
    } else {
      if (!Input) {
        throw new Error('enquirer Input not available');
      }

      const confirmPrompt = new Input({
        message: `${info} ${chalk.cyan('Press Enter to copy to clipboard...')}`,
        /*
        onCancel: () => {
          throw new CancelError();
        },
        */
      }) as EnquirerBasePrompt;

      try {
        await runPrompt(confirmPrompt);
        await clipboard.write(generatedCommand);
        console.log(success, chalk.green('Command copied to clipboard!'));
        if (!options.silent) {
          sendNotify('Minecraft-Command-Gen-CLI', '✅️ The command was copied successfully.');
        }
      } catch {
        console.log(error, chalk.red('Failed to copy command to clipboard'));
        if (!options.silent) {
          sendNotify('Minecraft-Command-Gen-CLI', '❌️ Failed to copy command.');
        }
      }
    }
  });

  return cmd;
}
