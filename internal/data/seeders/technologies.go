package seeders

import (
	"database/sql"
	"fmt"
	"strings"
)

type Technology struct {
	Name     string
	Category string
}

func normalizeSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))

	s = strings.ReplaceAll(s, ".js", "")
	s = strings.ReplaceAll(s, ".css", "")
	s = strings.ReplaceAll(s, " ", "-")

	return s
}

func SeedTechnologies(db *sql.DB) error {

	techs := []Technology{
		// frontend
		{Name: "React", Category: "frontend"},
		{Name: "Vue.js", Category: "frontend"},
		{Name: "Angular", Category: "frontend"},
		{Name: "Svelte", Category: "frontend"},
		{Name: "SolidJS", Category: "frontend"},
		{Name: "Next.js", Category: "frontend"},
		{Name: "Nuxt.js", Category: "frontend"},
		{Name: "Astro", Category: "frontend"},
		{Name: "Remix", Category: "frontend"},
		{Name: "Vite", Category: "tool"},
		{Name: "Tailwind CSS", Category: "frontend"},
		{Name: "TypeScript", Category: "language"},

		// backend
		{Name: "Node.js", Category: "backend"},
		{Name: "Express", Category: "backend"},
		{Name: "NestJS", Category: "backend"},
		{Name: "Laravel", Category: "backend"},
		{Name: "Django", Category: "backend"},
		{Name: "FastAPI", Category: "backend"},
		{Name: "Spring Boot", Category: "backend"},
		{Name: "Go", Category: "language"},
		{Name: "Gin", Category: "backend"},
		{Name: "Fiber", Category: "backend"},
		{Name: "Ruby on Rails", Category: "backend"},

		// devops
		{Name: "Docker", Category: "devops"},
		{Name: "Kubernetes", Category: "devops"},
		{Name: "Terraform", Category: "devops"},
		{Name: "Ansible", Category: "devops"},
		{Name: "GitHub Actions", Category: "devops"},
		{Name: "GitLab CI", Category: "devops"},
		{Name: "AWS", Category: "devops"},
		{Name: "Google Cloud Platform", Category: "devops"},
		{Name: "Azure", Category: "devops"},
		{Name: "Nginx", Category: "devops"},
		{Name: "Linux", Category: "devops"},

		// database
		{Name: "PostgreSQL", Category: "database"},
		{Name: "MySQL", Category: "database"},
		{Name: "MongoDB", Category: "database"},
		{Name: "Redis", Category: "database"},
		{Name: "SQLite", Category: "database"},
		{Name: "MariaDB", Category: "database"},
		{Name: "Cassandra", Category: "database"},
		{Name: "Elasticsearch", Category: "database"},

		// mobile
		{Name: "React Native", Category: "mobile"},
		{Name: "Flutter", Category: "mobile"},
		{Name: "Swift", Category: "mobile"},
		{Name: "Kotlin", Category: "mobile"},
	}

	for _, t := range techs {
		slug := normalizeSlug(t.Name)

		_, err := db.Exec(`
			INSERT INTO technologies (name, slug, category, isVerified)
			SELECT ?, ?, ?, true
			WHERE NOT EXISTS (
				SELECT 1 FROM technologies WHERE slug = ?
			)
		`, t.Name, slug, t.Category, slug)

		if err != nil {
			return fmt.Errorf("failed seeding %s: %w", t.Name, err)
		}
	}

	return nil
}
