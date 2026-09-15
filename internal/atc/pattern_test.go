package atc

import "testing"

func TestDetectPatternLeg(t *testing.T) {
	air := &AircraftState{OnGround: false, Callsign: "Genius 1-1"}
	gnd := &AircraftState{OnGround: true, Callsign: "Genius 1-1"}
	cases := []struct {
		in   string
		st   *AircraftState
		want string
	}{
		{"Switching to Al-Minad on 250.2.", air, ""},
		{"request taxi", gnd, ""},
		{"ready for departure", gnd, ""},
		{"Genius 1-1 inbound", air, legInitial},
		{"Aminad, Tower, Genius 1-1, initial runway 27.", air, legInitial},
		{"Aminad Tower, Genius 1-1, midfield break.", air, legBreak},
		{"Almanac Tower, Genius 1-1, gear down, downwind.", air, legDownwind},
		{"Aminad Tower, Genius 1-1, turning base.", air, legBase},
		{"I'm an entire genius one one on short final", air, legFinal},
		{"Thank you. That was accurate.", gnd, legThanks},
		{"Aminad Tower, Genius 1-1, no further assistance required.", gnd, legThanks},
		{"radio check", air, ""},
		{"go around", air, ""},
	}
	for _, c := range cases {
		got := detectPatternLeg(c.in, c.st)
		if got != c.want {
			t.Errorf("%q: got %q want %q", c.in, got, c.want)
		}
	}
}

func TestPatternReplySequence(t *testing.T) {
	st := &AircraftState{Callsign: "Genius 1-1"}
	pilot, cs, rwy, wind := "Genius 1-1", "Al Minhad Tower", "runway two-seven", "wind 270 at 26"

	msg, next, clear, _ := patternReply(legInitial, st, pilot, cs, rwy, wind)
	if next != legInitial || clear || !containsAny(msg, "report break") || containsAny(msg, "cleared to land") {
		t.Fatalf("initial: %q next=%s clear=%v", msg, next, clear)
	}
	st.Pattern = next

	msg, next, clear, _ = patternReply(legBreak, st, pilot, cs, rwy, wind)
	if next != legBreak || clear || !containsAny(msg, "report downwind") {
		t.Fatalf("break: %q", msg)
	}

	st.LandingGear = 1
	msg, next, clear, _ = patternReply(legDownwind, st, pilot, cs, rwy, wind)
	if clear || !containsAny(msg, "report base") {
		t.Fatalf("downwind: %q", msg)
	}

	msg, next, clear, _ = patternReply(legBase, st, pilot, cs, rwy, wind)
	if !clear || !containsAny(msg, "cleared to land") {
		t.Fatalf("base: %q clear=%v", msg, clear)
	}
	st.ClearedLand = true

	msg, _, _, _ = patternReply(legFinal, st, pilot, cs, rwy, wind)
	if !containsAny(msg, "continue") {
		t.Fatalf("final already cleared: %q", msg)
	}

	msg, _, _, _ = patternReply(legThanks, st, pilot, cs, rwy, wind)
	if containsAny(msg, "wind", "parked", "cleared") {
		t.Fatalf("thanks too long: %q", msg)
	}
}
