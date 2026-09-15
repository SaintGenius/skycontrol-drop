package atc

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/skycontrol/skycontrol/internal/airfield"
)

const (
	patternTrafficNM = 6.0 // airborne near the FIELD (pattern / final)
	buddyTrafficNM   = 5.0 // airborne near YOU
	parkedKt         = 12.0
	parkedMS         = 6.2 // 12 kt
)

func wantsTrafficCall(intent string) bool {
	switch strings.ToLower(strings.TrimSpace(intent)) {
	case "landing", "inbound", "takeoff", "go_around", "touch_and_go", "check_in", "traffic":
		return true
	default:
		return false
	}
}

// attachTraffic adds one short advisory if the transmission does not already mention traffic.
// Caller must hold t.mu (or snapshot RLock).
func (t *Tower) attachTraffic(text string, st *AircraftState, af *airfield.Airfield, intent string) string {
	if text == "" || !wantsTrafficCall(intent) {
		return text
	}
	if strings.Contains(strings.ToLower(text), "traffic") {
		return text
	}
	call := t.trafficPhraseLocked(st, af, intent)
	if call == "" {
		return text
	}
	fmt.Printf("  traffic: %s\n", call)
	text = strings.TrimRight(text, " \t.")
	return text + ". " + call
}

func (t *Tower) trafficPhraseLocked(st *AircraftState, af *airfield.Airfield, intent string) string {
	o, distField, distYou, onRunway := t.pickTrafficLocked(st, af)
	if onRunway {
		return "Traffic on the runway."
	}
	if o == nil {
		return ""
	}
	typ := shortACType(o.Type)
	fieldMiles := spokenMiles(distField)
	youMiles := spokenMiles(distYou)
	alt := int((o.AltitudeFt+50)/100) * 100
	if alt < 100 {
		alt = int(o.AltitudeFt)
	}
	intent = strings.ToLower(intent)
	phase := o.Phase
	if phase == "" {
		phase = classifyPhase(o)
	}

	// Heading toward the field vs away beats gear-down guesses.
	if phase == "runway" {
		return "Traffic on the runway."
	}
	if phase == "final" {
		return fmt.Sprintf("Traffic, %s on a %s mile final.", typ, fieldMiles)
	}
	if phase == "downwind" {
		return fmt.Sprintf("Traffic, %s downwind.", typ)
	}
	if phase == "base" {
		return fmt.Sprintf("Traffic, %s on base.", typ)
	}
	if phase == "departure" {
		if intent == "landing" || intent == "inbound" || intent == "go_around" || intent == "touch_and_go" {
			return fmt.Sprintf("Traffic, %s departing, %s miles.", typ, fieldMiles)
		}
		if intent == "takeoff" {
			return fmt.Sprintf("Traffic, %s airborne, %s miles.", typ, fieldMiles)
		}
	}
	if intent == "takeoff" && phase == "final" {
		return fmt.Sprintf("Traffic, %s on a %s mile final.", typ, fieldMiles)
	}
	if intent != "takeoff" && st != nil && !st.OnGround && af != nil &&
		phase == "final" && distField+0.4 < st.DistanceNM && distField <= patternTrafficNM {
		return fmt.Sprintf("Number 2. Traffic is a %s, %s miles.", typ, fieldMiles)
	}
	if distField > patternTrafficNM && distYou <= buddyTrafficNM {
		return fmt.Sprintf("Traffic, %s, %s miles, %d feet.", typ, youMiles, alt)
	}
	return fmt.Sprintf("Traffic, %s, %s miles, %d feet.", typ, fieldMiles, alt)
}

func spokenMiles(nm float64) string {
	switch {
	case nm < 0.7:
		return "less than 1"
	case nm < 1.5:
		return "1"
	default:
		return fmt.Sprintf("%.0f", nm)
	}
}

