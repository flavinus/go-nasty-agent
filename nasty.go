package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// WAVE FORM

func NewWaveform(filename string) *Waveform {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	w := Waveform{}
	json.Unmarshal(byteValue, &w)

	w.loadStates()

	return &w
}

type Waveform struct {
	Description string `json:"description"`
	ChanClient  []int  `json:"channel0"`
	ChanAgent   []int  `json:"channel1"`

	samples int

	stateClient []int
	stateAgent  []int
	discussion  string
}

func (w *Waveform) loadStates() {

	w.stateClient = []int{}
	w.stateAgent = []int{}

	w.samples = int(math.Min(float64(len(w.ChanClient)), float64(len(w.ChanAgent))))

	for i := 0; i < w.samples; i++ {
		sc := w.getChanStatus(w.ChanClient[i])
		w.stateClient = append(w.stateClient, sc)

		sa := w.getChanStatus(w.ChanAgent[i])
		w.stateAgent = append(w.stateAgent, sa)

		w.discussion = w.discussion + w.getDiscussionStatus(sc, sa)
	}
}

func (w *Waveform) getChanStatus(volume int) int {
	switch {
	case volume == 0:
		return 0
	case volume < 10:
		return 1
	default:
		return 2
	}
}

func (w *Waveform) getDiscussionStatus(sc, sa int) string {
	if sc < 2 && sa < 2 {
		return "_"
	} else if sc > 1 && sa > 1 {
		return "#"
	} else if sc > 1 {
		return "c"
	} else if sa > 1 {
		return "a"
	}
	return ""
}

// ANALYSER

type Analyser struct {
	score int
	debug bool
}

func (a *Analyser) raiseScore(amount int, message string) {
	a.score = a.score + amount
	fmt.Printf("\n - Raise score: +%d : %s", amount, message)
}

func (a *Analyser) run(w *Waveform) {

	fmt.Printf("\n<< ANALYSE : %s ", w.Description)
	if a.debug {
		fmt.Printf("\n Client : states=%v", w.stateClient)
		fmt.Printf("\n Agent  : states=%v", w.stateAgent)
	}

	a.score = 0

	psc, psa := -1, -1 // previous state
	dsc, dsa := 0, 0   // beginning of state

	// Watch client and agent states changes
	for i := 0; i < w.samples; i++ {
		psc, dsc = a.watchStatus("Client", i, w.stateClient[i], psc, dsc)
		psa, dsa = a.watchStatus("Agent", i, w.stateAgent[i], psa, dsa)
	}

	// We watch status once again at the end
	a.watchStatus("Client", w.samples, -1, psc, dsc)
	a.watchStatus("Agent", w.samples, -1, psa, dsa)

	// Results
	a.results(w)
}

func (a *Analyser) watchStatus(who string, i, sc, psc, dsc int) (int, int) {

	if sc != psc { // state changed
		count := i - dsc
		if psc == 0 { // mute ended
			a.raiseScore(count*5, fmt.Sprintf("%s remained muted during %.2fs", who, float64(count)*0.1))
			if sc == -1 && dsc < 30 { // muted since more than 3s at end of call
				a.raiseScore(1000, fmt.Sprintf("%s was muted at the end of conversation", who))
			}
		}

		psc = sc
		dsc = i
	}
	return psc, dsc
}

func (a *Analyser) results(w *Waveform) {

	// Eval repartitions ( int % )
	pnobody := float64(strings.Count(w.discussion, "_")) / float64(w.samples) * 100.0
	pclient := float64(strings.Count(w.discussion, "c")) / float64(w.samples) * 100.0
	pagent := float64(strings.Count(w.discussion, "a")) / float64(w.samples) * 100.0
	pboth := float64(strings.Count(w.discussion, "#")) / float64(w.samples) * 100.0

	// Update score according to repartions stats

	// Si il y a eu beaucoup de collisions
	if pboth > 20 {
		a.raiseScore(int(30*pboth), "many collisions during discussion")
	}

	// Si l'agent à très peu parlé
	if pagent < 10 {
		a.raiseScore(1000, "agent didn't speak very much")
	}

	ratio := float64(a.score) / float64(w.samples)

	if a.debug {
		fmt.Printf("\nDISCUSSION: %s", w.discussion)
	}
	fmt.Printf("\nREPARTITION: silent=%.2f%%  client=%.2f%%  agent=%.2f%%  both=%.2f%%", pnobody, pclient, pagent, pboth)
	fmt.Printf("\nRESULT     : samples=%d  score=%d  => [ RATIO = %.2f ]", w.samples, a.score, ratio)
	fmt.Printf("\n>>\n")
}

func main() {

	a := Analyser{debug: false}

	//a.run(NewWaveform("samples/data_ok.json"))
	//a.run(NewWaveform("samples/data_muted_agent_long.json"))

	files, _ := filepath.Glob("samples/*.json")
	for _, file := range files {
		a.run(NewWaveform(file))
	}
}
