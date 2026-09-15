package atc

import "testing"

func TestEmergencyKindAndSouls(t *testing.T) {
	if emergencyKind("declaring an emergency. low fuel state. direct approach") != "fuel" {
		t.Fatal("fuel")
	}
	if parseSouls("one soul on board") != 1 {
		t.Fatal("souls")
	}
	if !isEmergencyCall("mayday mayday engine out") {
		t.Fatal("mayday")
	}
}

func TestEmergencyReplyAsksSoulsOnce(t *testing.T) {
	st := &AircraftState{Callsign: "Genius 1-1", EmergKind: "fuel", FuelState: "low"}
	msg := emergencyReply(st, true, "Genius 1-1", "Al Minhad Tower", "runway two-seven", "wind 270 at 26")
	if !containsAny(msg, "cleared to land") || !containsAny(msg, "souls") {
		t.Fatalf("first: %s", msg)
	}
	if containsAny(msg, "fuel remaining") {
		t.Fatalf("should not ask fuel when already low: %s", msg)
	}
	st.Souls = 1
	msg = emergencyReply(st, false, "Genius 1-1", "Al Minhad Tower", "runway two-seven", "")
	if containsAny(msg, "cleared to land") {
		t.Fatalf("follow-up recleared: %s", msg)
	}
	if !containsAny(msg, "one soul") {
		t.Fatalf("follow-up: %s", msg)
	}
}
