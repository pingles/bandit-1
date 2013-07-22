package bandit

import (
	"math"
	"math/rand"
	"time"
)

// Arm simulates a single bandit arm pull with every execution. Returns {0,1}.
type Arm func() float64

// Bernoulli returns an Arm function such that a ~ Bern(x|μ)
func Bernoulli(μ float64) Arm {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() float64 {
		res := 0.0
		if r.Float64() <= μ {
			res = 1.0
		}

		return res
	}
}

// BanditNew is a curried constructor function.
type BanditNew func() (Bandit, error)

// MonteCarlo runs a monte carlo experiment with the given bandit and arms.
func MonteCarlo(sims, trials int, bandit BanditNew, arms []Arm) (Sim, error) {
	s := Sim{
		Sims:       sims,
		Trials:     trials,
		Sim:        make([]int, sims*trials),
		Trial:      make([]int, sims*trials),
		Selected:   make([]int, sims*trials),
		Reward:     make([]float64, sims*trials),
		Cumulative: make([]float64, sims*trials),
	}

	for sim := 0; sim < sims; sim++ {
		b, err := bandit()
		if err != nil {
			return Sim{}, err
		}

		for trial := 0; trial < trials; trial++ {
			selected := b.SelectArm()
			reward := arms[selected-1]()
			b.Update(selected, reward)

			// record this trial into column i
			i := sim*trials + trial
			s.Sim[i] = sim + 1
			s.Trial[i] = trial + 1
			s.Selected[i] = selected
			s.Reward[i] = reward
			s.Cumulative[i] = s.Cumulative[int(math.Max(float64(i-1), 0.0))] + reward
		}
	}

	return s, nil
}

// Sim is a matrix of simulation results. Columns represent individual trial
// results that grow to the right with each trial
type Sim struct {
	Sims       int
	Trials     int
	Sim        []int
	Trial      []int
	Selected   []int
	Reward     []float64
	Cumulative []float64
}

// Performance returns an array of average rewards at each trial point.
// Averaged over sims
func Performance(s Sim) []float64 {
	t := make([]float64, s.Trials)
	for trial := 0; trial < s.Trials; trial++ {
		accum, count := 0.0, 0
		for sim := 0; sim < s.Sims; sim++ {
			i := sim*s.Trials + trial
			if s.Trial[i] != trial+1 {
				panic("impossible trial access")
			}

			accum = accum + s.Reward[i]
			count = count + 1
		}

		t[trial] = accum / float64(count)
	}

	return t
}
