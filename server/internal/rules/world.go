package rules

import (
	"math"

	"realmforge/server/internal/pack"
)

// RNG is a small deterministic generator (SplitMix64). All randomness in the
// game (map generation, scouting detection, item drops) derives from explicit
// seeds so outcomes are reproducible and replay-safe.
type RNG struct{ s uint64 }

func NewRNG(seed uint64) *RNG { return &RNG{s: seed} }

// SeedFrom mixes several integers into one seed.
func SeedFrom(parts ...int64) uint64 {
	var h uint64 = 0x9E3779B97F4A7C15
	for _, p := range parts {
		h ^= uint64(p) + 0x9E3779B97F4A7C15 + (h << 6) + (h >> 2)
		h = mix(h)
	}
	return h
}

func mix(z uint64) uint64 {
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

func (r *RNG) Uint64() uint64 {
	r.s += 0x9E3779B97F4A7C15
	return mix(r.s)
}

// Float returns a uniform value in [0,1).
func (r *RNG) Float() float64 { return float64(r.Uint64()>>11) / (1 << 53) }

// Intn returns a uniform int in [0,n).
func (r *RNG) Intn(n int) int { return int(r.Uint64() % uint64(n)) }

// Distance is the Euclidean tile distance.
func Distance(x1, y1, x2, y2 int) float64 {
	dx, dy := float64(x1-x2), float64(y1-y2)
	return math.Sqrt(dx*dx + dy*dy)
}

// TravelSeconds (GAME_RULES R-MOV-1):
//
//	t = max(ceil(minSeconds/worldSpeed), ceil(distance / (slowestSpeed × (1+speedPct/100)) × 3600 / worldSpeed))
func TravelSeconds(dist, slowestSpeed, speedPct, worldSpeed float64, minSeconds int) int64 {
	v := slowestSpeed * math.Max(0.1, 1+speedPct/100)
	if v <= 0 {
		return math.MaxInt32
	}
	t := int64(math.Ceil(dist / v * 3600 / worldSpeed))
	return max(int64(math.Ceil(float64(minSeconds)/worldSpeed)), t, 1)
}

// Tile is one generated map cell.
type Tile struct {
	X, Y         int
	Terrain      string
	Neutral      string
	NeutralLevel int
}

// GenerateMap deterministically builds the world map from a seed.
// Neutral object level grows with distance from the map centre so that the
// centre (where players spawn first) is safer (PROPOSED, see WORLD design).
func GenerateMap(seed uint64, p *pack.Loaded) []Tile {
	w, h := p.Params.MapWidth, p.Params.MapHeight
	rng := NewRNG(seed)
	totalW := 0
	for _, t := range p.Terrains {
		totalW += t.Weight
	}
	cx, cy := float64(w-1)/2, float64(h-1)/2
	maxD := math.Hypot(cx, cy)
	tiles := make([]Tile, 0, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pick := rng.Intn(totalW)
			var ter *pack.Terrain
			for i := range p.Terrains {
				pick -= p.Terrains[i].Weight
				if pick < 0 {
					ter = &p.Terrains[i]
					break
				}
			}
			t := Tile{X: x, Y: y, Terrain: ter.ID}
			if ter.Passable && len(p.Neutrals) > 0 && rng.Intn(100) < p.Params.NeutralDensityPct {
				n := &p.Neutrals[rng.Intn(len(p.Neutrals))]
				d := math.Hypot(float64(x)-cx, float64(y)-cy) / maxD // 0 centre .. 1 corner
				maxL := len(n.Levels)
				lvl := 1 + int(math.Floor(d*float64(maxL)*0.999))
				if rng.Intn(4) == 0 && lvl > 1 { // some variance
					lvl--
				}
				t.Neutral, t.NeutralLevel = n.ID, lvl
			}
			tiles = append(tiles, t)
		}
	}
	return tiles
}

// XPForLevel returns total XP needed to reach level L (L>=1):
// xp(1)=0, xp(L) = Σ_{k=1}^{L-1} round(base × growth^(k-1)).
func XPForLevel(base int64, growth float64, level int) int64 {
	var total int64
	for k := 1; k < level; k++ {
		total += int64(math.Round(float64(base) * math.Pow(growth, float64(k-1))))
	}
	return total
}

// LevelForXP returns the level reached with xp, capped at maxLevel.
func LevelForXP(base int64, growth float64, maxLevel int, xp int64) int {
	l := 1
	for l < maxLevel && xp >= XPForLevel(base, growth, l+1) {
		l++
	}
	return l
}
