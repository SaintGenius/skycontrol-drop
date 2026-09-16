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
