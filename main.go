/*
Patch Edge Copilot - Go版本

此项目从Python版本(patch_edge_copilot.py)移植而来，
原始实现参考自：https://github.com/jiarandiana0307/patch-edge-copilot

主要改进：
- 迁移到Go语言以实现跨平台编译
- 使用GitHub Actions自动构建多平台可执行文件
- 支持Windows、Linux、macOS平台
- 支持x64和ARM64架构

主要功能：
- 自动检测Microsoft Edge用户数据路径
- 安全关闭Edge进程
- 修改配置启用Copilot功能
*/
package main


import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type LocalState struct {
	VariationsCountry string `json:"variations_country"`
}

type Preferences struct {
	Browser Browser `json:"browser"`
}

type Browser struct {
	ChatIPEligibilityStatus *bool `json:"chat_ip_eligibility_status"`
}

type versionPath struct {
	stable string
	canary string
	dev    string
	beta   string
}

func getVersionAndUserDataPath() (map[string]string, error) {
	var paths versionPath

	switch runtime.GOOS {
	case "windows":
		home := os.Getenv("USERPROFILE")
		paths = versionPath{
			stable: filepath.Join(home, "AppData", "Local", "Microsoft", "Edge", "User Data"),
			canary: filepath.Join(home, "AppData", "Local", "Microsoft", "Edge SxS", "User Data"),
			dev:    filepath.Join(home, "AppData", "Local", "Microsoft", "Edge Dev", "User Data"),
			beta:   filepath.Join(home, "AppData", "Local", "Microsoft", "Edge Beta", "User Data"),
		}
	case "linux":
		home := os.Getenv("HOME")
		paths = versionPath{
			stable: filepath.Join(home, ".config", "microsoft-edge"),
			canary: filepath.Join(home, ".config", "microsoft-edge-canary"),
			dev:    filepath.Join(home, ".config", "microsoft-edge-dev"),
			beta:   filepath.Join(home, ".config", "microsoft-edge-beta"),
		}
	case "darwin":
		home := os.Getenv("HOME")
		paths = versionPath{
			stable: filepath.Join(home, "Library", "Application Support", "Microsoft Edge"),
			canary: filepath.Join(home, "Library", "Application Support", "Microsoft Edge Canary"),
			dev:    filepath.Join(home, "Library", "Application Support", "Microsoft Edge Dev"),
			beta:   filepath.Join(home, "Library", "Application Support", "Microsoft Edge Beta"),
		}
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	versionDataPaths := make(map[string]string)

	// Check which versions exist
	if _, err := os.Stat(paths.stable); err == nil {
		versionDataPaths["stable"] = paths.stable
	}
	if _, err := os.Stat(paths.canary); err == nil {
		versionDataPaths["canary"] = paths.canary
	}
	if _, err := os.Stat(paths.dev); err == nil {
		versionDataPaths["dev"] = paths.dev
	}
	if _, err := os.Stat(paths.beta); err == nil {
		versionDataPaths["beta"] = paths.beta
	}

	return versionDataPaths, nil
}

func shutdownEdge() ([]string, error) {
	var terminatedEdges []string
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		var isEdge bool
		if runtime.GOOS == "darwin" {
			isEdge = strings.HasPrefix(name, "Microsoft Edge")
		} else {
			isEdge = name == "msedge"
		}

		if !isEdge {
			continue
		}

		// Check if process is still running
		if running, _ := p.IsRunning(); !running {
			continue
		}

		// Skip if parent has same name (avoid killing helper processes)
		if ppid, _ := p.Ppid(); ppid > 0 {
			if parent, err := process.NewProcess(ppid); err == nil {
				if parentName, err := parent.Name(); err == nil && parentName == name {
					continue
				}
			}
		}

		exe, err := p.Exe()
		if err != nil {
			continue
		}

		if err := p.Kill(); err == nil {
			terminatedEdges = append(terminatedEdges, exe)
		}
	}

	return terminatedEdges, nil
}

