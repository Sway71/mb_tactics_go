package main

import "math"

var weaponMinRange = map[string]int{
	"shield":		0,
	"knife":		1,
	"sword":		1,
	"short_bow":	2,
}

var weaponMaxRange = map[string]int{
	"shield":		0,
	"knife":		1,
	"sword":		1,
	"short_bow":	4,
}

var weaponDamage = map[string]func(character Character, weaponStrength int) int {
	"knife":		func(character Character, weaponStrength int) int {
		return ((character.Strength + character.Speed) / 2 ) * weaponStrength
	},
	"sword":		func(character Character, weaponStrength int) int {
		return character.Strength * weaponStrength
	},
	"short_bow":	func(character Character, weaponStrength int) int {
		return character.Strength * weaponStrength
	},
}

func checkAttackRange(attackMinRange int, attackMaxRange int, attackerLocation Location, victimLocation Location) bool {
	graphRange := int(math.Abs(float64(attackerLocation.X - victimLocation.X))) + int(math.Abs(float64(attackerLocation.Y - victimLocation.Y)))
	if graphRange >= attackMinRange && graphRange <= attackMaxRange {
		return true
	}
	return false
}

// TODO: need to make sure that the attack function can be handled without having the weapon id passed
// as we need to just make sure it is simply the weapon currently equipped by the character
// TODO: implement method for keeping track of equipped weapons...
func calcWeaponDamage(weapon Weapon, character Character, attackerLocation Location, victimLocation Location) (int, string) {
	if checkAttackRange(
		weaponMinRange[weapon.Type],
		weaponMaxRange[weapon.Type],
		attackerLocation,
		victimLocation,
	) {
		// TODO: check to hit here as well maybe? Should have thought this through better...
		return weaponDamage[weapon.Type](character, weapon.WeaponStrength), "success"
	}
	return 0, "invalid"
}