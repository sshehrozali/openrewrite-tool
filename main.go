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

	if err := runRewrite("org.openrewrite.java.spring.boot3.UpgradeSpringBoot_3_3"); err != nil {
		log.Printf("❌ Rewrite failed: %v\n", err)
		os.Exit(1)
	}

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

func runRewrite(recipe string) error {
	cmd := exec.Command("rewrite", "run", "--activeRecipes", recipe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runBuild(projectType string) error {
	var cmd *exec.Cmd
	switch projectType {
	case "maven":
		cmd = exec.Command("./mvnw", "test")
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
