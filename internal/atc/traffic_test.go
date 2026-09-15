package atc

import (
	"testing"

	"github.com/skycontrol/skycontrol/internal/airfield"
)

func TestHeadingErr(t *testing.T) {
	if headingErr(10, 350) > 21 {
		t.Fatalf("wrap: got %v", headingErr(10, 350))
	}
	if headingErr(0, 180) < 179 {
		t.Fatalf("opposite: got %v", headingErr(0, 180))
	}
}

func TestInboundOutbound(t *testing.T) {
	af := &airfield.Airfield{Name: "Senaki", Latitude: 42.2406, Longitude: 42.0483}
	// South of field, heading north → inbound
	st := &AircraftState{Latitude: 42.20, Longitude: 42.0483, Heading: 0}
	if !st.inbound(af) {
		t.Fatalf("expected inbound, bearing=%v err=%v", st.bearingTo(af), headingErr(st.Heading, st.bearingTo(af)))
	}
	if st.outbound(af) {
		t.Fatal("south heading north should not be outbound")
	}
	st.Heading = 180
	if !st.outbound(af) {
		t.Fatalf("expected outbound, bearing=%v err=%v", st.bearingTo(af), headingErr(st.Heading, st.bearingTo(af)))
	}
	if st.inbound(af) {
		t.Fatal("heading away should not be inbound")
	}
}

func TestClassifyPhase(t *testing.T) {
	af := &airfield.Airfield{
		Name: "Senaki", Latitude: 42.2406, Longitude: 42.0483, ElevationFt: 43,
		Runways: []airfield.Runway{{Name: "09", HeadingTrue: 90}},
	}
	parked := &AircraftState{OnGround: true, SpeedMS: 4, Nearest: af, DistanceNM: 0.2, Latitude: af.Latitude, Longitude: af.Longitude}
	if p := classifyPhase(parked); p != "parked" {
		t.Fatalf("parked: %s", p)
	}
	cold := &AircraftState{OnGround: true, SpeedMS: 8, HasRPM: true, EngineRPM: 0, Nearest: af, DistanceNM: 0.2, Latitude: af.Latitude, Longitude: af.Longitude}
	if p := classifyPhase(cold); p != "parked" {
		t.Fatalf("cold: %s", p)
	}
	taxi := &AircraftState{OnGround: true, SpeedMS: 12, Nearest: af, DistanceNM: 0.3, Latitude: af.Latitude, Longitude: af.Longitude} // ~23 kt
	if p := classifyPhase(taxi); p != "taxi" {
		t.Fatalf("taxi: %s", p)
	}
	// North of field, heading south = inbound (toward field)
	fin := &AircraftState{
		OnGround: false, Latitude: 42.29, Longitude: 42.0483, Heading: 180,
		Nearest: af, DistanceNM: 3.0, AGLFt: 1200, LandingGear: 1, AltitudeFt: 1243,
	}
	fin.DistanceNM = airfield.DistanceNM(fin.Latitude, fin.Longitude, af.Latitude, af.Longitude)
	if p := classifyPhase(fin); p != "final" {
		t.Fatalf("final: %s inbound=%v bearing=%v dist=%v", p, fin.inbound(af), fin.bearingTo(af), fin.DistanceNM)
	}
	dep := &AircraftState{
		OnGround: false, Latitude: 42.29, Longitude: 42.0483, Heading: 0,
		Nearest: af, DistanceNM: 3.0, AGLFt: 1500, LandingGear: 1, MotionVS: 10, AltitudeFt: 1543,
	}
	dep.DistanceNM = airfield.DistanceNM(dep.Latitude, dep.Longitude, af.Latitude, af.Longitude)
	if p := classifyPhase(dep); p != "departure" {
		t.Fatalf("departure with gear down should not be final: %s outbound=%v", p, dep.outbound(af))
	}
}
