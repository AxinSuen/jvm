# Java Version Manager (jvm) 🚀

**Java Version Manager (jvm)** is a Java multi-version switching tool designed specifically for Windows, inspired by `nvm`. it pursues **extreme simplicity** and **engineering-grade stability**.

---

## 💎 Why choose jvm?

### 🚀 High-Speed Experience
With built-in mirror acceleration, downloading a JDK takes only a moment. Use `jvm ls a` to quickly browse all GA versions released by Adoptium.

### 🛡️ Smart Privilege Management (NEW!)
No more searching for "Run as Administrator"! `jvm` intelligently identifies sensitive operations like `use` or `setup` and automatically requests UAC authorization when needed.

### 🔗 Flexible Symlink Mapping
You can map the JDK to any location you prefer (e.g., `C:\Program Files\Java\jdk`). Changing the mapping path is just one `jvm symlink` command away—environment variables synchronize automatically, and IDEs require no reconfiguration.

### 🍃 Purely Portable
All JDKs and configurations are contained within the tool directory. If you want to switch computers, just copy the `bin` directory and you're good to go.

---

## ⚡ Common Commands

```bash
jvm ls a         # See which JDKs are available in the cloud
jvm i 17         # Install JDK 17
jvm use 17       # Switch to JDK 17 (automatically requests admin rights)
jvm current      # Confirm current version
jvm ls           # See which versions are installed locally
jvm -v           # View tool version
```

---

## 📦 How to Distribute
If you are a developer, simply compile `src/main.go` and place the generated `jvm.exe` into the `bin` folder along with `install.bat` and `uninstall.bat`. Users only need the `bin` directory to enjoy full functionality.

---

**Powered by Go, Empowering Java.**
