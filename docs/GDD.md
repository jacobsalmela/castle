# Game Design Document

## The Chronovian Thanatome

### Inspirations

- Super Metroid
- Castlevania
- Rebel Transmute
- Axiom Verge
- Cube
- The Antikythera Mechanism

## Lore

> “It does not tell time. It tells where time will choose to exist next.”
>   — Fragment 11, Codex Vornhelm

Long before the fall of the Luminarch Cartographers, there existed a city that refused to remain still. Its towers drifted through dimensions like wandering constellations; its streets would rearrange themselves between breaths. Scholars called it the Hall of Aevum, though its true name was lost each dawn. Inside it were said to lie archives of every age — pasts not yet written, futures already forgotten.

To find it, one needed the Chronovian Thanatome — a relic bound in living metal and inscribed with temporal bone. The device was neither clock nor book, but both: an orrery whose hands etched shifting runes across translucent plates. Those who learned to read it could trace the trajectories of the Hall’s rooms, predicting when a given chamber would align with the waking world.

But the Thanatome demanded more than intellect. Each reading required a tithe of memory — a moment of the observer’s own life, erased and fed into the mechanism. Over time, its keepers became hollow historians, their minds filled with places they had never been, and faces that no longer remembered them.

In the current age, only fragments remain. The last known copy — or perhaps the original itself — was seen pulsing beneath the ruins of Iskareth, its plates turning without a hand to move them. Rumor claims it is once again active, mapping the reappearance of the Hall of Aevum as its rooms blink into new realities.

Those who seek it say the Thanatome’s light no longer measures time.
It measures you — how long until you, too, are part of its history.

⸻

Would you like me to follow this up with a visual concept description (how the device looks, materials, how it animates when used) or a gameplay concept (how a player interacts with it to reveal the moving rooms)?

### Objective

- Detailed envionments
- Super-metroid style map
- Shifting rooms?

### Themes

- Time
- Ancient Relics
- Traps
- Science Fiction
- Magic

### Possible areas

- Aqueducts
- Metal factory
- Ruins
- Prototype lab
- Jungle


### Loose ideas

- There are humans that managed to free themselves of the first sin, though a rite, which grant them immortality, these humans became gods-like in the kingdom.
- The king (Arawn) of the castle is one of these "gods". In his governance some major event happened where the wisdom of becoming immortal leaked, making everyone in the castle immortal. That's when an outsider god placed a barrier on the castle and let all it's inhabitants prisoned.
- There was a person who didn't use the wisdom to became immortal, which let him die and later became like a saint and prophet for one of the covenants which named him "the first death". This covenant seeks a way to undo the rite somehow and become sinners again. Maybe as a plot twist, this prophet did become immortal and he just faked his dead and appears in the game?

#### Area Enemies

### Objectives

- Find all seven pieces of the Chronovian Thanatome to unlock the key of finding out what rooms will be when
  - Each component unlocks a new ability

### Main game mechanics

### Mechanics introduction

#### Stamina

The player will have a stamina bar which:

- Depletes with every attack.
- Recovers by idling.
- Recover rate is slower when guarding.
- If stamina reaches less than zero, the player staggers for a moment making him vulnerable.

#### Poise

The player will have a poise stat which depletes with every taken hit:

- Independent of guarding, every taken hit will deplete the poise bar.
- Fully recovers on a timer with fixed time.
- If the poise stat reaches zero, the player's animation (can be attacking) is interrupted and staggers for a moment.
- Better armor have bigger poise bars.

#### Guard

The player can guard with their shield anytime they want:

- Player moves slower when guarding.
- Negates damage taken.
- Every hit taken while guarding will deplete the stamina bar.
- Better shield drain less stamina per hit.

#### Checkpoints

To add dread and difficulty, the player can only save at designated points (like Bonfires):

- If the player dies, they loses their currency, he must go back for them to retrieve it.

#### Misc

- Tiled base maps, with slopes and ladders.
- Player can jump, and double-jump.  
- There will be a jumping attack.
- 2 type of attack (for now), light attack (short windup) and heavy attack (long windup), by holding and charging the attack button.
- Enemy death is permanent.
- Enemies plot and plan, build reinforcements, communicate with each other resulting in a unique game every time

#### Mechanics imposed limitations

- Lots of leveling up.
- Perks
- Bosses are not brutal, just fun to beat

## Enemies

### First area

#### Crawler

The first enemy, an insignificant, slow and weak obstacle, they can occasionally attack.
A good introduction for combat.
Lore:

- There are poor lost souls that are too tired and spent from living.

#### Ghoul

These are well-rounded enemies that can throw rocks from higher places. They have a 2 attack combo that can trick you, but they have weak poise.
Lore:

- IDK

#### Skeleman

These ones are the first real challenge a player will face. They have mid poise and can spin their swords to make a wall of hitboxes. They will force you to use the shield.
Lore:

- IDK

#### Abomination

These are health walls (not too much). But can jump and fall on you. They hit hard but are slow and well telegraphed.
Lore:

- They come from the mass of bodies under the dungeon. Where people believe that pulping and mutilation oneself can lead to a state of death. These are the ones that decide to leave the mass.

## TODOs and unsorted ideas

- Opening scene is the player falling into the chasm
  - When loading is complete, the title of the game is shown as rocks that the player dives through and smashes to pieces
  - Falling is the mini-game to play while any loading happens--no slow fading intro scenes: must get to the gameplay as fast as possible
- Player levels up skills automatically:
  - If they jump a lot, they can slowly jump higher
  - If they always block, they can defend against more attacks
  - No menus, just gradual progression
- No slow release text dialogues--all text is instant and dismissiable (but can be viewed again any time)
- No cutscenes--story is told through experiences of the player
- Ruins for save points
- https://www.gamedeveloper.com/design/the-fundamental-pillars-of-a-combat-system

Level progression:

- Hints are exposed as player presses buttongs and tries things--teach them when they fail
- Level design steps [How to Design Great Metroidvania Levels | Game Design](https://www.youtube.com/watch?v=bAHXYfP38CA)
