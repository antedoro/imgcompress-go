# Project Context: ImgCompress-go (v2.2)

## Project Overview
`ImgCompress-go` is a professional CLI utility for optimizing and converting images. It provides a rich interactive interface for configuring compression parameters and batch processing files via drag-and-drop.

## Tech Stack
*   **Language:** Go (Golang)
*   **CLI Framework:** `github.com/manifoldco/promptui`
*   **Image Processing:** Standard `image` lib + `golang.org/x/image/draw` for high-quality scaling.
*   **Persistence:** JSON configuration stored in system-standard paths (`os.UserConfigDir`).

## Key Features
*   **Interactive Menu:** Settings management and process control.
*   **Smart Resize:** Downscale images exceeding a specific width while maintaining aspect ratio.
*   **Format Conversion:** Convert between JPEG and PNG (handling transparency automatically).
*   **Preview Mode:** Detailed table showing potential savings before finalizing the operation.
*   **macOS Integration:** Packaged as a `.app` bundle that launches a dedicated terminal window.

## Build System
*   `build_mac.sh`: Generates `ImgCompress.app`. It also handles converting `icon.png` into a native macOS `.icns` file.
*   `build_win.sh`: Cross-compiles binaries for Windows (`x64` and `ARM64`) into the `dist/` folder.

## Project Structure
*   `main.go`: Core application logic and UI.
*   `icon.png`: Source for the application icon.
*   `build_mac.sh` / `build_win.sh`: Automation scripts for deployment.
*   `config.json`: (Generated at runtime) Stores user preferences.