func (t *Tower) pickTrafficLocked(st *AircraftState, af *airfield.Airfield) (best *AircraftState, distField, distYou float64, onRunway bool) {
	if st == nil {
		return nil, 0, 0, false
	}
	bestField := 1e9
	bestYou := 1e9
	var nearField, nearYou *AircraftState
	for _, o := range t.aircraft {
		if o == nil || o.ID == st.ID {
			continue
		}
		if strings.HasPrefix(o.ID, "demo-") {
			continue
		}
		if time.Since(o.LastSeen) > 12*time.Second {
			continue
		}
		if strings.EqualFold(o.Callsign, st.Callsign) || (o.Pilot != "" && strings.EqualFold(o.Pilot, st.Pilot)) {
			continue
		}
		if o.Phase == "parked" || (o.OnGround && o.SpeedMS < parkedMS && o.Afterburner < 0.2) {
			continue
		}
		dy := airfield.DistanceNM(st.Latitude, st.Longitude, o.Latitude, o.Longitude)
		df := dy
		if af != nil {
			df = airfield.DistanceNM(o.Latitude, o.Longitude, af.Latitude, af.Longitude)
		}
		if o.Phase == "runway" || (o.OnGround && o.Afterburner >= 0.2 && af != nil && df < 0.8) {
			return o, df, dy, true
		}
		if o.OnGround && o.SpeedMS > 15 && af != nil && df < 0.5 {
			return o, df, dy, true
		}
		if o.OnGround {
			continue
		}
		if df <= patternTrafficNM && df < bestField {
			bestField = df
			nearField = o
		}
		if dy <= buddyTrafficNM && dy < bestYou {
			bestYou = dy
			nearYou = o
		}
	}
	if nearField != nil {
		dy := airfield.DistanceNM(st.Latitude, st.Longitude, nearField.Latitude, nearField.Longitude)
		return nearField, bestField, dy, false
	}
	if nearYou != nil {
		df := bestYou
		if af != nil {
			df = airfield.DistanceNM(nearYou.Latitude, nearYou.Longitude, af.Latitude, af.Longitude)
		}
		return nearYou, df, bestYou, false
	}
	return nil, 0, 0, false
}

func headingErr(a, b float64) float64 {
	d := math.Mod(math.Abs(a-b), 360)
	if d > 180 {
		d = 360 - d
	}
	return d
}

func (st *AircraftState) bearingTo(af *airfield.Airfield) float64 {
	if st == nil || af == nil {
		return 0
	}
	return airfield.BearingDeg(st.Latitude, st.Longitude, af.Latitude, af.Longitude)
}

func (st *AircraftState) inbound(af *airfield.Airfield) bool {
	if st == nil || af == nil {
		return false
	}
	return headingErr(st.Heading, st.bearingTo(af)) <= 50
}

func (st *AircraftState) outbound(af *airfield.Airfield) bool {
	if st == nil || af == nil {
		return false
	}
	return headingErr(st.Heading, st.bearingTo(af)) >= 125
}

func runwayHeading(af *airfield.Airfield) float64 {
	if af == nil || len(af.Runways) == 0 {
		return 0
	}
	r := af.Runways[0]
	if r.HeadingTrue != 0 {
		return r.HeadingTrue
	}
	return r.HeadingMag
}

func (st *AircraftState) aglFt(af *airfield.Airfield) float64 {
	if st == nil {
		return 0
	}
	if st.AGLFt > 0 {
		return st.AGLFt
	}
	if af != nil {
		return st.AltitudeFt - af.ElevationFt
	}
	return st.AltitudeFt
}

