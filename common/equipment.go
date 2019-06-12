package common

// EqNameToPrice a map to lookup weapon prices
var EqNameToPrice map[string]int

func init() {
	EqNameToPrice = make(map[string]int)
	EqNameToPrice["AK-47"] = 2700
	EqNameToPrice["AUG"] = 3300
	EqNameToPrice["AWP"] = 4750
	EqNameToPrice["PP-Bizon"] = 1400
	EqNameToPrice["Desert Eagle"] = 700
	EqNameToPrice["Dual Barettas"] = 400
	EqNameToPrice["FAMAS"] = 2250
	EqNameToPrice["Five-SeveN"] = 500
	EqNameToPrice["G3SG1"] = 5000
	EqNameToPrice["Galil AR"] = 2000
	EqNameToPrice["Glock-18"] = 200
	EqNameToPrice["P2000"] = 200
	EqNameToPrice["M249"] = 5200
	EqNameToPrice["M4A1"] = 3100
	EqNameToPrice["MAC-10"] = 1050
	EqNameToPrice["MAG-7"] = 1300
	EqNameToPrice["MP7"] = 1500
	EqNameToPrice["MP9"] = 1250
	EqNameToPrice["Negev"] = 1700
	EqNameToPrice["Nova"] = 1050
	EqNameToPrice["p250"] = 300
	EqNameToPrice["P90"] = 2350
	EqNameToPrice["Sawed-Off"] = 1100
	EqNameToPrice["SCAR-20"] = 5000
	EqNameToPrice["SG 553"] = 2750
	EqNameToPrice["SSG 08"] = 1700
	EqNameToPrice["Tec-9"] = 500
	EqNameToPrice["UMP-45"] = 1200
	EqNameToPrice["xm1014"] = 2000
	EqNameToPrice["CZ75 Auto"] = 500
	EqNameToPrice["USP-S"] = 200
	EqNameToPrice["R8 Revolver"] = 600
}
