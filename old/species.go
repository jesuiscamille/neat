/*


species.go implementation of species of genomes.

@licstart   The following is the entire license notice for
the Go code in this page.

Copyright (C) 2016 jin yeom, whitewolf.studio

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

As additional permission under GNU GPL version 3 section 7, you
may distribute non-source (e.g., minimized or compacted) forms of
that code without the copy of the GNU GPL normally required by
section 4, provided you include this license notice and a URL
through which recipients can access the Corresponding Source.

@licend    The above is the entire license notice
for the Go code in this page.


*/

package neat

import (
	"math"
	"math/rand"
	"sort"
)

// Species is an implementation of species of genomes in NEAT, which
// are separated by measuring compatibility distance among genomes
// within a population.
type Species struct {
	sid            int       // species ID
	age            int       // species age (in generations)
	prevFitness    float64   // previous average fitness
	representative *Genome   // species representative
	members        []*Genome // genomes in this species
}

// NewSpecies creates a new species given a species ID, and the genome
// that first populates the new species.
func NewSpecies(sid int, g *Genome) *Species {
	return &Species{
		sid:            sid,
		age:            0,
		prevFitness:    0.0,
		representative: g,
		members:        []*Genome{},
	}
}

// SID returns this species' species ID.
func (s *Species) SID() int {
	return s.sid
}

// Age returns this species' age.
func (s *Species) Age() int {
	return s.age
}

// Representative returns this species' representative.
func (s *Species) Representative() *Genome {
	return s.representative
}

// Members returns this species' member genomes.
func (s *Species) Members() []*Genome {
	return s.members
}

// AddGenome adds a new genome to this species.
func (s *Species) AddMember(g *Genome) {
	g.sid = s.sid
	s.members = append(s.members, g)
}

// Select sorts the members by their fitness values and update them based on
// the survival rate; return the remaining members.
func (s *Species) Select() []*Genome {
	sort.Sort(byFitness(s.members))
	survived := int(math.Ceil(float64(len(s.members)) * param.SurvivalRate))
	s.members = s.members[:survived]
	return s.members
}

// Champion returns the genome with the best fitness value in this species.
func (s *Species) Champion() *Genome {
	champion := s.members[0]
	for i := range s.members {
		if toolbox.Comparison(s.members[i], champion) == 1 {
			champion = s.members[i]
		}
	}
	return champion
}

// AvgFitness returns the species' average fitness.
func (s *Species) AvgFitness() float64 {
	fitnessSum := 0.0
	for i := range s.members {
		fitnessSum += s.members[i].fitness
	}
	return fitnessSum / float64(len(s.members))
}

// IsStagnant checks whether a species is stagnant based on comparison between
// previous and current average fitnesses; this function call also updates its
// previous average fitness to the current fitness.
func (s *Species) IsStagnant() bool {
	avgFitness := s.AvgFitness()
	if s.prevFitness < avgFitness {
		s.prevFitness = avgFitness
		return true
	}
	s.prevFitness = avgFitness
	return false
}

// sh implements a part of the explicit fitness sharing function, sh.
// If a compatibility distance 'd' is larger than the compatibility
// threshold 'dt', return 0; otherwise, return 1.
func sh(d float64) float64 {
	if d > param.DistThreshold {
		return 0.0
	}
	return 1.0
}

// FitnessShare computes and assigns the shared fitness of genomes,
// via explicit fitness sharing.
func (s *Species) FitnessShare() {
	adjusted := make(map[int]float64)
	for _, g0 := range s.members {
		adjustment := 0.0
		for _, g1 := range s.members {
			adjustment += sh(g0.Distance(g1))
		}
		if adjustment != 0.0 {
			adjusted[g0.gid] = g0.fitness / adjustment
		}
	}
	for _, member := range s.members {
		member.fitness = adjusted[member.gid]
	}
}

// VarMembers selects n parent genomes and reproduce len(species) - n
// number of children genomes; n is determined by survival rate from
// parameter. Update the members.
func (s *Species) VarMembers() {
	numMembers := len(s.members)
	survived := s.Select()
	numSurvived := len(survived)

	numChildren := numMembers - numSurvived
	for i := 0; i < numChildren; i++ {
		parent0 := survived[rand.Intn(numSurvived)]
		parent1 := survived[rand.Intn(numSurvived)]
		child := Crossover(parent0, parent1, 0)
		survived = append(survived, child)
	}

	s.members = survived
}