package characters

/*
var (
	debugCharacters = []string{
		"Genius",
		"n/a",
		"n/a",
		"n/a",
		"n/a",
		"n/a",
	}
	debugCharacterMetadata = []int{
		-1,
		-1,
		-1,
		-1,
		-1,
		-1,
	}
	debugUsernames = []string{
		"test",
		"test1",
		"test2",
		"test3",
		"test4",
		"test5",
		"test6",
		"test7",
	}
)

func charactersGenerate(g *Game) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	// Local variables
	variant := variants[g.Options.VariantName]

	// If predefined character selections were specified, use those
	if g.ExtraOptions.CustomCharacterAssignments != nil &&
		len(g.ExtraOptions.CustomCharacterAssignments) != 0 {

		if len(g.ExtraOptions.CustomCharacterAssignments) != len(g.Players) {
			hLog.Errorf(
				"There are %v predefined characters, but there are %v players in the game.",
				len(g.ExtraOptions.CustomCharacterAssignments),
				len(g.Players),
			)
			return
		}

		for i, p := range g.Players {
			p.Character = g.ExtraOptions.CustomCharacterAssignments[i].Name
			p.CharacterMetadata = g.ExtraOptions.CustomCharacterAssignments[i].Metadata
		}
		return
	}

	// This is not a replay,
	// so we must generate new random character selections based on the game's seed
	setSeed(g.Seed) // Seed the random number generator

	for i, p := range g.Players {
		// Set the character
		if stringInSlice(p.Name, debugUsernames) {
			// Hard-code some character assignments for testing purposes
			p.Character = debugCharacters[i]
		} else {
			for {
				// Get a random character assignment
				// We don't have to seed the PRNG,
				// since that was done just a moment ago when the deck was shuffled
				randomIndex := rand.Intn(len(characterNames)) // nolint: gosec
				p.Character = characterNames[randomIndex]

				// Check to see if any other players have this assignment already
				alreadyAssigned := false
				for j, p2 := range g.Players {
					if i == j {
						break
					}

					if p2.Character == p.Character {
						alreadyAssigned = true
						break
					}
				}
				if alreadyAssigned {
					continue
				}

				// Check to see if this character is restricted from 2-player games
				if characters[p.Character].Not2P && len(g.Players) == 2 {
					continue
				}

				break
			}
		}

		// Initialize the metadata to -1
		p.CharacterMetadata = -1

		// Specific characters also have secondary attributes that are stored in the character
		// metadata field
		if stringInSlice(p.Name, debugUsernames) {
			p.CharacterMetadata = debugCharacterMetadata[i]
		} else {
			if p.Character == "Fuming" { // 0
				// A random number from 0 to the number of colors in this variant
				p.CharacterMetadata = rand.Intn(len(variant.ClueColors)) // nolint: gosec
			} else if p.Character == "Dumbfounded" { // 1
				// A random number from 1 to 5
				p.CharacterMetadata = rand.Intn(4) + 1 // nolint: gosec
			} else if p.Character == "Inept" { // 2
				// A random number from 0 to the number of colors in this variant
				p.CharacterMetadata = rand.Intn(len(variant.ClueColors)) // nolint: gosec
			} else if p.Character == "Awkward" { // 3
				// A random number from 1 to 5
				p.CharacterMetadata = rand.Intn(4) + 1 // nolint: gosec
			}
		}
	}
}

// characterValidateAction returns true if validation fails
func characterValidateAction(s *Session, d *CommandData, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	if p.Character == "Vindictive" && // 9
		p.CharacterMetadata == 0 &&
		(d.Type != constants.ActionTypeColorClue && d.Type != constants.ActionTypeRankClue) {

		s.Warningf(
			"You are %v, so you must give a clue if you have been given a clue on this go-around.",
			p.Character,
		)
		return true
	} else if p.Character == "Insistent" && // 13
		p.CharacterMetadata != -1 &&
		(d.Type != constants.ActionTypeColorClue && d.Type != constants.ActionTypeRankClue) {

		s.Warningf(
			"You are %v, so you must continue to clue the same card until it is played or discarded.",
			p.Character,
		)
		return true
	} else if p.Character == "Impulsive" && // 17
		p.CharacterMetadata == 0 &&
		(d.Type != constants.ActionTypePlay ||
			d.Target != p.Hand[len(p.Hand)-1].Order) {

		s.Warningf(
			"You are %v, so you must play your slot 1 card after it has been clued.",
			p.Character,
		)
		return true
	} else if p.Character == "Indolent" && // 18
		d.Type == constants.ActionTypePlay &&
		p.CharacterMetadata == 0 {

		s.Warningf(
			"You are %v, so you cannot play a card if you played one in the last round.",
			p.Character,
		)
		return true
	} else if p.Character == "Stubborn" && // 28
		(d.Type == p.CharacterMetadata ||
			(d.Type == constants.ActionTypeColorClue && p.CharacterMetadata == constants.ActionTypeRankClue) ||
			(d.Type == constants.ActionTypeRankClue && p.CharacterMetadata == constants.ActionTypeColorClue)) {

		s.Warningf(
			"You are %v, so you cannot perform the same kind of action that the previous player did.",
			p.Character,
		)
		return true
	}

	return false
}

// characterValidateSecondAction returns true if validation fails
func characterValidateSecondAction(s *Session, d *CommandData, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	if p.CharacterMetadata == -1 {
		return false
	}

	if p.Character == "Genius" { // 24
		if d.Type != ActionTypeRankClue {
			s.Warningf("You are %v, so you must now give a rank clue.", p.Character)
			return true
		}

		if d.Target != p.CharacterMetadata {
			s.Warningf(
				"You are %v, so you must give the second clue to the same player.",
				p.Character,
			)
			return true
		}
	} else if p.Character == "Panicky" && // 26
		d.Type != ActionTypeDiscard {

		s.Warningf(
			"You are %v, so you must discard again since there are 4 or less clues available.",
			p.Character,
		)
		return true
	}

	return false
}

// characterValidateClue returns true if validation fails
func characterValidateClue(s *Session, d *CommandData, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	// Local variables
	variant := variants[g.Options.VariantName]
	clue := NewClue(d)        // Convert the incoming data to a clue object
	p2 := g.Players[d.Target] // Get the target of the clue

	if p.Character == "Fuming" && // 0
		clue.Type == ClueTypeColor &&
		clue.Value != p.CharacterMetadata {

		s.Warningf("You are %v, so you can not give that type of clue.", p.Character)
		return true
	} else if p.Character == "Dumbfounded" && // 1
		clue.Type == ClueTypeRank &&
		clue.Value != p.CharacterMetadata {

		s.Warningf("You are %v, so you can not give that type of clue.", p.Character)
		return true
	} else if p.Character == "Inept" { // 2
		cardsTouched := p2.FindCardsTouchedByClue(clue)
		for _, order := range cardsTouched {
			c := g.Deck[order]
			if c.SuitIndex == p.CharacterMetadata {
				s.Warningf(
					"You are %v, so you cannot give clues that touch a specific suit.",
					p.Character,
				)
				return true
			}
		}
	} else if p.Character == "Awkward" { // 3
		cardsTouched := p2.FindCardsTouchedByClue(clue)
		for _, order := range cardsTouched {
			c := g.Deck[order]
			if c.Rank == p.CharacterMetadata {
				s.Warningf(
					"You are %v, so you cannot give clues that touch cards with a rank of %v.",
					p.Character,
					p.CharacterMetadata,
				)
				return true
			}
		}
	} else if p.Character == "Conservative" && // 4
		len(p2.FindCardsTouchedByClue(clue)) != 1 {

		s.Warningf(
			"You are %v, so you can only give clues that touch a single card.",
			p.Character,
		)
		return true
	} else if p.Character == "Greedy" && // 5
		len(p2.FindCardsTouchedByClue(clue)) < 2 {

		s.Warningf(
			"You are %v, so you can only give clues that touch 2+ cards.",
			p.Character,
		)
		return true
	} else if p.Character == "Picky" && // 6
		((clue.Type == ClueTypeRank &&
			clue.Value%2 == 0) ||
			(clue.Type == ClueTypeColor &&
				(clue.Value+1)%2 == 0)) {

		s.Warningf(
			"You are %v, so you can only clue odd numbers or odd colors.",
			p.Character,
		)
		return true
	} else if p.Character == "Spiteful" { // 7
		leftIndex := p.Index + 1
		if leftIndex == len(g.Players) {
			leftIndex = 0
		}
		if d.Target == leftIndex {
			s.Warningf(
				"You are %v, so you cannot clue the player to your left.",
				p.Character,
			)
			return true
		}
	} else if p.Character == "Insolent" { // 8
		rightIndex := p.Index - 1
		if rightIndex == -1 {
			rightIndex = len(g.Players) - 1
		}
		if d.Target == rightIndex {
			s.Warningf("You are %v, so you cannot clue the player to your right.", p.Character)
			return true
		}
	} else if p.Character == "Miser" && // 10
		g.ClueTokens < variant.GetAdjustedClueTokens(4) {

		s.Warningf(
			"You are %v, so you cannot give a clue unless there are 4 or more clues available.",
			p.Character,
		)
		return true
	} else if p.Character == "Compulsive" && // 11
		!p2.IsFirstCardTouchedByClue(clue) &&
		!p2.IsLastCardTouchedByClue(clue) {

		s.Warningf(
			"You are %v, so you can only give a clue if it touches either the newest or oldest card in a hand.",
			p.Character,
		)
		return true
	} else if p.Character == "Mood Swings" && // 12
		p.CharacterMetadata == clue.Type {

		s.Warningf("You are %v, so cannot give the same clue type twice in a row.", p.Character)
		return true
	} else if p.Character == "Insistent" && // 13
		p.CharacterMetadata != -1 {

		cardsTouched := p2.FindCardsTouchedByClue(clue)
		touchedInsistentCard := false
		for _, order := range cardsTouched {
			c := g.Deck[order]
			if c.InsistentTouched {
				touchedInsistentCard = true
				break
			}
		}
		if !touchedInsistentCard {
			s.Warningf(
				"You are %v, so you must continue to clue a card until it is played or discarded.",
				p.Character,
			)
			return true
		}
	} else if p.Character == "Genius" && // 24
		p.CharacterMetadata == -1 {

		if g.ClueTokens < variant.GetAdjustedClueTokens(2) {
			s.Warningf(
				"You are %v, so there needs to be at least two clues available for you to give a clue.",
				p.Character,
			)
			return true
		}

		if clue.Type != ClueTypeColor {
			s.Warningf("You are %v, so you must give a color clue first.", p.Character)
			return true
		}
	}

	if p2.Character == "Vulnerable" && // 14
		clue.Type == ClueTypeRank &&
		(clue.Value == 2 || clue.Value == 5) {

		s.Warningf("You cannot give a number 2 or number 5 clue to a %v character.", p2.Character)
		return true
	} else if p2.Character == "Color-Blind" && // 15
		clue.Type == ClueTypeColor {

		s.Warningf("You cannot give that color clue to a %v character.", p2.Character)
		return true
	}

	return false
}

// characterCheckPlay returns true if the card cannot be played
func characterCheckPlay(s *Session, d *CommandData, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	if p.Character == "Hesitant" && // 19
		p.GetCardSlot(d.Target) == 1 {

		s.Warningf("You cannot play that card since you are a %v character.", p.Character)
		return true
	}

	return false
}

// characterCheckMisplay returns true if the card should misplay
func characterCheckMisplay(g *Game, p *GamePlayer, c *Card) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	if p.Character == "Follower" { // 31
		// Look through the stacks to see if two cards of this rank have already been played
		numPlayedOfThisRank := 0
		for _, s := range g.Stacks {
			if s >= c.Rank {
				numPlayedOfThisRank++
			}
		}
		if numPlayedOfThisRank < 2 {
			return true
		}
	}

	return false
}

// characterCheckDiscard returns true if the player cannot currently discard
func characterCheckDiscard(s *Session, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	// Local variables
	variant := variants[g.Options.VariantName]

	if p.Character == "Anxious" && // 21
		g.ClueTokens%2 == 0 { // Even amount of clues

		s.Warningf(
			"You are %v, so you cannot discard when there is an even number of clues available.",
			p.Character,
		)
		return true
	} else if p.Character == "Traumatized" && // 22
		g.ClueTokens%2 == 1 { // Odd amount of clues

		s.Warningf(
			"You are %v, so you cannot discard when there is an odd number of clues available.",
			p.Character,
		)
		return true
	} else if p.Character == "Wasteful" && // 23
		g.ClueTokens >= variant.GetAdjustedClueTokens(2) {

		s.Warningf(
			"You are %v, so you cannot discard if there are 2 or more clues available.",
			p.Character,
		)
		return true
	}

	return false
}

func characterPostClue(d *CommandData, g *Game, p *GamePlayer) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	clue := NewClue(d)        // Convert the incoming data to a clue object
	p2 := g.Players[d.Target] // Get the target of the clue

	if p.Character == "Mood Swings" { // 12
		p.CharacterMetadata = clue.Type
	} else if p.Character == "Insistent" { // 13
		// Don't do anything if they are already in their "Insistent" state
		if p.CharacterMetadata == -1 {
			// Mark that the cards that they clued must be continue to be clued
			cardsTouched := p2.FindCardsTouchedByClue(clue)
			for _, order := range cardsTouched {
				c := g.Deck[order]
				c.InsistentTouched = true
			}
			p.CharacterMetadata = 0 // 0 means that the "Insistent" state is activated
		}
	}

	if p2.Character == "Vindictive" { // 9
		// Store that they have had at least one clue given to them on this go-around of the table
		p2.CharacterMetadata = 0
	} else if p2.Character == "Impulsive" && // 17
		p2.IsFirstCardTouchedByClue(clue) {

		// Store that they had their slot 1 card clued
		p2.CharacterMetadata = 0
	}
}

func characterPostRemoveCard(g *Game, p *GamePlayer, c *Card) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	if !c.InsistentTouched {
		return
	}

	for _, c2 := range p.Hand {
		c2.InsistentTouched = false
	}

	// Find the "Insistent" player and reset their state so that
	// they are not forced to give a clue on their subsequent turn
	for _, p2 := range g.Players {
		if p2.Character == "Insistent" { // 13
			p2.CharacterMetadata = -1
			break // Only one player should be Insistent
		}
	}
}

func characterPostAction(d *CommandData, g *Game, p *GamePlayer) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	// Clear the counter for characters that have abilities relating to
	// a single go-around of the table
	if p.Character == "Vindictive" { // 9
		p.CharacterMetadata = -1
	} else if p.Character == "Impulsive" { // 17
		p.CharacterMetadata = -1
	} else if p.Character == "Indolent" { // 18
		if d.Type == constants.ActionTypePlay {
			p.CharacterMetadata = 0
		} else {
			p.CharacterMetadata = -1
		}
	} else if p.Character == "Contrarian" { // 27
		g.TurnsInverted = !g.TurnsInverted
	}

	// Store the last action that was performed
	for _, p2 := range g.Players {
		if p2.Character == "Stubborn" { // 28
			p2.CharacterMetadata = d.Type
		}
	}
}

func characterNeedsToTakeSecondTurn(d *CommandData, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	// Local variables
	variant := variants[g.Options.VariantName]

	if p.Character == "Genius" { // 24
		// Must clue both a color and a number (uses 2 clues)
		// The clue target is stored in "p.CharacterMetadata"
		if d.Type == constants.ActionTypeColorClue {
			p.CharacterMetadata = d.Target
			return true
		} else if d.Type == constants.ActionTypeRankClue {
			p.CharacterMetadata = -1
			return false
		}
	} else if p.Character == "Panicky" && // 26
		d.Type == ActionTypeDiscard {

		// After discarding, discards again if there are 4 clues or less
		// "p.CharacterMetadata" represents the state, which alternates between -1 and 0
		if p.CharacterMetadata == -1 && g.ClueTokens <= variant.GetAdjustedClueTokens(4) {
			p.CharacterMetadata = 0
			return true
		} else if p.CharacterMetadata == 0 {
			p.CharacterMetadata = -1
			return false
		}
	}

	return false
}

func characterHideCard(a *ActionDraw, g *Game, p *GamePlayer) bool {
	if !g.Options.DetrimentalCharacters {
		return false
	}

	if p.Character == "Blind Spot" && a.PlayerIndex == p.GetNextPlayer() { // 29
		return true
	} else if p.Character == "Oblivious" && a.PlayerIndex == p.GetPreviousPlayer() { // 30
		return true
	} else if p.Character == "Slow-Witted" { // 33
		return true
	}

	return false
}

func characterSendCardIdentityOfSlot2(g *Game, playerIndexDrawingCard int) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	// Local variables
	t := g.Table
	p := g.Players[playerIndexDrawingCard]

	if len(p.Hand) <= 1 {
		return
	}

	hasSlowWitted := false
	for _, p2 := range g.Players {
		if p2.Character == "Slow-Witted" { // 33
			hasSlowWitted = true
			break
		}
	}

	if hasSlowWitted {
		// Card information will be scrubbed from the action in the "CheckScrub()" function
		c := p.Hand[len(p.Hand)-2] // Slot 2
		g.Actions = append(g.Actions, ActionCardIdentity{
			Type:        "cardIdentity",
			PlayerIndex: p.Index,
			Order:       c.Order,
			SuitIndex:   c.SuitIndex,
			Rank:        c.Rank,
		})
		t.NotifyGameAction()
	}
}

func characterAdjustEndTurn(g *Game) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	// Check to see if anyone is playing as a character that will adjust
	// the final go-around of the table
	for _, p := range g.Players {
		if p.Character == "Contrarian" { // 27
			// 3 instead of 2 because it should be 2 turns after the final card is drawn
			g.EndTurn = g.Turn + 3
		}
	}
}

func characterHasTakenLastTurn(g *Game) bool {
	if g.EndTurn == -1 {
		return false
	}
	originalPlayer := g.ActivePlayerIndex
	activePlayer := g.ActivePlayerIndex
	turnsInverted := g.TurnsInverted
	for turn := g.Turn + 1; turn <= g.EndTurn; turn++ {
		if turnsInverted {
			activePlayer += len(g.Players)
			activePlayer = (activePlayer - 1) % len(g.Players)
		} else {
			activePlayer = (activePlayer + 1) % len(g.Players)
		}
		if activePlayer == originalPlayer {
			return false
		}
		if g.Players[activePlayer].Character == "Contrarian" { // 27
			turnsInverted = !turnsInverted
		}
	}
	return true
}

func characterCheckSoftlock(g *Game, p *GamePlayer) {
	if !g.Options.DetrimentalCharacters {
		return
	}

	// Local variables
	variant := variants[g.Options.VariantName]

	if g.ClueTokens < variant.GetAdjustedClueTokens(1) &&
		p.CharacterMetadata == 0 && // The character's "special ability" is currently enabled
		(p.Character == "Vindictive" || // 9
			p.Character == "Insistent") { // 13

		g.EndCondition = EndConditionCharacterSoftlock
		g.EndPlayer = p.Index
	}
}

func characterSeesCard(g *Game, p *GamePlayer, p2 *GamePlayer, cardOrder int) bool {
	if !g.Options.DetrimentalCharacters {
		return true
	}

	if p.Character == "Blind Spot" && p2.Index == p.GetNextPlayer() { // 29
		// Cannot see the cards of the next player
		return false
	}

	if p.Character == "Oblivious" && p2.Index == p.GetPreviousPlayer() { // 30
		// Cannot see the cards of the previous player
		return false
	}

	if p.Character == "Slow-Witted" && p2.GetCardSlot(cardOrder) == 1 { // 33
		// Cannot see cards in slot 1
		return false
	}

	return true
}
*/
