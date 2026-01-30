import { Command } from 'commander';
import chalk from 'chalk';
import clipboard from 'clipboardy';
import { createInterface } from 'readline';

function createQuestion(query: string): Promise<string> {
  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(query, (answer) => {
      rl.close();
      resolve(answer);
    });
  });
}

async function selectFromList(message: string, choices: string[]): Promise<string> {
  console.log(chalk.cyan(message));
  choices.forEach((choice, index) => {
    console.log(`  ${chalk.yellow(index + 1)}. ${choice}`);
  });

  let selection: number = 0;
  let validInput = false;

  while (!validInput) {
    const input = (await createQuestion(chalk.cyan('Choose: '))).trim();

    // Ensure the input is a strict integer string (digits only) before parsing.
    if (!/^[0-9]+$/.test(input)) {
      console.log(chalk.red(`Invalid selection. Please choose 1-${choices.length}`));
      continue;
    }
    selection = parseInt(input, 10);

    if (selection >= 1 && selection <= choices.length) {
      validInput = true;
    } else {
      console.log(chalk.red(`Invalid selection. Please choose 1-${choices.length}`));
    }
  }

  return choices[selection - 1];
}

export function createCommand(): Command {
  const cmd = new Command('create');
  cmd.description('Generate Minecraft commands');
  cmd.action(async () => {
    const supportedTypes = ['give', 'teleport', 'setblock', 'fill', 'say', 'execute'];

    // Q1: Select command type
    const commandType = await selectFromList('Select a command type:', supportedTypes);

    console.log(chalk.blue(`Generate target:`), `${chalk.green(`${chalk.bold(commandType)}`)}`);

    let generatedCommand = '';
    let selector = '';
    let addSelectors = '';

    console.log('\n');

    switch (commandType) {
      case 'give':
        // Q2: Choose Target
        const target = [
          '@p - Near Player',
          '@a - All Player',
          '@s - Myself',
          '@r - Random Player',
          "@n - A Nearest Player (1.21+, same '@p[sort=nearest, limit=1]')",
        ];
        const targetType = (await selectFromList('Select a target selector type:', target)).split(
          ' '
        )[0];

        console.log(chalk.blue(`Target:`), `${chalk.green(`${chalk.bold(targetType)}`)}`);
        console.log('\n');

        // Q2.5: Ask to refine target selector
        const refineSelector = await createQuestion(
          chalk.cyan('Want to further refine your target selector? (y/N): ')
        );
        const shouldRefine = refineSelector.toLowerCase() === 'y';

        if (shouldRefine) {
          console.log(
            `${chalk.blue('Further target selector:')} ${chalk.green(`${chalk.bold('Yes')}`)}`
          );

          const allSelectorTypes = [
            'distance',
            'score',
            'tag',
            'team',
            'limit',
            'level',
            'gamemode',
            'advancements',
            'predicate',
            'sort',
          ];

          const addedSelectors: string[] = [];
          let continueAdding = true;

          while (continueAdding) {
            const availableSelectors = allSelectorTypes.filter(
              (selector) => !addedSelectors.some((added) => added.split('=')[0] === selector)
            );

            const selectorOptions = [
              ...availableSelectors.map((s) => {
                const descriptions: { [key: string]: string } = {
                  distance: 'Distance to Entity(=Player)',
                  score: 'The score value or range which the entity has',
                  tag: 'The tag which the entity has',
                  team: 'The team which the entity joins',
                  limit: 'Amount limit',
                  level: 'Experience level',
                  gamemode: 'Player gamemode',
                  advancements: 'The advancements which the player has',
                  predicate: 'Match predicates(required datapacks)',
                  sort: 'Specify the order in which to select targets',
                };
                return `${s} - ${descriptions[s] || s}`;
              }),
              'OK',
            ];

            // Q2.55 Selector Type
            const addSelectorsType = (
              await selectFromList('Additional selectors: ', selectorOptions)
            ).split(' ')[0];

            console.log(
              chalk.blue(`Additional selector:`),
              `${chalk.green(`${chalk.bold(addSelectorsType)}`)}`
            );
            console.log('\n');

            switch (addSelectorsType) {
              case 'distance':
                const distance = await createQuestion(
                  chalk.cyan(
                    'Distance to entity (int or range e.g., 1..5 = 1~5). Type "back" to go back: '
                  )
                );
                if (distance.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!distance.trim()) {
                  console.log(chalk.red('Please enter a distance value or range.'));
                  process.exit(1);
                }
                console.log(chalk.blue(`Distance:`), `${chalk.green(`${chalk.bold(distance)}`)}`);
                console.log('\n');
                addedSelectors.push(`distance=${distance}`);
                break;

              case 'score':
                let score = await createQuestion(
                  chalk.cyan(
                    "Score (format: score value or range of A = 'A=1' or 'A=1..10'). Type \"back\" to go back: "
                  )
                );
                if (score.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!score.trim()) {
                  console.log(chalk.red('Please enter a score format.'));
                  process.exit(1);
                }
                console.log(chalk.blue(`Score:`), `${chalk.green(`${chalk.bold(score)}`)}`);
                console.log('\n');
                addedSelectors.push(`score=${score}`);
                break;

              case 'tag':
                const tag = await createQuestion(
                  chalk.cyan(
                    'Tag (format: <your-tag> | If it put \'!\' at the beginning, the tag will be excluded.). Type "back" to go back: '
                  )
                );
                if (tag.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!tag.trim()) {
                  console.log(chalk.red('Please enter a tag.'));
                  process.exit(1);
                }
                console.log(chalk.blue(`Tag:`), `${chalk.green(`${chalk.bold(tag)}`)}`);
                console.log('\n');
                addedSelectors.push(`tag=${tag}`);
                break;

              case 'team':
                const team = await createQuestion(
                  chalk.cyan(
                    'Team (format: <your-team> | If it put \'!\' at the beginning, the team will be excluded.). Type "back" to go back: '
                  )
                );
                if (team.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!team.trim()) {
                  console.log(chalk.red('Please enter a team.'));
                  process.exit(1);
                }
                console.log(chalk.blue(`Team:`), `${chalk.green(`${chalk.bold(team)}`)}`);
                console.log('\n');
                addedSelectors.push(`team=${team}`);
                break;

              case 'limit':
                const limit = await createQuestion(
                  chalk.cyan('Limit(int). Type "back" to go back: ')
                );
                if (limit.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!limit.trim()) {
                  console.log(chalk.red('Please enter a limit.'));
                  process.exit(1);
                }
                console.log(chalk.blue(`Limit:`), `${chalk.green(`${chalk.bold(limit)}`)}`);
                console.log('\n');
                addedSelectors.push(`limit=${limit}`);
                break;

              case 'level':
                const level = await createQuestion(
                  chalk.cyan(
                    "Exp Level(int or range format: '10' or '10..20'). Type \"back\" to go back: "
                  )
                );
                if (level.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!level.trim()) {
                  console.log(chalk.red('Please enter a level.'));
                  process.exit(1);
                }
                console.log(chalk.blue(`Level:`), `${chalk.green(`${chalk.bold(level)}`)}`);
                console.log('\n');
                addedSelectors.push(`level=${level}`);
                break;

              case 'gamemode':
                const gamemodeList = [
                  'survival',
                  'creative',
                  'adventure',
                  'spectator',
                  'back - Go back to selector selection',
                ];
                const gamemodeResult = await selectFromList(
                  chalk.cyan('Player Gamemode: '),
                  gamemodeList
                );
                const gamemode = gamemodeResult.split(' ')[0];

                if (gamemode.toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }

                console.log(chalk.blue(`Gamemode:`), `${chalk.green(`${chalk.bold(gamemode)}`)}`);
                console.log('\n');
                addedSelectors.push(`gamemode=${gamemode}`);
                break;

              case 'advancements':
                const advancement = await createQuestion(
                  chalk.cyan(
                    'Advancement(format: <advancement_ID>=true/false). Type "back" to go back: '
                  )
                );
                if (advancement.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!advancement.trim()) {
                  console.log(chalk.red('Please enter an advancement.'));
                  process.exit(1);
                }

                console.log(
                  chalk.blue(`Advancements:`),
                  `${chalk.green(`${chalk.bold(advancement)}`)}`
                );
                console.log('\n');
                addedSelectors.push(`advancements=${advancement}`);
                break;

              case 'predicate':
                const predicate = await createQuestion(
                  chalk.cyan(
                    'Predicate(predicate_id | If it put \'!\' at the beginning, the predicate will be excluded.). Type "back" to go back: '
                  )
                );
                if (predicate.trim().toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }
                if (!predicate.trim()) {
                  console.log(chalk.red('Please enter a predicate.'));
                  process.exit(1);
                }

                console.log(chalk.blue(`Predicate:`), `${chalk.green(`${chalk.bold(predicate)}`)}`);
                console.log('\n');
                addedSelectors.push(`predicate=${predicate}`);
                break;

              case 'sort':
                const sortList = [
                  'nearest - Select from the nearest entity first',
                  'furthest - Select from the furthest entity first',
                  'random - Select random',
                  'arbitrary - Select by spawn time',
                  'back - Go back to selector selection',
                ];
                const sortResult = await selectFromList(chalk.cyan('Sort: '), sortList);
                const sort = sortResult.split(' ')[0];

                if (sort.toLowerCase() === 'back') {
                  console.log(chalk.yellow('Cancelled. Back to selector selection.'));
                  console.log('\n');
                  break;
                }

                console.log(chalk.blue(`Sort:`), `${chalk.green(`${chalk.bold(sort)}`)}`);
                console.log('\n');
                addedSelectors.push(`sort=${sort}`);
                break;

              case 'OK':
                continueAdding = false;
                break;
            }
          }

          addSelectors = addedSelectors.join(',');
          console.log(
            chalk.blue(`All selectors:`),
            `${chalk.green(`${chalk.bold(addSelectors || null)}`)}`
          );
          console.log('\n');
        } else {
          console.log(
            `${chalk.blue('Further target selector:')} ${chalk.green(`${chalk.bold('No')}`)}`
          );
        }

        if (addSelectors) {
          selector = `${targetType}[${addSelectors}]`;
        } else {
          selector = `${targetType}`;
        }

        console.log('\n');

        // Q3: Enter item name
        const itemName = await createQuestion(chalk.cyan('Item name (e.g., diamond): '));
        if (!itemName.trim()) {
          console.log(chalk.red('Please enter an item name.'));
          process.exit(1);
        }
        console.log(chalk.blue(`Item name:`), `${chalk.green(`${chalk.bold(itemName)}`)}`);
        console.log('\n');

        // Q4: Amount
        let amount = await createQuestion(
          chalk.cyan("Item amount(How many? If empty, it'll set 1.): ")
        );
        if (!amount.trim()) {
          amount = '1';
        }
        console.log(chalk.blue(`Item amount:`), `${chalk.green(`${chalk.bold(amount)}`)}`);
        console.log('\n');

        generatedCommand = `/give ${selector} ${itemName} ${amount}`;
        break;

      case 'teleport':
        const destination = await createQuestion(
          chalk.cyan('Destination player/entity or coordinates (e.g., @p or 0 64 0): ')
        );
        if (!destination.trim()) {
          console.log(chalk.red('Please enter a destination.'));
          process.exit(1);
        }
        generatedCommand = `/teleport ${destination}`;
        break;

      case 'setblock':
        const sbPosition = await createQuestion(chalk.cyan('Position (e.g., 0 64 0): '));
        const sbBlock = await createQuestion(chalk.cyan('Block (e.g., diamond_block): '));
        if (!sbPosition.trim() || !sbBlock.trim()) {
          console.log(chalk.red('Please enter position and block.'));
          process.exit(1);
        }
        generatedCommand = `/setblock ${sbPosition} ${sbBlock}`;
        break;

      case 'fill':
        const fillFrom = await createQuestion(chalk.cyan('From position (e.g., 0 64 0): '));
        const fillTo = await createQuestion(chalk.cyan('To position (e.g., 10 64 10): '));
        const fillBlock = await createQuestion(chalk.cyan('Block (e.g., stone): '));
        if (!fillFrom.trim() || !fillTo.trim() || !fillBlock.trim()) {
          console.log(chalk.red('Please enter all positions and block.'));
          process.exit(1);
        }
        generatedCommand = `/fill ${fillFrom} ${fillTo} ${fillBlock}`;
        break;

      case 'say':
        const message = await createQuestion(chalk.cyan('Message: '));
        if (!message.trim()) {
          console.log(chalk.red('Please enter a message.'));
          process.exit(1);
        }
        generatedCommand = `/say ${message}`;
        break;

      case 'execute':
        const execTarget = await createQuestion(chalk.cyan('Target selector (e.g., @a): '));
        const execCommand = await createQuestion(chalk.cyan('Command to execute: '));
        if (!execTarget.trim() || !execCommand.trim()) {
          console.log(chalk.red('Please enter target and command.'));
          process.exit(1);
        }
        generatedCommand = `/execute as ${execTarget} at @s run ${execCommand}`;
        break;
      default:
        console.log(chalk.red(`✗ Unknown command type: ${commandType}`));
        console.log(chalk.red(`✗ "${commandType}" is not yet supported. Sorry!`));
        process.exit(1);
    }

    console.log(`${chalk.green('Generated! Command:')} ${chalk.blue(`${generatedCommand}`)}`);

    // Wait for Enter key
    const rl = createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    rl.question(chalk.cyan('Press Enter to copy to clipboard...'), async () => {
      try {
        await clipboard.write(generatedCommand);
        console.log(chalk.green('✓ Command copied to clipboard!'));
      } catch (error) {
        console.log(chalk.red('✗ Failed to copy command to clipboard'));
      } finally {
        rl.close();
      }
    });
  });

  return cmd;
}
