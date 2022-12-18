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

type Sample struct {
	client, agent int
}

// WAVE FORM

// todo : error handling
func LoadWaveform(filename string) *Waveform {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	w := Waveform{}
	json.Unmarshal(byteValue, &w)

	return &w
}

type Waveform struct {
	ChanClient []int `json:"channel0"`
	ChanAgent  []int `json:"channel1"`
}

func (w *Waveform) ToChannel() chan Sample {

	output := make(chan Sample)

	count := int(math.Min(float64(len(w.ChanClient)), float64(len(w.ChanAgent))))
	go func() {
		for i := 0; i < count; i++ {
			output <- Sample{client: w.ChanClient[i], agent: w.ChanAgent[i]}
		}
		output <- Sample{client: -1, agent: -1} // really usefull ?

		close(output)
	}()

	return output
}

// ANALYSER
//todo ? use counters in place of discussion string

type Analyser struct {
	score      int
	count      int
	discussion string
}

func (a *Analyser) run(input chan Sample) float32 {

	a.score = 0
	a.count = 0
	a.discussion = ""

	prevClientStatus, prevAgentStatus := int(-1), int(-1)
	beginClientStatus, beginAgentStatus := 0, 0

	for sample := range input {
		clientStatus := getChanStatus(sample.client)
		agentStatus := getChanStatus(sample.agent)
		a.discussion = a.discussion + fmt.Sprint(getDiscussionStatus(clientStatus, agentStatus))

		prevClientStatus, beginClientStatus = a.watchStatus("Client", clientStatus, prevClientStatus, beginClientStatus)
		prevAgentStatus, beginAgentStatus = a.watchStatus("Agent", agentStatus, prevAgentStatus, beginAgentStatus)

		a.count = a.count + 1
	}

	return a.results()

}

func (a *Analyser) watchStatus(who string, status int, previousStatus int, beginStatus int) (int, int) {
	if status != previousStatus { // state changed
		if previousStatus == 0 { // mute ended
			a.eventMuteEnd(who, status, beginStatus)
		}
		return status, a.count
	}
	return previousStatus, beginStatus
}

func (a *Analyser) eventMuteEnd(who string, state int, beginStatus int) {

	count := a.count - beginStatus
	a.raiseScore(count*5, fmt.Sprintf("%s remained muted during %.2fs", who, float32(count)*0.1))

	if state == -1 && beginStatus < 30 { // muted since more than 3s at end of call
		a.raiseScore(1000, fmt.Sprintf("%s was muted at the end of conversation", who))
	}
}

func (a *Analyser) raiseScore(amount int, message string) {
	a.score = a.score + amount
	fmt.Printf("\n - Raise score: +%d : %s", amount, message)
}

func (a *Analyser) getRepartitions() (float32, float32, float32, float32) {

	return float32(strings.Count(a.discussion, "0")) / float32(a.count) * 100.0,
		float32(strings.Count(a.discussion, "1")) / float32(a.count) * 100.0,
		float32(strings.Count(a.discussion, "2")) / float32(a.count) * 100.0,
		float32(strings.Count(a.discussion, "3")) / float32(a.count) * 100.0
}

func (a *Analyser) results() float32 {

	silent, client, agent, collision := a.getRepartitions()

	if collision > 20 {
		a.raiseScore(int(50*collision), "Too many collisions during discussion")
	}

	if agent < 10 {
		a.raiseScore(1000, "Agent didn't speak very much")
	}

	ratio := float32(a.score) / float32(a.count)

	//fmt.Printf("\nd %s     : ", a.discussion)
	fmt.Printf("\nREPARTITION: silent=%.2f%%  client=%.2f%%  agent=%.2f%%  collision=%.2f%%", silent, client, agent, collision)
	fmt.Printf("\nRESULT     : count=%d  score=%d  => [ RATIO = %.2f ]", a.count, a.score, ratio)
	fmt.Printf("\n>>\n")

	return ratio
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

func getDiscussionStatus(clientStatus, agentStatus int) int {
	switch {
	case clientStatus > 1 && agentStatus > 1:
		return 3
	case clientStatus > 1:
		return 1
	case agentStatus > 1:
		return 2
	default:
		return 0
	}
}

func main() {

	analyzer := Analyser{}

	//analyzer.run(LoadWaveform("samples/data_ok.json").ToChannel())
	//analyzer.run(LoadWaveform("samples/data_muted_agent_long.json").ToChannel())

	files, _ := filepath.Glob("samples/*.json")
	for _, file := range files {
		fmt.Printf("\n<< Analyse %s", file)
		analyzer.run(LoadWaveform(file).ToChannel())
	}

}
