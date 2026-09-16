package atc

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/skycontrol/skycontrol/internal/airfield"
	"github.com/skycontrol/skycontrol/internal/radio"
	"github.com/skycontrol/skycontrol/internal/weather"
)

const atisRangeNM = 40.0

type ttsFile interface {
	SayTempFile(text string) (path string, err error)
}

// ATIS is a second SRS radio: one frequency, cached WAV, loop.
// Never stacked onto tower TX.
type ATIS struct {
	log     *slog.Logger
	tower   *Tower
	radio   radio.Client
	speaker ttsFile

	letter  byte
	lastKey string
	wavPath string
	hold    time.Duration
	freq    radio.Frequency
	name    string
	text    string
}

func NewATIS(log *slog.Logger, tower *Tower, rad radio.Client, speaker ttsFile) *ATIS {
	if log == nil {
		log = slog.Default()
	}
	return &ATIS{log: log, tower: tower, radio: rad, speaker: speaker, letter: 'A'}
}

func (a *ATIS) Run(ctx context.Context) {
	if a == nil || a.radio == nil || a.tower == nil {
		return
	}
	fmt.Println("  ATIS transmitter: on (nearest field, COM2)")
	for {
		if ctx.Err() != nil {
			return
		}
		wait := a.once()
		if wait < 2*time.Second {
			wait = 2 * time.Second
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (a *ATIS) Status() string {
	if a == nil || a.name == "" || a.freq.Hz < 1 {
		return "ATIS off — sit in a jet within 40 nm of a field"
	}
	return fmt.Sprintf("%s on %.3f (tune COM2)", a.name, a.freq.Hz/1_000_000)
}

func (a *ATIS) once() time.Duration {
	st := a.tower.PrimaryAircraft()
	if st == nil || st.Nearest == nil || st.DistanceNM > atisRangeNM {
		if a.name != "" {
			fmt.Println("  ATIS off (out of range)")
			a.name, a.text, a.wavPath = "", "", ""
			a.freq = radio.Frequency{}
			a.hold = 0
			a.lastKey = ""
		}
		return 2 * time.Second
	}
	af := st.Nearest
	atisStr := af.PrimaryATISFreq()
	if atisStr == "" {
		return 2 * time.Second
	}
	freq, err := radio.ParseFrequency(atisStr)
	if err != nil || freq.Hz < 1_000_000 {
		return 2 * time.Second
	}
	if freq.Modulation == "" {
		freq.Modulation = "AM"
	}
	runway := a.tower.activeName(af, st.Heading)
	w := a.tower.windSample()
	body := atisBody(af, runway, w)
	key := af.ID + "|" + body
	if key != a.lastKey {
		if a.lastKey != "" {
			a.letter++
			if a.letter > 'Z' {
				a.letter = 'A'
			}
		}
		a.lastKey = key
		info := natoLetter(a.letter)
		text := fmt.Sprintf("%s information %s. %s Advise on initial contact you have information %s.",
			af.Name, info, body, info)
		hold, path, err := a.render(text)
		if err != nil {
			a.log.Warn("ATIS render failed", "error", err, "field", af.Name)
			fmt.Printf("  ATIS render failed: %v\n", err)
			a.lastKey = ""
			return 5 * time.Second
		}
		a.wavPath = path
		a.hold = hold
		a.freq = freq
		a.name = af.Name + " Information"
		a.text = text
		mhz := freq.Hz / 1_000_000
		fmt.Printf("  ATIS: %s on %.3f  (tune COM2)\n", a.name, mhz)
		fmt.Printf("  ATIS tape: %s\n", text)
		a.log.Info("ATIS clip ready", "field", af.Name, "freq_mhz", mhz, "info", info, "sec", hold.Seconds())
	}
	if a.wavPath == "" {
		return 2 * time.Second
	}
	a.radio.Transmit(radio.Transmission{
		Callsign:   a.name,
		Text:       a.text,
		Spoken:     a.text,
		Frequency:  a.freq,
		ExtraFreqs: []radio.Frequency{a.freq},
		AudioFile:  a.wavPath,
		KeepFile:   true,
		Priority:   -1,
	})
	wait := a.hold + 2500*time.Millisecond
	if wait < 5*time.Second {
		wait = 5 * time.Second
	}
	return wait
}

func (a *ATIS) render(text string) (time.Duration, string, error) {
	if a.speaker == nil {
		return 0, "", fmt.Errorf("no speaker")
	}
	wav, err := a.speaker.SayTempFile(text)
	if err != nil || wav == "" {
		if err == nil {
			err = fmt.Errorf("empty wav")
		}
		return 0, "", err
	}
	dst := "atis-loop.wav"
	hold, err := radio.PrepareLoopWav(wav, dst, 0.8)
	_ = os.Remove(wav)
	if err != nil {
		return 0, "", err
	}
	return hold, dst, nil
}

func atisBody(af *airfield.Airfield, runway string, w weather.Sample) string {
	var b strings.Builder
	if w.OK && w.SpeedKt >= 1 {
		b.WriteString("Wind ")
		b.WriteString(speakHeading(w.FromDeg))
		b.WriteString(" at ")
		b.WriteString(speakKnots(int(w.SpeedKt + 0.5)))
		b.WriteString(". ")
	} else {
		b.WriteString("Wind calm. ")
	}
	if runway != "" {
		b.WriteString("Landing and departing ")
		b.WriteString(SpeakRunway(runway))
		b.WriteString(". ")
	}
	if af.TACAN != "" {
		b.WriteString("TACAN ")
		b.WriteString(SpeakTACAN(af.TACAN))
		b.WriteString(". ")
	}
	return b.String()
}

func natoLetter(c byte) string {
	if c >= 'A' && c <= 'Z' {
		if w, ok := nato[rune(c)]; ok {
			return w
		}
	}
	return "alpha"
}

func speakHeading(deg float64) string {
	for deg < 0 {
		deg += 360
	}
	n := int(deg+0.5) % 360
	s := fmt.Sprintf("%03d", n)
	var parts []string
	for _, r := range s {
		if w, ok := digits[r]; ok {
			parts = append(parts, w)
		}
	}
	return strings.Join(parts, " ")
}

func speakKnots(n int) string {
	if n < 0 {
		n = 0
	}
	s := fmt.Sprintf("%d", n)
	var parts []string
	for _, r := range s {
		if w, ok := digits[r]; ok {
			parts = append(parts, w)
		}
	}
	if n == 1 {
		return strings.Join(parts, " ") + " knot"
	}
	return strings.Join(parts, " ") + " knots"
}
