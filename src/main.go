package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"archive/zip"
	"syscall"
	"unsafe"
)

const (
	AppVersion  = "0.2.4"
	AdoptiumAPI = "https://api.adoptium.net/v3/assets/feature_releases"

	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorCyan  = "\033[36m"
)

func enableConsoleColors() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")

	handle, _ := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	var mode uint32
	getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode.Call(uintptr(handle), uintptr(mode))
}

type Settings struct {
	Root      string `json:"root"`
	Current   string `json:"current"`
	Symlink   string `json:"symlink"`
	Proxy     string `json:"proxy"`
	UseMirror bool   `json:"useMirror"`
	MirrorURL string `json:"mirrorURL"`
}

var settings Settings

func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

func runAsAdmin() {
	exe, _ := os.Executable()
	args := strings.Join(os.Args[1:], " ")
	cmd := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs", exe, args)
	exec.Command("powershell", "-Command", cmd).Run()
	os.Exit(0)
}

type Release struct {
	VersionData struct {
		Semver string `json:"semver"`
	} `json:"version_data"`
	Binaries []struct {
		Package struct {
			Link string `json:"link"`
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"package"`
	} `json:"binaries"`
}

func main() {
	enableConsoleColors()
	loadSettings()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "list", "ls":
		if len(os.Args) > 2 && (os.Args[2] == "available" || os.Args[2] == "a") {
			listAvailable()
		} else {
			listInstalled()
		}
	case "install", "i":
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定要安装的版本。")
			return
		}
		installVersion(os.Args[2])
	case "uninstall":
		if !isAdmin() { runAsAdmin() }
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定要卸载的版本。")
			return
		}
		uninstallVersion(os.Args[2])
	case "use":
		if !isAdmin() { runAsAdmin() }
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定要切换的版本。")
			return
		}
		useVersion(os.Args[2])
	case "root":
		if len(os.Args) > 2 {
			path := os.Args[2]
			if path == "." || path == "current" {
				settings.Root = ""
				fmt.Println("根目录已设为动态 (随 jvm.exe 位置变化)")
			} else {
				absPath, _ := filepath.Abs(path)
				settings.Root = absPath
				fmt.Printf("根目录已设为固定路径: %s\n", settings.Root)
			}
			saveSettings()
		} else {
			fmt.Printf("当前根目录: %s\n", getEffectiveRoot())
		}
	case "mirror":
		if len(os.Args) > 2 {
			val := strings.ToLower(os.Args[2])
			if val == "on" || val == "true" {
				settings.UseMirror = true
				fmt.Println("镜像加速: 开启 (使用 ghfast.top 代理)")
			} else {
				settings.UseMirror = false
				fmt.Println("镜像加速: 关闭 (使用官方源)")
			}
			saveSettings()
		} else {
			status := "关闭"
			if settings.UseMirror {
				status = "开启"
			}
			fmt.Printf("当前镜像加速状态: %s\n", status)
		}
	case "current":
		ver := getCurrentVersion()
		if ver == "" {
			fmt.Println("当前未选择任何 JDK 版本。")
		} else {
			fmt.Printf("当前版本: %s\n", ver)
		}
	case "setup":
		if !isAdmin() { runAsAdmin() }
		customSymlink := ""
		if len(os.Args) > 2 {
			customSymlink = os.Args[2]
		}
		setupEnvironment(customSymlink)
	case "symlink":
		if !isAdmin() { runAsAdmin() }
		if len(os.Args) < 3 {
			fmt.Printf("当前软链接路径: %s\n", settings.Symlink)
			fmt.Println("如需修改，请使用: jvm symlink <新路径>")
			return
		}
		newPath := os.Args[2]
		fmt.Printf("正在将软链接路径更新为: %s\n", newPath)
		
		// 1. Remove old symlink
		runPowerShell(fmt.Sprintf("if (Test-Path '%s') { Remove-Item '%s' -Force }", settings.Symlink, settings.Symlink))
		
		// 2. Update setting and Environment
		setupEnvironment(newPath)
		
		// 3. Re-apply current version if any
		if settings.Current != "" {
			useVersion(settings.Current)
		}
		fmt.Println("软链接路径已更新，环境已刷新！")
	case "version", "-v", "v":
		fmt.Printf("jvm version %s\n", AppVersion)
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Printf("%sJava Version Manager %s%s\n\n", ColorCyan, AppVersion, ColorReset)
	fmt.Println("使用说明:")
	fmt.Println("  jvm list [ls] [a]       列出已安装或 [a] 云端可用版本")
	fmt.Println("  jvm install [i] <ver>   安装指定版本 (如 17.0.1+12)")
	fmt.Println("  jvm use <version>       切换到指定版本")
	fmt.Println("  jvm mirror [on|off]     开启/关闭国内镜像加速")
	fmt.Println("  jvm uninstall <version> 卸载指定版本")
	fmt.Println("  jvm root [path]         设置或查看 JDK 存储根目录")
	fmt.Println("  jvm current             显示当前生效的版本")
	fmt.Println("  jvm version [-v]        显示工具版本及署名")
}

func loadSettings() {
	exePath, _ := os.Executable()
	settingsPath := filepath.Join(filepath.Dir(exePath), "settings.txt")
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, &settings)
	} else {
		settings.UseMirror = false
	}

	if settings.MirrorURL == "" {
		settings.MirrorURL = "https://ghfast.top/"
	}
	if settings.Symlink == "" {
		settings.Symlink = `C:\Program Files\Java\jdk`
	}
}

func saveSettings() {
	exePath, _ := os.Executable()
	settingsPath := filepath.Join(filepath.Dir(exePath), "settings.txt")
	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(settingsPath, data, 0644)
}

func getEffectiveRoot() string {
	if settings.Root != "" {
		return settings.Root
	}
	exePath, _ := os.Executable()
	return filepath.Dir(exePath)
}

func setupEnvironment(customSymlink string) {
	fmt.Println("正在初始化 Java 版本管理环境...")
	
	if customSymlink != "" {
		absSym, _ := filepath.Abs(customSymlink)
		settings.Symlink = absSym
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	root := getEffectiveRoot()
	os.MkdirAll(filepath.Join(root, "versions"), 0755)

	// Ensure the parent directory of the symlink exists
	os.MkdirAll(filepath.Dir(settings.Symlink), 0755)
	
	fmt.Printf("%s正在将 JAVA_HOME 设置为 %s%s\n", ColorGreen, settings.Symlink, ColorReset)
	if output, err := runPowerShell(fmt.Sprintf("[Environment]::SetEnvironmentVariable('JAVA_HOME', '%s', 'Machine')", settings.Symlink)); err != nil {
		fmt.Printf("\n%s[错误] 无法设置 JAVA_HOME: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		os.Exit(1)
	}
	
	fmt.Printf("%s正在将 %%JAVA_HOME%%\\bin 和工具目录添加到系统 PATH 顶部...%s\n", ColorGreen, ColorReset)
	script := fmt.Sprintf(`
		$exeDir = '%s'
		$javaBin = '%%JAVA_HOME%%\bin'
		$path = [Environment]::GetEnvironmentVariable('Path', 'Machine')
		
		# Remove existing entries to avoid duplicates and ensure they move to the top
		$pathArray = $path -split ';' | Where-Object { $_ -ne $javaBin -and $_ -ne $exeDir -and $_ -ne ($exeDir + '\') -and $_ -ne '' }
		
		# Prepend our paths
		$newPath = $javaBin + ';' + $exeDir + ';' + ($pathArray -join ';')
		[Environment]::SetEnvironmentVariable('Path', $newPath, 'Machine')
	`, exeDir)
	
	if output, err := runPowerShell(script); err != nil {
		fmt.Printf("\n%s[错误] 无法更新 PATH: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		os.Exit(1)
	}

	saveSettings()
	fmt.Println("\n初始化完成！请重新启动终端以使更改生效。")
}

func runPowerShell(command string) (string, error) {
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func listAvailable() {
	fmt.Println("正在从 Adoptium (Temurin) 获取可用版本...")
	majors := []int{8, 11, 17, 21}
	for _, v := range majors {
		releases, err := fetchReleases(v)
		if err != nil {
			fmt.Printf("获取 JDK %d 出错: %v\n", v, err)
			continue
		}

		if len(releases) > 0 {
			versions := make(map[string]bool)
			for _, r := range releases {
				versions[r.VersionData.Semver] = true
			}
			var sorted []string
			for k := range versions {
				sorted = append(sorted, k)
			}
			sort.Strings(sorted)
			fmt.Printf("\nJDK %d 可用版本:\n", v)
			for _, ver := range sorted {
				fmt.Printf("  - %s\n", ver)
			}
		}
	}
}

func fetchReleases(major int) ([]Release, error) {
	url := fmt.Sprintf("%s/%d/ga?architecture=%s&image_type=jdk&os=windows", AdoptiumAPI, major, getArch())
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func installVersion(version string) {
	fmt.Printf("正在查找版本 %s...\n", version)
	
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		fmt.Println("无效的版本格式。")
		return
	}
	majorStr := parts[0]
	if majorStr == "jdk8u" { majorStr = "8" } 
	var major int
	fmt.Sscanf(majorStr, "%d", &major)
	if major == 0 { major = 8 } 

	releases, err := fetchReleases(major)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	var targetLink string
	var targetSize int64
	for _, r := range releases {
		if r.VersionData.Semver == version {
			if len(r.Binaries) > 0 {
				targetLink = r.Binaries[0].Package.Link
				targetSize = r.Binaries[0].Package.Size
				break
			}
		}
	}

	if targetLink == "" {
		fmt.Printf("%s未找到适用于 Windows %s 的版本 %s%s\n", ColorRed, getArch(), version, ColorReset)
		return
	}

	if settings.UseMirror {
		fmt.Printf("正在使用镜像加速: %s\n", settings.MirrorURL)
		targetLink = settings.MirrorURL + targetLink
	}

	versionsDir := filepath.Join(getEffectiveRoot(), "versions")
	installDir := filepath.Join(versionsDir, version)

	if _, err := os.Stat(installDir); err == nil {
		fmt.Printf("版本 %s 已经安装。\n", version)
		return
	}

	os.MkdirAll(versionsDir, 0755)
	
	tmpFile := filepath.Join(os.TempDir(), "jvm-jdk.zip")
	fmt.Printf("正在下载 %s (%.2f MB)...\n", version, float64(targetSize)/1024/1024)
	
	if err := downloadFile(tmpFile, targetLink, targetSize); err != nil {
		fmt.Printf("\n下载出错: %v\n", err)
		return
	}

	fmt.Println("\n正在解压...")
	if err := extractZip(tmpFile, installDir); err != nil {
		fmt.Printf("解压出错: %v\n", err)
		return
	}

	os.Remove(tmpFile)
	fmt.Printf("\n成功安装 JDK %s 至 %s\n", version, installDir)
}

func uninstallVersion(version string) {
	installDir := filepath.Join(getEffectiveRoot(), "versions", version)
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		fmt.Printf("版本 %s 未安装。\n", version)
		return
	}

	fmt.Printf("正在卸载 %s...\n", version)
	if err := os.RemoveAll(installDir); err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Println("卸载完成。")
}

type ProgressWriter struct {
	Total      int64
	Downloaded int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += int64(n)
	pw.printProgress()
	return n, nil
}

func (pw *ProgressWriter) printProgress() {
	percent := float64(pw.Downloaded) / float64(pw.Total) * 100
	width := 30
	completed := int(percent / 100 * float64(width))
	
	bar := make([]byte, width)
	for i := 0; i < width; i++ {
		if i < completed {
			bar[i] = '='
		} else if i == completed {
			bar[i] = '>'
		} else {
			bar[i] = ' '
		}
	}
	
	fmt.Printf("\r[%s] %.1f%% (%d/%d MB)", string(bar), percent, pw.Downloaded/1024/1024, pw.Total/1024/1024)
}

func downloadFile(filepath string, url string, totalSize int64) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("状态异常: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &ProgressWriter{Total: totalSize}
	_, err = io.Copy(io.MultiWriter(out, pw), resp.Body)
	return err
}

func extractZip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	var topDir string
	if len(r.File) > 0 {
		firstFile := r.File[0].Name
		idx := strings.Index(firstFile, "/")
		if idx != -1 {
			topDir = firstFile[:idx]
		}
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, strings.TrimPrefix(f.Name, topDir))

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func useVersion(version string) {
	root := getEffectiveRoot()
	versionDir := filepath.Join(root, "versions", version)
	
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		fmt.Printf("版本 %s 未安装。请先执行 'jvm install %s'。\n", version, version)
		return
	}

	linkPath := settings.Symlink
	
	// Create symlink using PowerShell (requires admin)
	fmt.Printf("正在创建软链接: %s -> %s\n", linkPath, versionDir)
	// Use double quotes for paths to handle spaces better
	script := fmt.Sprintf("if (Test-Path \"%s\") { Remove-Item \"%s\" -Force -Recurse }; New-Item -ItemType SymbolicLink -Path \"%s\" -Target \"%s\"", 
		linkPath, linkPath, linkPath, versionDir)
	
	output, err := runPowerShell(script)
	if err != nil {
		fmt.Printf("%s创建软链接出错: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		fmt.Println("提示: 请尝试以管理员身份运行。")
		return
	}

	settings.Current = version
	saveSettings()
	fmt.Printf("\n%s当前正在使用 JDK %s%s\n", ColorGreen, version, ColorReset)

	// 如果是提权运行的（即没有在控制台保持打开），增加一个等待或确保用户看见
	if isAdmin() && len(os.Args) > 1 && os.Args[1] == "use" {
		fmt.Println("\n请按回车键关闭此窗口...")
		var input string
		fmt.Scanln(&input)
	}
}

func listInstalled() {
	root := getEffectiveRoot()
	versionsPath := filepath.Join(root, "versions")
	
	if _, err := os.Stat(versionsPath); os.IsNotExist(err) {
		fmt.Println("尚未安装任何 JDK 版本。使用 'jvm install' 开始安装。")
		return
	}

	entries, err := os.ReadDir(versionsPath)
	if err != nil {
		fmt.Printf("读取版本列表出错: %v\n", err)
		return
	}

	current := getCurrentVersion()
	fmt.Println("已安装的 JDK 版本:")
	for _, entry := range entries {
		if entry.IsDir() {
			prefix := "  "
			if entry.Name() == current {
				prefix = "* "
			}
			fmt.Printf("%s %s\n", prefix, entry.Name())
		}
	}
}

func getArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return runtime.GOARCH
}

func getCurrentVersion() string {
	return settings.Current
}
