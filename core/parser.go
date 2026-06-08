package core

import (
	"fmt"
	"sort"
	"strings"
)

// Token represents a single parsed word token from the command line.
type Token struct {
	Text            string
	HasLeadingSpace bool
}

// Tokenize splits command input into tokens, taking braces/brackets and quotes protection into account,
// and separating direct-coupled component suffixes like netherite_sword[damage=0] into separate tokens.
func Tokenize(input string) []Token {
	var tokens []Token
	var current strings.Builder
	var bracketStack []rune
	inDoubleQuotes := false
	inSingleQuotes := false
	escaped := false
	hasLeadingSpace := false

	runes := []rune(input)

	addToken := func(text string, leadingSpace bool) {
		tokens = append(tokens, Token{
			Text:            text,
			HasLeadingSpace: leadingSpace,
		})
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			current.WriteRune(r)
			escaped = true
			continue
		}

		if inDoubleQuotes {
			current.WriteRune(r)
			if r == '"' {
				inDoubleQuotes = false
			}
			continue
		}

		if inSingleQuotes {
			current.WriteRune(r)
			if r == '\'' {
				inSingleQuotes = false
			}
			continue
		}

		// Not in quotes
		if r == '"' {
			inDoubleQuotes = true
			current.WriteRune(r)
			continue
		}
		if r == '\'' {
			inSingleQuotes = true
			current.WriteRune(r)
			continue
		}

		// Bracket handling
		if r == '{' || r == '[' {
			// If stack is empty and current token has text, split here!
			// The suffix starting with '{' or '[' gets HasLeadingSpace = false.
			if len(bracketStack) == 0 && current.Len() > 0 {
				addToken(current.String(), hasLeadingSpace)
				current.Reset()
				hasLeadingSpace = false // Directly attached suffix
			}
			bracketStack = append(bracketStack, r)
			current.WriteRune(r)
			continue
		}

		if r == '}' || r == ']' {
			current.WriteRune(r)
			if len(bracketStack) > 0 {
				top := bracketStack[len(bracketStack)-1]
				if (r == '}' && top == '{') || (r == ']' && top == '[') {
					bracketStack = bracketStack[:len(bracketStack)-1]
				}
			}
			continue
		}

		// Space handling
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if len(bracketStack) == 0 {
				if current.Len() > 0 {
					addToken(current.String(), hasLeadingSpace)
					current.Reset()
				}
				hasLeadingSpace = true
			} else {
				current.WriteRune(r)
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		addToken(current.String(), hasLeadingSpace)
	} else if len(runes) > 0 && (runes[len(runes)-1] == ' ' || runes[len(runes)-1] == '\t' || runes[len(runes)-1] == '\n' || runes[len(runes)-1] == '\r') {
		addToken("", true)
	}
	return tokens
}

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
	OriginalParser  string
	CurrentNode     *BrigadierNode
	Registry        string
	CurrentNodeName string
	IsExecutable    bool
}

// resolveRedirect resolves a redirect path (slice of strings) starting from the root tree.
func resolveRedirect(root *BrigadierNode, path []string) *BrigadierNode {
	curr := root
	for _, segment := range path {
		if curr.Children == nil {
			return nil
		}
		next, exists := curr.Children[segment]
		if !exists {
			return nil
		}
		curr = next
	}
	return curr
}