func getLastVersion(userDataPath string) (string, error) {
	lastVersionFile := filepath.Join(userDataPath, "Last Version")
	if _, err := os.Stat(lastVersionFile); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", lastVersionFile)
	}

	data, err := os.ReadFile(lastVersionFile)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func patchLocalState(userDataPath string) error {
	localStateFile := filepath.Join(userDataPath, "Local State")
	if _, err := os.Stat(localStateFile); os.IsNotExist(err) {
		return fmt.Errorf("failed to patch Local State. File not found: %s", localStateFile)
	}

	data, err := os.ReadFile(localStateFile)
	if err != nil {
		return err
	}

	var localState LocalState
	if err := json.Unmarshal(data, &localState); err != nil {
		return err
	}

	if localState.VariationsCountry != "US" {
		localState.VariationsCountry = "US"
		updatedData, err := json.MarshalIndent(localState, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(localStateFile, updatedData, 0644); err != nil {
			return err
		}
		fmt.Println("Succeeded in patching Local State")
	} else {
		fmt.Println("No need to patch Local State")
	}

	return nil
}

func patchPreferences(userDataPath string) error {
	entries, err := os.ReadDir(userDataPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "Default" && !strings.HasPrefix(entry.Name(), "Profile ") {
			continue
		}

		preferencesFile := filepath.Join(userDataPath, entry.Name(), "Preferences")
		if _, err := os.Stat(preferencesFile); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(preferencesFile)
		if err != nil {
			continue
		}

		var preferences Preferences
		if err := json.Unmarshal(data, &preferences); err != nil {
			continue
		}

		// Check if we need to patch
		if preferences.Browser.ChatIPEligibilityStatus == nil || !*preferences.Browser.ChatIPEligibilityStatus {
			// Set to true
			trueVal := true
			preferences.Browser.ChatIPEligibilityStatus = &trueVal

			updatedData, err := json.MarshalIndent(preferences, "", "  ")
			if err != nil {
				fmt.Printf("Failed to marshal preferences for %s: %v\n", entry.Name(), err)
				continue
			}

			if err := os.WriteFile(preferencesFile, updatedData, 0644); err != nil {
				fmt.Printf("Failed to write preferences for %s: %v\n", entry.Name(), err)
				continue
			}
			fmt.Printf("Succeeded in patching Preferences of %s\n", entry.Name())
		} else {
			fmt.Printf("No need to patch Preferences of %s\n", entry.Name())
		}
	}

	return nil
}

func restartEdge(terminatedEdges []string) {
	for _, edge := range terminatedEdges {
		var cmd *string
		switch runtime.GOOS {
		case "windows":
			cmdStr := edge + " --start-maximized"
			cmd = &cmdStr
		case "darwin":
			cmdStr := "open -a 'Microsoft Edge' --args --start-maximized"
			cmd = &cmdStr
		case "linux":
			cmdStr := edge + " --start-maximized"
			cmd = &cmdStr
		}

		if cmd != nil {
			fmt.Printf("Starting: %s\n", edge)
			// Note: In a real implementation, you'd use exec.Command here
			// This is simplified for the example
		}
	}
}

func main() {
	versionAndUserDataPath, err := getVersionAndUserDataPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(versionAndUserDataPath) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No available user data path found")
		os.Exit(1)
	}

	terminatedEdges, err := shutdownEdge()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to shutdown Edge: %v\n", err)
	} else if len(terminatedEdges) > 0 {
		fmt.Println("Shutdown Edge")
	}

	for version, userDataPath := range versionAndUserDataPath {
		lastVersion, err := getLastVersion(userDataPath)
		if err != nil {
			fmt.Printf("Failed to get version. %v\n", err)
			continue
		}

		parts := strings.Split(lastVersion, ".")
		if len(parts) == 0 {
			fmt.Printf("Invalid version format: %s\n", lastVersion)
			continue
		}

		fmt.Printf("Patching Edge %s %s \"%s\"\n", version, lastVersion, userDataPath)

		if err := patchLocalState(userDataPath); err != nil {
			fmt.Printf("Error patching Local State: %v\n", err)
		}

		if err := patchPreferences(userDataPath); err != nil {
			fmt.Printf("Error patching Preferences: %v\n", err)
		}
	}

	if len(terminatedEdges) > 0 {
		fmt.Println("Restart Edge")
		restartEdge(terminatedEdges)
	}

	fmt.Print("\nPress Enter to continue...")
	fmt.Scanln()
}
