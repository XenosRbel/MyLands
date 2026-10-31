package rules

import "time"

// EconState is the mutable economic state of one city between events.
type EconState struct {
	Stock      map[string]int64   // milli units
	Population int64
	Garrison   map[string]int64
	Carry      map[string]float64
	UpdatedAt  time.Time // instant up to which the state is valid
}

// AdvanceCity integrates a city's economy from s.UpdatedAt to `to` and
// returns the rates in effect at `to`. in.Population/in.Garrison are taken
// from s.
//
// Population, desertion and overflow decay change only at economy-tick
// boundaries aligned to absolute time (multiples of economyTickSeconds since
// the Unix epoch); resource accrual between boundaries uses constant rates.
// Because boundaries are absolute, the result does not depend on how often a
// city is settled (path independence up to ≤1 milli flooring per settlement).
// Buildings/techs/away units are constant inside the interval: callers split
// the interval at every scheduled event that changes them (GAME_RULES R-TIME-1).
func AdvanceCity(s *EconState, to time.Time, in EconInput) Rates {
	p := in.Pack
	if s.Carry == nil {
		s.Carry = map[string]float64{}
	}
	if s.Garrison == nil {
		s.Garrison = map[string]int64{}
	}
	ratesFn := func() Rates {
		x := in
		x.Population = s.Population
		x.Garrison = s.Garrison
		return ComputeRates(x)
	}
	rates := ratesFn()
	if !to.After(s.UpdatedAt) {
		return rates
	}
	tick := time.Duration(p.Params.EconomyTickSeconds) * time.Second
	t := s.UpdatedAt
	ts := &TickState{Population: s.Population, Garrison: s.Garrison, Stock: s.Stock, Carry: s.Carry}
	for t.Before(to) {
		next := t.Truncate(tick).Add(tick) // next absolute boundary strictly after t
		segEnd := next
		if to.Before(segEnd) {
			segEnd = to
		}
		Accrue(s.Stock, rates, p, segEnd.Sub(t).Milliseconds())
		t = segEnd
		if t.Equal(next) {
			if TickStep(ts, rates, p, in.Mods) {
				s.Population = ts.Population
				rates = ratesFn()
			}
		}
	}
	s.UpdatedAt = to
	return rates
}
