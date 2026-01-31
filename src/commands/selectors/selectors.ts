import chalk from 'chalk';
import { createQuestion, selectFromList } from '../create.js';

export async function addtionalSelectorsQuestion(): Promise<string> {
  console.log(`${chalk.blue('Further target selector:')} ${chalk.green(`${chalk.bold('Yes')}`)}`);

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
        const distance = await createQuestion(
          chalk.cyan('Distance to entity (int or range e.g., 1..5 = 1~5). Type "back" to go back: ')
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
      }
      case 'score': {
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
      }
      case 'tag': {
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
      }
      case 'team': {
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
      }
      case 'limit': {
        const limit = await createQuestion(chalk.cyan('Limit(int). Type "back" to go back: '));
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
      }
      case 'level': {
        const level = await createQuestion(
          chalk.cyan("Exp Level(int or range format: '10' or '10..20'). Type \"back\" to go back: ")
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
      }
      case 'gamemode': {
        const gamemodeList = [
          'survival',
          'creative',
          'adventure',
          'spectator',
          'back - Go back to selector selection',
        ];
        const gamemodeResult = await selectFromList(chalk.cyan('Player Gamemode: '), gamemodeList);
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
        const advancement = await createQuestion(
          chalk.cyan('Advancement(format: <advancement_ID>=true/false). Type "back" to go back: ')
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

        console.log(chalk.blue(`Advancements:`), `${chalk.green(`${chalk.bold(advancement)}`)}`);
        console.log('\n');
        addedSelectors.push(`advancements=${advancement}`);
        break;
      }
      case 'predicate': {
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
  const addSelectors = addedSelectors.join(',');

  console.log(
    chalk.blue(`All selectors:`),
    `${chalk.green(`${chalk.bold(addSelectors || null)}`)}`
  );
  console.log('\n');

  return addSelectors;
}
