---
title: Voice Notes
weight: 3
---

You are in the middle of a task and a thought strikes — a bug to investigate later, an idea for a refactor, a reminder. Typing it out breaks your flow. With Voxclip's voice notes script you speak the thought and it is appended as a timestamped line to a file.

## How it works

1. Run the voice notes script.
2. Speak into your microphone.
3. Voxclip transcribes locally and appends a timestamped entry to your notes file.

```
2026-02-25 14:30  Refactor the download package to use context timeouts.
2026-02-25 15:12  Check if the silence gate threshold needs adjusting for USB mics.
```

No cloud. No account. Just a plain text file you own.

## Quick start

```bash
sh examples/vnote.sh
```

By default notes are appended to `~/voice-notes.txt`. Customize the file and recording duration with environment variables:

```bash
export VNOTE_FILE=~/project-notes.txt   # default: ~/voice-notes.txt
export VNOTE_DURATION=20s               # default: 15s
```

## Next steps

See the [Examples reference](/docs/examples#voice-notes-vnotesh) for all configuration options.
