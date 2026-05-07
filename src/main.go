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
	AppVersion  = "0.2.3"
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
			fmt.Println("Error: Please specify a version to install.")
			return
		}
		installVersion(os.Args[2])
	case "uninstall":
		if !isAdmin() { runAsAdmin() }
		if len(os.Args) < 3 {
			fmt.Println("Error: Please specify a version to uninstall.")
			return
		}
		uninstallVersion(os.Args[2])
	case "use":
		if !isAdmin() { runAsAdmin() }
		if len(os.Args) < 3 {
			fmt.Println("Error: Please specify a version to use.")
			return
		}
		useVersion(os.Args[2])
	case "root":
		if len(os.Args) > 2 {
			path := os.Args[2]
			if path == "." || path == "current" {
				settings.Root = ""
				fmt.Println("Root set to dynamic (follow jvm.exe location)")
			} else {
				absPath, _ := filepath.Abs(path)
				settings.Root = absPath
				fmt.Printf("Root set to fixed path: %s\n", settings.Root)
			}
			saveSettings()
		} else {
			fmt.Printf("Current Root: %s\n", getEffectiveRoot())
		}
	case "mirror":
		if len(os.Args) > 2 {
			val := strings.ToLower(os.Args[2])
			if val == "on" || val == "true" {
				settings.UseMirror = true
				fmt.Println("Mirror acceleration: ON (Using ghfast.top proxy)")
			} else {
				settings.UseMirror = false
				fmt.Println("Mirror acceleration: OFF (Using Official)")
			}
			saveSettings()
		} else {
			status := "OFF"
			if settings.UseMirror {
				status = "ON"
			}
			fmt.Printf("Mirror acceleration is currently: %s\n", status)
		}
	case "current":
		ver := getCurrentVersion()
		if ver == "" {
			fmt.Println("No JDK is currently in use.")
		} else {
			fmt.Printf("Current Version: %s\n", ver)
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
			fmt.Printf("Current Symlink Path: %s\n", settings.Symlink)
			fmt.Println("To change, use: jvm symlink <path>")
			return
		}
		newPath := os.Args[2]
		fmt.Printf("Updating Symlink Path to: %s\n", newPath)
		
		// 1. Remove old symlink
		runPowerShell(fmt.Sprintf("if (Test-Path '%s') { Remove-Item '%s' -Force }", settings.Symlink, settings.Symlink))
		
		// 2. Update setting and Environment
		setupEnvironment(newPath)
		
		// 3. Re-apply current version if any
		if settings.Current != "" {
			useVersion(settings.Current)
		}
		fmt.Println("Symlink path updated and environment refreshed!")
	case "version", "-v", "v":
		fmt.Printf("jvm version %s\n", AppVersion)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Printf("%sJava Version Manager %s%s\n", ColorCyan, AppVersion, ColorReset)
	fmt.Println("Usage:")
	fmt.Println("  jvm list [ls] [a]       List installed or [a]vailable versions")
	fmt.Println("  jvm install [i] <ver>   Install a specific version (e.g. 17.0.1+12)")
	fmt.Println("  jvm use <version>       Switch to a specific version")
	fmt.Println("  jvm mirror [on|off]     Toggle Chinese mirror acceleration")
	fmt.Println("  jvm uninstall <version> Remove a specific version")
	fmt.Println("  jvm root [path]         Set or view the JDK storage root")
	fmt.Println("  jvm current             Show current active version")
	fmt.Println("  jvm version [-v]        Show tool version")
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
	fmt.Println("Initializing Java Version Manager environment...")
	
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
	
	fmt.Printf("%sSetting JAVA_HOME to %s%s\n", ColorGreen, settings.Symlink, ColorReset)
	if output, err := runPowerShell(fmt.Sprintf("[Environment]::SetEnvironmentVariable('JAVA_HOME', '%s', 'Machine')", settings.Symlink)); err != nil {
		fmt.Printf("\n%s[ERROR] Failed to set JAVA_HOME: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		os.Exit(1)
	}
	
	fmt.Printf("%sAdding %%JAVA_HOME%%\\bin and JVM tool directory to the TOP of System PATH...%s\n", ColorGreen, ColorReset)
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
		fmt.Printf("\n%s[ERROR] Failed to update PATH: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		os.Exit(1)
	}

	saveSettings()
	fmt.Println("\nSetup complete! Please RESTART your terminal for changes to take effect.")
}

