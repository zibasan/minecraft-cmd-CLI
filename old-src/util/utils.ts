export const COMPONENT_DESCRIPTIONS: Record<string, string> = {
  item_name: 'Item Name(Override the original name)',
  custom_name:
    'Item Name(looks like it was edited with an anvil, do not override the original name)',
  lore: 'Item Lore',
  damage: 'How much to reduce the durability',
  enchantment_glint_override: 'Whether show glint of enchantment(no enchantments)',
  enchantments: 'Item Enchantments',
  food: 'Setting edible items',
  break_sound: 'The sound played when the item is broken',
  max_damage: 'The maximum durability value of that item',
  max_stack_size: 'The maximum stack size value of that item',
  can_break: 'Specify breakable blocks in adventure mode',
  can_place_on: 'Specify the blocks on witch this can be placed in adventure mode',
  rarity: 'Item Rarity',
};

export const TP_COMMAND_DESCRIPTIONS: {
  cmd: string;
  description: string;
}[] = [
  {
    cmd: '1. tp <destination>',
    description: 'Teleport to a specific entity',
  },
  {
    cmd: '2. tp <targets> <destination>',
    description: 'Teleport specific targets to a specific entity',
  },
  {
    cmd: '3. tp <location>',
    description: 'Teleport to a specific location',
  },
  {
    cmd: '4. tp <targets> <location>',
    description: 'Teleport specific targets to a specific location',
  },
  {
    cmd: '5. tp <targets> <location> <rotation>',
    description: 'Teleport specific targets to a specific location with rotation(angle)',
  },
  {
    cmd: '6. tp <targets> <location> facing <facingRotation>',
    description: 'Teleport specific targets to a specific location with rotation(coordinate)',
  },
  {
    cmd: '7. tp <targets> <location> facing entity <facingEntity> [facingAnchor]',
    description: 'Teleport specific targets to a specific location with rotation(facing entity)',
  },
];

export const SETBLOCK_OPTIONS_DESCRIPTIONS: {
  options: string;
  description: string;
}[] = [
  {
    options: 'destroy',
    description:
      'Destroys the original block as if it were destroyed by the player, dropping items',
  },
  {
    options: 'keep',
    description: 'A new block is placed only if the original block is air',
  },
  {
    options: 'replace',
    description: '(default) Simply place the new block',
  },
  {
    options: 'strict',
    description: 'Places the block as is without triggering a block update or geometry update',
  },
  {
    options: 'Skip',
    description: 'Skip this question and use the default option "replace"',
  },
];

type Slot = {
  slot: string; // スロットのキー名
  description: string; // そのスロットの説明
};

type DependentInfo =
  | { hasSlotNumber: true; slotNumberRange: string } // スロット番号が必要かどうかと、その範囲（例: "0-26"）
  | { hasSlotNumber: false; slotNumberRange?: never }; // スロット番号が不要な場合

type SlotDesc = Slot & DependentInfo;

export const SLOTS_DESCRIPTIONS: SlotDesc[] = [
  {
    slot: 'contents',
    description:
      'An entity that has only one item slot; item, item_frame etc (This option isn\'t available when targeting "block").',
    hasSlotNumber: false,
  },
  {
    slot: 'container',
    description:
      'A block that has multiple item slots; chest, hopper etc (This option isn\'t available when targeting "entity").',
    hasSlotNumber: true,
    slotNumberRange: '0-53',
  },
  {
    slot: 'hotbar',
    description:
      'Player inventory hotbar slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: true,
    slotNumberRange: '0-8',
  },
  {
    slot: 'inventory',
    description:
      'Player inventory slots (excluding hotbar) (This option isn\'t available when targeting "block")',
    hasSlotNumber: true,
    slotNumberRange: '0-26',
  },
  {
    slot: 'enderchest',
    description: 'Ender chest slots (This option isn\'t available when targeting "entity")',
    hasSlotNumber: true,
    slotNumberRange: '0-26',
  },
  {
    slot: 'villager',
    description: 'Villager trading slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: true,
    slotNumberRange: '0-7',
  },
  {
    slot: 'player.crafting',
    description: 'Player crafting grid slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: true,
    slotNumberRange: '0-3',
  },
  {
    slot: 'horse',
    description: 'Horse inventory slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: true,
    slotNumberRange: '0-14',
  },
  {
    slot: 'weapon',
    description: 'Player weapon slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: false,
  },
  {
    slot: 'armor',
    description: 'Player armor slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: false,
  },
  {
    slot: 'saddle',
    description: 'Saddle slot (This option isn\'t available when targeting "block")',
    hasSlotNumber: false,
  },
  {
    slot: 'horse.chest',
    description: 'Horse chest slots (This option isn\'t available when targeting "block")',
    hasSlotNumber: false,
  },
  {
    slot: 'player.cursor',
    description:
      'The item the cursor is holding (This option isn\'t available when targeting "block")',
    hasSlotNumber: false,
  },
];

export const ITEM_COMMANDS_DESCRIPTIONS: {
  options: string;
  description: string;
}[] = [
  {
    options: 'with',
    description: 'Replaces the item in the specified slot with the specified item',
  },
  {
    options: 'from',
    description: 'Copies the item from the specified slot to the target slot',
  },
];
