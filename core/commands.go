package core

import (
	"fmt"
	"regexp"
	"sort"
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

type NbtChoice struct {
	Value string
	Label string
}

type NbtTagOption struct {
	Key          string      // NBTのキー名（例: "NoAI"）
	Label        string      // 画面に表示する日本語名
	Description  string      // タグの詳しい解説
	ApplicableTo []string    // 適用可能なエンティティIDのリスト（空配列はすべて）
	Type         string      // "boolean", "int", "string", "raw", "select"
	Choices      []NbtChoice // 選択肢リスト ("select" の場合)
}

var breedableMobs = []string{
	"axolotl", "bee", "camel", "cat", "chicken", "cow", "donkey", "fox", "frog", "goat",
	"hoglin", "horse", "llama", "mooshroom", "mule", "ocelot", "panda", "pig", "rabbit",
	"sheep", "sniffer", "strider", "trader_llama", "turtle", "villager", "wolf",
}

var ageLockedMobs = []string{
	"axolotl", "bee", "camel", "cat", "chicken", "cow", "donkey", "fox", "frog", "goat",
	"hoglin", "horse", "llama", "mooshroom", "mule", "ocelot", "panda", "pig", "rabbit",
	"sheep", "sniffer", "strider", "trader_llama", "turtle", "wolf",
}

var NbtMasterList = []NbtTagOption{
	// 1. 全エンティティ共通 NBT
	{Key: "NoAI", Label: "NoAI - Disable entity AI (does not move or attack)", Description: "AIを無効化 (NoAI)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Invulnerable", Label: "Invulnerable - Immune to all damage", Description: "無敵化 (Invulnerable)", ApplicableTo: nil, Type: "boolean"},
	{Key: "NoGravity", Label: "NoGravity - Disable gravity", Description: "重力無効 (NoGravity)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Silent", Label: "Silent - Silence all sounds from this entity", Description: "消音 (Silent)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Glowing", Label: "Glowing - Make the entity outline glow", Description: "発光 (Glowing)", ApplicableTo: nil, Type: "boolean"},
	{Key: "PersistenceRequired", Label: "PersistenceRequired - Prevent despawning naturally", Description: "デスポーン防止 (PersistenceRequired)", ApplicableTo: nil, Type: "boolean"},
	{Key: "CustomName", Label: "CustomName - Custom name of the entity", Description: "カスタム名 (CustomName)", ApplicableTo: nil, Type: "string"},
	{Key: "CustomNameVisible", Label: "CustomNameVisible - Always show custom name", Description: "カスタム名を常時表示 (CustomNameVisible)", ApplicableTo: nil, Type: "boolean"},
	{Key: "Fire", Label: "Fire - Number of ticks the entity is on fire", Description: "炎上時間（ticks）", ApplicableTo: nil, Type: "int"},
	{Key: "Tags", Label: "Tags - List of custom tags on this entity", Description: "カスタムタグ（例: [\"tag1\", \"tag2\"]）", ApplicableTo: nil, Type: "raw"},

	// 2. Mob / Living Entity 共通 NBT
	{Key: "Health", Label: "Health - Current health value of the entity", Description: "体力値 (Health)", ApplicableTo: nil, Type: "int"},
	{Key: "AbsorptionAmount", Label: "AbsorptionAmount - Extra absorption health", Description: "吸収体力値 (AbsorptionAmount)", ApplicableTo: nil, Type: "int"},
	{Key: "LeftHanded", Label: "LeftHanded - Whether the entity is left-handed", Description: "左利きにする (LeftHanded)", ApplicableTo: nil, Type: "boolean"},
	{Key: "HandItems", Label: "HandItems - List of items held in hands", Description: "手持ちアイテム（例: [{id:\"minecraft:diamond_sword\",Count:1b},{}]）", ApplicableTo: nil, Type: "raw"},
	{Key: "ArmorItems", Label: "ArmorItems - List of equipped armor items", Description: "防具装備（例: [{},{},{},{id:\"minecraft:iron_helmet\",Count:1b}]）", ApplicableTo: nil, Type: "raw"},

	// 3. 繁殖可能Mob共通 NBT
	{Key: "Age", Label: "Age - Age of the mob (negative for baby, positive for breeding cooldown)", Description: "年齢/成長度 (Age)", ApplicableTo: breedableMobs, Type: "int"},
	{Key: "ForcedAge", Label: "ForcedAge - Ticks to force baby size until automatic growth resume", Description: "強制年齢維持時間 (ForcedAge)", ApplicableTo: breedableMobs, Type: "int"},
	{Key: "LoveCause", Label: "LoveCause - UUID of the player who caused the breeding mode", Description: "繁殖原因プレイヤーUUID (LoveCause)", ApplicableTo: breedableMobs, Type: "string"},
	{Key: "AgeLocked", Label: "AgeLocked - Prevent baby from growing up automatically", Description: "子供の自動成長停止 (AgeLocked)", ApplicableTo: ageLockedMobs, Type: "boolean"},

	// 4. 特定Mob固有 NBT
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
	{Key: "CollarColor", Label: "CollarColor - Dye color of wolf/cat collar (0-15)", Description: "首輪の色（0-15）(CollarColor)", ApplicableTo: []string{"wolf", "cat"}, Type: "select", Choices: []NbtChoice{
		{Value: "0", Label: "0 - 白 (White)"},
		{Value: "1", Label: "1 - 橙 (Orange)"},
		{Value: "2", Label: "2 - マゼンタ (Magenta)"},
		{Value: "3", Label: "3 - 空色 (Light Blue)"},
		{Value: "4", Label: "4 - 黄 (Yellow)"},
		{Value: "5", Label: "5 - 黄緑 (Lime)"},
		{Value: "6", Label: "6 - 桃 (Pink)"},
		{Value: "7", Label: "7 - 灰色 (Gray)"},
		{Value: "8", Label: "8 - 薄灰色 (Light Gray)"},
		{Value: "9", Label: "9 - 青緑 (Cyan)"},
		{Value: "10", Label: "10 - 紫 (Purple)"},
		{Value: "11", Label: "11 - 青 (Blue)"},
		{Value: "12", Label: "12 - 茶 (Brown)"},
		{Value: "13", Label: "13 - 緑 (Green)"},
		{Value: "14", Label: "14 - 赤 (Red)"},
		{Value: "15", Label: "15 - 黒 (Black)"},
	}},
	{Key: "variant", Label: "variant - Frog variant type", Description: "カエルのバリアント (variant)", ApplicableTo: []string{"frog"}, Type: "select", Choices: []NbtChoice{
		{Value: `'"minecraft:temperate"'`, Label: "minecraft:temperate (温帯/茶)"},
		{Value: `'"minecraft:warm"'`, Label: "minecraft:warm (熱帯/白)"},
		{Value: `'"minecraft:cold"'`, Label: "minecraft:cold (寒冷/緑)"},
	}},
	{Key: "variant", Label: "variant - Cat variant type", Description: "猫のバリアント (variant)", ApplicableTo: []string{"cat"}, Type: "select", Choices: []NbtChoice{
		{Value: `'"minecraft:tabby"'`, Label: "minecraft:tabby (トラ猫)"},
		{Value: `'"minecraft:tuxedo"'`, Label: "minecraft:tuxedo (タキシード)"},
		{Value: `'"minecraft:red"'`, Label: "minecraft:red (茶トラ)"},
		{Value: `'"minecraft:siamese"'`, Label: "minecraft:siamese (シャム)"},
		{Value: `'"minecraft:british_shorthair"'`, Label: "minecraft:british_shorthair (ブリティッシュショートヘア)"},
		{Value: `'"minecraft:calico"'`, Label: "minecraft:calico (三毛猫)"},
		{Value: `'"minecraft:persian"'`, Label: "minecraft:persian (ペルシャ)"},
		{Value: `'"minecraft:ragdoll"'`, Label: "minecraft:ragdoll (ラグドール)"},
		{Value: `'"minecraft:white"'`, Label: "minecraft:white (白猫)"},
		{Value: `'"minecraft:black"'`, Label: "minecraft:black (黒猫)"},
		{Value: `'"minecraft:all_black"'`, Label: "minecraft:all_black (黒/ジェリー)"},
	}},
	{Key: "variant", Label: "variant - Llama variant type", Description: "ラマのバリアント (variant)", ApplicableTo: []string{"llama", "trader_llama"}, Type: "select", Choices: []NbtChoice{
		{Value: "0", Label: "0 - クリーム色 (Creamy)"},
		{Value: "1", Label: "1 - 白色 (White)"},
		{Value: "2", Label: "2 - 茶色 (Brown)"},
		{Value: "3", Label: "3 - 灰色 (Gray)"},
	}},
	{Key: "Angry", Label: "Angry - Whether the wolf is angry", Description: "怒り状態か (Angry)", ApplicableTo: []string{"wolf"}, Type: "boolean"},
	{Key: "Tamed", Label: "Tamed - Whether the animal is tamed", Description: "手なずけ状態か (Tamed)", ApplicableTo: []string{"wolf", "cat"}, Type: "boolean"},
	{Key: "ChestedHorse", Label: "ChestedHorse - Whether the horse has chests equipped", Description: "チェスト装着済みか (ChestedHorse)", ApplicableTo: []string{"donkey", "mule", "llama", "trader_llama"}, Type: "boolean"},
	{Key: "Variant", Label: "Variant - Axolotl variant type", Description: "アホロートルのバリアント (Variant)", ApplicableTo: []string{"axolotl"}, Type: "select", Choices: []NbtChoice{
		{Value: "0", Label: "0 - 桃 (Lucy)"},
		{Value: "1", Label: "1 - 茶 (Wild)"},
		{Value: "2", Label: "2 - 金 (Gold)"},
		{Value: "3", Label: "3 - 水色 (Cyan)"},
		{Value: "4", Label: "4 - 青 (Blue)"},
	}},
	{Key: "PlayerCreated", Label: "PlayerCreated - Whether the iron golem was created by player", Description: "プレイヤー製ゴーレムか (PlayerCreated)", ApplicableTo: []string{"iron_golem"}, Type: "boolean"},
	{Key: "VillagerData", Label: "VillagerData - Professional stats of villager", Description: "村人のデータ（例: {profession:\"minecraft:farmer\",level:1}）", ApplicableTo: []string{"villager"}, Type: "raw"},

	// 5. カスタム NBT (汎用)
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

// -----------------------------------------------------------------------------
// Command Parser and TUI Guide Logic for Free Mode
// -----------------------------------------------------------------------------

// ParsedWord represents a single parsed word with its metadata.
type ParsedWord struct {
	Text      string
	IsLiteral bool
	ArgIndex  int
	IsError   bool
}

// ParseResult holds the result of parsing a command line.
type ParseResult struct {
	Words           []ParsedWord
	Suggestions     []string
	SyntaxPreview   string
	ErrorIdx        int
	CurrentParser   string
	CurrentNode     *BrigadierNode
	Registry        string
	CurrentNodeName string
}

// ParserTypeNames maps Brigadier parser types to user-friendly Japanese explanations.
var ParserTypeNames = map[string]string{
	"minecraft:entity":            "対象（プレイヤー名や @p, @a など）",
	"minecraft:game_profile":      "ゲームプロフィール（プレイヤー名）",
	"minecraft:score_holder":      "スコアホルダー（プレイヤー名または対象エンティティ）",
	"minecraft:item_stack":        "アイテムID (例: minecraft:diamond)",
	"minecraft:item_parser":       "アイテム (例: minecraft:diamond)",
	"minecraft:item_enchantment":  "エンチャントするアイテム",
	"minecraft:item_component":    "アイテムコンポーネント（例: damage=10）",
	"minecraft:item_slot":         "インベントリスロット (例: container.0)",
	"minecraft:entity_summon":     "召喚するエンティティID (例: zombie)",
	"minecraft:mob_effect":        "ステータス効果ID (例: speed)",
	"minecraft:enchantment":       "エンチャント効果ID (例: sharpness)",
	"minecraft:block_pos":         "ブロック座標 (X Y Z)",
	"minecraft:vec3":              "3次元座標 (X Y Z)",
	"minecraft:vec2":              "2次元座標 (X Z)",
	"minecraft:column_pos":        "列座標 (X Z)",
	"minecraft:block_state":       "ブロックIDと状態 (例: air, stone)",
	"minecraft:block_input":       "ブロックIDと状態",
	"minecraft:resource_key":      "リソースキー ID",
	"minecraft:resource":          "リソース ID",
	"minecraft:resource_location": "リソースの場所 / パス",
	"minecraft:dimension":         "ディメンション名 (例: minecraft:overworld)",
	"minecraft:nbt_compound_tag":  "NBTデータ (例: {CustomName:'\"Zombie\"'})",
	"minecraft:message":           "メッセージテキスト",
	"minecraft:component":         "JSON テキストコンポーネント",
	"minecraft:item_predicate":    "アイテム判定条件",
	"minecraft:block_predicate":   "ブロック判定条件",
	"brigadier:string":            "文字列",
	"brigadier:integer":           "整数値",
	"brigadier:double":            "浮動小数点数値",
	"brigadier:bool":              "真偽値 (true / false)",
}

// CommandDescriptions maps Minecraft commands to their Japanese descriptions.
var CommandDescriptions = map[string]string{
	"ability":         "プレイヤーのアビリティを設定または確認します。",
	"advancement":     "プレイヤーの進捗を進めたり戻したりします。",
	"attribute":       "エンティティの属性（最大体力や移動速度など）を操作します。",
	"ban":             "プレイヤーをサーバーからBAN（アクセス禁止）にします。",
	"ban-ip":          "IPアドレスをサーバーからBANします。",
	"banlist":         "BANされたプレイヤーまたはIPの一覧を表示します。",
	"bossbar":         "ボスバーを作成、削除、または設定します。",
	"clear":           "プレイヤーのインベントリからアイテムを消去します。",
	"clone":           "指定した範囲のブロックを別の場所に複製します。",
	"damage":          "エンティティにダメージを与えます。",
	"data":            "ブロックやエンティティのNBTデータを取得、結合、変更、消去します。",
	"datapack":        "データパックを有効化、無効化、または一覧表示します。",
	"debug":           "デバッグセッションを開始または停止します。",
	"defaultgamemode": "ワールドのデフォルトのゲームモードを設定します。",
	"difficulty":      "ゲームの難易度を設定します。",
	"effect":          "エンティティにステータス効果（ポーション効果）を付与または消去します。",
	"enchant":         "プレイヤーの持っているアイテムにエンチャントを付与します。",
	"execute":         "他のコマンドを指定した条件や実行者、位置で実行します。",
	"experience":      "プレイヤーに経験値またはレベルを与えたり奪ったりします。",
	"fill":            "指定した範囲を特定のブロックで満たします。",
	"fillbiome":       "指定した範囲のバイオームを変更します。",
	"forceload":       "チャンクを常時ロードするように設定または解除します。",
	"gamemode":        "プレイヤーのゲームモード（サバイバル、クリエイティブなど）を変更します。",
	"gamerule":        "ゲームルールの値を設定または取得します。",
	"give":            "プレイヤーにアイテムを与えます。",
	"help":            "コマンドのヘルプを表示します。",
	"item":            "ブロックやエンティティのインベントリ内のアイテムを変更・置換します。",
	"kick":            "プレイヤーをサーバーからキック（強制退出）します。",
	"kill":            "エンティティをキル（消去）します。",
	"list":            "サーバーに接続しているプレイヤーの一覧を表示します。",
	"locate":          "最も近い構造物、バイオーム、または関心のある場所(POI)の座標を表示します。",
	"loot":            "指定したソース（チェストやエンティティなど）からドロップアイテムをインベントリやワールドに排出します。",
	"me":              "プレイヤーのアクションメッセージを表示します。",
	"msg":             "他のプレイヤーにプライベートメッセージを送信します。",
	"op":              "プレイヤーにオペレーター権限（管理者権限）を付与します。",
	"pardon":          "プレイヤーのBANを解除します。",
	"pardon-ip":       "IPアドレスのBANを解除します。",
	"particle":        "パーティクル（粒子エフェクト）を表示します。",
	"playground":      "プレイグラウンドモードを開始します。",
	"playsound":       "指定したサウンドを再生します。",
	"recipe":          "プレイヤーにレシピを与えたり剥奪したりします。",
	"reload":          "データパック、レシピ、戦利品テーブルなどを再読み込みします。",
	"ride":            "エンティティを他のエンティティに乗せたり降ろしたりします。",
	"say":             "チャット欄にメッセージを送信します。",
	"schedule":        "指定時間後にアクション（関数）を実行するようにスケジュールします。",
	"scoreboard":      "スコアボードのオブジェクトやプレイヤーのスコアを管理します。",
	"seed":            "ワールドのシード値を表示します。",
	"setblock":        "指定した座標のブロックを別のブロックに変更します。",
	"setworldspawn":   "ワールドのスポーン地点を設定します。",
	"spawnpoint":      "プレイヤーの個別スポーン地点を設定します。",
	"spectate":        "プレイヤーに他のエンティティを観戦させます。",
	"spreadplayers":   "エンティティをランダムな位置に分散させます。",
	"stopsound":       "再生中のサウンドを停止します。",
	"summon":          "エンティティ（Mobやアイテムなど）を召喚します。",
	"tag":             "エンティティのカスタムタグを管理します。",
	"team":            "チームの作成、削除、設定、およびメンバーの管理を行います。",
	"teammsg":         "所属しているチームのメンバーのみにメッセージを送信します。",
	"teleport":        "エンティティを指定した位置または他のエンティティの位置へテレポートします。",
	"tell":            "他のプレイヤーにプライベートメッセージを送信します。",
	"tellraw":         "プレイヤーにJSONフォーマットのテキストメッセージを表示します。",
	"time":            "ワールドの時間を変更または照会します。",
	"title":           "プレイヤーの画面にタイトルテキストを表示します。",
	"tp":              "エンティティをテレポートします（teleportの短縮形）。",
	"trigger":         "トリガー型のスコアボード目的を有効化します。",
	"weather":         "天候（晴れ、雨、雷雨）を設定します。",
	"worldborder":     "ワールド境界線を管理します。",
	"xp":              "経験値を操作します（experienceの短縮形）。",
}

// ArgumentDescriptions maps common argument names to descriptive Japanese summaries.
var ArgumentDescriptions = map[string]string{
	"targets":          "効果やアクションの対象となるプレイヤーまたはエンティティ",
	"target":           "効果やアクションの対象となるプレイヤーまたはエンティティ（単一）",
	"destination":      "テレポート先となる座標またはエンティティ",
	"location":         "座標位置 (X Y Z)",
	"pos":              "ブロック座標 (X Y Z)",
	"block":            "配置または指定するブロックID",
	"item":             "アイテムID",
	"count":            "アイテムの個数",
	"amount":           "数量",
	"entity":           "召喚または指定するエンティティID",
	"nbt":              "NBTデータ（JSON形式）",
	"components":       "アイテムに付与するコンポーネント（属性、耐久度、カスタム名など）",
	"slot":             "インベントリのスロット名（例: container.0）",
	"modifier":         "属性のモディファイア（修飾子）ID",
	"effect":           "ステータス効果ID",
	"enchantment":      "エンチャント効果ID",
	"level":            "エンチャントレベルや経験値レベルなどの数値",
	"duration":         "効果の持続時間（秒数または ticks）",
	"amplifier":        "効果の強さ（レベル、0がレベル1に相当）",
	"hideParticles":    "パーティクル（泡のエフェクト）を非表示にするかどうか (true/false)",
	"scale":            "倍率（数値）",
	"reason":           "BANやキックの理由メッセージ",
	"message":          "送信するメッセージテキスト",
	"command":          "実行するコマンド文字列",
	"visible":          "表示するかどうか (true/false)",
	"value":            "設定する値",
	"id":               "IDまたはリソースの場所",
	"name":             "名前（JSON形式のコンポーネントなど）",
	"speed":            "速度（数値）",
}

// GetPropertiesDescription formats limits (min/max) defined in properties.
func (n *BrigadierNode) GetPropertiesDescription() string {
	if n.Properties == nil {
		return ""
	}
	var parts []string
	if min, exists := n.Properties["min"]; exists {
		parts = append(parts, fmt.Sprintf("最小値: %v", min))
	}
	if max, exists := n.Properties["max"]; exists {
		parts = append(parts, fmt.Sprintf("最大値: %v", max))
	}
	if len(parts) > 0 {
		return " (" + strings.Join(parts, ", ") + ")"
	}
	return ""
}

// GetDynamicSuggestions returns list of values for Minecraft argument types.
func GetDynamicSuggestions(parser string, registry string) []string {
	if registry != "" {
		switch registry {
		case "minecraft:entity_type":
			return Entities
		case "minecraft:mob_effect":
			return Effects
		case "minecraft:enchantment":
			var enchNames []string
			for _, e := range Enchantments {
				enchNames = append(enchNames, e.Name)
			}
			return enchNames
		case "minecraft:attribute":
			return []string{"generic.max_health", "generic.movement_speed", "generic.attack_damage"}
		case "minecraft:dimension":
			return []string{"minecraft:overworld", "minecraft:the_nether", "minecraft:the_end"}
		}
	}

	switch parser {
	case "minecraft:entity", "minecraft:game_profile", "minecraft:score_holder":
		return []string{"@p", "@a", "@r", "@s", "@e"}
	case "minecraft:item_stack", "minecraft:item_parser", "minecraft:item_enchantment", "minecraft:item_slot":
		return Items
	case "minecraft:entity_summon":
		return Entities
	case "minecraft:mob_effect":
		return Effects
	case "minecraft:enchantment":
		var enchNames []string
		for _, e := range Enchantments {
			enchNames = append(enchNames, e.Name)
		}
		return enchNames
	case "minecraft:block_pos", "minecraft:vec3", "minecraft:vec2", "minecraft:column_pos":
		return []string{"~ ~ ~"}
	case "minecraft:block_state", "minecraft:block_input":
		return Blocks
	case "minecraft:resource_key", "minecraft:resource":
		if registry == "minecraft:attribute" {
			return []string{"generic.max_health", "generic.movement_speed", "generic.attack_damage"}
		}
		if registry == "minecraft:dimension" {
			return []string{"minecraft:overworld", "minecraft:the_nether", "minecraft:the_end"}
		}
	}
	return nil
}

// combineCoordinates joins consecutive coordinate tokens into a single word token.
func combineCoordinates(rawWords []string) []string {
	var result []string
	n := len(rawWords)
	for i := 0; i < n; {
		w := rawWords[i]
		if w == "" {
			result = append(result, w)
			i++
			continue
		}

		if IsValidPositionToken(w) {
			posParts := []string{w}
			idx := i + 1
			for idx < n && len(posParts) < 3 {
				nextW := rawWords[idx]
				if nextW != "" && IsValidPositionToken(nextW) {
					posParts = append(posParts, nextW)
					idx++
				} else {
					break
				}
			}
			result = append(result, strings.Join(posParts, " "))
			i = idx
		} else {
			result = append(result, w)
			i++
		}
	}
	return result
}

// isValidArgValue checks if a given word is valid for a brigadier parser.
func isValidArgValue(w string, parser string, registry string, isLast bool) bool {
	if parser == "minecraft:block_pos" || parser == "minecraft:vec3" || parser == "minecraft:vec2" || parser == "minecraft:column_pos" {
		parts := strings.Split(w, " ")
		for _, part := range parts {
			if part == "" {
				continue
			}
			if !IsValidPositionToken(part) {
				return false
			}
		}
		return true
	}

	suggestions := GetDynamicSuggestions(parser, registry)
	if suggestions != nil {
		for _, s := range suggestions {
			cleanS := strings.TrimPrefix(s, "minecraft:")
			if isLast {
				if strings.HasPrefix(s, w) || strings.HasPrefix(cleanS, w) {
					return true
				}
			} else {
				if s == w || cleanS == w {
					return true
				}
			}
		}
		return false
	}

	if parser == "brigadier:integer" || parser == "brigadier:double" {
		if isLast {
			return true
		}
		if parser == "brigadier:integer" {
			var val int
			_, err := fmt.Sscanf(w, "%d", &val)
			return err == nil
		} else {
			var val float64
			_, err := fmt.Sscanf(w, "%f", &val)
			return err == nil
		}
	}
	if parser == "brigadier:bool" {
		if isLast {
			return strings.HasPrefix("true", w) || strings.HasPrefix("false", w)
		}
		return w == "true" || w == "false"
	}

	return true
}

// ParseCommand parses a raw input command line dynamically against the commands tree.
func ParseCommand(input string) ParseResult {
	cleanInput := strings.TrimPrefix(input, "/")
	rawWords := combineCoordinates(strings.Split(cleanInput, " "))
	var words []ParsedWord
	currentNode := &CommandTree
	argIdx := 0
	errorIdx := -1
	var currentParser string
	var registry string
	var lastNodeName string

	for i, w := range rawWords {
		isLast := i == len(rawWords)-1
		if isLast {
			break
		}
		if w == "" {
			continue
		}

		if errorIdx != -1 {
			dispText := w
			if i == 0 && strings.HasPrefix(input, "/") {
				dispText = "/" + w
			}
			words = append(words, ParsedWord{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
			continue
		}

		var matched *BrigadierNode
		var matchedName string
		if currentNode.Children != nil {
			if child, exists := currentNode.Children[w]; exists && child.Type == "literal" {
				matched = child
				matchedName = w
			} else {
				for name, child := range currentNode.Children {
					if child.Type == "argument" {
						if isValidArgValue(w, child.Parser, child.GetRegistry(), false) {
							matched = child
							matchedName = name
							break
						}
					}
				}
			}
		}

		dispText := w
		if i == 0 && strings.HasPrefix(input, "/") {
			dispText = "/" + w
		}

		if matched != nil {
			isLiteral := matched.Type == "literal"
			words = append(words, ParsedWord{
				Text:      dispText,
				IsLiteral: isLiteral,
				ArgIndex:  argIdx,
				IsError:   false,
			})
			if !isLiteral {
				argIdx++
			}
			currentNode = matched
			lastNodeName = matchedName
		} else {
			errorIdx = i
			words = append(words, ParsedWord{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
		}
	}

	lastWord := rawWords[len(rawWords)-1]
	var suggestions []string
	var syntaxParts []string

	isFirstWord := len(rawWords) == 1
	hasSlash := strings.HasPrefix(input, "/")

	if errorIdx != -1 {
		dispText := lastWord
		if len(rawWords) == 1 && hasSlash {
			dispText = "/" + lastWord
		}
		words = append(words, ParsedWord{
			Text:      dispText,
			IsLiteral: false,
			ArgIndex:  argIdx,
			IsError:   true,
		})
		return ParseResult{
			Words:         words,
			Suggestions:   nil,
			SyntaxPreview: "",
			ErrorIdx:      errorIdx,
			CurrentParser: "",
			CurrentNode:   currentNode,
		}
	}

	// item component suggestion detection
	if lastIdx := strings.LastIndex(lastWord, "["); lastIdx != -1 && !strings.Contains(lastWord[lastIdx:], "]") {
		prefixPart := lastWord[:lastIdx+1]
		componentPart := lastWord[lastIdx+1:]
		segments := strings.Split(componentPart, ",")
		lastSeg := segments[len(segments)-1]

		if !strings.Contains(lastSeg, "=") {
			argPrefix := ""
			if len(segments) > 1 {
				argPrefix = strings.Join(segments[:len(segments)-1], ",") + ","
			}

			for _, comp := range Components {
				cleanComp := strings.TrimPrefix(comp, "minecraft:")
				if strings.HasPrefix(comp, lastSeg) || strings.HasPrefix(cleanComp, lastSeg) {
					suggestVal := prefixPart + argPrefix + comp
					if isFirstWord && hasSlash {
						suggestVal = "/" + suggestVal
					}
					suggestions = append(suggestions, suggestVal)
				}
			}

			if len(suggestions) > 0 {
				sort.Strings(suggestions)
				
				lastIsLiteral := false
				dispText := lastWord
				if len(rawWords) == 1 && hasSlash {
					dispText = "/" + lastWord
				}
				words = append(words, ParsedWord{
					Text:      dispText,
					IsLiteral: lastIsLiteral,
					ArgIndex:  argIdx,
					IsError:   false,
				})

				return ParseResult{
					Words:           words,
					Suggestions:     suggestions,
					SyntaxPreview:   "item_component",
					ErrorIdx:        errorIdx,
					CurrentParser:   "minecraft:item_component",
					CurrentNode:     currentNode,
					Registry:        registry,
					CurrentNodeName: "components",
				}
			}
		}
	}

	var matchedLast *BrigadierNode
	var matchedLastName string
	if currentNode.Children != nil {
		if child, exists := currentNode.Children[lastWord]; exists && child.Type == "literal" {
			matchedLast = child
			matchedLastName = lastWord
		} else {
			for name, child := range currentNode.Children {
				if child.Type == "argument" {
					if isValidArgValue(lastWord, child.Parser, child.GetRegistry(), true) {
						matchedLast = child
						matchedLastName = name
						break
					}
				}
			}
		}
	}

	if currentNode.Children != nil {
		var keys []string
		for k := range currentNode.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			child := currentNode.Children[key]
			if child.Type == "literal" {
				if strings.HasPrefix(key, lastWord) {
					suggestVal := key
					if isFirstWord && hasSlash {
						suggestVal = "/" + key
					}
					suggestions = append(suggestions, suggestVal)
				}
				syntaxParts = append(syntaxParts, key)
			} else if child.Type == "argument" {
				list := GetDynamicSuggestions(child.Parser, child.GetRegistry())
				var sortedList []string
				if list != nil {
					sortedList = make([]string, len(list))
					copy(sortedList, list)
					sort.Strings(sortedList)
				}
				for _, item := range sortedList {
					cleanItem := strings.TrimPrefix(item, "minecraft:")
					if strings.HasPrefix(item, lastWord) || strings.HasPrefix(cleanItem, lastWord) {
						suggestions = append(suggestions, item)
					}
				}
				syntaxParts = append(syntaxParts, "<"+key+">")
			}
		}
	}

	if matchedLast != nil && matchedLast.Type == "argument" {
		currentParser = matchedLast.Parser
		registry = matchedLast.GetRegistry()
	}

	lastIsError := false
	if lastWord != "" {
		hasValidMatch := false
		if matchedLast != nil {
			hasValidMatch = true
		} else if currentNode.Children != nil {
			for key, child := range currentNode.Children {
				if child.Type == "literal" && strings.HasPrefix(key, lastWord) {
					hasValidMatch = true
					break
				}
			}
		}
		if !hasValidMatch {
			lastIsError = true
			errorIdx = len(rawWords) - 1
		}
	}

	lastIsLiteral := false
	if matchedLast != nil && matchedLast.Type == "literal" {
		lastIsLiteral = true
	}

	if lastWord != "" {
		dispText := lastWord
		if len(rawWords) == 1 && hasSlash {
			dispText = "/" + lastWord
		}
		words = append(words, ParsedWord{
			Text:      dispText,
			IsLiteral: lastIsLiteral,
			ArgIndex:  argIdx,
			IsError:   lastIsError,
		})
	}

	var nextSyntaxPreview string
	if len(syntaxParts) > 0 {
		nextSyntaxPreview = strings.Join(syntaxParts, " | ")
	}

	activeNode := currentNode
	activeNodeName := lastNodeName
	if matchedLast != nil && !lastIsError {
		activeNode = matchedLast
		activeNodeName = matchedLastName
	}

	return ParseResult{
		Words:           words,
		Suggestions:     suggestions,
		SyntaxPreview:   nextSyntaxPreview,
		ErrorIdx:        errorIdx,
		CurrentParser:   currentParser,
		CurrentNode:     activeNode,
		Registry:        registry,
		CurrentNodeName: activeNodeName,
	}
}

// GetSyntaxGuides returns possible command paths from a given node.
func GetSyntaxGuides(node *BrigadierNode, prefix string) []string {
	if node == nil {
		return nil
	}

	var results []string
	var dfs func(n *BrigadierNode, currentPath []string)
	dfs = func(n *BrigadierNode, currentPath []string) {
		if n.Executable {
			results = append(results, strings.Join(currentPath, " "))
		}

		if n.Children == nil || len(n.Children) == 0 {
			return
		}

		var keys []string
		for k := range n.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			child := n.Children[k]
			var segment string
			if child.Type == "literal" {
				segment = k
			} else {
				segment = "<" + k + ">"
			}
			dfs(child, append(currentPath, segment))
		}
	}

	var initialPath []string
	if prefix != "" {
		initialPath = []string{prefix}
	}

	if node.Children != nil {
		var keys []string
		for k := range node.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := node.Children[k]
			var segment string
			if child.Type == "literal" {
				segment = k
			} else {
				segment = "<" + k + ">"
			}
			dfs(child, append(initialPath, segment))
		}
	}

	return results
}