func runPowerShell(command string) (string, error) {
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func listAvailable() {
	fmt.Println("Fetching available versions from Adoptium (Temurin)...")
	majors := []int{8, 11, 17, 21}
	for _, v := range majors {
		releases, err := fetchReleases(v)
		if err != nil {
			fmt.Printf("Error fetching JDK %d: %v\n", v, err)
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
			fmt.Printf("\nJDK %d Versions:\n", v)
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
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func installVersion(version string) {
	fmt.Printf("Finding version %s...\n", version)
	
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		fmt.Println("Invalid version format.")
		return
	}
	majorStr := parts[0]
	if majorStr == "jdk8u" { majorStr = "8" } 
	var major int
	fmt.Sscanf(majorStr, "%d", &major)
	if major == 0 { major = 8 } 

	releases, err := fetchReleases(major)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
		fmt.Printf("%sCould not find version %s for Windows %s%s\n", ColorRed, version, getArch(), ColorReset)
		return
	}

	if settings.UseMirror {
		fmt.Printf("Using Mirror Acceleration: %s\n", settings.MirrorURL)
		targetLink = settings.MirrorURL + targetLink
	}

	versionsDir := filepath.Join(getEffectiveRoot(), "versions")
	installDir := filepath.Join(versionsDir, version)

	if _, err := os.Stat(installDir); err == nil {
		fmt.Printf("Version %s is already installed.\n", version)
		return
	}

	os.MkdirAll(versionsDir, 0755)
	
	tmpFile := filepath.Join(os.TempDir(), "jvm-jdk.zip")
	fmt.Printf("Downloading %s (%.2f MB)...\n", version, float64(targetSize)/1024/1024)
	
	if err := downloadFile(tmpFile, targetLink, targetSize); err != nil {
		fmt.Printf("\nDownload error: %v\n", err)
		return
	}

	fmt.Println("\nExtracting...")
	if err := extractZip(tmpFile, installDir); err != nil {
		fmt.Printf("Extraction error: %v\n", err)
		return
	}

	os.Remove(tmpFile)
	fmt.Printf("\nSuccessfully installed JDK %s to %s\n", version, installDir)
}

func uninstallVersion(version string) {
	installDir := filepath.Join(getEffectiveRoot(), "versions", version)
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		fmt.Printf("Version %s is not installed.\n", version)
		return
	}

	fmt.Printf("Uninstalling %s...\n", version)
	if err := os.RemoveAll(installDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Done.")
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
		return fmt.Errorf("bad status: %s", resp.Status)
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
		fmt.Printf("Version %s is not installed. Use 'jvm install %s' first.\n", version, version)
		return
	}

	linkPath := settings.Symlink
	
	// Create symlink using PowerShell (requires admin)
	fmt.Printf("Creating symbolic link: %s -> %s\n", linkPath, versionDir)
	// Use double quotes for paths to handle spaces better
	script := fmt.Sprintf("if (Test-Path \"%s\") { Remove-Item \"%s\" -Force -Recurse }; New-Item -ItemType SymbolicLink -Path \"%s\" -Target \"%s\"", 
		linkPath, linkPath, linkPath, versionDir)
	
	output, err := runPowerShell(script)
	if err != nil {
		fmt.Printf("%sError creating symbolic link: %v%s\n", ColorRed, err, ColorReset)
		fmt.Println(output)
		fmt.Println("Tip: Please run as Administrator to allow symlink creation.")
		return
	}

	settings.Current = version
	saveSettings()
	fmt.Printf("\n%sNow using JDK %s%s\n", ColorGreen, version, ColorReset)
}

func listInstalled() {
	root := getEffectiveRoot()
	versionsPath := filepath.Join(root, "versions")
	
	if _, err := os.Stat(versionsPath); os.IsNotExist(err) {
		fmt.Println("No JDK versions installed yet. Use 'jvm install' to get started.")
		return
	}

	entries, err := os.ReadDir(versionsPath)
	if err != nil {
		fmt.Printf("Error reading versions: %v\n", err)
		return
	}

	current := getCurrentVersion()
	fmt.Println("Installed JDK versions:")
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
	root := getEffectiveRoot()
	currentLink := filepath.Join(root, "current")
	target, err := os.Readlink(currentLink)
	if err != nil {
		return ""
	}
	return filepath.Base(target)
}