// getCoordinateArgValTokens extracts consecutive valid coordinate tokens up to the maximum required.
func getCoordinateArgValTokens(tokens []Token, start int, parser string) (string, int) {
	var maxRequired int
	if parser == "minecraft:block_pos" || parser == "minecraft:vec3" {
		maxRequired = 3
	} else if parser == "minecraft:vec2" || parser == "minecraft:column_pos" {
		maxRequired = 2
	} else {
		return "", 0
	}

	var parts []string
	idx := start
	n := len(tokens)
	lastConsumedIdx := start

	for idx < n && len(parts) < maxRequired {
		tok := tokens[idx]
		if tok.Text == "" {
			idx++
			continue
		}
		if IsValidPositionToken(tok.Text) {
			parts = append(parts, tok.Text)
			idx++
			lastConsumedIdx = idx
		} else {
			break
		}
	}
	return strings.Join(parts, " "), lastConsumedIdx - start
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
			wLower := strings.ToLower(w)
			sLower := strings.ToLower(s)
			cleanSLower := strings.ToLower(cleanS)
			if isLast {
				if strings.HasPrefix(sLower, wLower) || strings.HasPrefix(cleanSLower, wLower) {
					return true
				}
			} else {
				if sLower == wLower || cleanSLower == wLower {
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
		wLower := strings.ToLower(w)
		if isLast {
			return strings.HasPrefix("true", wLower) || strings.HasPrefix("false", wLower)
		}
		return wLower == "true" || wLower == "false"
	}

	return true
}

// isIncompleteBrackets checks if the string has unclosed brackets [ or {
func isIncompleteBrackets(s string) bool {
	squareCount := 0
	curlyCount := 0
	for _, r := range s {
		if r == '[' {
			squareCount++
		} else if r == ']' {
			squareCount--
		} else if r == '{' {
			curlyCount++
		} else if r == '}' {
			curlyCount--
		}
	}
	return squareCount > 0 || curlyCount > 0
}

// isCompletedBrackets checks if the string contains brackets [ or { and all are closed
func isCompletedBrackets(s string) bool {
	hasOpen := strings.Contains(s, "[") || strings.Contains(s, "{")
	if !hasOpen {
		return false
	}
	squareCount := 0
	curlyCount := 0
	for _, r := range s {
		if r == '[' {
			squareCount++
		} else if r == ']' {
			squareCount--
		} else if r == '{' {
			curlyCount++
		} else if r == '}' {
			curlyCount--
		}
	}
	return squareCount == 0 && curlyCount == 0
}

// ParseCommand parses a raw input command line dynamically against the commands tree.
func ParseCommand(input string) ParseResult {
	cleanInput := strings.TrimPrefix(input, "/")
	tokens := Tokenize(cleanInput)
	var words []ParsedWord
	currentNode := &CommandTree
	argIdx := 0
	errorIdx := -1
	var currentParser string
	var originalParser string
	var registry string
	var lastNodeName string

	// 現在の座標引数の状態管理
	currentPosTokensCount := 0
	currentPosTokensRequired := 0

	p := 0
	n := len(tokens)

	for p < n {
		isLast := p == n-1
		tok := tokens[p]

		// 1. カッコや引数直結のコンポーネントなどの修飾トークンの場合
		isMod := strings.HasPrefix(tok.Text, "[") || strings.HasPrefix(tok.Text, "{") || isIncompleteBrackets(tok.Text) || isCompletedBrackets(tok.Text)
		if p > 0 && !isLast && isMod {
			prevIsError := false
			prevArgIdx := 0
			if len(words) > 0 {
				prevIsError = words[len(words)-1].IsError
				prevArgIdx = words[len(words)-1].ArgIndex
			}
			words = append(words, ParsedWord{
				Text:      tok.Text,
				IsLiteral: false,
				ArgIndex:  prevArgIdx,
				IsError:   prevIsError || errorIdx != -1,
			})
			p++
			continue
		}

		// 2. 直前の座標引数が未完成のまま、新しい非空トークンが入力された場合のチェック
		if errorIdx == -1 && tok.Text != "" && currentPosTokensRequired > 0 && currentPosTokensCount < currentPosTokensRequired {
			errorIdx = p
		}

		// 3. すでにエラーが検出されている場合
		if errorIdx != -1 {
			dispText := tok.Text
			if p == 0 && strings.HasPrefix(input, "/") {
				dispText = "/" + dispText
			}
			words = append(words, ParsedWord{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
			p++
			continue
		}

		// 4. 最後のトークンはループ外で処理
		if isLast {
			break
		}

		var matched *BrigadierNode
		var matchedName string
		consumedTokens := 1
		var matchedVal string
		var newPosCount int
		var newPosRequired int
		deferToLastWord := false

		if currentNode.Children != nil {
			tokLower := strings.ToLower(tok.Text)
			var literalChild *BrigadierNode
			var literalName string
			for name, child := range currentNode.Children {
				if child.Type == "literal" && strings.ToLower(name) == tokLower {
					literalChild = child
					literalName = name
					break
				}
			}

			if literalChild != nil {
				matched = literalChild
				matchedName = literalName
				matchedVal = tok.Text
			} else {
				for name, child := range currentNode.Children {
					if child.Type == "argument" {
						isCoord := child.Parser == "minecraft:block_pos" || child.Parser == "minecraft:vec3" || child.Parser == "minecraft:vec2" || child.Parser == "minecraft:column_pos"
						if isCoord {
							val, count := getCoordinateArgValTokens(tokens, p, child.Parser)
							if count > 0 {
								requiredCount := 3
								if child.Parser == "minecraft:vec2" || child.Parser == "minecraft:column_pos" {
									requiredCount = 2
								}

								// 最後(n-1)のトークンを含む、かつ座標数が足りない（未完成）場合は、
								// ここではマッチさせずループ外で処理する
								if p+count-1 == n-1 && count < requiredCount {
									deferToLastWord = true
									break
								}

								if isValidArgValue(val, child.Parser, child.GetRegistry(), false) {
									matched = child
									matchedName = name
									consumedTokens = count
									matchedVal = val
									newPosCount = count
									newPosRequired = requiredCount
									break
								}
							}
						} else {
							if isValidArgValue(tok.Text, child.Parser, child.GetRegistry(), false) {
								matched = child
								matchedName = name
								consumedTokens = 1
								matchedVal = tok.Text
								break
							}
						}
					}
				}
			}
		}

		if deferToLastWord {
			break
		}

		dispText := matchedVal
		if dispText == "" {
			dispText = tok.Text
		}
		if p == 0 && strings.HasPrefix(input, "/") {
			dispText = "/" + dispText
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

			// redirect または run リテラルによる遷移の追従
			if len(matched.Redirect) > 0 {
				target := resolveRedirect(&CommandTree, matched.Redirect)
				if target != nil {
					currentNode = target
				} else {
					currentNode = matched
				}
			} else if matched.Type == "literal" && matchedName == "run" && (matched.Children == nil || len(matched.Children) == 0) {
				currentNode = &CommandTree
			} else {
				currentNode = matched
			}

			lastNodeName = matchedName
			p += consumedTokens
			currentPosTokensCount = newPosCount
			currentPosTokensRequired = newPosRequired
		} else {
			errorIdx = p
			words = append(words, ParsedWord{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
			p++
		}
	}

	var lastWord string
	var lastToken Token
	if p < n {
		lastToken = tokens[p]
		var parts []string
		for idx := p; idx < n; idx++ {
			if tokens[idx].Text != "" {
				parts = append(parts, tokens[idx].Text)
			}
		}
		lastWord = strings.Join(parts, " ")
	}

	hasSlash := strings.HasPrefix(input, "/")

	// 最後のトークン処理に進む前のエラーチェック
	if errorIdx == -1 && currentPosTokensRequired > 0 && currentPosTokensCount < currentPosTokensRequired {
		isCoord := true
		for _, w := range strings.Fields(lastWord) {
			if !IsValidPositionToken(w) {
				isCoord = false
				break
			}
		}
		if !isCoord {
			errorIdx = p
		}
	}

	if errorIdx != -1 {
		for idx := p; idx < n; idx++ {
			tok := tokens[idx]
			dispText := tok.Text
			if idx == 0 && hasSlash {
				dispText = "/" + dispText
			}
			words = append(words, ParsedWord{
				Text:      dispText,
				IsLiteral: false,
				ArgIndex:  argIdx,
				IsError:   true,
			})
		}
		return ParseResult{
			Words:         words,
			Suggestions:   nil,
			SyntaxPreview: "",
			ErrorIdx:      errorIdx,
			CurrentParser: "",
			CurrentNode:   currentNode,
			IsExecutable:  false,
		}
	}

	var matchedLast *BrigadierNode
	var matchedLastName string
	var lastIsError bool
	var matchedLastVal string

	isCompletedModifier := isCompletedBrackets(lastWord)
	isIncompleteModifier := isIncompleteBrackets(lastWord)

	if isCompletedModifier || isIncompleteModifier {
		matchedLast = currentNode
		matchedLastName = lastNodeName
		matchedLastVal = lastWord
	} else if currentNode.Children != nil {
		lastWordLower := strings.ToLower(lastWord)
		if lastWord != "" {
			for name, child := range currentNode.Children {
				if child.Type == "literal" && strings.ToLower(name) == lastWordLower {
					matchedLast = child
					matchedLastName = name
					matchedLastVal = name
					break
				}
			}
		}
		if matchedLast == nil {
			for name, child := range currentNode.Children {
				if child.Type == "argument" {
					isCoord := child.Parser == "minecraft:block_pos" || child.Parser == "minecraft:vec3" || child.Parser == "minecraft:vec2" || child.Parser == "minecraft:column_pos"
					if isCoord {
						parts := strings.Fields(lastWord)
						validCount := 0
						allValid := true
						for _, p := range parts {
							if IsValidPositionToken(p) {
								validCount++
							} else {
								allValid = false
							}
						}
						requiredCount := 3
						if child.Parser == "minecraft:vec2" || child.Parser == "minecraft:column_pos" {
							requiredCount = 2
						}
						if allValid && validCount <= requiredCount && (validCount > 0 || lastWord == "") {
							matchedLastVal = lastWord
							matchedLast = child
							matchedLastName = name
							currentPosTokensCount = validCount
							currentPosTokensRequired = requiredCount
							break
						}
					} else {
						if isValidArgValue(lastWord, child.Parser, child.GetRegistry(), true) {
							matchedLast = child
							matchedLastName = name
							matchedLastVal = lastWord
							break
						}
					}
				}
			}
		}
	}

	if matchedLast != nil && matchedLast.Type == "argument" {
		currentParser = matchedLast.Parser
		registry = matchedLast.GetRegistry()
		originalParser = matchedLast.Parser
	}

	if lastWord != "" {
		hasValidMatch := false
		if matchedLast != nil {
			hasValidMatch = true
		} else if matchedLastVal != "" {
			hasValidMatch = true
		} else if currentNode.Children != nil {
			lastWordLower := strings.ToLower(lastWord)
			for key, child := range currentNode.Children {
				if child.Type == "literal" && strings.HasPrefix(strings.ToLower(key), lastWordLower) {
					hasValidMatch = true
					break
				}
			}
		}
		if !hasValidMatch {
			lastIsError = true
			errorIdx = p
		}
	}

	lastIsLiteral := false
	if matchedLast != nil && matchedLast.Type == "literal" {
		lastIsLiteral = true
	}

	if lastWord != "" {
		dispText := lastWord
		if p == 0 && hasSlash {
			dispText = "/" + dispText
		}
		words = append(words, ParsedWord{
			Text:      dispText,
			IsLiteral: lastIsLiteral,
			ArgIndex:  argIdx,
			IsError:   lastIsError,
		})
	}

	var syntaxParts []string
	var activeNode *BrigadierNode
	var activeNodeName string

	if matchedLast != nil && !lastIsError {
		if len(matchedLast.Redirect) > 0 {
			target := resolveRedirect(&CommandTree, matchedLast.Redirect)
			if target != nil {
				activeNode = target
			} else {
				activeNode = matchedLast
			}
		} else if matchedLast.Type == "literal" && matchedLastName == "run" && (matchedLast.Children == nil || len(matchedLast.Children) == 0) {
			activeNode = &CommandTree
		} else {
			activeNode = matchedLast
		}
		activeNodeName = matchedLastName
	} else {
		activeNode = currentNode
		activeNodeName = lastNodeName
	}
	if isIncompleteModifier || isCompletedModifier {
		if currentParser == "minecraft:item_stack" || currentParser == "minecraft:item_parser" {
			currentParser = "minecraft:item_component"
			activeNodeName = "components"
		}
	}
	var prependPart string
	if p > 0 && !lastToken.HasLeadingSpace {
		prependPart = tokens[p-1].Text
	}

	var suggestions []string
	if isIncompleteModifier {
		suggestions = getModifierSuggestions(lastWord, currentParser, prependPart)
	} else if currentNode.Children != nil {
		var keys []string
		for k := range currentNode.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		lastWordLower := strings.ToLower(lastWord)
		for _, key := range keys {
			child := currentNode.Children[key]
			if child.Type == "literal" {
				if strings.HasPrefix(strings.ToLower(key), lastWordLower) {
					suggestVal := key
					if p == 0 && hasSlash {
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
					if strings.HasPrefix(strings.ToLower(item), lastWordLower) || strings.HasPrefix(strings.ToLower(cleanItem), lastWordLower) {
						suggestions = append(suggestions, item)
					}
				}
				syntaxParts = append(syntaxParts, "<"+key+">")
			}
		}
	}

	var nextSyntaxPreview string
	if len(syntaxParts) > 0 {
		nextSyntaxPreview = strings.Join(syntaxParts, " | ")
	}

	// 適切なタイミングでのみサジェストを行う文脈制御
	inputTrimmed := strings.TrimRight(input, " \t\r\n")
	hasTrailingSpace := len(input) > len(inputTrimmed)

	isLastWordFullyMatched := false
	if matchedLast != nil && lastWord != "" {
		if matchedLast.Type == "literal" {
			isLastWordFullyMatched = strings.ToLower(lastWord) == strings.ToLower(matchedLastVal)
		} else {
			isLastWordFullyMatched = isValidArgValue(lastWord, matchedLast.Parser, matchedLast.GetRegistry(), false)
		}
	}

	isCoordsCompleted := currentPosTokensRequired > 0 && currentPosTokensCount == currentPosTokensRequired
	hasNoChildren := activeNode != nil && (activeNode.Children == nil || len(activeNode.Children) == 0)

	if !hasTrailingSpace {
		// 1. 引数やリテラルが完全に確定しており、かつ末尾にスペースがない場合
		// 2. または座標引数が完了している場合
		if isLastWordFullyMatched || isCoordsCompleted {
			suggestions = nil
		}
	} else {
		// 末尾にスペースがあるが、次の子ノードが存在しない場合（コマンド終端）
		if hasNoChildren {
			suggestions = nil
		}
	}

	// executable 判定（エラーがなく、座標が完全で、かつ現在のアクティブノードが executable の場合）
	isExecutable := errorIdx == -1 && !lastIsError && !isIncompleteModifier && activeNode != nil && activeNode.Executable &&
		(currentPosTokensRequired == 0 || currentPosTokensCount == currentPosTokensRequired)

	return ParseResult{
		Words:           words,
		Suggestions:     suggestions,
		SyntaxPreview:   nextSyntaxPreview,
		ErrorIdx:        errorIdx,
		CurrentParser:   currentParser,
		OriginalParser:  originalParser,
		CurrentNode:     activeNode,
		Registry:        registry,
		CurrentNodeName: activeNodeName,
		IsExecutable:    isExecutable,
	}
}

// selector keys that can be used inside @p[...]
var selectorKeys = []string{
	"sort=", "gamemode=", "limit=", "distance=", "level=", "team=",
	"name=", "tag=", "type=", "x=", "y=", "z=", "dx=", "dy=", "dz=", "scores=",
}

// selector sort values
var selectorSortValues = []string{"nearest", "furthest", "random", "arbitrary"}

// selector gamemode values
var selectorGamemodeValues = []string{"survival", "creative", "adventure", "spectator"}

func getModifierSuggestions(lastWord string, parser string, prependPart string) []string {
	// find the last '[' or '{'
	idx := strings.LastIndex(lastWord, "[")
	isBrace := false
	if idx == -1 {
		idx = strings.LastIndex(lastWord, "{")
		isBrace = true
	}
	if idx == -1 {
		return nil
	}

	prefixPart := lastWord[:idx+1]
	innerPart := lastWord[idx+1:]

	segments := strings.Split(innerPart, ",")
	lastSeg := segments[len(segments)-1]

	var innerPrefix string
	if len(segments) > 1 {
		innerPrefix = strings.Join(segments[:len(segments)-1], ",") + ","
	}
	fullPrefix := prependPart + prefixPart + innerPrefix

	var suggestions []string

	if !strings.Contains(lastSeg, "=") {
		// Typing key
		lastSegLower := strings.ToLower(lastSeg)
		if isBrace {
			for _, tag := range NbtMasterList {
				key := tag.Key + ":"
				if strings.HasPrefix(strings.ToLower(key), lastSegLower) {
					suggestions = append(suggestions, fullPrefix+key)
				}
			}
		} else {
			if parser == "minecraft:entity" || parser == "minecraft:game_profile" || parser == "minecraft:score_holder" {
				for _, key := range selectorKeys {
					if strings.HasPrefix(strings.ToLower(key), lastSegLower) {
						suggestions = append(suggestions, fullPrefix+key)
					}
				}
			} else if parser == "minecraft:item_stack" || parser == "minecraft:item_parser" || parser == "minecraft:item_component" {
				for _, comp := range Components {
					cleanComp := strings.TrimPrefix(comp, "minecraft:")
					compLower := strings.ToLower(comp)
					cleanCompLower := strings.ToLower(cleanComp)
					if strings.HasPrefix(compLower, lastSegLower) || strings.HasPrefix(cleanCompLower, lastSegLower) {
						suggestions = append(suggestions, fullPrefix+comp+"=")
						if strings.HasPrefix(cleanCompLower, lastSegLower) && comp != cleanComp {
							suggestions = append(suggestions, fullPrefix+cleanComp+"=")
						}
					}
				}
			} else if parser == "minecraft:block_state" || parser == "minecraft:block_input" {
				blockProps := []string{"snowy=", "waterlogged=", "facing=", "powered=", "lit=", "half=", "shape=", "axis="}
				for _, key := range blockProps {
					if strings.HasPrefix(strings.ToLower(key), lastSegLower) {
						suggestions = append(suggestions, fullPrefix+key)
					}
				}
			}
		}
	} else {
		// Typing value (has '=') or NBT (has ':')
		parts := strings.SplitN(lastSeg, "=", 2)
		var key, valInput string
		sep := "="
		if len(parts) < 2 && isBrace {
			parts = strings.SplitN(lastSeg, ":", 2)
			sep = ":"
		}
		if len(parts) >= 2 {
			key = parts[0]
			valInput = parts[1]
		} else {
			key = lastSeg
			valInput = ""
		}

		var valCandidates []string
		keyLower := strings.ToLower(key)

		if parser == "minecraft:entity" || parser == "minecraft:game_profile" || parser == "minecraft:score_holder" {
			if keyLower == "sort" {
				valCandidates = selectorSortValues
			} else if keyLower == "gamemode" {
				valCandidates = selectorGamemodeValues
			} else if keyLower == "type" {
				valCandidates = Entities
			}
		} else if parser == "minecraft:item_stack" || parser == "minecraft:item_parser" || parser == "minecraft:item_component" {
			if keyLower == "minecraft:unbreakable" || keyLower == "unbreakable" {
				valCandidates = []string{"{}"}
			}
		}

		valInputLower := strings.ToLower(valInput)
		for _, val := range valCandidates {
			if strings.HasPrefix(strings.ToLower(val), valInputLower) {
				suggestions = append(suggestions, fullPrefix+key+sep+val)
			}
		}
	}

	sort.Strings(suggestions)
	return suggestions
}
