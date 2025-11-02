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
	fmt.Println("🔍 Detected project type: ", projectType)

	fetchRecipes()
	runOpenRewrite()

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

func fetchRecipes() {
	yamlURL := "https://raw.githubusercontent.com/sshehrozali/openrewrite-tool/main/java-spring-recipes/recipes.yml"
	tmpFile := "rewrite.yml"

	resp, _ := http.Get(yamlURL)

	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	os.WriteFile(tmpFile, data, 0644)
	fmt.Println("📥 Downloaded recipe file from", yamlURL)
}

func runOpenRewrite() {
	recipeArtifact := "org.openrewrite.recipe:rewrite-spring:RELEASE"
	recipeYAML := "java-spring-recipes"

	cmd := exec.Command(
		"mvn", "-U",
		"org.openrewrite.maven:rewrite-maven-plugin:run",
		fmt.Sprintf("-Drewrite.recipeArtifactCoordinates=%s", recipeArtifact),
		fmt.Sprintf("-Drewrite.activeRecipes=%s", recipeYAML),
		"-Drewrite.exportDatatables=true",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🚀 Running Custom OpenRewrite recipes: %s\n", recipeYAML)

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
