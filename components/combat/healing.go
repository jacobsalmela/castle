package combat

// Healing tracks consumable healing charges for an entity.
// Healing charges are typically consumed by the player to restore health.
// The HealAmount specifies how much health is restored per charge used.
type Healing struct {
	Count      int     // Current number of healing charges
	MaxCount   int     // Maximum number of healing charges
	HealAmount float64 // Health restored per charge used
}
