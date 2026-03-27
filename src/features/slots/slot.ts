import chalk from 'chalk';
import { createQuestion, selectFromList } from '../../util/questionsFunc.js';
import { error } from '../../util/symbols.js';
import { SLOTS_DESCRIPTIONS } from '../../util/utils.js';

export async function getSlot(): Promise<string> {
  const slotTargets = SLOTS_DESCRIPTIONS.map((s) => `${s.slot} - ${s.description}`);
  const selectedSlotTarget = await selectFromList(chalk.cyan('Select slot type:'), slotTargets);
  const slotKey = selectedSlotTarget.split(' - ')[0];
  const slotInfo = SLOTS_DESCRIPTIONS.find((s) => s.slot === slotKey);
  let slotNumberPart = '';

  // slotInfo.hasSlotNumber が true の場合のみ、スロット番号を聞く
  // TypeScriptの型推論により、このif文の中では slotInfo.slotNumberRange が確実に存在すると判定されます
  if (slotInfo?.hasSlotNumber) {
    // '0-53' のような文字列を '-' で分割し、最小値と最大値を取得する
    const [minStr, maxStr] = slotInfo.slotNumberRange.split('-');
    const min = parseInt(minStr, 10);
    const max = parseInt(maxStr, 10);

    while (true) {
      const slotNumStr = await createQuestion(chalk.cyan(`Enter slot number (${min}-${max}): `));

      // 空入力のチェック
      if (!slotNumStr.trim()) {
        console.log(error, chalk.red('Please enter a slot number.'));
        continue;
      }

      const slotNum = parseInt(slotNumStr.trim(), 10);

      // バリデーション: 数値であること、かつ min 以上 max 以下であること
      if (Number.isNaN(slotNum) || slotNum < min || slotNum > max) {
        console.log(error, chalk.red(`Please enter a valid number between ${min} and ${max}.`));
        continue;
      }

      // Minecraftのコマンド仕様に合わせて、ドットで繋ぐ準備をする (例: .5)
      slotNumberPart = `.${slotNum}`;
      break;
    }
  }

  if (slotInfo?.slot === 'weapon' || slotInfo?.slot === 'armor') {
    switch (slotInfo?.slot) {
      case 'weapon': {
        const weaponChoices = ['mainhand', 'offhand'];
        const selectedWeaponSlot = await selectFromList(
          chalk.cyan('Select weapon slot:'),
          weaponChoices
        );
        slotNumberPart = `.${selectedWeaponSlot}`;
        break;
      }
      case 'armor': {
        const armorChoices = ['head', 'chest', 'legs', 'feet'];
        const selectedArmorSlot = await selectFromList(
          chalk.cyan('Select armor slot:'),
          armorChoices
        );
        slotNumberPart = `.${selectedArmorSlot}`;
        break;
      }
    }
  }

  const finalSlot = `${slotKey}${slotNumberPart}`;
  return finalSlot;
}
