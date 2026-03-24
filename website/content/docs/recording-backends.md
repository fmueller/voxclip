---
title: Recording Backends
weight: 4
---

Voxclip supports multiple recording backends and automatically selects the best available one.

## Linux backend order

1. **`pw-record`** (PipeWire) — preferred
2. **`arecord`** (ALSA utils) — fallback
3. **`ffmpeg`** — last resort

## macOS backend

1. **`ffmpeg`** (`avfoundation`)

## Diagnostics

Use `voxclip devices` to see which backends are available and which devices are detected:

```bash
voxclip devices
```

## Forcing a backend

Use `--backend` to override automatic selection:

```bash
voxclip --backend pw-record --language en
voxclip --backend arecord --language en
voxclip --backend ffmpeg --language en
```
