# Changelog

All notable changes to this project will be documented in this file.

## [1.0.1] - 2026-01-01

### 🛠 Fixes & Improvements
- **Windows Path Support:** Fixed a bug where backslashes in Windows paths were incorrectly treated as escape characters during drag-and-drop operations.
- **Windows Icon Integration:** Corrected the build process to properly embed the application icon into Windows executables.
- **Improved Input Parsing:** Enhanced the command-line input parser to be more robust across different operating systems.

## [1.0.0] - 2026-01-01

First stable public release.

### 🚀 New Features
- **Interactive Interface:** Keyboard-navigable CLI menu for configuration and process control.
- **Drag & Drop Support:**
  - Drag files/folders onto the app icon (macOS) to process immediately.
  - Drag files directly into the open terminal window.
- **Compression Engine:**
  - Optimized **JPEG** support (configurable quality).
  - **PNG** support (maximum compression).
  - Automatic transparency handling (white background) for PNG -> JPEG conversions.
- **Resizing:**
  - `Max Width` option to automatically downscale large images while preserving aspect ratio (High-quality Catmull-Rom algorithm).
- **Format Conversion:**
  - Option to force output to JPEG or PNG, or keep the original format.
- **Analysis & Preview:**
  - Detailed pre-save table showing original size, estimated size, and savings percentage.
  - Total saved MB calculation.
  - Automatic skip if the compressed file is larger than the original.
- **Persistent Configuration:**
  - User preferences (Quality, Format, Resize) are saved in `config.json` within system folders (`Application Support` / `AppData`).
- **Build System:**
  - `build_mac.sh` script to generate native `.app` bundles with custom icons and terminal launchers (Intel & Silicon).
  - `build_win.sh` script to cross-compile Windows executables (x64 and ARM64).

### 🛠 Fixes & Improvements
- Fixed immediate terminal closure issue on macOS.
- Correct handling of paths with spaces and special characters (escaping).
- Eliminated system beep sounds during menu navigation.