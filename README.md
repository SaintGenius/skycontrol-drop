# Sky Control drop — backlog 1–4

This is **not** the whole app. Copy these files **on top of** your existing folder:

`C:\Users\Rob\Downloads\skycontrol\skycontrol\`

Keep the folder names (`internal\atc\`, `internal\telemetry\`).

## Files

| Copy this | Onto your PC here |
|-----------|-------------------|
| `internal/atc/traffic.go` | `internal\atc\traffic.go` *(new)* |
| `internal/atc/tower.go` | `internal\atc\tower.go` |
| `internal/atc/agent.go` | `internal\atc\agent.go` |
| `internal/atc/phraseology.go` | `internal\atc\phraseology.go` |
| `internal/atc/handoff.go` | `internal\atc\handoff.go` |
| `internal/telemetry/tacview.go` | `internal\telemetry\tacview.go` |

`traffic_test.go` is optional.

## Then

1. Run **BUILD.bat**
2. Start DCS + Tacview + SRS
3. Run **LAUNCH.bat**

**Download all as a zip:**  
https://github.com/SaintGenius/skycontrol-drop/archive/refs/heads/main.zip
