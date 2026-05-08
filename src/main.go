package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	AppVersion  = "0.3.0"
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
	SourceURL string `json:"sourceURL"`
	JavaHome  string `json:"javaHome"`
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
	cmd := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb RunAs -Wait", exe, args)
	exec.Command("powershell", "-Command", cmd).Run()

	loadSettings()
	if len(os.Args) > 1 && (os.Args[1] == "use" || os.Args[1] == "uninstall") {
		listInstalled()
	}
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

type RemoteRelease struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	Size    int64  `json:"size"`
}

type JDKProvider interface {
	FetchAvailable(major int) ([]RemoteRelease, error)
}

// ---------------------------------------------------------
// Adoptium Provider
// ---------------------------------------------------------
type AdoptiumProvider struct {
	APIUrl string
}

func (p *AdoptiumProvider) FetchAvailable(major int) ([]RemoteRelease, error) {
	url := fmt.Sprintf("%s/%d/ga?architecture=%s&image_type=jdk&os=windows", p.APIUrl, major, getArch())
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

	var results []RemoteRelease
	for _, r := range releases {
		if len(r.Binaries) > 0 {
			results = append(results, RemoteRelease{
				Version: r.VersionData.Semver,
				URL:     r.Binaries[0].Package.Link,
				Size:    r.Binaries[0].Package.Size,
			})
		}
	}
	return results, nil
}

// ---------------------------------------------------------
// Zulu Provider
// ---------------------------------------------------------
type ZuluPackage struct {
	JavaVersion []int  `json:"java_version"`
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
}

type ZuluProvider struct {
	APIUrl string
}

func (p *ZuluProvider) FetchAvailable(major int) ([]RemoteRelease, error) {
	// Azul API: hw_bitness=64 for x64, arch=x86 (means intel architecture)
	arch := "x86"
	if getArch() != "x64" {
		arch = getArch() // Best effort
	}
	url := fmt.Sprintf("%s?os=windows&arch=%s&hw_bitness=64&ext=zip&java_version=%d&release_status=ga&java_package_type=jdk", p.APIUrl, arch, major)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态: %s", resp.Status)
	}

	var packages []ZuluPackage
	if err := json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return nil, err
	}

	var results []RemoteRelease
	for _, pkg := range packages {
		// Filter out JRE or FX packages if they sneaked in
		if strings.Contains(pkg.Name, "-jre") || strings.Contains(pkg.Name, "-fx-") {
			continue
		}

		// Reconstruct version string like "17.0.2"
		verStr := ""
		if len(pkg.JavaVersion) > 0 {
			var strParts []string
			for _, v := range pkg.JavaVersion {
				strParts = append(strParts, strconv.Itoa(v))
			}
			verStr = strings.Join(strParts, ".")
		} else {
			continue
		}

		results = append(results, RemoteRelease{
			Version: verStr,
			URL:     pkg.DownloadURL,
			Size:    0, // Zulu API doesn't provide size in the list
		})
	}
	return results, nil
}

// ---------------------------------------------------------
// Custom Provider
// ---------------------------------------------------------
type CustomProvider struct {
	APIUrl string
}