// classifyPhase: parked / taxi / runway / departure / downwind / base / final / airborne.
func classifyPhase(st *AircraftState) string {
	if st == nil {
		return ""
	}
	af := st.Nearest
	gsKt := st.SpeedMS * 1.94384
	cold := st.HasRPM && st.EngineRPM < 0.5
	idle := st.HasThrottle && st.Throttle <= 0.22 && st.Afterburner < 0.2
	ab := st.Afterburner >= 0.25

	if st.OnGround {
		df := 99.0
		if af != nil {
			df = airfield.DistanceNM(st.Latitude, st.Longitude, af.Latitude, af.Longitude)
		}
		if ab && df < 0.8 {
			return "runway"
		}
		if cold || gsKt < parkedKt || (idle && gsKt < 14) {
			return "parked"
		}
		if df < 0.5 && gsKt >= 30 {
			return "runway"
		}
		return "taxi"
	}
	if af == nil {
		return "airborne"
	}
	dist := st.DistanceNM
	agl := st.aglFt(af)
	in := st.inbound(af)
	out := st.outbound(af)

	// Pointing away from the field is a departure, even with gear still down.
	if out && dist < 10 {
		if st.LandingGear < 0.5 || agl > 400 || st.MotionVS > 2 {
			return "departure"
		}
		if agl > 200 {
			return "departure"
		}
	}
	if in && dist <= 7 && agl < 3500 {
		if st.LandingGear >= 0.5 || st.Flaps >= 0.25 || dist <= 3.2 {
			return "final"
		}
		if dist <= 5 && agl < 2000 {
			return "final"
		}
	}
	rwy := runwayHeading(af)
	if rwy != 0 && dist >= 1.2 && dist <= 5 && agl >= 500 && agl <= 3200 {
		if headingErr(st.Heading, rwy+180) <= 35 {
			return "downwind"
		}
		if headingErr(st.Heading, rwy+90) <= 30 || headingErr(st.Heading, rwy+270) <= 30 {
			return "base"
		}
	}
	if !in && !out && dist <= 5 && agl < 3000 {
		brg := st.bearingTo(af)
		if headingErr(st.Heading, brg+90) <= 40 || headingErr(st.Heading, brg-90) <= 40 {
			return "downwind"
		}
	}
	return "airborne"
}

func shortACType(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "hornet") || strings.Contains(n, "fa-18") || strings.Contains(n, "f/a-18"):
		return "Hornet"
	case strings.Contains(n, "f-16") || strings.Contains(n, "f16") || strings.Contains(n, "viper"):
		return "Viper"
	case strings.Contains(n, "f-15") || strings.Contains(n, "eagle"):
		return "Eagle"
	case strings.Contains(n, "f-14") || strings.Contains(n, "tomcat"):
		return "Tomcat"
	case strings.Contains(n, "a-10") || strings.Contains(n, "warthog") || strings.Contains(n, "thunderbolt"):
		return "Hog"
	case strings.Contains(n, "av-8") || strings.Contains(n, "harrier"):
		return "Harrier"
	case strings.Contains(n, "f-5") || strings.Contains(n, "tiger"):
		return "Tiger"
	case strings.Contains(n, "m-2000") || strings.Contains(n, "mirage"):
		return "Mirage"
	case strings.Contains(n, "su-27") || strings.Contains(n, "su-33") || strings.Contains(n, "su-30") || strings.Contains(n, "flanker"):
		return "Flanker"
	case strings.Contains(n, "su-25") || strings.Contains(n, "frogfoot"):
		return "Frogfoot"
	case strings.Contains(n, "mig-29") || strings.Contains(n, "fulcrum"):
		return "Fulcrum"
	case strings.Contains(n, "mig-21"):
		return "Fishbed"
	case strings.Contains(n, "ah-64") || strings.Contains(n, "apache"):
		return "Apache"
	case strings.Contains(n, "uh-60") || strings.Contains(n, "blackhawk") || strings.Contains(n, "black hawk"):
		return "Blackhawk"
	case strings.Contains(n, "ka-50"):
		return "Shark"
	case strings.Contains(n, "mi-8") || strings.Contains(n, "mi-24"):
		return "Hip"
	case strings.Contains(n, "c-130") || strings.Contains(n, "hercules"):
		return "Hercules"
	case strings.Contains(n, "jf-17"):
		return "Thunder"
	case strings.Contains(n, "f-4") || strings.Contains(n, "phantom"):
		return "Phantom"
	case strings.Contains(n, "ajs37") || strings.Contains(n, "viggen"):
		return "Viggen"
	default:
		return "aircraft"
	}
}
