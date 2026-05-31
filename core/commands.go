package core

import (
	"fmt"
	"regexp"
	"strings"
)

var posTokenRegex = regexp.MustCompile(`^(?:[~^]-?\d+|[~^]|-?\d+)$`)

// IsValidPositionToken checks if a coordinate token is valid (e.g. ~ ~5 -10 ^ ^3).
func IsValidPositionToken(token string) bool {
	return posTokenRegex.MatchString(token)
}

// IsValidPosition checks if a position string contains exactly 3 valid coordinate tokens.
func IsValidPosition(pos string) bool {
	tokens := strings.Fields(pos)
	if len(tokens) != 3 {
		return false
	}
	for _, t := range tokens {
		if !IsValidPositionToken(t) {
			return false
		}
	}
	return true
}

// FormatCommand adds or removes the leading slash based on preferences.
func FormatCommand(cmd string, includeSlash bool) string {
	if includeSlash {
		if !strings.HasPrefix(cmd, "/") {
			return "/" + cmd
		}
		return cmd
	}
	return strings.TrimPrefix(cmd, "/")
}

// 1. Give Command
type GiveCommand struct {
	Selector   string
	Item       string
	Components string // e.g. "item_name='\"Sword\"',damage=10"
	Amount     string
}

func (g *GiveCommand) Build() string {
	itemStr := g.Item
	if g.Components != "" {
		itemStr = fmt.Sprintf("%s[%s]", g.Item, g.Components)
	}
	amountStr := "1"
	if g.Amount != "" {
		amountStr = g.Amount
	}
	return fmt.Sprintf("give %s %s %s", g.Selector, itemStr, amountStr)
}

// 2. Teleport Command
type TeleportCommand struct {
	Type           string // "1" to "7"
	Targets        string
	Destination    string
	Location       string
	Rotation       string
	FacingLocation string
	FacingEntity   string
	FacingAnchor   string
}

func (t *TeleportCommand) Build() string {
	switch t.Type {
	case "1":
		return fmt.Sprintf("teleport %s", t.Destination)
	case "2":
		return fmt.Sprintf("teleport %s %s", t.Targets, t.Destination)
	case "3":
		return fmt.Sprintf("teleport %s", t.Location)
	case "4":
		return fmt.Sprintf("teleport %s %s", t.Targets, t.Location)
	case "5":
		return fmt.Sprintf("teleport %s %s %s", t.Targets, t.Location, t.Rotation)
	case "6":
		return fmt.Sprintf("teleport %s %s facing %s", t.Targets, t.Location, t.FacingLocation)
	case "7":
		anchorStr := ""
		if t.FacingAnchor != "" {
			anchorStr = " " + t.FacingAnchor
		}
		return fmt.Sprintf("teleport %s %s facing entity %s%s", t.Targets, t.Location, t.FacingEntity, anchorStr)
	default:
		return ""
	}
}

// 3. Setblock Command
type SetblockCommand struct {
	Position string
	Block    string
	Option   string // destroy, keep, replace, strict, etc.
}

func (s *SetblockCommand) Build() string {
	blockStr := s.Block
	if !strings.HasPrefix(blockStr, "minecraft:") {
		blockStr = "minecraft:" + blockStr
	}
	optStr := ""
	if s.Option != "" {
		optStr = " " + s.Option
	}
	return fmt.Sprintf("setblock %s %s%s", s.Position, blockStr, optStr)
}

// 4. Fill Command
type FillCommand struct {
	From  string
	To    string
	Block string
}

func (f *FillCommand) Build() string {
	blockStr := f.Block
	if !strings.HasPrefix(blockStr, "minecraft:") {
		blockStr = "minecraft:" + blockStr
	}
	return fmt.Sprintf("fill %s %s %s", f.From, f.To, blockStr)
}

// 5. Say Command
type SayCommand struct {
	Message string
}

func (s *SayCommand) Build() string {
	return fmt.Sprintf("say %s", s.Message)
}

// 6. Execute Command
type ExecuteCommand struct {
	Target  string
	Command string
}

func (e *ExecuteCommand) Build() string {
	return fmt.Sprintf("execute as %s at @s run %s", e.Target, e.Command)
}

// 7. Item Command
type ItemCommand struct {
	TargetType     string // "block" or "entity"
	TargetPosOrSel string // coordinates for block, selector for entity
	Slot           string
	SourceType     string // "block" or "entity"
	SourcePosOrSel string // coordinates for block, selector for entity
	SourceSlot     string
	Item           string // only used if replacing with item
	IsFrom         bool   // true: replace from ..., false: replace with ...
}

func (i *ItemCommand) Build() string {
	target := fmt.Sprintf("%s %s", i.TargetType, i.TargetPosOrSel)

	if i.IsFrom {
		source := fmt.Sprintf("%s %s", i.SourceType, i.SourcePosOrSel)
		return fmt.Sprintf("item replace %s %s from %s %s", target, i.Slot, source, i.SourceSlot)
	}

	return fmt.Sprintf("item replace %s %s with %s", target, i.Slot, i.Item)
}

// 8. Effect Command
type EffectCommand struct {
	Type          string // "give" or "clear"
	Target        string
	Effect        string // used in give or clear specific
	Duration      string // used in give
	Amplifier     string // used in give
	HideParticles bool   // used in give
	ClearAll      bool   // used in clear
}

func (e *EffectCommand) Build() string {
	if e.Type == "clear" {
		if e.ClearAll {
			return fmt.Sprintf("effect clear %s", e.Target)
		}
		effectStr := e.Effect
		if !strings.HasPrefix(effectStr, "minecraft:") {
			effectStr = "minecraft:" + effectStr
		}
		return fmt.Sprintf("effect clear %s %s", e.Target, effectStr)
	}

	effectStr := e.Effect
	if !strings.HasPrefix(effectStr, "minecraft:") {
		effectStr = "minecraft:" + effectStr
	}

	durStr := "infinite"
	if e.Duration != "" {
		durStr = e.Duration
	}

	ampStr := "0"
	if e.Amplifier != "" {
		ampStr = e.Amplifier
	}

	particleStr := "false"
	if e.HideParticles {
		particleStr = "true"
	}

	return fmt.Sprintf("effect give %s %s %s %s %s", e.Target, effectStr, durStr, ampStr, particleStr)
}

// 9. Enchant Command
type EnchantCommand struct {
	Target      string
	Enchantment string
	Level       string
}

func (e *EnchantCommand) Build() string {
	enchStr := e.Enchantment
	if !strings.HasPrefix(enchStr, "minecraft:") {
		enchStr = "minecraft:" + enchStr
	}
	return fmt.Sprintf("enchant %s %s %s", e.Target, enchStr, e.Level)
}

// 10. Summon Command
type SummonCommand struct {
	Entity   string
	Position string
	NBT      string
}

func (s *SummonCommand) Build() string {
	entityStr := s.Entity
	if !strings.HasPrefix(entityStr, "minecraft:") {
		entityStr = "minecraft:" + entityStr
	}
	posStr := ""
	if s.Position != "" {
		posStr = " " + s.Position
	}
	nbtStr := ""
	if s.NBT != "" {
		nbtStr = " " + s.NBT
	}
	return fmt.Sprintf("summon %s%s%s", entityStr, posStr, nbtStr)
}

type NbtTagOption struct {
	Key          string   // NBTのキー名（例: "NoAI"）
	Label        string   // 画面に表示する日本語名
	Description  string   // タグの詳しい解説
	ApplicableTo []string // 適用可能なエンティティIDのリスト（空配列の場合はすべてのエンティティで共通して利用可能とする）
	Type         string   // "boolean", "string", "int", "raw"
}

var breedableMobs = []string{
	"axolotl", "bee", "camel", "cat", "chicken", "cow", "donkey", "fox", "frog", "goat",
	"hoglin", "horse", "llama", "mooshroom", "mule", "ocelot", "panda", "pig", "rabbit",
	"sheep", "sniffer", "strider", "trader_llama", "turtle", "villager", "wolf",
}

