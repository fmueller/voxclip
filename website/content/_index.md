---
title: ""
toc: false
---

<div class="hx-mt-6 hx-mb-6">
{{< hextra/hero-badge link="https://github.com/fmueller/voxclip/releases" >}}
  <span>Latest Release</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
</div>

{{< hextra/hero-headline >}}
  Turn your voice into text,&nbsp;<br class="sm:hx-block hx-hidden" />right from the terminal
{{< /hextra/hero-headline >}}

{{< hextra/hero-subtitle >}}
  Record, transcribe, and paste — in a single command.&nbsp;<br class="sm:hx-block hx-hidden" />Everything runs locally. No cloud APIs, no accounts, works offline.
{{< /hextra/hero-subtitle >}}

<div class="hx-mb-12 hx-mt-8">
{{< hextra/hero-button text="Get Started" link="/docs/getting-started" >}}
</div>

<div class="install-block">

<p class="install-label">Quick install (Homebrew):</p>

```bash
brew tap fmueller/tap && brew install --cask voxclip
```

<p class="install-label">Quick install (script):</p>

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
```

</div>

## Why Voxclip?

{{< hextra/feature-grid >}}
  {{< hextra/feature-card
    title="Private by design"
    icon="lock-closed"
    subtitle="No one hears what you say except your own machine. No accounts, no cloud, no data collection — ever."
  >}}
  {{< hextra/feature-card
    title="Works offline"
    icon="cloud"
    subtitle="Download a small speech model once, then forget about the internet. Works on planes, in cafés, anywhere."
  >}}
  {{< hextra/feature-card
    title="One command"
    icon="play"
    subtitle="`voxclip` — that's it. Speak, press Enter, and the transcript is on your clipboard ready to paste."
  >}}
{{< /hextra/feature-grid >}}

## How it works

{{% steps %}}

### Install Voxclip

**Homebrew:**

```bash
brew tap fmueller/tap && brew install --cask voxclip
```

**Script:**

```bash
curl -fsSL https://raw.githubusercontent.com/fmueller/voxclip/main/scripts/install-voxclip.sh | sh
```

Downloads the correct binary for your OS and architecture. No build tools needed.

### Download a speech model

```bash
voxclip setup
```

Downloads and verifies the default speech model. This is a one-time download — after that, everything runs offline.

### Record and transcribe

```bash
voxclip
```

Speak into your microphone, press Enter to stop. The transcript prints in your terminal and is copied to your clipboard, ready to paste.

{{% /steps %}}

## Built for real workflows

{{< workflow-grid >}}
  {{< workflow-card
    title="Think faster than you type"
    icon="terminal"
    subtitle="Speak complex instructions to Claude Code, aider, or any AI tool instead of typing paragraphs. The transcript lands on your clipboard."
    link="/use-cases/voice-prompting"
  >}}
  {{< workflow-card
    title="Dictate anywhere"
    icon="microphone"
    subtitle="Press a hotkey, speak, and the transcript is pasted into whatever app is active — browser, email, chat. No cloud, no app switching."
    link="/use-cases/dictation"
  >}}
  {{< workflow-card
    title="Never lose a thought"
    icon="document-text"
    subtitle="Speak a thought mid-task and it's saved as a timestamped line in a plain text file. No context switch, no friction."
    link="/use-cases/voice-notes"
  >}}
{{< /workflow-grid >}}

## Features

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    title="Cross-platform"
    icon="desktop-computer"
    subtitle="Works on Linux and macOS with automatic backend detection for PipeWire, ALSA, and ffmpeg."
    link="/docs/recording-backends"
  >}}
  {{< hextra/feature-card
    title="Clipboard integration"
    icon="clipboard-copy"
    subtitle="Transcripts land on your clipboard automatically — paste into your editor, chat, or terminal."
    link="/docs/commands"
  >}}
  {{< hextra/feature-card
    title="Composable CLI"
    icon="terminal"
    subtitle="Subcommands for recording, transcription, and device management. Pipe and script them as you need."
    link="/docs/commands"
  >}}
  {{< hextra/feature-card
    title="Open-source models"
    icon="code"
    subtitle="Powered by OpenAI's Whisper models running locally via whisper.cpp. No proprietary dependencies."
    link="/docs/getting-started"
  >}}
{{< /hextra/feature-grid >}}

## Get started

<div class="hx-mt-6">
{{< hextra/hero-button text="Installation Guide" link="/docs/installation" >}}
{{< hextra/hero-button text="Documentation" link="/docs" style="background: transparent; color: inherit; border: 1px solid #d1d5db;" >}}
</div>
