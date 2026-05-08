# Java Version Manager (jvm) 🚀

A lightweight, high-speed, and portable Java Version Manager for Windows.

---

## 🌟 Key Features
- **Minimalist Commands**: Supports shortcuts like `ls a` (list cloud), `i` (install), and `-v` (version).
- **Intelligent Elevation**: Automatically identifies operations requiring higher privileges and triggers UAC prompts—no more "Right-click, Run as Administrator."
- **Customizable Output**: Supports custom global JDK junction paths (defaults to `C:\Program Files\Java\jdk`), making it perfectly compatible with all IDEs.
- **Mirror Acceleration**: Built-in GitHub mirror support allows JDK downloads to complete in seconds.
- **Zero-Footprint**: No registry clutter. Environment variables are managed automatically. Simply delete the directory to uninstall.

---

## 📦 Directory Structure
- `src/`: Source code (Go)
- `bin/`: **[Distribution Build]** Users only need to download and use this folder.
- `scripts/`: Original script templates.

---

## 🛠️ Quick Start
1. Go to the `bin` directory and run `install.bat` as an administrator.
2. Set your **JAVA_HOME Junction Path** (Press Enter to use the default path).
3. Restart your terminal and start using the tool:
   - `jvm ls a` (View available JDKs)
   - `jvm i 17` (Download and install JDK 17)
   - `jvm use 17` (Switch between versions)

---

## 📖 Command Reference

| Command | Alias | Description |
| :--- | :--- | :--- |
| `list [a]` | `ls [a]` | List local or [a]vailable remote versions |
| `install <v>` | `i <v>` | Install a specific version |
| `use <v>` | - | Switch to a specific version (supports name or index) |
| `javahome <p>` | - | Migrate JAVA_HOME junction path (e.g., `jvm javahome D:\Java`) |
| `source [url]` | - | View/Switch JDK download source (supports adoptium/zulu/custom) |
| `mirror [on\|off]` | - | Enable/Disable download acceleration |
| `current` | - | View the currently active version |
| `version` | `-v` | View tool version |

---

**Made with ❤️ for Developers.**
