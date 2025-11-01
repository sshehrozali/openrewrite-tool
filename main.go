package main

import (
	"fmt"
	"log"
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
	fmt.Println("🔍 Detected project type: ", projectType)

	runRewrite()

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

func runRewrite() {
	// Hardcoded recipe details
	recipeArtifact := "org.openrewrite.recipe:rewrite-spring:RELEASE"
	recipeYAML := "com.org.Custom"

	// Construct the Maven command
	cmd := exec.Command(
		"mvn", "-U",
		"org.openrewrite.maven:rewrite-maven-plugin:run",
		fmt.Sprintf("-Drewrite.recipeArtifactCoordinates=%s", recipeArtifact),
		fmt.Sprintf("-Drewrite.activeRecipes=%s", recipeYAML),
		"-Drewrite.exportDatatables=true",
	)

	// Output to terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🚀 Running Custom OpenRewrite recipes: %s\n", recipeYAML)

	// Execute the command
	if err := cmd.Run(); err != nil {

		fmt.Printf("❌ Failed to run OpenRewrite: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅  OpenRewrite executed successfully.")
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
