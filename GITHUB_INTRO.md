# Java Version Manager (jvm) 🚀

**Java Version Manager (jvm)** is a Java version switching tool specifically designed for Windows, inspired by `nvm`. It aims for **extreme simplicity** and **enterprise-grade stability**.

---

## 💎 Why Choose jvm?

### 🚀 High Speed & Multi-Source Support (NEW!)
With built-in mirror acceleration, downloading a JDK takes just a few seconds. Support for `jvm source zulu/adoptium` allows seamless switching between different distribution APIs, and even supports private custom download sources.

### 🛡️ Intelligent Privilege Management (NEW!)
No more hunting for "Run as Administrator"! `jvm` intelligently identifies sensitive operations like `use` or `setup` and automatically requests UAC authorization when needed.

### 🔗 Flexible JAVA_HOME Mapping
You can map the JDK to any location you prefer (e.g., `C:\Program Files\Java\jdk`). Changing the mapping path is as simple as running `jvm javahome`, using Directory Junction technology—which is highly compatible even with Windows Home editions. Environment variables sync automatically, so your IDE requires no reconfiguration.

### 🍃 Pure Portable Software
All JDKs and configurations are kept within the tool directory. If you want to move to another computer, simply copy the `bin` directory and you're good to go.

---

## ⚡ Common Commands Quick Start

```bash
jvm ls a         # View available cloud JDKs
jvm i 17         # Install JDK 17
jvm use 17       # Switch to JDK 17 (Auto-requests admin privileges)
jvm source zulu  # Switch to Azul Zulu download source
jvm current      # Confirm the current active version
jvm ls           # View locally installed versions
jvm -v           # View tool version
```

---

## 📦 Distribution
If you are a developer, simply compile `src/main.go` and place the generated `jvm.exe` along with `install.bat` and `uninstall.bat` into the `bin` folder. Users only need the `bin` directory to enjoy full functionality.

---

**Driven by Go, empowering Java.**
