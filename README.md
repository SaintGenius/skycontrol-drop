# Sky Control drop — ATIS

This is **not** the whole app. Copy these files **on top of** your existing folder:

`C:\\Users\\Rob\\Downloads\\skycontrol\\skycontrol\\`

Keep the folder names (`internal\\atc\\`, `internal\\airfield\\`, …).

**Download all as a zip:**  
https://github.com/SaintGenius/skycontrol-drop/archive/refs/heads/main.zip

## ATIS (this drop)

Second SRS radio. Loops a cached tape on the field ATIS frequency (Senaki = **260.900**). Tower stays on **261.000**.

Ask Ground: **what is the ATIS frequency?**

Then: BUILD.bat → LAUNCH.bat. COM1 261 talk, COM2 260.900 listen.
