# Java Version Manager (jvm) 🚀

**Java Version Manager (jvm)** 是一个专为 Windows 打造的、受 `nvm` 启发开发的 Java 多版本切换工具。它追求**极致的简单**与**工程级的稳定**。

---

## 💎 为什么选择 jvm？

### 🚀 极速体验
通过内置的国内镜像加速，下载一个 JDK 只需要喝一口水的时间。支持 `jvm ls a` 快速浏览 Adoptium 官方发布的各个 GA 版本。

### 🛡️ 智能权限管理 (NEW!)
再也不用满地找“以管理员身份运行”了！`jvm` 能够智能识别 `use` 或 `setup` 等敏感操作，并在需要时自动请求 UAC 授权。

### 🔗 灵活的 Symlink 映射
你可以将 JDK 映射到任何你喜欢的地方（如 `C:\Program Files\Java\jdk`）。更换映射路径只需一个 `jvm symlink` 命令，环境变量会自动同步，IDE 无需重新配置。

### 🍃 纯粹的绿色软件
所有 JDK 和配置都锁在工具目录下。如果你想换台电脑，直接把 `bin` 目录拷走就能用。

---

## ⚡ 常用指令速查

```bash
jvm ls a         # 看看云端有哪些 JDK
jvm i 17         # 安装 JDK 17
jvm use 17       # 切换到 JDK 17 (自动请求管理员权限)
jvm current      # 确认当前版本
jvm ls           # 看看本地装了哪些
jvm -v           # 看看工具版本
```

---

## 📦 如何分发
如果你是开发者，只需编译 `src/main.go` 并将生成的 `jvm.exe` 与 `install.bat` / `uninstall.bat` 一起放入 `bin` 文件夹。用户只需拥有 `bin` 目录即可享受完整功能。

---

**由 Go 驱动，为 Java 赋能。**
