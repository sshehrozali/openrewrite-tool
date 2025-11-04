package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("🔍 Running OpenRewrite pre-push checks...")

	projectType := detectProjectType()
	if projectType == "" {
		log.Println("❌ Not a recognized Spring Boot project (no pom.xml or build.gradle found).")
		os.Exit(1)
	}
	fmt.Println("🔍 Detected project type:", projectType)

	// Step 1: Fetch recipe
	tmpFile, err := fetchRecipes()
	if err != nil {
		log.Fatalf("❌ Failed to fetch recipes: %v", err)
	}
	// Step 2: Defer cleanup after successful fetch
	defer func() {
		os.Remove(tmpFile)
		fmt.Println("🧹 Cleaned up recipe YAML file:", tmpFile)
	}()

	// Step 3: Run rewrite + build
	runOpenRewrite(tmpFile)

	if err := runBuild(projectType); err != nil {
		log.Printf("❌ Build/tests failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Rewrite successful and build passed. Proceeding with push.")
}

func detectProjectType() string {
	if _, err := os.Stat("pom.xml"); err == nil {
		return "maven"
	}
	if _, err := os.Stat("build.gradle"); err == nil {
		return "gradle"
	}
	if _, err := os.Stat("build.gradle.kts"); err == nil {
		return "gradle"
	}
	return ""
}

func fetchRecipes() (string, error) {
	yamlURL := "https://raw.githubusercontent.com/sshehrozali/openrewrite-tool/main/java-spring-recipes/recipes.yml"
	tmpFile := "rewrite.yml"

	fmt.Println("📥 Downloading recipe file from", yamlURL)
	resp, err := http.Get(yamlURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch YAML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read YAML: %w", err)
	}

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save YAML: %w", err)
	}

	fmt.Println("✅ Recipe file saved to:", tmpFile)
	return tmpFile, nil
}

func runOpenRewrite(configFile string) {
	recipeArtifact := "org.openrewrite.recipe:rewrite-spring:RELEASE"

	cmd := exec.Command(
		"mvn", "-U",
		"org.openrewrite.maven:rewrite-maven-plugin:run",
		fmt.Sprintf("-Drewrite.recipeArtifactCoordinates=%s", recipeArtifact),
		fmt.Sprintf("-Drewrite.activeRecipes=%s", "java-spring-recipes"),
		"-Drewrite.exportDatatables=true",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🚀 Running OpenRewrite using config: %s\n", configFile)

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to run OpenRewrite: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ OpenRewrite executed successfully.")
}

func runBuild(projectType string) error {
	var cmd *exec.Cmd
	switch projectType {
	case "maven":
		cmd = exec.Command("mvn", "clean", "install")
	case "gradle":
		cmd = exec.Command("./gradlew", "test")
	default:
		return fmt.Errorf("unsupported project type: %s", projectType)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir, _ = filepath.Abs(".")
	return cmd.Run()
}