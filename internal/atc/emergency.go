package atc

import (
	"fmt"
	"strings"
	"time"

	"github.com/skycontrol/skycontrol/internal/radio"
)

func (t *Tower) handleEmergencyCall(call radio.ReceivedCall) bool {
	text := strings.ToLower(strings.TrimSpace(call.Transcript))
	if text == "" {
		return false
	}
	if containsAny(text, "say again", "see again", "last transmission") {
		return false
	}

	t.mu.Lock()
	st := t.primaryLocked()
	if st == nil {
		st = t.findByPilot(call.Pilot)
	}
	cancel := containsAny(text, "cancel emergency", "cancelling emergency",
		"emergency cancelled", "able to continue")
	declaring := isEmergencyCall(text)
	follow := st != nil && st.Emergency && (isEmergencyFollowUp(text) || cancel)
	if !declaring && !follow {
		t.mu.Unlock()
		return false
	}

	t.seedOwnerLocked(st)
	af := ownerOrNearest(st)
	pilot := "Aircraft"
	if st != nil && st.Callsign != "" {
		pilot = st.Callsign
	}
	role := RoleTower
	if st != nil && st.OnGround {
		role = t.airRole(af, true)
	}
	cs := t.cfgCallsign(af, role)
	runway := t.activeSpoken(af, st)
	wind := t.windPhrase()

	kind := emergencyKind(text)
	souls := parseSouls(text)
	fuel := parseFuel(text)
	first := st == nil || !st.Emergency

	if st != nil {
		if cancel {
			st.Emergency = false
			st.EmergKind = ""
			st.Pattern = ""
			st.LastClearance = time.Now()
		} else {
			st.Emergency = true
			if first {
				st.EmergAt = time.Now()
			}
			if kind != "" {
				st.EmergKind = kind
			}
			if st.EmergKind == "" {
				st.EmergKind = "general"
			}
			if souls > 0 {
				st.Souls = souls
			}
			if fuel != "" {
				st.FuelState = fuel
			}
			st.ClearedLand = true
			st.Pattern = "emergency"
			st.LastClearance = time.Now()
		}
	}

	var msg string
	if cancel {
		msg = fmt.Sprintf("%s, %s, roger, emergency cancelled, resume normal.", pilot, cs)
	} else {
		msg = emergencyReply(st, first, pilot, cs, runway, wind)
	}
	t.lastIntent = Intent("emergency")
	t.lastIntentAt = time.Now()
	t.mu.Unlock()

	t.log.Info("emergency", "first", first, "cancel", cancel, "text", msg)
	fmt.Printf("  pattern: emergency\n")
	if af != nil {
		t.sayAs(af, role, msg)
	} else {
		t.say(call.Frequency, cs, msg)
	}
	return true
}

func isEmergencyCall(t string) bool {
	return containsAny(t,
		"mayday", "pan pan", "pan-pan", "declaring an emergency", "declaring emergency",
		"emergency", "low fuel", "minimum fuel", "bingo",
		"engine out", "flameout", "bird strike", "birdstrike",
		"hydraulic", "eject")
}

func isEmergencyFollowUp(t string) bool {
	return containsAny(t, "soul", "souls", "solo", "on board", "onboard", "just me",
		"fuel", "pounds", "bingo")
}

func emergencyKind(t string) string {
	switch {
	case containsAny(t, "low fuel", "minimum fuel", "bingo"):
		return "fuel"
	case containsAny(t, "engine out", "flameout", "engine failure"):
		return "engine"
	case containsAny(t, "bird strike", "birdstrike"):
		return "bird"
	case containsAny(t, "hydraulic"):
		return "hydraulic"
	case containsAny(t, "mayday"):
		return "mayday"
	case containsAny(t, "pan pan", "pan-pan"):
		return "pan"
	case containsAny(t, "emergency"):
		return "general"
	}
	return ""
}

func parseSouls(t string) int {
	switch {
	case containsAny(t, "two souls", "souls two", "souls 2", "two on board"):
		return 2
	case containsAny(t, "one soul", "souls one", "souls 1", "1 soul", "solo",
		"just me", "single seat", "one on board", "one onboard"):
		return 1
	}
	return 0
}

func parseFuel(t string) string {
	switch {
	case containsAny(t, "bingo"):
		return "bingo"
	case containsAny(t, "low fuel", "minimum fuel"):
		return "low"
	case containsAny(t, "fuel remaining", "pounds of fuel"):
		return "reported"
	}
	return ""
}

func emergencyReply(st *AircraftState, first bool, pilot, station, runway, wind string) string {
	kind, souls, fuel, onGround := "general", 0, "", false
	if st != nil {
		if st.EmergKind != "" {
			kind = st.EmergKind
		}
		souls, fuel, onGround = st.Souls, st.FuelState, st.OnGround
	}
	nature := "emergency"
	switch kind {
	case "fuel":
		nature = "low fuel"
	case "engine":
		nature = "engine failure"
	case "bird":
		nature = "bird strike"
	case "hydraulic":
		nature = "hydraulic emergency"
	case "mayday":
		nature = "mayday"
	case "pan":
		nature = "pan-pan"
	}

	if onGround && first {
		return fmt.Sprintf("%s, %s, roger %s, hold position, equipment rolling.", pilot, station, nature)
	}
	if !first {
		if souls == 1 {
			return fmt.Sprintf("%s, %s, roger, one soul, equipment standing by.", pilot, station)
		}
		if souls > 1 {
			return fmt.Sprintf("%s, %s, roger, %d souls, equipment standing by.", pilot, station, souls)
		}
		if fuel != "" {
			return fmt.Sprintf("%s, %s, roger, fuel state copied, equipment standing by.", pilot, station)
		}
	}

	msg := fmt.Sprintf("%s, %s, roger %s, %s, cleared to land, full stop.", pilot, station, nature, runway)
	if wind != "" && wind != "wind calm" {
		msg += " " + strings.TrimSuffix(wind, ".") + "."
	}
	if kind != "pan" {
		msg += " Equipment standing by."
	}
	if souls == 0 {
		msg += " Say souls on board."
	}
	if fuel == "" && kind != "fuel" {
		msg += " Say fuel remaining."
	}
	return msg
}
