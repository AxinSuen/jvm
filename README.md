# Java Version Manager (jvm) 🚀

A lightweight, high-speed, and portable Java version management tool for Windows.

---

## 🌟 Key Features
- **Minimalist Commands**: Supports shortcuts like `ls a` (list remote), `i` (install), and `-v` (version).
- **Smart Elevation**: Automatically detects high-privilege operations and triggers UAC prompts—no more "Right-click, Run as Administrator".
- **Customizable Symlink**: Supports custom JDK symlink locations (default: `C:\Program Files\Java\jdk`), fully compatible with various IDEs.
- **Download Acceleration**: Built-in GitHub mirror support allows JDK downloads to complete in seconds.
- **Purely Portable**: No registry entries, automatic environment variable management—simply delete the directory to uninstall.

---

## 📦 Directory Structure
- `src/`: Source code (Go)
- `bin/`: **[Distribution]** Users only need this folder.
- `scripts/`: Original script templates.

---

## 🛠️ Quick Start
1. Enter the `bin` directory and run `install.bat` as an administrator.
2. Set your **Symlink Path** (press Enter for the default path).
3. Restart your terminal to start the magic:
   - `jvm ls a` (Check available JDKs)
   - `jvm i 17` (Download and install JDK 17)
   - `jvm use 17` (Switch versions)

---

## 📖 Command Reference

| Command | Alias | Description |
| :--- | :--- | :--- |
| `list [a]` | `ls [a]` | List local or [a]vailable remote versions |
| `install <v>` | `i <v>` | Install a specified version |
| `use <v>` | - | Switch to a specified version (with smart elevation) |
| `symlink <p>` | - | Change symlink export path (e.g., `jvm symlink D:\Java`) |
| `mirror [on\|off]` | - | Enable/Disable download acceleration |
| `current` | - | View currently active version |
| `version` | `-v` | View tool version |

---

**Made with ❤️ for Developers.**