var NbtMasterList = []NbtTagOption{
	{Key: "NoAI", Label: "NoAI - Disable entity AI (does not move or attack)", Description: "AIを無効化 (NoAI)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Invulnerable", Label: "Invulnerable - Immune to all damage", Description: "無敵化 (Invulnerable)", ApplicableTo: nil, Type: "boolean"},
	{Key: "NoGravity", Label: "NoGravity - Disable gravity", Description: "重力無効 (NoGravity)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Silent", Label: "Silent - Silence all sounds from this entity", Description: "消音 (Silent)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Glowing", Label: "Glowing - Make the entity outline glow", Description: "発光 (Glowing)", ApplicableTo: nil, Type: "boolean"},
	{Key: "PersistenceRequired", Label: "PersistenceRequired - Prevent despawning naturally", Description: "デスポーン防止 (PersistenceRequired)", ApplicableTo: nil, Type: "boolean"},
	{Key: "CustomName", Label: "CustomName - Custom name of the entity", Description: "カスタム名 (CustomName)", ApplicableTo: nil, Type: "string"},
	{Key: "CustomNameVisible", Label: "CustomNameVisible - Always show custom name", Description: "カスタム名を常時表示 (CustomNameVisible)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Fire", Label: "Fire - Number of ticks the entity is on fire", Description: "炎上時間（ticks）", ApplicableTo: nil, Type: "int"},
	{Key: "PortalCooldown", Label: "PortalCooldown - Ticks before entity can use portal again", Description: "ポータル移動クールダウン（ticks）", ApplicableTo: nil, Type: "int"},
	{Key: "Tags", Label: "Tags - List of custom tags on this entity", Description: "カスタムタグ（例: [\"tag1\", \"tag2\"]）", ApplicableTo: nil, Type: "raw"},
	{Key: "Health", Label: "Health - Current health value of the entity", Description: "体力値 (Health)", ApplicableTo: nil, Type: "int"},
	{Key: "AbsorptionAmount", Label: "AbsorptionAmount - Extra absorption health", Description: "吸収体力値 (AbsorptionAmount)", ApplicableTo: nil, Type: "int"},
	{Key: "LeftHanded", Label: "LeftHanded - Whether the entity is left-handed", Description: "左利きにする (LeftHanded)", ApplicableTo: nil, Type: "boolean"},
	{Key: "HandItems", Label: "HandItems - List of items held in hands", Description: "手持ちアイテム（例: [{id:\"minecraft:diamond_sword\",Count:1b},{}]）", ApplicableTo: nil, Type: "raw"},
	{Key: "ArmorItems", Label: "ArmorItems - List of equipped armor items", Description: "防具装備（例: [{},{},{},{id:\"minecraft:iron_helmet\",Count:1b}]）", ApplicableTo: nil, Type: "raw"},
	{Key: "Age", Label: "Age - Age of the mob (negative for baby, positive for breeding cooldown)", Description: "年齢/成長度 (Age)", ApplicableTo: breedableMobs, Type: "int"},
	{Key: "ForcedAge", Label: "ForcedAge - Ticks to force baby size until automatic growth resume", Description: "強制年齢維持時間 (ForcedAge)", ApplicableTo: breedableMobs, Type: "int"},
	{Key: "InLove", Label: "InLove - Ticks the mob remains in breeding mode", Description: "繁殖モード残り時間 (InLove)", ApplicableTo: breedableMobs, Type: "int"},
	{Key: "LoveCause", Label: "LoveCause - UUID of the player who caused the breeding mode", Description: "繁殖原因プレイヤーUUID (LoveCause)", ApplicableTo: breedableMobs, Type: "string"},
	{Key: "IsBaby", Label: "IsBaby - Spawn as a baby mob", Description: "子供化 (IsBaby)", ApplicableTo: []string{"zombie", "zombie_villager", "husk", "drowned", "piglin", "zombified_piglin", "hoglin", "zoglin"}, Type: "boolean"},
	{Key: "powered", Label: "powered - Charge the creeper", Description: "帯電状態 (powered)", ApplicableTo: []string{"creeper"}, Type: "boolean"},
	{Key: "ExplosionRadius", Label: "ExplosionRadius - Explosion radius of creeper", Description: "爆発半径 (ExplosionRadius)", ApplicableTo: []string{"creeper"}, Type: "int"},
	{Key: "Fuse", Label: "Fuse - Explosion delay ticks", Description: "爆発までの時間 (Fuse)", ApplicableTo: []string{"creeper", "tnt"}, Type: "int"},
	{Key: "ignited", Label: "ignited - Auto ignite the creeper", Description: "即時点火 (ignited)", ApplicableTo: []string{"creeper"}, Type: "boolean"},
	{Key: "Size", Label: "Size - Set size of slime/magmacube", Description: "サイズ変更 (Size)", ApplicableTo: []string{"slime", "magma_cube"}, Type: "int"},
	{Key: "ShowArms", Label: "ShowArms - Display arms on armor stand", Description: "腕を表示する (ShowArms)", ApplicableTo: []string{"armor_stand"}, Type: "boolean"},
	{Key: "Invisible", Label: "Invisible - Make armor stand invisible", Description: "透明化 (Invisible)", ApplicableTo: []string{"armor_stand"}, Type: "boolean"},
	{Key: "NoBasePlate", Label: "NoBasePlate - Remove bottom plate of armor stand", Description: "底板非表示 (NoBasePlate)", ApplicableTo: []string{"armor_stand"}, Type: "boolean"},
	{Key: "Small", Label: "Small - Make armor stand small", Description: "サイズを小さくする (Small)", ApplicableTo: []string{"armor_stand"}, Type: "boolean"},
	{Key: "carriedBlockState", Label: "carriedBlockState - Block carried by enderman", Description: "持ち運んでいるブロック", ApplicableTo: []string{"enderman"}, Type: "raw"},
	{Key: "HasNectar", Label: "HasNectar - Whether the bee has nectar", Description: "花蜜を持っているか (HasNectar)", ApplicableTo: []string{"bee"}, Type: "boolean"},
	{Key: "HasStung", Label: "HasStung - Whether the bee has stung a player (will die soon)", Description: "プレイヤー刺傷済みか (HasStung)", ApplicableTo: []string{"bee"}, Type: "boolean"},
	{Key: "TicksSincePollination", Label: "TicksSincePollination - Ticks since the bee last pollinated a flower", Description: "最終受粉からの時間 (TicksSincePollination)", ApplicableTo: []string{"bee"}, Type: "int"},
	{Key: "HasLeftHorn", Label: "HasLeftHorn - Whether the goat has its left horn", Description: "左の角があるか (HasLeftHorn)", ApplicableTo: []string{"goat"}, Type: "boolean"},
	{Key: "HasRightHorn", Label: "HasRightHorn - Whether the goat has its right horn", Description: "右の角があるか (HasRightHorn)", ApplicableTo: []string{"goat"}, Type: "boolean"},
	{Key: "IsScreamingGoat", Label: "IsScreamingGoat - Whether the goat is a screaming goat variant", Description: "叫ぶヤギか (IsScreamingGoat)", ApplicableTo: []string{"goat"}, Type: "boolean"},
	{Key: "Sleeping", Label: "Sleeping - Whether the fox is sleeping", Description: "睡眠中か (Sleeping)", ApplicableTo: []string{"fox"}, Type: "boolean"},
	{Key: "Sitting", Label: "Sitting - Whether the fox is sitting", Description: "お座り中か (Sitting)", ApplicableTo: []string{"fox"}, Type: "boolean"},
	{Key: "MainGene", Label: "MainGene - Dominant genetic trait of the panda", Description: "主遺伝子 (MainGene)", ApplicableTo: []string{"panda"}, Type: "string"},
	{Key: "HiddenGene", Label: "HiddenGene - Recessive genetic trait of the panda", Description: "隠れ遺伝子 (HiddenGene)", ApplicableTo: []string{"panda"}, Type: "string"},
	{Key: "Angry", Label: "Angry - Whether the wolf is angry", Description: "怒り状態か (Angry)", ApplicableTo: []string{"wolf"}, Type: "boolean"},
	{Key: "Tamed", Label: "Tamed - Whether the animal is tamed", Description: "手懐け状態か (Tamed)", ApplicableTo: []string{"wolf", "cat"}, Type: "boolean"},
	{Key: "CollarColor", Label: "CollarColor - Dye color of wolf/cat collar (0-15)", Description: "首輪の色（0-15）(CollarColor)", ApplicableTo: []string{"wolf", "cat"}, Type: "int"},
	{Key: "variant", Label: "variant - Variant type of frog, cat, or llama", Description: "種類/バリアント（例: minecraft:temperate）", ApplicableTo: []string{"frog", "llama", "trader_llama", "cat"}, Type: "string"},
	{Key: "Tame", Label: "Tame - Whether the horse is tame", Description: "手懐け状態か (Tame)", ApplicableTo: []string{"horse", "donkey", "mule", "llama", "trader_llama", "camel", "skeleton_horse", "zombie_horse"}, Type: "boolean"},
	{Key: "Temper", Label: "Temper - Temper value (0-100), higher makes taming easier", Description: "気性値 (Temper)", ApplicableTo: []string{"horse", "donkey", "mule", "llama", "trader_llama", "camel", "skeleton_horse", "zombie_horse"}, Type: "int"},
	{Key: "SaddleItem", Label: "SaddleItem - Item component for the horse's saddle", Description: "鞍アイテム (SaddleItem)", ApplicableTo: []string{"horse", "donkey", "mule", "llama", "trader_llama", "camel", "skeleton_horse", "zombie_horse"}, Type: "raw"},
	{Key: "ChestedHorse", Label: "ChestedHorse - Whether the horse has chests equipped", Description: "チェスト装着済みか (ChestedHorse)", ApplicableTo: []string{"donkey", "mule", "llama", "trader_llama"}, Type: "boolean"},
	{Key: "State", Label: "State - Current behavioral state of the sniffer", Description: "行動状態 (State)", ApplicableTo: []string{"sniffer"}, Type: "string"},
	{Key: "PlayerCreated", Label: "PlayerCreated - Whether the iron golem was created by player", Description: "プレイヤー製ゴーレムか (PlayerCreated)", ApplicableTo: []string{"iron_golem"}, Type: "boolean"},
	{Key: "Variant", Label: "Variant - Color variant of axolotl (0: pink, 1: brown, 2: gold, 3: cyan, 4: blue)", Description: "色バリアント（0-4）(Variant)", ApplicableTo: []string{"axolotl"}, Type: "int"},
	{Key: "CanDuplicate", Label: "CanDuplicate - Whether the allay can duplicate using amethyst", Description: "複製可能か (CanDuplicate)", ApplicableTo: []string{"allay"}, Type: "boolean"},
	{Key: "DuplicationCooldown", Label: "DuplicationCooldown - Cooldown ticks before allay can duplicate again", Description: "複製クールダウン時間 (DuplicationCooldown)", ApplicableTo: []string{"allay"}, Type: "int"},
	{Key: "RabbitType", Label: "RabbitType - Type of rabbit (0-5, 99 for killer rabbit)", Description: "ウサギの種類 (RabbitType)", ApplicableTo: []string{"rabbit"}, Type: "int"},
	{Key: "VillagerData", Label: "VillagerData - Professional stats of villager", Description: "村人のデータ（例: {profession:\"minecraft:farmer\",level:1}）", ApplicableTo: []string{"villager"}, Type: "raw"},
	{Key: "DataVersion", Label: "DataVersion - [Int]: Version of the chunk data.", Description: "[Int]: Version of the chunk data.", ApplicableTo: nil, Type: "int"},
	{Key: "AirAirAirAir", Label: "AirAirAirAir - [Short]: How much air the entity has, in game ticks. Decreases when unable to breathe (except suffocating in a block)...", Description: "[Short]: How much air the entity has, in game ticks. Decreases when unable to breathe (except suffocating in a block)...", ApplicableTo: nil, Type: "int"},
	{Key: "data", Label: "data - [NBT Compound / JSON Object]: Optional arbitrary NBT data. Is removed if empty. Represents the minecraft:custom_data ...", Description: "[NBT Compound / JSON Object]: Optional arbitrary NBT data. Is removed if empty. Represents the minecraft:custom_data ...", ApplicableTo: nil, Type: "raw"},
	{Key: "fall_distance", Label: "fall_distance - [Double]: Distance the entity has fallen. Larger values cause more damage when the entity lands.", Description: "[Double]: Distance the entity has fallen. Larger values cause more damage when the entity lands.", ApplicableTo: nil, Type: "int"},
	{Key: "HasVisualFire", Label: "HasVisualFire - [Boolean]: 1 or 0 (true/false) - if true, the entity visually appears on fire, even if it is not actually on fire.", Description: "[Boolean]: 1 or 0 (true/false) - if true, the entity visually appears on fire, even if it is not actually on fire.", ApplicableTo: nil, Type: "raw"},
	{Key: "OnGround", Label: "OnGround - [Boolean]: 1 or 0 (true/false) - if true, the entity is touching the ground.", Description: "[Boolean]: 1 or 0 (true/false) - if true, the entity is touching the ground.", ApplicableTo: nil, Type: "raw"},
	{Key: "TicksFrozen", Label: "TicksFrozen - [Int]: Optional. How many game ticks the entity has been freezing. Although this tag is defined for all entities, it ...", Description: "[Int]: Optional. How many game ticks the entity has been freezing. Although this tag is defined for all entities, it ...", ApplicableTo: nil, Type: "int"},
	{Key: "UUIDInts", Label: "UUIDInts - [Int Array]: This entity's Universally Unique IDentifier. The 128-bit UUID is stored as four 32-bit integers ([Int]),...", Description: "[Int Array]: This entity's Universally Unique IDentifier. The 128-bit UUID is stored as four 32-bit integers ([Int]),...", ApplicableTo: nil, Type: "int"},
	{Key: "ambient", Label: "ambient - [Boolean]: 1 or 0 (true/false) - if true, this effect is provided by a Beacon and therefore should be less intrusive ...", Description: "[Boolean]: 1 or 0 (true/false) - if true, this effect is provided by a Beacon and therefore should be less intrusive ...", ApplicableTo: nil, Type: "raw"},
	{Key: "amplifier", Label: "amplifier - [Byte]: The status effect level. 0 is level 1.", Description: "[Byte]: The status effect level. 0 is level 1.", ApplicableTo: nil, Type: "int"},
	{Key: "duration", Label: "duration - [Int]: The number of game ticks before the effect wears off. -1 means infinite duration.", Description: "[Int]: The number of game ticks before the effect wears off. -1 means infinite duration.", ApplicableTo: nil, Type: "int"},
	{Key: "hidden_effect", Label: "hidden_effect - [NBT Compound / JSON Object]: Lower amplifier effect of the same type, this replaces the above effect when it expires...", Description: "[NBT Compound / JSON Object]: Lower amplifier effect of the same type, this replaces the above effect when it expires...", ApplicableTo: nil, Type: "raw"},
	{Key: "show_icon", Label: "show_icon - [Boolean]: 1 or 0 (true/false) - if true, effect icon is shown; if false, no icon is shown.", Description: "[Boolean]: 1 or 0 (true/false) - if true, effect icon is shown; if false, no icon is shown.", ApplicableTo: nil, Type: "boolean"},
	{Key: "show_particles", Label: "show_particles - [Boolean]: 1 or 0 (true/false) - if true, particles are shown (affected by ambient); if false, no particles are shown.", Description: "[Boolean]: 1 or 0 (true/false) - if true, particles are shown (affected by ambient); if false, no particles are shown.", ApplicableTo: nil, Type: "boolean"},
	{Key: "base", Label: "base - [Double]: The base value of this attribute.", Description: "[Double]: The base value of this attribute.", ApplicableTo: nil, Type: "int"},
	{Key: "amount", Label: "amount - [Double]: The amount by which this modifier modifies the base value in calculations.", Description: "[Double]: The amount by which this modifier modifies the base value in calculations.", ApplicableTo: nil, Type: "int"},
	{Key: "CanPickUpLoot", Label: "CanPickUpLoot - [Boolean]: 1 or 0 (true/false) - if true, the mob can pick up loot (wear armor it picks up, use weapons it picks up).", Description: "[Boolean]: 1 or 0 (true/false) - if true, the mob can pick up loot (wear armor it picks up, use weapons it picks up).", ApplicableTo: nil, Type: "boolean"},
	{Key: "DeathTime", Label: "DeathTime - [Short]: Number of ticks the mob has been dead for. Controls death animations. 0 when alive.", Description: "[Short]: Number of ticks the mob has been dead for. Controls death animations. 0 when alive.", ApplicableTo: nil, Type: "int"},
	{Key: "head", Label: "head - [Float] : Chance value for the head item to drop.", Description: "[Float] : Chance value for the head item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "chest", Label: "chest - [Float] : Chance value for the chest item to drop.", Description: "[Float] : Chance value for the chest item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "legs", Label: "legs - [Float] : Chance value for the legs item to drop.", Description: "[Float] : Chance value for the legs item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "feet", Label: "feet - [Float] : Chance value for the feet item to drop.", Description: "[Float] : Chance value for the feet item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "mainhand", Label: "mainhand - [Float] : Chance value for the mainhand item to drop.", Description: "[Float] : Chance value for the mainhand item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "offhand", Label: "offhand - [Float] : Chance value for the offhand item to drop.", Description: "[Float] : Chance value for the offhand item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "body", Label: "body - [Float] : Chance value for the body item to drop.", Description: "[Float] : Chance value for the body item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "saddle", Label: "saddle - [Float] : Chance value for the saddle item to drop.", Description: "[Float] : Chance value for the saddle item to drop.", ApplicableTo: nil, Type: "int"},
	{Key: "FallFlying", Label: "FallFlying - [Byte]: Setting to 1 for non-player entities causes the entity to glide as long as they are wearing elytra in the che...", Description: "[Byte]: Setting to 1 for non-player entities causes the entity to glide as long as they are wearing elytra in the che...", ApplicableTo: nil, Type: "boolean"},
	{Key: "home_pos", Label: "home_pos - [Int Array]: The mob's \"home\" position. Mobs will limit their pathfinding to stay within the indicated area. Some mob...", Description: "[Int Array]: The mob's \"home\" position. Mobs will limit their pathfinding to stay within the indicated area. Some mob...", ApplicableTo: nil, Type: "raw"},
	{Key: "home_radius", Label: "home_radius - [Int]: Max radius of the data home_pos.", Description: "[Int]: Max radius of the data home_pos.", ApplicableTo: nil, Type: "int"},
	{Key: "HurtByTimestamp", Label: "HurtByTimestamp - [Int]: The last time the mob was damaged, measured in the number of ticks since the mob's creation. Updates to a new ...", Description: "[Int]: The last time the mob was damaged, measured in the number of ticks since the mob's creation. Updates to a new ...", ApplicableTo: nil, Type: "int"},
	{Key: "HurtTime", Label: "HurtTime - [Short]: Number of ticks the mob turns red for after being hit. 0 when not recently hit.", Description: "[Short]: Number of ticks the mob turns red for after being hit. 0 when not recently hit.", ApplicableTo: nil, Type: "int"},
	{Key: "last_hurt_by_mob", Label: "last_hurt_by_mob - [Int Array]: The UUID of the last mob that attacked this mob. Clears when the attacking mob dies or despawns.", Description: "[Int Array]: The UUID of the last mob that attacked this mob. Clears when the attacking mob dies or despawns.", ApplicableTo: nil, Type: "raw"},
	{Key: "last_hurt_by_player", Label: "last_hurt_by_player - [Int Array]: The UUID of the last player that attacked this mob.", Description: "[Int Array]: The UUID of the last player that attacked this mob.", ApplicableTo: nil, Type: "raw"},
	{Key: "color", Label: "color - [Int]: The waypoint's color stored as 32-bit signed integer using two's complement, assuming the color is fully opaque.", Description: "[Int]: The waypoint's color stored as 32-bit signed integer using two's complement, assuming the color is fully opaque.", ApplicableTo: nil, Type: "int"},
	{Key: "style", Label: "style - [String]: The waypoint's style name from waypoint_style directory in a resource pack.", Description: "[String]: The waypoint's style name from waypoint_style directory in a resource pack.", ApplicableTo: nil, Type: "string"},
	{Key: "sleeping_pos", Label: "sleeping_pos - [Int Array]: The coordinate of where the entity is sleeping, absent if not sleeping.", Description: "[Int Array]: The coordinate of where the entity is sleeping, absent if not sleeping.", ApplicableTo: nil, Type: "raw"},
	{Key: "Team", Label: "Team - [String]: This tag is actually not part of the NBT data of a mob, but instead used when spawning it, so it cannot be ...", Description: "[String]: This tag is actually not part of the NBT data of a mob, but instead used when spawning it, so it cannot be ...", ApplicableTo: nil, Type: "string"},
	{Key: "distance", Label: "distance - [Int]: Nonnegative integer.", Description: "[Int]: Nonnegative integer.", ApplicableTo: nil, Type: "int"},
	{Key: "game_event", Label: "game_event - [String]: A resource location of the game event.", Description: "[String]: A resource location of the game event.", ApplicableTo: nil, Type: "string"},
	{Key: "projectile_owner", Label: "projectile_owner - [Int Array]: Optional. The projectile owner's UUID. The 128-bit UUID is stored as four 32-bit integers, ordered from ...", Description: "[Int Array]: Optional. The projectile owner's UUID. The 128-bit UUID is stored as four 32-bit integers, ordered from ...", ApplicableTo: nil, Type: "raw"},
	{Key: "source", Label: "source - [Int Array]: Optional. The source entity's UUID. The 128-bit UUID is stored as four 32-bit integers, ordered from mos...", Description: "[Int Array]: Optional. The source entity's UUID. The 128-bit UUID is stored as four 32-bit integers, ordered from mos...", ApplicableTo: nil, Type: "raw"},
	{Key: "event_delay", Label: "event_delay - [Int]: Nonnegative integer.", Description: "[Int]: Nonnegative integer.", ApplicableTo: nil, Type: "int"},
	{Key: "event_distance", Label: "event_distance - [Int]: Nonnegative integer.", Description: "[Int]: Nonnegative integer.", ApplicableTo: nil, Type: "int"},
	{Key: "range", Label: "range - [Int]: Nonnegative integer.", Description: "[Int]: Nonnegative integer.", ApplicableTo: nil, Type: "int"},
	{Key: "type", Label: "type - [String]: A resource location of the position source type.", Description: "[String]: A resource location of the position source type.", ApplicableTo: nil, Type: "string"},
	{Key: "source_entity", Label: "source_entity - [Int Array]: The entity's UUID. The 128-bit UUID is stored as four 32-bit integers, ordered from most to least signif...", Description: "[Int Array]: The entity's UUID. The 128-bit UUID is stored as four 32-bit integers, ordered from most to least signif...", ApplicableTo: nil, Type: "raw"},
	{Key: "y_offset", Label: "y_offset - [Float]:", Description: "[Float]:", ApplicableTo: nil, Type: "int"},
	{Key: "scute_time", Label: "scute_time - [Int]: The number of ticks until the armadillo drops a scute. A scute is dropped at 0 and this timer gets reset to a ...", Description: "[Int]: The number of ticks until the armadillo drops a scute. A scute is dropped at 0 and this timer gets reset to a ...", ApplicableTo: nil, Type: "int"},
	{Key: "DisabledSlots", Label: "DisabledSlots - [Int]: Bit field allowing disable place/replace/remove of armor elements. For example, the value 16191 or4144959 disa...", Description: "[Int]: Bit field allowing disable place/replace/remove of armor elements. For example, the value 16191 or4144959 disa...", ApplicableTo: nil, Type: "int"},
	{Key: "Marker", Label: "Marker - [Byte]: 1 or 0 (true/false) - if true, ArmorStand's size is set to 0, has a tiny hitbox, and disables interactions wi...", Description: "[Byte]: 1 or 0 (true/false) - if true, ArmorStand's size is set to 0, has a tiny hitbox, and disables interactions wi...", ApplicableTo: nil, Type: "boolean"},
	{Key: "FromBucket", Label: "FromBucket - [Byte]: 1 or 0 (true/false) – if true, indicates the axolotl has been released from a bucket.", Description: "[Byte]: 1 or 0 (true/false) – if true, indicates the axolotl has been released from a bucket.", ApplicableTo: []string{"axolotl"}, Type: "boolean"},
	{Key: "BatFlags", Label: "BatFlags - [Byte]: 1 or 0 (true/false) - true if the bat is hanging upside-down from a block, false if the bat is flying.", Description: "[Byte]: 1 or 0 (true/false) - true if the bat is hanging upside-down from a block, false if the bat is flying.", ApplicableTo: nil, Type: "boolean"},
	{Key: "CannotEnterHiveTicks", Label: "CannotEnterHiveTicks - [Int]: Time left in ticks until the bee can enter a beehive. Used when the bee is angered and released from the hive ...", Description: "[Int]: Time left in ticks until the bee can enter a beehive. Used when the bee is angered and released from the hive ...", ApplicableTo: []string{"bee"}, Type: "int"},
	{Key: "CropsGrownSincePollination", Label: "CropsGrownSincePollination - [Int]: How many crops the bee has grown since its last pollination. Used to limit number of crops it can grow.", Description: "[Int]: How many crops the bee has grown since its last pollination. Used to limit number of crops it can grow.", ApplicableTo: []string{"bee"}, Type: "int"},
	{Key: "flower_pos", Label: "flower_pos - [Int Array]: Block location, as 3 integers, of the flower that the bee is circling.", Description: "[Int Array]: Block location, as 3 integers, of the flower that the bee is circling.", ApplicableTo: []string{"bee"}, Type: "raw"},
	{Key: "hive_pos", Label: "hive_pos - [Int Array]: Block location, as 3 integers, of the bee's hive.", Description: "[Int Array]: Block location, as 3 integers, of the bee's hive.", ApplicableTo: []string{"bee"}, Type: "raw"},
	{Key: "Bred", Label: "Bred - [Byte]: 1 or 0 (true/false) – Unknown. Remains 0 after breeding. If true, causes it to stay near other horses with th...", Description: "[Byte]: 1 or 0 (true/false) – Unknown. Remains 0 after breeding. If true, causes it to stay near other horses with th...", ApplicableTo: nil, Type: "boolean"},
	{Key: "EatingHaystack", Label: "EatingHaystack - [Byte]: 1 or 0 (true/false) – true if the mob is eating grass.", Description: "[Byte]: 1 or 0 (true/false) – true if the mob is eating grass.", ApplicableTo: nil, Type: "boolean"},
	{Key: "LastPoseTick", Label: "LastPoseTick - [Long]: The tick when the camel started changing its pose.", Description: "[Long]: The tick when the camel started changing its pose.", ApplicableTo: []string{"camel"}, Type: "int"},
	{Key: "EggLayTime", Label: "EggLayTime - [Int]: Number of ticks until the chicken lays its egg. Laying occurs at 0 and this timer gets reset to a new random v...", Description: "[Int]: Number of ticks until the chicken lays its egg. Laying occurs at 0 and this timer gets reset to a new random v...", ApplicableTo: []string{"chicken"}, Type: "int"},
	{Key: "IsChickenJockey", Label: "IsChickenJockey - [Boolean]: 1 or 0 (true/false) - Whether or not the chicken is a jockey for a baby zombie. If true, the chicken can n...", Description: "[Boolean]: 1 or 0 (true/false) - Whether or not the chicken is a jockey for a baby zombie. If true, the chicken can n...", ApplicableTo: []string{"zombie", "chicken"}, Type: "raw"},
	{Key: "weather_state", Label: "weather_state - [String]: unaffected, exposed, weathered, or oxidized - the oxidation level of the copper golem", Description: "[String]: unaffected, exposed, weathered, or oxidized - the oxidation level of the copper golem", ApplicableTo: nil, Type: "string"},
	{Key: "next_weather_age", Label: "next_weather_age - [Long]: The number of ticks gametime must reach for the copper golem's oxidation level to change. Setting the value t...", Description: "[Long]: The number of ticks gametime must reach for the copper golem's oxidation level to change. Setting the value t...", ApplicableTo: nil, Type: "int"},
	{Key: "Moistness", Label: "Moistness - [Int]: How moist this dolphin is. Set to 2400 when in water or rain. Decreases by 1 every tick otherwise. The dolphin...", Description: "[Int]: How moist this dolphin is. Set to 2400 when in water or rain. Decreases by 1 every tick otherwise. The dolphin...", ApplicableTo: nil, Type: "int"},
	{Key: "GotFish", Label: "GotFish - [Byte]: 1 or 0 (true/false) - if true, this dolphin got fish from a player.", Description: "[Byte]: 1 or 0 (true/false) - if true, this dolphin got fish from a player.", ApplicableTo: nil, Type: "boolean"},
	{Key: "CanBreakDoors", Label: "CanBreakDoors - [Byte]: 1 or 0 (true/false) - true if the zombie can break doors (default value is 0).", Description: "[Byte]: 1 or 0 (true/false) - true if the zombie can break doors (default value is 0).", ApplicableTo: []string{"zombie"}, Type: "boolean"},
	{Key: "DrownedConversionTime", Label: "DrownedConversionTime - [Int]: The number of ticks until this zombie converts to a drowned, or husk to zombie. (default value is -1, when no ...", Description: "[Int]: The number of ticks until this zombie converts to a drowned, or husk to zombie. (default value is -1, when no ...", ApplicableTo: []string{"zombie"}, Type: "int"},
	{Key: "InWaterTime", Label: "InWaterTime - [Int]: The number of ticks this zombie or husk has been under water, used to start the drowning conversion. (default ...", Description: "[Int]: The number of ticks this zombie or husk has been under water, used to start the drowning conversion. (default ...", ApplicableTo: []string{"zombie"}, Type: "int"},
	{Key: "DragonDeathTime", Label: "DragonDeathTime - [Int]: Number of ticks the dragon has been dead for. At 150, the dragon begins spawning experience orbs every 5 ticks...", Description: "[Int]: Number of ticks the dragon has been dead for. At 150, the dragon begins spawning experience orbs every 5 ticks...", ApplicableTo: nil, Type: "int"},
	{Key: "DragonPhase", Label: "DragonPhase - [Int]: A number indicating the dragon's current state. 0 means circling. 1 means strafing (preparing to shoot a fireb...", Description: "[Int]: A number indicating the dragon's current state. 0 means circling. 1 means strafing (preparing to shoot a fireb...", ApplicableTo: nil, Type: "int"},
	{Key: "Lifetime", Label: "Lifetime - [Int]: How long the endermite has existed in ticks. Disappears when this reaches around 2400.", Description: "[Int]: How long the endermite has existed in ticks. Disappears when this reaches around 2400.", ApplicableTo: nil, Type: "int"},
	{Key: "SpellTicks", Label: "SpellTicks - [Int]: Number of ticks until a spell can be cast. Set to a positive value when a spell is cast, and decreases by 1 pe...", Description: "[Int]: Number of ticks until a spell can be cast. Set to a positive value when a spell is cast, and decreases by 1 pe...", ApplicableTo: nil, Type: "int"},
	{Key: "Crouching", Label: "Crouching - [Byte]: 1 or 0 (true/false) – Whether the fox is crouching.", Description: "[Byte]: 1 or 0 (true/false) – Whether the fox is crouching.", ApplicableTo: []string{"fox"}, Type: "boolean"},
	{Key: "ExplosionPower", Label: "ExplosionPower - [Byte]: The radius of the explosion created by the fireballs the ghast fires. Default value is 1.", Description: "[Byte]: The radius of the explosion created by the fireballs the ghast fires. Default value is 1.", ApplicableTo: nil, Type: "int"},
	{Key: "DarkTicksRemaining", Label: "DarkTicksRemaining - [Int]: Countdown of ticks remaining until the glow squid starts glowing. Not glowing while positive, glowing when cou...", Description: "[Int]: Countdown of ticks remaining until the glow squid starts glowing. Not glowing while positive, glowing when cou...", ApplicableTo: nil, Type: "int"},
	{Key: "still_timeout", Label: "still_timeout - [Int]: Prevents the Happy Ghast from moving when greater than 0. Set to 10 when a player is less than 2 blocks above ...", Description: "[Int]: Prevents the Happy Ghast from moving when greater than 0. Set to 10 when a player is less than 2 blocks above ...", ApplicableTo: nil, Type: "int"},
	{Key: "CannotBeHunted", Label: "CannotBeHunted - [Boolean]: 1 or 0 (true/false) - if true, piglins do not attack the hoglin. Set to true for hoglins spawned as a part...", Description: "[Boolean]: 1 or 0 (true/false) - if true, piglins do not attack the hoglin. Set to true for hoglins spawned as a part...", ApplicableTo: []string{"hoglin"}, Type: "raw"},
	{Key: "IsImmuneToZombification", Label: "IsImmuneToZombification - [Boolean]: 1 or 0 (true/false) – if true, the hoglin does not transform to a zoglin when in the Overworld and TimeInO...", Description: "[Boolean]: 1 or 0 (true/false) – if true, the hoglin does not transform to a zoglin when in the Overworld and TimeInO...", ApplicableTo: []string{"hoglin", "zoglin"}, Type: "raw"},
	{Key: "TimeInOverworld", Label: "TimeInOverworld - [Int]: The number of ticks that the hoglin has existed in the Overworld; the hoglin converts to a zoglin when this is...", Description: "[Int]: The number of ticks that the hoglin has existed in the Overworld; the hoglin converts to a zoglin when this is...", ApplicableTo: []string{"hoglin", "zoglin"}, Type: "int"},
	{Key: "DespawnDelay", Label: "DespawnDelay - [Int]: A timer for trader llamas to despawn, present only in trader_llama. The trader llama despawns when this value ...", Description: "[Int]: A timer for trader llamas to despawn, present only in trader_llama. The trader llama despawns when this value ...", ApplicableTo: []string{"llama", "trader_llama"}, Type: "int"},
	{Key: "exceptfor", Label: "exceptfor - the tags: CanPickUpLoot, DeathLootTable, DeathLootTableSeed, drop_chances, home_pos, home_radius, leash, LeftHanded, ...", Description: "the tags: CanPickUpLoot, DeathLootTable, DeathLootTableSeed, drop_chances, home_pos, home_radius, leash, LeftHanded, ...", ApplicableTo: nil, Type: "raw"},
	{Key: "immovable", Label: "immovable - [Boolean] - Optional boolean specifying that the mannequin cannot be moved (defaults to false).", Description: "[Boolean] - Optional boolean specifying that the mannequin cannot be moved (defaults to false).", ApplicableTo: nil, Type: "raw"},
	{Key: "Trusting", Label: "Trusting - [Byte]: 1 or 0 (true/false) - true if the ocelot trusts players.", Description: "[Byte]: 1 or 0 (true/false) - true if the ocelot trusts players.", ApplicableTo: nil, Type: "boolean"},
	{Key: "size", Label: "size - [Int]: The size of the phantom. Ranges from 0 to 64, similar to slimes. Unlike slimes, phantoms always have a constan...", Description: "[Int]: The size of the phantom. Ranges from 0 to 64, similar to slimes. Unlike slimes, phantoms always have a constan...", ApplicableTo: []string{"phantom"}, Type: "int"},
	{Key: "anchor_pos", Label: "anchor_pos - [Int Array]: The phantom, when not actively attacking, attempts to circle around X,Y,Z. Appears to reset to a point a...", Description: "[Int Array]: The phantom, when not actively attacking, attempts to circle around X,Y,Z. Appears to reset to a point a...", ApplicableTo: []string{"phantom"}, Type: "raw"},
	{Key: "CannotHunt", Label: "CannotHunt - [Byte]: 1 or 0 (true/false) – if true, the piglin does not attack hoglins. Set to true for piglins spawned as a part ...", Description: "[Byte]: 1 or 0 (true/false) – if true, the piglin does not attack hoglins. Set to true for piglins spawned as a part ...", ApplicableTo: []string{"piglin"}, Type: "boolean"},
	{Key: "flying", Label: "flying - [Byte]: 1 or 0 (true/false) - true if the player is currently flying.", Description: "[Byte]: 1 or 0 (true/false) - true if the player is currently flying.", ApplicableTo: nil, Type: "boolean"},
	{Key: "flySpeed", Label: "flySpeed - [Float]: The flying speed, set to 0.05.", Description: "[Float]: The flying speed, set to 0.05.", ApplicableTo: nil, Type: "int"},
	{Key: "instabuild", Label: "instabuild - [Byte]: 1 or 0 (true/false) - If true, the player can place blocks without depleting them. This is true for Creative ...", Description: "[Byte]: 1 or 0 (true/false) - If true, the player can place blocks without depleting them. This is true for Creative ...", ApplicableTo: nil, Type: "boolean"},
	{Key: "invulnerable", Label: "invulnerable - [Byte]: 1 or 0 (true/false) - Behavior is not the same as the invulnerable tag on other entities. If true, the player...", Description: "[Byte]: 1 or 0 (true/false) - Behavior is not the same as the invulnerable tag on other entities. If true, the player...", ApplicableTo: nil, Type: "boolean"},
	{Key: "mayBuild", Label: "mayBuild - [Byte]: 1 or 0 (true/false) - If true, the player can place blocks. true when in Creative or Survival mode, and false...", Description: "[Byte]: 1 or 0 (true/false) - If true, the player can place blocks. true when in Creative or Survival mode, and false...", ApplicableTo: nil, Type: "boolean"},
	{Key: "mayfly", Label: "mayfly - [Byte]: 1 or 0 (true/false) - If true, the player can fly and doesn't take fall damage. true when in Creative and Spe...", Description: "[Byte]: 1 or 0 (true/false) - If true, the player can fly and doesn't take fall damage. true when in Creative and Spe...", ApplicableTo: nil, Type: "boolean"},
	{Key: "walkSpeed", Label: "walkSpeed - [Float]: The walking speed, set to 0.1.", Description: "[Float]: The walking speed, set to 0.1.", ApplicableTo: nil, Type: "int"},
	{Key: "current_explosion_impact_pos", Label: "current_explosion_impact_pos - [NBT List / JSON Array]: Position where the player was when the last explosion happened. Used for wind charge fall d...", Description: "[NBT List / JSON Array]: Position where the player was when the last explosion happened. Used for wind charge fall da...", ApplicableTo: nil, Type: "raw"},
	{Key: "foodExhaustionLevel", Label: "foodExhaustionLevel - [Float]: See Hunger § Mechanics.", Description: "[Float]: See Hunger § Mechanics.", ApplicableTo: nil, Type: "int"},
	{Key: "foodLevelhunger", Label: "foodLevelhunger - [Int]: The value of the hunger bar. Referred to as . See Hunger.", Description: "[Int]: The value of the hunger bar. Referred to as . See Hunger.", ApplicableTo: nil, Type: "int"},
	{Key: "foodSaturationLevelsaturation", Label: "foodSaturationLevelsaturation - [Float]: Referred to as . See Hunger § Mechanics.", Description: "[Float]: Referred to as . See Hunger § Mechanics.", ApplicableTo: nil, Type: "int"},
	{Key: "foodTickTimer", Label: "foodTickTimer - [Int]: See Hunger.", Description: "[Int]: See Hunger.", ApplicableTo: nil, Type: "int"},
	{Key: "ignore_fall_damage_from_current_explosion", Label: "ignore_fall_damage_from_current_explosion - [Boolean]: 1 or 0 (true/false) - true if the current explosion should apply a fall damage reduction. On...", Description: "[Boolean]: 1 or 0 (true/false) - true if the current explosion should apply a fall damage reduction. Only used by exp...", ApplicableTo: nil, Type: "raw"},
	{Key: "playerGameType", Label: "playerGameType - [Int]: The current game mode of the player. 0 means Survival, 1 means Creative, 2 means Adventure, and 3 means Specta...", Description: "[Int]: The current game mode of the player. 0 means Survival, 1 means Creative, 2 means Adventure, and 3 means Specta...", ApplicableTo: nil, Type: "int"},
	{Key: "previousPlayerGameType", Label: "previousPlayerGameType - [Int]: The previous game mode of the player.", Description: "[Int]: The previous game mode of the player.", ApplicableTo: nil, Type: "int"},
	{Key: "Attach", Label: "Attach - [Int Array]: The UUID of the entity the player is riding, stored as four ints.", Description: "[Int Array]: The UUID of the entity the player is riding, stored as four ints.", ApplicableTo: nil, Type: "int"},
	{Key: "seenCredits", Label: "seenCredits - [Byte]: 1 or 0 (true/false) - true if the player has entered the exit portal in the End at least once.", Description: "[Byte]: 1 or 0 (true/false) - true if the player has entered the exit portal in the End at least once.", ApplicableTo: nil, Type: "boolean"},
	{Key: "SelectedItemSlot", Label: "SelectedItemSlot - [Int]: The selected hotbar slot of the player. Values are 0-indexed, so the leftmost slot is 0 and the rightmost slot...", Description: "[Int]: The selected hotbar slot of the player. Values are 0-indexed, so the leftmost slot is 0 and the rightmost slot...", ApplicableTo: nil, Type: "int"},
	{Key: "SleepTimer", Label: "SleepTimer - [Short]: The number of game ticks the player had been in bed. 0 when the player is not sleeping. When in bed, increas...", Description: "[Short]: The number of game ticks the player had been in bed. 0 when the player is not sleeping. When in bed, increas...", ApplicableTo: nil, Type: "int"},
	{Key: "forced", Label: "forced - [Boolean]: true if this spawn was set through commands (default: false)", Description: "[Boolean]: true if this spawn was set through commands (default: false)", ApplicableTo: nil, Type: "raw"},
	{Key: "warning_level", Label: "warning_level - [Int]: A warning level between 0, and 3 (inclusive). The warden spawns at level 3.", Description: "[Int]: A warning level between 0, and 3 (inclusive). The warden spawns at level 3.", ApplicableTo: nil, Type: "int"},
	{Key: "cooldown_ticks", Label: "cooldown_ticks - [Int]: The number of game ticks before the warning_level can be increased again. Decreases by 1 every tick. It is set...", Description: "[Int]: The number of game ticks before the warning_level can be increased again. Decreases by 1 every tick. It is set...", ApplicableTo: nil, Type: "int"},
	{Key: "ticks_since_last_warning", Label: "ticks_since_last_warning - [Int]: The number of game ticks since the player was warned for warden spawning. Increases by 1 every tick. After 120...", Description: "[Int]: The number of game ticks since the player was warned for warden spawning. Increases by 1 every tick. After 120...", ApplicableTo: nil, Type: "int"},
	{Key: "XpLevel", Label: "XpLevel - [Int]: The level shown on the experience bar.", Description: "[Int]: The level shown on the experience bar.", ApplicableTo: nil, Type: "int"},
	{Key: "XpP", Label: "XpP - [Float]: The progress across the experience bar to the next level, stored as a percentage.[verify]", Description: "[Float]: The progress across the experience bar to the next level, stored as a percentage.[verify]", ApplicableTo: nil, Type: "int"},
	{Key: "XpSeed", Label: "XpSeed - [Int]: The seed used for the next enchantment in enchanting tables.", Description: "[Int]: The seed used for the next enchantment in enchanting tables.", ApplicableTo: nil, Type: "int"},
	{Key: "XpTotal", Label: "XpTotal - [Int]: The total amount of experience the player has collected over time; used for the score upon death.", Description: "[Int]: The total amount of experience the player has collected over time; used for the score upon death.", ApplicableTo: nil, Type: "int"},
	{Key: "MoreCarrotTicks", Label: "MoreCarrotTicks - [Int]: Set to 40 when a carrot crop is eaten, decreases by 0–2 every tick until it reaches 0. Rabbit can eat another ...", Description: "[Int]: Set to 40 when a carrot crop is eaten, decreases by 0–2 every tick until it reaches 0. Rabbit can eat another ...", ApplicableTo: []string{"rabbit"}, Type: "int"},
	{Key: "AttackTick", Label: "AttackTick - [Int]: Attack cooldown for this ravager.", Description: "[Int]: Attack cooldown for this ravager.", ApplicableTo: []string{"ravager"}, Type: "int"},
	{Key: "RoarTick", Label: "RoarTick - [Int]: Roar attack cooldown for this ravager.", Description: "[Int]: Roar attack cooldown for this ravager.", ApplicableTo: []string{"ravager"}, Type: "int"},
	{Key: "StunTick", Label: "StunTick - [Int]: Stun attack cooldown for this ravager.", Description: "[Int]: Stun attack cooldown for this ravager.", ApplicableTo: []string{"ravager"}, Type: "int"},
	{Key: "Sheared", Label: "Sheared - [Byte]: 1 or 0 (true/false) - true if the sheep has been shorn.", Description: "[Byte]: 1 or 0 (true/false) - true if the sheep has been shorn.", ApplicableTo: []string{"sheep"}, Type: "boolean"},
	{Key: "StrayConversionTime", Label: "StrayConversionTime - [Int]: The number of ticks until this skeleton converts to a stray (default value is -1, when no conversion is under ...", Description: "[Int]: The number of ticks until this skeleton converts to a stray (default value is -1, when no conversion is under ...", ApplicableTo: []string{"skeleton"}, Type: "int"},
	{Key: "SkeletonTrap", Label: "SkeletonTrap - [Byte]: 1 or 0 (true/false) - true if the horse is a trapped skeleton horse. Does not affect horse type.", Description: "[Byte]: 1 or 0 (true/false) - true if the horse is a trapped skeleton horse. Does not affect horse type.", ApplicableTo: []string{"skeleton", "horse"}, Type: "boolean"},
	{Key: "SkeletonTrapTime", Label: "SkeletonTrapTime - [Int]: Incremented each tick when SkeletonTrap is set to 1. The horse automatically despawns when it reaches 18000 (1...", Description: "[Int]: Incremented each tick when SkeletonTrap is set to 1. The horse automatically despawns when it reaches 18000 (1...", ApplicableTo: []string{"horse"}, Type: "int"},
	{Key: "Pumpkin", Label: "Pumpkin - [Byte] : 1 or 0 (true/false) - whether or not the Snow Golem has a pumpkin on its head.", Description: "[Byte] : 1 or 0 (true/false) - whether or not the Snow Golem has a pumpkin on its head.", ApplicableTo: []string{"snow_golem"}, Type: "boolean"},
	{Key: "ExplosionRadius", Label: "ExplosionRadius - [Byte]: The radius of the explosion created by this creeper. (default value is 3)", Description: "[Byte]: The radius of the explosion created by this creeper. (default value is 3)", ApplicableTo: []string{"creeper"}, Type: "int"},
	{Key: "ignited", Label: "ignited - [Byte]: 1 or 0 (true/false) - Whether the creeper has been ignited with flint and steel.", Description: "[Byte]: 1 or 0 (true/false) - Whether the creeper has been ignited with flint and steel.", ApplicableTo: []string{"creeper"}, Type: "boolean"},
	{Key: "Fuse", Label: "Fuse - [Short]: The number of ticks the creeper will flash before exploding after being ignited. (default value is 30)", Description: "[Short]: The number of ticks the creeper will flash before exploding after being ignited. (default value is 30)", ApplicableTo: []string{"creeper", "tnt"}, Type: "int"},
	{Key: "powered", Label: "powered - [Byte]: 1 or 0 (true/false) - Whether this creeper is charged by lightning. (default value is 0)", Description: "[Byte]: 1 or 0 (true/false) - Whether this creeper is charged by lightning. (default value is 0)", ApplicableTo: []string{"creeper"}, Type: "boolean"},
	{Key: "Size", Label: "Size - [Int]: The size of the slime/magmacube. Set to 1, 2, or 4 for natural spawning. Maximum size is 127. If the slime is ...", Description: "[Int]: The size of the slime/magmacube. Set to 1, 2, or 4 for natural spawning. Maximum size is 127. If the slime is ...", ApplicableTo: []string{"slime", "magma_cube"}, Type: "int"},
	{Key: "wasOnGround", Label: "wasOnGround - [Byte]: 1 or 0 (true/false) - Whether this slime was on ground on last tick.", Description: "[Byte]: 1 or 0 (true/false) - Whether this slime was on ground on last tick.", ApplicableTo: []string{"slime"}, Type: "boolean"},
	{Key: "Item", Label: "Item - [NBT Compound / JSON Object]: Represents the item tag. Represents the minecraft:item component.", Description: "[NBT Compound / JSON Object]: Represents the item tag. Represents the minecraft:item component.", ApplicableTo: []string{"item"}, Type: "raw"},
	{Key: "OwnerUUID", Label: "OwnerUUID - [Int Array]: The UUID of the player who threw this item. (default value is absent)", Description: "[Int Array]: The UUID of the player who threw this item. (default value is absent)", ApplicableTo: nil, Type: "int"},
	{Key: "DespawnTime", Label: "DespawnTime - [Int]: Counter of ticks since this item spawned. Range: -32768 to 32767. When it reaches 6000 (5 minutes), the ite...", Description: "[Int]: Counter of ticks since this item spawned. Range: -32768 to 32767. When it reaches 6000 (5 minutes), the ite...", ApplicableTo: nil, Type: "int"},
	{Key: "LootTable", Label: "LootTable - [String]: The resource location of the loot table that will be used when opening this chest.", Description: "[String]: The resource location of the loot table that will be used when opening this chest.", ApplicableTo: nil, Type: "string"},
	{Key: "LootTableSeed", Label: "LootTableSeed - [Long]: The seed for generating the loot table. Ignored if LootTable is absent.", Description: "[Long]: The seed for generating the loot table. Ignored if LootTable is absent.", ApplicableTo: nil, Type: "int"},
	{Key: "Ageable", Label: "Ageable - Age", Description: "Age", ApplicableTo: nil, Type: "raw"},
	{Key: "HasRightHorn", Label: "HasRightHorn - [Boolean]: 1 or 0 (true/false) - true if the goat has its right horn.", Description: "[Boolean]: 1 or 0 (true/false) - true if the goat has its right horn.", ApplicableTo: []string{"goat"}, Type: "raw"},
	{Key: "HasLeftHorn", Label: "HasLeftHorn - [Boolean]: 1 or 0 (true/false) - true if the goat has its left horn.", Description: "[Boolean]: 1 or 0 (true/false) - true if the goat has its left horn.", ApplicableTo: []string{"goat"}, Type: "raw"},
	{Key: "IsScreamingGoat", Label: "IsScreamingGoat - [Boolean]: 1 or 0 (true/false) - true if the goat is the screaming variant.", Description: "[Boolean]: 1 or 0 (true/false) - true if the goat is the screaming variant.", ApplicableTo: []string{"goat"}, Type: "raw"},
	{Key: "AngryAt", Label: "AngryAt - [Int Array]: The UUID of the entity that this neutral mob is targeted to attack.", Description: "[Int Array]: The UUID of the entity that this neutral mob is targeted to attack.", ApplicableTo: nil, Type: "int"},
	{Key: "AngerTime", Label: "AngerTime - [Int]: The number of ticks that this neutral mob will remain angry for.", Description: "[Int]: The number of ticks that this neutral mob will remain angry for.", ApplicableTo: []string{"bee", "wolf"}, Type: "int"},
	{Key: "Sitting", Label: "Sitting - [Byte]: 1 or 0 (true/false) – Whether the animal is sitting.", Description: "[Byte]: 1 or 0 (true/false) – Whether the animal is sitting.", ApplicableTo: []string{"fox"}, Type: "boolean"},
	{Key: "Sleeping", Label: "Sleeping - [Byte]: 1 or 0 (true/false) – Whether the fox is sleeping.", Description: "[Byte]: 1 or 0 (true/false) – Whether the fox is sleeping.", ApplicableTo: []string{"fox"}, Type: "boolean"},
	{Key: "CollarColor", Label: "CollarColor - [Byte]: The color of the collar on a tamed wolf/cat. Defaults to 14 (red). Represents the minecraft:wolf/coll...", Description: "[Byte]: The color of the collar on a tamed wolf/cat. Defaults to 14 (red). Represents the minecraft:wolf/coll...", ApplicableTo: []string{"wolf", "cat"}, Type: "int"},
	{Key: "variant", Label: "variant - [String]: The variant of the frog. temperate, warm, or cold. Represents the minecraft:frog/variant component.", Description: "[String]: The variant of the frog. temperate, warm, or cold. Represents the minecraft:frog/variant component.", ApplicableTo: []string{"frog", "llama", "trader_llama", "cat"}, Type: "string"},
	{Key: "varianttype", Label: "varianttype - [String]: The variant type of llama. creamy, white, brown, or gray. Represents the minecraft:llama/variant component.", Description: "[String]: The variant type of llama. creamy, white, brown, or gray. Represents the minecraft:llama/variant component.", ApplicableTo: []string{"llama"}, Type: "string"},
	{Key: "Tame", Label: "Tame - [Byte]: 1 or 0 (true/false) - true if the horse is tamed.", Description: "[Byte]: 1 or 0 (true/false) - true if the horse is tamed.", ApplicableTo: []string{"horse", "donkey", "mule", "llama", "trader_llama", "camel", "skeleton_horse", "zombie_horse"}, Type: "boolean"},
	{Key: "Temper", Label: "Temper - [Int]: The temper of the horse. Range 0 to 100. Higher temper makes the horse easier to tame.", Description: "[Int]: The temper of the horse. Range 0 to 100. Higher temper makes the horse easier to tame.", ApplicableTo: []string{"horse", "donkey", "mule", "llama", "trader_llama", "camel", "skeleton_horse", "zombie_horse"}, Type: "int"},
	{Key: "SaddleItem", Label: "SaddleItem - [NBT Compound / JSON Object]: Represents the saddle item on the horse. Represents the minecraft:saddle co...", Description: "[NBT Compound / JSON Object]: Represents the saddle item on the horse. Represents the minecraft:saddle co...", ApplicableTo: []string{"horse", "donkey", "mule", "llama", "trader_llama", "camel", "skeleton_horse", "zombie_horse"}, Type: "raw"},
	{Key: "ChestedHorse", Label: "ChestedHorse - [Byte]: 1 or 0 (true/false) - true if the horse has chest(s).", Description: "[Byte]: 1 or 0 (true/false) - true if the horse has chest(s).", ApplicableTo: []string{"donkey", "mule", "llama", "trader_llama"}, Type: "boolean"},
	{Key: "State", Label: "State - [String]: Current behavior state of the sniffer. idling, feeling_happy, scenting, sniffing, searching, digging, or rising...", Description: "[String]: Current behavior state of the sniffer. idling, feeling_happy, scenting, sniffing, searching, digging, or rising...", ApplicableTo: []string{"sniffer"}, Type: "string"},
	{Key: "PlayerCreated", Label: "PlayerCreated - [Byte]: 1 or 0 (true/false) - true if this iron golem was created by player.", Description: "[Byte]: 1 or 0 (true/false) - true if this iron golem was created by player.", ApplicableTo: []string{"iron_golem"}, Type: "boolean"},
	{Key: "Variant", Label: "Variant - [Int]: The color variant of this axolotl. (0: lucy (pink), 1: wild (brown), 2: gold (yellow), 3: cyan, 4: blue). Defau...", Description: "[Int]: The color variant of this axolotl. (0: lucy (pink), 1: wild (brown), 2: gold (yellow), 3: cyan, 4: blue). Defau...", ApplicableTo: []string{"axolotl"}, Type: "int"},
	{Key: "CanDuplicate", Label: "CanDuplicate - [Byte]: 1 or 0 (true/false) - true if this allay is allowed to duplicate.", Description: "[Byte]: 1 or 0 (true/false) - true if this allay is allowed to duplicate.", ApplicableTo: []string{"allay"}, Type: "boolean"},
	{Key: "DuplicationCooldown", Label: "DuplicationCooldown - [Long]: The amount of ticks left before this allay is allowed to duplicate again. Value is 0 if it is allowed ...", Description: "[Long]: The amount of ticks left before this allay is allowed to duplicate again. Value is 0 if it is allowed ...", ApplicableTo: []string{"allay"}, Type: "int"},
	{Key: "Angry", Label: "Angry - [Byte]: 1 or 0 (true/false) - true if this wolf is angry and its eyes are red.", Description: "[Byte]: 1 or 0 (true/false) - true if this wolf is angry and its eyes are red.", ApplicableTo: []string{"wolf"}, Type: "boolean"},
	{Key: "Tamed", Label: "Tamed - [Byte]: 1 or 0 (true/false) - true if this animal has been tamed.", Description: "[Byte]: 1 or 0 (true/false) - true if this animal has been tamed.", ApplicableTo: []string{"wolf", "cat"}, Type: "boolean"},
	{Key: "RabbitType", Label: "RabbitType - [Int]: The type of rabbit. (0: brown, 1: white, 2: black, 3: black & white, 4: gold, 5: salt & pepper, 99: Killer R...", Description: "[Int]: The type of rabbit. (0: brown, 1: white, 2: black, 3: black & white, 4: gold, 5: salt & pepper, 99: Killer R...", ApplicableTo: []string{"rabbit"}, Type: "int"},
	{Key: "VillagerData", Label: "VillagerData - [NBT Compound / JSON Object]: Data representing the villager's attributes (type, profession, level).", Description: "[NBT Compound / JSON Object]: Data representing the villager's attributes (type, profession, level).", ApplicableTo: []string{"villager"}, Type: "raw"},
	{Key: "Custom NBT", Label: "Custom NBT - Enter raw custom NBT string", Description: "カスタムNBT", ApplicableTo: nil, Type: "raw"},
}

func AssembleNbtString(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:1b", k))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func FilterBlocksByCategory(allBlocks []string, categoryKeywords []string) []string {
	if len(categoryKeywords) == 0 {
		return allBlocks
	}

	var filtered []string
	for _, block := range allBlocks {
		for _, keyword := range categoryKeywords {
			if strings.Contains(block, keyword) {
				filtered = append(filtered, block)
				break
			}
		}
	}
	return filtered
}
