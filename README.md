# UK-US TTS Player

Simple personal application built with **Go + Fyne**. It takes English text and plays it back with **British (UK)** or **American (US)** pronunciation using **Piper TTS** and **mpv**.

---

## TODO

* [x] Start the project with basic features
* [ ] Organise the project using modules
* [ ] Set up a local database (SQLite)
* [ ] Finish the documentation

---

## Features

* Graphical interface built with **Fyne**
* Offline audio generation with **Piper TTS**
* Automatic playback with **mpv**
* Support for multiple British voices
* Audio caching to avoid regenerating repeated text

---

## Requirements

### 1. Install Piper

Download the Linux binary from:

https://github.com/rhasspy/piper/releases

Extract it so the executable is available at:

```text
./piper/piper
```

If you use a different location, update the `piperBin` constant in `main.go`.

---

### 2. Download British voices

Official voice list:

https://github.com/rhasspy/piper/blob/master/VOICES.md

Look for voices beginning with `en_GB-`, for example:

* `en_GB-alba-medium`
* `en_GB-alan-medium`

Place both the `.onnx` and `.onnx.json` files in:

```text
./piper-voices/
```

If the filenames differ, update the `britishVoices` map in `main.go`.

---

### 3. Install mpv

#### Void Linux

```bash
sudo xbps-install -S mpv
```

#### Debian / Ubuntu

```bash
sudo apt install mpv
```

---

### 4. Install Fyne system dependencies

#### Void Linux

```bash
sudo xbps-install -S \
  gcc pkg-config \
  libX11-devel libXcursor-devel \
  libXrandr-devel libXinerama-devel \
  mesa-devel libXi-devel libXxf86vm-devel
```

---

## Expected project structure

```text
desktop-audio-english-US-GB/
├── main.go
├── go.mod
├── piper/
│   └── piper
├── piper-voices/
│   ├── en_GB-alba-medium.onnx
│   ├── en_GB-alba-medium.onnx.json
│   └── ...
└── audios/
```

The `audios/` directory is created automatically.

---

## Running the application

```bash
go mod tidy
go run main.go
```

---

## Usage

1. Enter English text.
2. Select the desired British voice.
3. Click **Play**.

The application will:

1. Generate the audio using **Piper**
2. Save it in `./audios/`
3. Play it automatically through **mpv**

Generated files are named from a normalised version of the text (lowercase, without spaces or accents) to prevent duplicate generation.

---

## Notes

* **Piper always produces WAV files**.
* If playback fails, run the application from a terminal to inspect the logs:

```bash
go run main.go
```

* To add new voices, include additional entries in the `britishVoices` map pointing to the corresponding `.onnx` files.

---

## Quick installation

```bash
curl -fsSL https://raw.githubusercontent.com/DavidSilva-S/desktop-audio-english-US-GB/main/install.sh | sh
```

---

## Built with

* **Go**
* **Fyne**
* **Piper TTS**
* **mpv**

---

## Licence

This project is intended for personal use and learning purposes. 

Because apparently even tiny desktop tools need proper documentation before anyone trusts them. Civilisation is held together by Markdown files and cautious developers.
