# SpudPad

A simple, lightweight notepad.exe replacement written in Go + Fyne.  
It's a potato — no bloat, no Markdown traps, no nonsense.

**Why?** Modern Notepad added features nobody asked for and vulnerabilities nobody wanted. SpudPad offers less — and that's the point.

### Features (so far)

- Open, save, save-as (with dialogs)
- Basic toolbar: new, open, save, cut/copy/paste (coming soon), help
- Live line/column + character count status bar
- Toggle word wrap
- Unsaved changes confirmation on new document
- Classic Notepad-like behavior (no wrap by default, horizontal scroll)

### Screenshots
None yet

### Downloads
Latest-dev version has been build by the runner and uploaded here:
https://github.com/EatPowerNode/spudpad/releases/download/latest-dev/SpudPad.exe

### Building

Requires Go 1.21+ and a C compiler (MSYS2/MinGW on Windows).

```bash
# Quick dev build
go build -o spudpad.exe .

# Or use Task (recommended)
go install github.com/go-task/task/v3/cmd/task@latest
task build-windows
