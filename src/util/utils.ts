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
