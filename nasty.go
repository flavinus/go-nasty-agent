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

// SAMPLE

type Sample struct {
	c int
	a int
}

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

	return &w
}

type Waveform struct {
	Description string `json:"description"`
	ChanClient  []int  `json:"channel0"`
	ChanAgent   []int  `json:"channel1"`
}

func (w *Waveform) ToChannel(output chan Sample) {
	samples := int(math.Min(float64(len(w.ChanClient)), float64(len(w.ChanAgent))))
	go func() {
		for i := 0; i < samples; i++ {
			output <- Sample{c: w.ChanClient[i], a: w.ChanAgent[i]}
		}
		output <- Sample{c: -1, a: -1} // really usefull ? i'm not sure it still works...

		close(output)
	}()
}

// ANALYSER

type Analyser struct {
	score      int
	samples    int
	discussion string
	debug      bool
}

func (a *Analyser) raiseScore(amount int, message string) {
	a.score = a.score + amount
	fmt.Printf("\n - Raise score: +%d : %s", amount, message)
}

func (a *Analyser) run(input chan Sample) bool {

	fmt.Printf("\n<< ANALYSE")

	a.score = 0
	a.samples = 0
	a.discussion = ""

	psc, psa := -1, -1 // previous state
	dsc, dsa := 0, 0   // beginning of state

	for sample := range input {
		//fmt.Printf("\n<< receive : %+v ", sample)
		sc := getChanStatus(sample.c)
		sa := getChanStatus(sample.a)
		d := getDiscussionStatus(sc, sa)

		psc, dsc = a.watchStatus("Client", a.samples, sc, psc, dsc)
		psa, dsa = a.watchStatus("Agent", a.samples, sa, psa, dsa)
		a.discussion = a.discussion + d

		a.samples = a.samples + 1
	}

	return a.results()

}

func (a *Analyser) watchStatus(who string, i, sc, psc, dsc int) (int, int) {

	if sc != psc { // state changed
		count := i - dsc

		//fmt.Printf("\n<< state changed %s : status %d  count %d ", who, sc, count)

		if psc == 0 { // mute ended
			a.raiseScore(count*5, fmt.Sprintf("%s remained muted during %.2fs", who, float64(count)*0.1))

			// TODO: BUG: this event is not raised anymore
			if sc == -1 && dsc < 30 { // muted since more than 3s at end of call
				a.raiseScore(1000, fmt.Sprintf("%s was muted at the end of conversation", who))
			}
		}

		psc = sc
		dsc = i
	}
	return psc, dsc
}

func (a *Analyser) results() bool {

	// Eval repartitions ( int % )
	pnobody := float64(strings.Count(a.discussion, "_")) / float64(a.samples) * 100.0
	pclient := float64(strings.Count(a.discussion, "c")) / float64(a.samples) * 100.0
	pagent := float64(strings.Count(a.discussion, "a")) / float64(a.samples) * 100.0
	pboth := float64(strings.Count(a.discussion, "#")) / float64(a.samples) * 100.0

	// Update score according to repartions stats

	// Si il y a eu beaucoup de collisions
	if pboth > 20 {
		a.raiseScore(int(50*pboth), "many collisions during discussion")
	}

	// Si l'agent à très peu parlé
	if pagent < 10 {
		a.raiseScore(1000, "agent didn't speak very much")
	}

	// Eval overall result and display
	ratio := float64(a.score) / float64(a.samples)

	if a.debug {
		fmt.Printf("\nDISCUSSION: %s", a.discussion)
	}
	fmt.Printf("\nREPARTITION: silent=%.2f%%  client=%.2f%%  agent=%.2f%%  both=%.2f%%", pnobody, pclient, pagent, pboth)
	fmt.Printf("\nRESULT     : samples=%d  score=%d  => [ RATIO = %.2f ]", a.samples, a.score, ratio)
	fmt.Printf("\n>>\n")

	return ratio < 2
}

// OTHER FUNCS

func getChanStatus(volume int) int {
	switch {
	case volume == 0:
		return 0
	case volume < 0:
		return -1
	case volume < 10:
		return 1
	default:
		return 2
	}
}

func getDiscussionStatus(sc, sa int) string {
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

func analyseFile(filename string) bool {

	w := NewWaveform(filename)

	schan := make(chan Sample)
	w.ToChannel(schan)

	a := Analyser{debug: false}
	return a.run(schan)
}

func main() {

	//analyseFile("samples/data_ok.json")
	//analyseFile("samples/data_muted_agent_long.json")

	files, _ := filepath.Glob("samples/*.json")
	for _, file := range files {
		analyseFile(file)
	}

}
