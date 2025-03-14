# Simulator Autostart

Simulator Autostart is a utility designed to automatically start specified programs when a target process (e.g., a racing simulator) is detected as running. This tool is useful for launching applications like CrewChief, SimHub, or telemetry recorders automatically when you start your racing game.

## Download

See [Releases](https://github.com/snipem/simulator_autostart/releases) for download links.

## Configuration

The configuration file is stored at:

```
%APPDATA%\simulator_autostart\config.ini
```

To open this folder in Windows Explorer, press `Win + R`, type the following command, and press Enter:

```
explorer %APPDATA%\simulator_autostart
```

### Example `config.ini` file

```ini
[AMS2AVX.exe]
programs=C:\Program Files (x86)\Britton IT Ltd\CrewChiefV4\CrewChiefV4.exe,
C:\Program Files (x86)\SimHub\SimHubWPF.exe,
C:\Users\mail\AppData\Local\popometer\popometer-recorder.exe,
C:\Users\mail\work\sim-to-motec\ams2-cli.bat

[iRacingUI.exe]
programs=C:\Program Files (x86)\Britton IT Ltd\CrewChiefV4\CrewChiefV4.exe,
C:\Program Files (x86)\SimHub\SimHubWPF.exe,
C:\Users\mail\AppData\Roaming\garage61-install\garage61-launcher.exe,
C:\Users\mail\AppData\Local\racelabapps\RacelabApps.exe
```

## How It Works

1. The program continuously checks if a specified process (like `AMS2AVX.exe` or `iRacingUI.exe`) is running.
2. If the process is found and the associated applications are not already running, it starts them automatically.
3. The applications are started in their respective working directories to ensure proper execution.

## Usage

Simply run the compiled executable of `simulator_autostart`. The program will automatically detect the defined processes and start the associated applications when needed.

## Logging

The program logs its activity, showing which processes were detected and whether an application was started or skipped due to already running.

## Notes
- Ensure the paths in `config.ini` are correct.
- If you modify `config.ini`, restart the program for changes to take effect.
- The tool runs continuously in the background, checking every second for process changes.

Enjoy your automated Simulator setup!

