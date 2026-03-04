---
title: Voxclip
toc: false
---

<div class="hx-mt-6 hx-mb-6">
{{< hextra/hero-badge link="https://github.com/fmueller/voxclip/releases" >}}
  <span>Latest Release</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
</div>

<div class="hx-mt-6 hx-mb-6">
<h1 class="hx-mt-2 hx-text-4xl hx-font-bold hx-tracking-tight hx-text-slate-900 dark:hx-text-slate-100 md:hx-text-5xl">Voice capture and transcription<br class="sm:hx-block hx-hidden" /> from your terminal</h1>
</div>

<div class="hx-mb-12">
<p class="hx-mt-6 hx-text-lg hx-text-gray-600 dark:hx-text-gray-400 sm:hx-text-xl">
Record audio, transcribe it with open-source speech models, and copy the result to your clipboard — all locally, all in one command.
</p>
</div>

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh

# Setup model and transcribe
voxclip setup
voxclip
```

<div class="hx-mt-6"></div>

## Features

{{< cards >}}
  {{< card link="docs" title="Fully Local" icon="lock-closed" subtitle="No cloud APIs, no data leaves your machine. Powered by open-source whisper.cpp speech models." >}}
  {{< card link="docs" title="One Command Flow" icon="play" subtitle="Record, transcribe, and copy to clipboard in a single command. No intermediate steps required." >}}
  {{< card link="docs" title="Cross-Platform" icon="desktop-computer" subtitle="Works on Linux and macOS with automatic backend detection for PipeWire, ALSA, and ffmpeg." >}}
  {{< card link="docs" title="Clipboard Ready" icon="clipboard-copy" subtitle="Transcripts are automatically copied to your clipboard, ready to paste into any application." >}}
  {{< card link="docs/installation" title="Easy Install" icon="download" subtitle="One-line installer script that detects your OS and architecture. No build tools needed." >}}
  {{< card link="docs/commands" title="Flexible CLI" icon="terminal" subtitle="Composable commands for recording, transcription, device listing, and model management." >}}
{{< /cards >}}

## Quick Start

{{% steps %}}

### Install Voxclip

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
```

### Download a speech model

```bash
voxclip setup
```

### Record and transcribe

```bash
voxclip
```

Speak into your microphone, press `Ctrl+C` to stop, and the transcript prints in the terminal and is copied to your clipboard.

{{% /steps %}}

## Learn More

{{< cards >}}
  {{< card link="docs" title="Documentation" icon="book-open" subtitle="Full command reference, configuration options, and troubleshooting." >}}
  {{< card link="https://github.com/fmueller/voxclip" title="GitHub" icon="github" subtitle="Source code, issues, and releases." >}}
{{< /cards >}}
