# Sky Control — backlog

**Use this in chat:** say *“roadmap”* or *“backlog item N”*.  
Do not treat this as done work. Items stay here until you fly them and say they work.

**Last updated:** 2026-09-14  
**On your PC:** not in the last zip you flew unless noted.

---

## How to read status

| Tag | Meaning |
|-----|---------|
| **FLYING** | In the build you have been using |
| **READY** | Written here, zip exists, **you have not installed it yet** |
| **NEXT** | Agreed direction, not coded |
| **LATER** | Wanted, not this week |
| **PARKED** | Decided no / wait |

---

## Open — fly / telemetry

| # | Item | Status | Notes |
|---|------|--------|-------|
| 1 | Pattern traffic callouts (downwind / base / final / 3-mile / runway) | **READY** | In `skycontrol-traffic-files.zip`. Gear + heading + position. |
| 2 | Parked vs taxi speed | **READY** | Same zip. On the ground, **under 12 kt = 0**. Idle/cold RPM also parked. |
| 3 | Approach vs departure by heading | **READY** | Pointing at the field = final/inbound. Pointing away = departure (even gear down). |
| 4 | RPM / throttle / afterburner / flaps as extra hints | **READY** | RPM ≈ 0 → cold parked. Idle throttle + slow → parked. Afterburner on the pad → runway. Missing for many other players/AI. |
| 5 | **SRS uplink — talk to Sky Control on the radio** | **NEXT** | You (and friends) PTT in SRS; SC hears that audio, not a second Windows mic. See below. |
| 5b | Native SRS receive (decode other humans) | *(same as 5)* | External AWACS / SRS TCP + UDP Opus. |
| 6 | Deliberate handoff (“contact Senaki”) vs auto nearest field | **LATER** | Auto nearest + identity is flying. Explicit switch not done. |
| 7 | ATIS | **LATER** | |
| 8 | Maps not in the DB (Normandy, Channel, South Atlantic, Iraq, CW Germany) | **LATER** | Easy to add JSON later. |

---

## Open — radio / voice

| # | Item | Status | Notes |
|---|------|--------|-------|
| 9 | SRS clipping (cuts after “Genius 1-1” / first sentence) | **WATCH** | Better on one UHF. Comes back with stacked freqs. Not declared dead. |
| 10 | Piper run-on / no pause between sentences | **WATCH** | Filler + speak-for-radio. Still easy to smash “zero-niner. You're”. |
| 11 | PTT beeps | **FLYING** | Worked when Windows default playback = G733. |
| 12 | ElevenLabs | **PARKED** | Extra delay + cost. Piper stays. |

---

## Open — product

| # | Item | Status | Notes |
|---|------|--------|-------|
| 13 | Friend uses Sky Control | **LATER** | Today: same PC as DCS+SRS+Tacview, or you host SRS and they tune the field freq. Not a per-player install that talks for them. |
| 14 | GUI polish | **FLYING** | Local page `http://127.0.0.1:8080`. Not a full settings/admin rewrite. |
| 15 | Wingman / GCI / other roles | **LATER** | Original idea. ATC first. |
| 16 | Whisper vs Windows STT | **FLYING** | Whisper path is what you use. Still mis-hears some calls. |

---

## Ready zip (not installed)

`skycontrol-traffic-files.zip` — backlog **1–4**

Drop onto `C:\Users\Rob\Downloads\skycontrol\skycontrol\`:

- `internal\atc\traffic.go` *(new)*
- `internal\atc\traffic_test.go` *(optional)*
- `internal\atc\tower.go`
- `internal\atc\agent.go`
- `internal\atc\phraseology.go`
- `internal\atc\handoff.go`
- `internal\telemetry\tacview.go`

Then **BUILD.bat** → DCS + Tacview + SRS → **LAUNCH.bat**.

---

## Test when you install that zip

1. Sitting still — ask speed. Should be **0**, not ~8 kt.  
2. Taxi — tens of knots, not hundreds.  
3. Pattern — downwind / base / final as you fly it.  
4. Gear down, ~3 nm, **heading toward the field** — “on a 3 mile final.”  
5. Gear down, heading **away**, climbing — “departing”, **not** final.  
6. Fast on the runway / afterburner — “traffic on the runway.”  
7. Parked jet (or engines off) — **no** airborne traffic call.

---

## Not doing unless you say so

- Rewrite from scratch  
- Mission-file ATC scripts  
- Buying a second Tacview license as a requirement  
- Cloud voice as default  

---

## Chat shortcuts

- *“backlog 1”* → traffic callouts  
- *“backlog 2”* → parked / taxi speed  
- *“backlog 3”* → heading in vs out  
- *“backlog 9”* → SRS clip  
- *“backlog 5”* → SRS uplink (talk to ATC on SRS)  
- *“mark 1 flying”* → you tested it, we move the tag  

---

## Backlog 5 — talk to Sky Control over SRS

**Today:** you PTT an Xbox button; Sky Control records the **Windows mic**. SRS is only how ATC **talks back**. Two pipes. Friends on another PC cannot talk to your ATC.

**Wanted:** you (or a friend) press **SRS radio PTT**, speak, Sky Control hears **that** transmission, then answers on the field freq. One radio.

**How (no extra DCS mission file):**

1. Sky Control joins the SRS **server** as a silent radio — same idea as ExternalAudio, but **listen**.
2. Best fit: SRS **External AWACS (EAM)** password so a program can RX without sitting in a cockpit.
3. Tune the same freqs we already TX (nearest tower UHF, plus any overlay).
4. When someone PTTs on that freq, SRS sends **UDP Opus**. We decode → Whisper → same ATC brain → ExternalAudio TX.
5. Ignore our own TX so we do not answer ourselves.

**Not this:** a virtual audio cable, or “whatever is playing on the headset.” That fights DCS and already caused mic lock / clipping.

**Order:** keep the current Xbox+mic path until EAM RX works, then make SRS the uplink and keep the mic as fallback.

**Depends on:** SRS server running (you already do this). Does not need LotATC. Friends only need SRS + the field freq.  
