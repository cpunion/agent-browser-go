package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	agentbrowser "github.com/cpunion/agent-browser-go"
)

func main() {
	// Get script directory
	cwd, _ := os.Getwd()
	dataDir := filepath.Join(cwd, "data")

	// Clean up data directory
	os.RemoveAll(dataDir)

	fmt.Println("==========================================")
	fmt.Println("UserDataDir API Test")
	fmt.Println("==========================================")

	// Test both backends
	backends := []agentbrowser.BackendType{
		agentbrowser.BackendChromedp,
		agentbrowser.BackendPlaywright,
	}

	allPassed := true

	for _, backend := range backends {
		backendName := "chromedp"
		if backend == agentbrowser.BackendPlaywright {
			backendName = "playwright"
		}

		fmt.Printf("\n========================================\n")
		fmt.Printf("Testing backend: %s\n", backendName)
		fmt.Printf("========================================\n")

		// Clean data dir for this test
		os.RemoveAll(dataDir)

		// Create browser manager
		bm := agentbrowser.NewBrowserManagerWithBackend(backend)

		// Launch with UserDataDir
		fmt.Printf("→ Launching browser with UserDataDir=%s\n", dataDir)
		err := bm.Launch(agentbrowser.LaunchOptions{
			Headless:    false,
			UserDataDir: dataDir,
		})
		if err != nil {
			fmt.Printf("✗ Failed to launch: %v\n", err)
			allPassed = false
			continue
		}

		// Navigate to example.com
		fmt.Println("→ Navigating to https://example.com")
		_, _, err = bm.Navigate("https://example.com", "load")
		if err != nil {
			fmt.Printf("✗ Failed to navigate: %v\n", err)
			bm.Close()
			allPassed = false
			continue
		}

		// Wait for Chrome to write files
		time.Sleep(2 * time.Second)

		// Check if data directory was created and has files
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			fmt.Println("✗ UserDataDir was not created!")
			bm.Close()
			allPassed = false
			continue
		}

		// Count files
		var fileCount int
		filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				fileCount++
			}
			return nil
		})

		if fileCount > 0 {
			fmt.Printf("✓ UserDataDir created with %d files\n", fileCount)
		} else {
			fmt.Println("✗ UserDataDir created but empty!")
			bm.Close()
			allPassed = false
			continue
		}

		// Close browser
		fmt.Println("→ Closing browser...")
		bm.Close()
		fmt.Printf("✓ Backend %s: PASSED\n", backendName)
	}

	// Summary
	fmt.Println("\n==========================================")
	fmt.Println("Test Summary")
	fmt.Println("==========================================")

	if allPassed {
		fmt.Println("✓ ALL TESTS PASSED!")
		os.Exit(0)
	} else {
		fmt.Println("✗ SOME TESTS FAILED!")
		os.Exit(1)
	}
}