func (p *CustomProvider) FetchAvailable(major int) ([]RemoteRelease, error) {
	resp, err := http.Get(p.APIUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 返回状态: %s", resp.Status)
	}

	var results []RemoteRelease
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	// Optional filtering by major version could be done here if the custom API returns all versions.
	// We'll trust the custom API to either return the right versions or we'll just filter it locally.
	var filtered []RemoteRelease
	prefix := fmt.Sprintf("%d.", major)
	prefix8 := fmt.Sprintf("1.%d.", major)
	for _, r := range results {
		if strings.HasPrefix(r.Version, prefix) || strings.HasPrefix(r.Version, prefix8) || strconv.Itoa(major) == "8" && strings.HasPrefix(r.Version, "8") {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) > 0 {
		return filtered, nil
	}
	return results, nil // Fallback to returning all if filtering yields nothing
}

// ---------------------------------------------------------
// Provider Router
// ---------------------------------------------------------
func getProvider() JDKProvider {
	url := settings.SourceURL
	if strings.Contains(url, "api.adoptium.net") {
		return &AdoptiumProvider{APIUrl: "https://api.adoptium.net/v3/assets/feature_releases"}
	} else if strings.Contains(url, "api.azul.com") {
		return &ZuluProvider{APIUrl: "https://api.azul.com/metadata/v1/zulu/packages"}
	} else {
		return &CustomProvider{APIUrl: url}
	}
}

func main() {
	enableConsoleColors()
	loadSettings()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	// 如果是作为独立提权窗口运行的，防止报错或执行完毕后一闪而过导致闪退，全局添加拦截。
	// (普通的非管理员窗口通过 os.Exit() 退出，不会触发此 defer)
	defer func() {
		if isAdmin() && (command == "use" || command == "uninstall" || command == "javahome") {
			fmt.Println("\n(需要操作) 请按回车键关闭此窗口...")
			var input string
			fmt.Scanln(&input)
		}
	}()

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
		if !isAdmin() {
			runAsAdmin()
		}
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定要卸载的版本。")
			return
		}
		uninstallVersion(resolveVersion(os.Args[2]))
	case "use":
		if !isAdmin() {
			runAsAdmin()
		}
		if len(os.Args) < 3 {
			fmt.Println("错误: 请指定要切换的版本。")
			return
		}
		useVersion(resolveVersion(os.Args[2]))
	case "source":
		if len(os.Args) > 2 {
			val := os.Args[2]
			if val == "adoptium" {
				settings.SourceURL = AdoptiumAPI
				fmt.Println("已将下载源切换为: Adoptium (Temurin)")
			} else if val == "zulu" {
				settings.SourceURL = "https://api.azul.com/metadata/v1/zulu/packages"
				fmt.Println("已将下载源切换为: Azul Zulu")
			} else if strings.HasPrefix(val, "http") {
				settings.SourceURL = val
				fmt.Printf("已将下载源切换为自定义地址: %s\n", val)
			} else {
				fmt.Println("未知的源格式。请使用 'adoptium', 'zulu' 或提供以 http 开头的完整 URL。")
				return
			}
			saveSettings()
		} else {
			sourceName := "自定义源 (Custom)"
			if strings.Contains(settings.SourceURL, "api.adoptium.net") {
				sourceName = "Adoptium (Temurin)"
			} else if strings.Contains(settings.SourceURL, "api.azul.com") {
				sourceName = "Azul Zulu"
			}
			fmt.Printf("当前下载源: %s\n", sourceName)
			fmt.Printf("源接口地址: %s\n", settings.SourceURL)
		}
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
		if !isAdmin() {
			runAsAdmin()
		}
		customJavaHome := ""
		if len(os.Args) > 2 {
			customJavaHome = os.Args[2]
		}
		setupEnvironment(customJavaHome)
	case "javahome":
		if !isAdmin() {
			runAsAdmin()
		}
		if len(os.Args) < 3 {
			fmt.Printf("当前 JAVA_HOME 路径: %s\n", settings.JavaHome)
			fmt.Println("如需修改，请使用: jvm javahome <新路径>")
			return
		}
		newPath := os.Args[2]
		fmt.Printf("正在将软链接路径更新为: %s\n", newPath)

		oldVersion := getCurrentVersion()

		// 1. Remove old symlink
		exec.Command("cmd", "/c", "rmdir", "/q", settings.JavaHome).Run()

		// 2. Update setting and Environment
		setupEnvironment(newPath)

		// 3. Re-apply current version if any
		if oldVersion != "" {
			useVersion(oldVersion)
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
	fmt.Println("  jvm source [url|name]   设置或查看 JDK 下载源 (adoptium/zulu/自定义URL)")
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
	if settings.JavaHome == "" {
		settings.JavaHome = `C:\Program Files\Java\jdk`
	}
	if settings.SourceURL == "" {
		settings.SourceURL = AdoptiumAPI
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

func setupEnvironment(customJavaHome string) {
	fmt.Println("正在初始化 Java 版本管理环境...")

	if customJavaHome != "" {
		absSym, _ := filepath.Abs(customJavaHome)
		settings.JavaHome = absSym
	}

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	root := getEffectiveRoot()
	os.MkdirAll(filepath.Join(root, "versions"), 0755)

	// Ensure the parent directory of the symlink exists
	os.MkdirAll(filepath.Dir(settings.JavaHome), 0755)

	fmt.Printf("%s正在将 JAVA_HOME 设置为 %s%s\n", ColorGreen, settings.JavaHome, ColorReset)
	if output, err := runPowerShell(fmt.Sprintf("[Environment]::SetEnvironmentVariable('JAVA_HOME', '%s', 'Machine')", settings.JavaHome)); err != nil {
		fmt.Printf("\n%s[错误] 无法设置 JAVA_HOME: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		os.Exit(1)
	}

	fmt.Printf("%s正在将 %%JAVA_HOME%%\\bin 和工具目录添加到系统 PATH 顶部...%s\n", ColorGreen, ColorReset)
	script := fmt.Sprintf(`
		$exeDir = '%s'
		$javaBin = '%s\bin'
		$path = [Environment]::GetEnvironmentVariable('Path', 'Machine')
		
		# Remove existing entries to avoid duplicates and ensure they move to the top
		$pathArray = $path -split ';' | Where-Object { $_ -ne $javaBin -and $_ -ne '%%JAVA_HOME%%\bin' -and $_ -ne '%%JAVA_HOME%%' -and $_ -ne $exeDir -and $_ -ne ($exeDir + '\') -and $_ -ne '' }
		
		# Prepend our paths
		$newPath = $javaBin + ';' + $exeDir + ';' + ($pathArray -join ';')
		[Environment]::SetEnvironmentVariable('Path', $newPath, 'Machine')
	`, exeDir, settings.JavaHome)

	if output, err := runPowerShell(script); err != nil {
		fmt.Printf("\n%s[错误] 无法更新 PATH: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		os.Exit(1)
	}

	saveSettings()
	fmt.Printf("\n%s[成功] Java 环境初始化完成！%s\n", ColorGreen, ColorReset)
	fmt.Println("请重启终端以使更改生效。")
}

func runPowerShell(command string) (string, error) {
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func listAvailable() {
	fmt.Println("正在从当前源获取可用版本...")
	majors := []int{8, 11, 17, 21}
	provider := getProvider()
	for _, v := range majors {
		releases, err := provider.FetchAvailable(v)
		if err != nil {
			fmt.Printf("获取 JDK %d 出错: %v\n", v, err)
			continue
		}

		if len(releases) > 0 {
			versions := make(map[string]bool)
			for _, r := range releases {
				versions[r.Version] = true
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

func installVersion(version string) {
	fmt.Printf("正在查找版本 %s...\n", version)

	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		fmt.Println("无效的版本格式。")
		return
	}
	majorStr := parts[0]
	if majorStr == "jdk8u" || majorStr == "1" {
		majorStr = "8"
	}
	var major int
	fmt.Sscanf(majorStr, "%d", &major)
	if major == 0 {
		major = 8
	}

	provider := getProvider()
	releases, err := provider.FetchAvailable(major)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	var targetLink string
	var targetSize int64
	for _, r := range releases {
		if r.Version == version {
			targetLink = r.URL
			targetSize = r.Size
			break
		}
	}

	if targetLink == "" {
		fmt.Printf("%s未找到适用于 Windows %s 的版本 %s%s\n", ColorRed, getArch(), version, ColorReset)
		return
	}

	if settings.UseMirror && strings.Contains(settings.SourceURL, "api.adoptium.net") {
		fmt.Printf("正在使用镜像加速: %s\n", settings.MirrorURL)
		targetLink = settings.MirrorURL + targetLink
	}

	if targetSize == 0 {
		resp, err := http.Head(targetLink)
		if err == nil && resp.StatusCode == http.StatusOK {
			targetSize = resp.ContentLength
		}
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
	fmt.Printf("%s[成功] 版本 %s 已卸载。%s\n", ColorGreen, version, ColorReset)
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

	linkPath := settings.JavaHome

	// 确保父目录存在，防止首次使用时因为目录缺失导致链接失败
	os.MkdirAll(filepath.Dir(linkPath), 0755)

	// Create symlink
	fmt.Printf("正在创建软链接: %s -> %s\n", linkPath, versionDir)

	// 使用原生的 cmd 命令执行，彻底避免 PowerShell 对路径解析或空格转义导致的不稳定问题
	exec.Command("cmd", "/c", "rmdir", "/q", linkPath).Run()
	// 在 Windows 家庭版中，/D (软链接) 会因为严格的安全策略 (无 SeCreateSymbolicLinkPrivilege 权限) 频繁失败。
	// 改用 /J (目录联接 Directory Junction) 是最完美的解决方案，它不需要任何特权即可创建，且功能完全等同。
	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, versionDir)
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)

	if err != nil {
		fmt.Printf("%s创建软链接出错: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		fmt.Println("提示: 确保路径中没有非法字符，且拥有管理员权限。")
		return
	}
	fmt.Printf("\n%s当前正在使用 JDK %s%s\n", ColorGreen, version, ColorReset)
}

func getInstalledVersions() []string {
	root := getEffectiveRoot()
	versionsPath := filepath.Join(root, "versions")
	entries, err := os.ReadDir(versionsPath)
	if err != nil {
		return nil
	}
	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	return versions
}

func resolveVersion(input string) string {
	versions := getInstalledVersions()
	if len(versions) == 0 {
		return input
	}

	if idx, err := strconv.Atoi(input); err == nil {
		if idx >= 1 && idx <= len(versions) {
			return versions[idx-1]
		}
	}
	return input
}

func listInstalled() {
	versions := getInstalledVersions()

	if len(versions) == 0 {
		fmt.Println("尚未安装任何 JDK 版本。使用 'jvm install' 开始安装。")
		return
	}

	current := getCurrentVersion()
	fmt.Println("已安装的 JDK 版本:")
	for i, ver := range versions {
		prefix := "  "
		if ver == current {
			prefix = "* "
		}
		fmt.Printf("[%d] %s %s\n", i+1, prefix, ver)
	}
}

func getArch() string {
	if runtime.GOARCH == "amd64" {
		return "x64"
	}
	return runtime.GOARCH
}

func getCurrentVersion() string {
	// 优先尝试读取软链接/目录联接指向的真实物理路径，作为唯一的“事实真相”
	if settings.JavaHome != "" {
		if realPath, err := os.Readlink(settings.JavaHome); err == nil {
			baseName := filepath.Base(realPath)
			if baseName != "" && baseName != "." && baseName != string(filepath.Separator) {
				return baseName
			}
		}
	}

	// 如果软链接不存在或无法读取，说明当前没有生效的 JDK
	return ""
}
