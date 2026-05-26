import chalk from 'chalk';
import ora from 'ora';
import { createQuestion, selectFromList, toggleQuestion } from '../../util/questionsFunc.js';
import { error, warn } from '../../util/symbols.js';

let addSelectors: string = '';
let selector: string = '';

export async function addSelectorsQuestion(): Promise<string> {
  let targetType: string = '';
  while (true) {
    // Choose Target
    const target = [
      '@p - Near Player',
      '@a - All Player',
      '@s - Myself',
      '@r - Random Player',
      "@n - A Nearest Player (1.21+, same '@p[sort=nearest, limit=1]')",
      'PlayerName',
    ];
    targetType = (await selectFromList('Select a target selector type:', target)).split(' ')[0];

    if (targetType === 'PlayerName') {
      while (true) {
        const playerName = await createQuestion(
          chalk.cyan('Enter the player name. Type "back" to go back.\n') +
            chalk.italic.gray(
              'Note: When it is entered, the existence of the player name will be checked in Mojang API.\nIf you skip this check, please add "!" at the beginning of the player name.\n'
            )
        );
        if (!playerName.trim()) {
          console.log(error, chalk.red('Please enter a player name.'));
          continue;
        }
        if (playerName.trim().toLowerCase() === 'back') {
          console.log(warn, chalk.yellow(' Cancelled. Back to components selection.'));
          console.log('\n');
          targetType = 'backed';
          break;
        }
        targetType = playerName.trim();

        if (!targetType.startsWith('!')) {
          // ===== 追加: Mojang API で存在確認 =====
          const spinner = ora('Checking player existence...').start();
          try {
            const response = await fetch(
              `https://api.mojang.com/users/profiles/minecraft/${targetType}`
            );
            if (response.status === 404 || response.status === 204) {
              spinner.warn(
                chalk.bgYellow.black(' WARN ') +
                  chalk.yellow(` Player "${targetType}" does not exist in Mojang database.`)
              );
              const proceed = await toggleQuestion(chalk.cyan('Use this name anyway?'));
              if (!proceed) {
                continue; // もう一度入力をやり直す
              }
            } else if (response.ok) {
              spinner.succeed(
                chalk.green(
                  `Player "${targetType}" is existed in Mojang database! Using this name.`
                )
              );
            } else {
              spinner.stop(); // レート制限(429)などの場合はスルー
            }

            // eslint-disable-next-line @typescript-eslint/no-unused-vars
          } catch (_e) {
            spinner.stop(); // ネットワークエラー等もとりあえずスルー
          }
          // =======================================
        }
        if (targetType.startsWith('!')) {
          targetType = targetType.slice(1); // 先頭の'!'を削除してプレイヤー名として使用
        }

        break;
      }
    }

    if (targetType === 'backed') {
      continue;
    }

    break;
  }
  console.log(chalk.blue(`Target:`), `${chalk.green(`${chalk.bold(targetType)}`)}`);
  console.log('\n');

  if (!targetType.startsWith('@')) {
    console.log(
      `${chalk.blue('Additional target selectors:')} ${chalk.green(`${chalk.bold('Skipped (Specific Player Name)')}`)}`
    );
    console.log('\n');
    return targetType;
  }

  // Q2.5: Ask to refine target selector
  const refineSelector = await toggleQuestion(
    chalk.cyan(
      'Do you want to add more target selectors? (e.g., distance, score, tag, team, etc.)'
    ),
    'Yes',
    'No'
  );

  const shouldRefine = refineSelector === true;

  if (shouldRefine) {
    console.log(
      `${chalk.blue('Additional target selectors:')} ${chalk.green(`${chalk.bold('Yes')}`)}`
    );

    const addedSelectors: string[] = [];
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
        case 'distance': {
          while (true) {
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
              continue;
            }
            console.log(chalk.blue(`Distance:`), `${chalk.green(`${chalk.bold(distance)}`)}`);
            console.log('\n');
            addedSelectors.push(`distance=${distance}`);
            break;
          }
          break;
        }
        case 'score': {
          while (true) {
            const score = await createQuestion(
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
              continue;
            }
            console.log(chalk.blue(`Score:`), `${chalk.green(`${chalk.bold(score)}`)}`);
            console.log('\n');
            addedSelectors.push(`score=${score}`);
            break;
          }
          break;
        }
        case 'tag': {
          while (true) {
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
              continue;
            }
            console.log(chalk.blue(`Tag:`), `${chalk.green(`${chalk.bold(tag)}`)}`);
            console.log('\n');
            addedSelectors.push(`tag=${tag}`);
            break;
          }
          break;
        }
        case 'team': {
          while (true) {
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
              continue;
            }
            console.log(chalk.blue(`Team:`), `${chalk.green(`${chalk.bold(team)}`)}`);
            console.log('\n');
            addedSelectors.push(`team=${team}`);
            break;
          }
          break;
        }
        case 'limit': {
          while (true) {
            const limit = await createQuestion(chalk.cyan('Limit(int). Type "back" to go back: '));
            if (limit.trim().toLowerCase() === 'back') {
              console.log(chalk.yellow('Cancelled. Back to selector selection.'));
              console.log('\n');
              break;
            }
            if (!limit.trim()) {
              console.log(chalk.red('Please enter a limit.'));
              continue;
            }
            console.log(chalk.blue(`Limit:`), `${chalk.green(`${chalk.bold(limit)}`)}`);
            console.log('\n');
            addedSelectors.push(`limit=${limit}`);
            break;
          }
          break;
        }
        case 'level': {
          while (true) {
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
              continue;
            }
            console.log(chalk.blue(`Level:`), `${chalk.green(`${chalk.bold(level)}`)}`);
            console.log('\n');
            addedSelectors.push(`level=${level}`);
            break;
          }
          break;
        }
        case 'gamemode': {
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
        }
        case 'advancements': {
          while (true) {
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
              continue;
            }

            console.log(
              chalk.blue(`Advancements:`),
              `${chalk.green(`${chalk.bold(advancement)}`)}`
            );
            console.log('\n');
            addedSelectors.push(`advancements=${advancement}`);
            break;
          }
          break;
        }
        case 'predicate': {
          while (true) {
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
              continue;
            }

            console.log(chalk.blue(`Predicate:`), `${chalk.green(`${chalk.bold(predicate)}`)}`);
            console.log('\n');
            addedSelectors.push(`predicate=${predicate}`);
            break;
          }
          break;
        }
        case 'sort': {
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
        }
        case 'OK': {
          continueAdding = false;
          break;
        }
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
      `${chalk.blue('Additional target selectors:')} ${chalk.green(`${chalk.bold('No')}`)}`
    );
    console.log('\n');
  }

  const addedSelectorsTF: boolean = !!addSelectors;

  if (addedSelectorsTF) {
    selector = `${targetType}[${addSelectors}]`;
  } else {
    selector = `${targetType}`;
  }

  return selector;
}
