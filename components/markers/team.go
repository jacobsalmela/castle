package markers

// TeamType represents the affiliation of an entity for collision, targeting, and interaction.
type TeamType string

const (
	// TeamPlayer indicates player-controlled entities.
	TeamPlayer TeamType = "player"

	// TeamEnemy indicates hostile entities that target the player.
	TeamEnemy TeamType = "enemy"

	// TeamNeutral indicates non-hostile entities (NPCs, interactives, projectiles).
	TeamNeutral TeamType = "neutral"
)

// Team component identifies which team an entity belongs to.
type Team struct {
	Type TeamType

	// Owner optionally tracks which entity spawned this entity (useful for projectiles).
	// Not used for team membership logic, but available for damage attribution.
	Owner int // EntityId if needed, or 0 for no owner
}
