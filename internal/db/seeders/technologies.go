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
		// languages
		{Name: "JavaScript", Category: "language"},
		{Name: "TypeScript", Category: "language"},
		{Name: "Python", Category: "language"},
		{Name: "Go", Category: "language"},
		{Name: "Rust", Category: "language"},
		{Name: "Java", Category: "language"},
		{Name: "Kotlin", Category: "language"},
		{Name: "C#", Category: "language"},
		{Name: "C++", Category: "language"},
		{Name: "C", Category: "language"},
		{Name: "Swift", Category: "language"},
		{Name: "Ruby", Category: "language"},
		{Name: "PHP", Category: "language"},
		{Name: "Scala", Category: "language"},
		{Name: "Dart", Category: "language"},
		{Name: "Elixir", Category: "language"},
		{Name: "Clojure", Category: "language"},
		{Name: "Haskell", Category: "language"},
		{Name: "Lua", Category: "language"},
		{Name: "Perl", Category: "language"},
		{Name: "R", Category: "language"},
		{Name: "Julia", Category: "language"},
		{Name: "Zig", Category: "language"},
		{Name: "Nim", Category: "language"},
		{Name: "Crystal", Category: "language"},
		{Name: "Erlang", Category: "language"},
		{Name: "Groovy", Category: "language"},
		{Name: "F#", Category: "language"},
		{Name: "OCaml", Category: "language"},
		{Name: "D", Category: "language"},
		{Name: "Solidity", Category: "language"},
		{Name: "Shell", Category: "language"},
		{Name: "PowerShell", Category: "language"},
		{Name: "GraphQL", Category: "language"},

		// frontend frameworks & libraries
		{Name: "React", Category: "frontend"},
		{Name: "Vue.js", Category: "frontend"},
		{Name: "Angular", Category: "frontend"},
		{Name: "Svelte", Category: "frontend"},
		{Name: "SolidJS", Category: "frontend"},
		{Name: "Qwik", Category: "frontend"},
		{Name: "Preact", Category: "frontend"},
		{Name: "Alpine.js", Category: "frontend"},
		{Name: "Lit", Category: "frontend"},
		{Name: "HTMX", Category: "frontend"},
		{Name: "Stimulus", Category: "frontend"},
		{Name: "Ember.js", Category: "frontend"},
		{Name: "Backbone.js", Category: "frontend"},
		{Name: "jQuery", Category: "frontend"},
		{Name: "Bootstrap", Category: "frontend"},
		{Name: "Tailwind CSS", Category: "frontend"},
		{Name: "Sass", Category: "frontend"},
		{Name: "Less", Category: "frontend"},
		{Name: "Styled Components", Category: "frontend"},
		{Name: "CSS Modules", Category: "frontend"},

		// meta-frameworks & full-stack
		{Name: "Next.js", Category: "frontend"},
		{Name: "Nuxt.js", Category: "frontend"},
		{Name: "Astro", Category: "frontend"},
		{Name: "Remix", Category: "frontend"},
		{Name: "Gatsby", Category: "frontend"},
		{Name: "Eleventy", Category: "frontend"},
		{Name: "Docusaurus", Category: "frontend"},
		// backend frameworks & runtimes
		{Name: "Node.js", Category: "backend"},
		{Name: "Deno", Category: "backend"},
		{Name: "Bun", Category: "backend"},
		{Name: "Express", Category: "backend"},
		{Name: "NestJS", Category: "backend"},
		{Name: "Fastify", Category: "backend"},
		{Name: "Koa", Category: "backend"},
		{Name: "Hapi", Category: "backend"},
		{Name: "Socket.io", Category: "backend"},
		{Name: "Laravel", Category: "backend"},
		{Name: "Symfony", Category: "backend"},
		{Name: "CakePHP", Category: "backend"},
		{Name: "CodeIgniter", Category: "backend"},
		{Name: "Django", Category: "backend"},
		{Name: "Flask", Category: "backend"},
		{Name: "FastAPI", Category: "backend"},
		{Name: "Tornado", Category: "backend"},
		{Name: "Spring Boot", Category: "backend"},
		{Name: "Spring", Category: "backend"},
		{Name: "Micronaut", Category: "backend"},
		{Name: "Quarkus", Category: "backend"},
		{Name: "Gin", Category: "backend"},
		{Name: "Fiber", Category: "backend"},
		{Name: "Echo", Category: "backend"},
		{Name: "Chi", Category: "backend"},
		{Name: "Revel", Category: "backend"},
		{Name: "Buffalo", Category: "backend"},
		{Name: "Actix", Category: "backend"},
		{Name: "Rocket", Category: "backend"},
		{Name: "Axum", Category: "backend"},
		{Name: "Ruby on Rails", Category: "backend"},
		{Name: "Sinatra", Category: "backend"},
		{Name: "Phoenix", Category: "backend"},
		{Name: "Plug", Category: "backend"},
		{Name: "ASP.NET", Category: "backend"},
		{Name: "ASP.NET Core", Category: "backend"},
		{Name: ".NET", Category: "backend"},

		// devops & infrastructure
		{Name: "Docker", Category: "devops"},
		{Name: "Podman", Category: "devops"},
		{Name: "Kubernetes", Category: "devops"},
		{Name: "Nomad", Category: "devops"},
		{Name: "Terraform", Category: "devops"},
		{Name: "Pulumi", Category: "devops"},
		{Name: "Ansible", Category: "devops"},
		{Name: "Puppet", Category: "devops"},
		{Name: "Chef", Category: "devops"},
		{Name: "SaltStack", Category: "devops"},
		{Name: "GitHub Actions", Category: "devops"},
		{Name: "GitLab CI", Category: "devops"},
		{Name: "Jenkins", Category: "devops"},
		{Name: "CircleCI", Category: "devops"},
		{Name: "Travis CI", Category: "devops"},
		{Name: "ArgoCD", Category: "devops"},
		{Name: "Flux", Category: "devops"},
		{Name: "Helm", Category: "devops"},
		{Name: "Kustomize", Category: "devops"},
		{Name: "AWS", Category: "devops"},
		{Name: "Google Cloud Platform", Category: "devops"},
		{Name: "Azure", Category: "devops"},
		{Name: "DigitalOcean", Category: "devops"},
		{Name: "Linode", Category: "devops"},
		{Name: "Vercel", Category: "devops"},
		{Name: "Netlify", Category: "devops"},
		{Name: "Cloudflare", Category: "devops"},
		{Name: "Nginx", Category: "devops"},
		{Name: "Apache", Category: "devops"},
		{Name: "Caddy", Category: "devops"},
		{Name: "Traefik", Category: "devops"},
		{Name: "HAProxy", Category: "devops"},
		{Name: "Linux", Category: "devops"},
		{Name: "Ubuntu", Category: "devops"},
		{Name: "Debian", Category: "devops"},
		{Name: "Alpine", Category: "devops"},
		{Name: "Prometheus", Category: "devops"},
		{Name: "Grafana", Category: "devops"},
		{Name: "Datadog", Category: "devops"},
		{Name: "Sentry", Category: "devops"},
		{Name: "OpenTelemetry", Category: "devops"},
		{Name: "Istio", Category: "devops"},
		{Name: "Linkerd", Category: "devops"},
		{Name: "Consul", Category: "devops"},
		{Name: "Vault", Category: "devops"},
		{Name: "Vagrant", Category: "devops"},
		{Name: "Packer", Category: "devops"},

		// databases
		{Name: "PostgreSQL", Category: "database"},
		{Name: "MySQL", Category: "database"},
		{Name: "MariaDB", Category: "database"},
		{Name: "SQLite", Category: "database"},
		{Name: "MongoDB", Category: "database"},
		{Name: "Redis", Category: "database"},
		{Name: "Elasticsearch", Category: "database"},
		{Name: "Cassandra", Category: "database"},
		{Name: "DynamoDB", Category: "database"},
		{Name: "Firestore", Category: "database"},
		{Name: "CockroachDB", Category: "database"},
		{Name: "TiDB", Category: "database"},
		{Name: "ClickHouse", Category: "database"},
		{Name: "InfluxDB", Category: "database"},
		{Name: "TimescaleDB", Category: "database"},
		{Name: "Neo4j", Category: "database"},
		{Name: "ArangoDB", Category: "database"},
		{Name: "Supabase", Category: "database"},
		{Name: "PlanetScale", Category: "database"},

		// ORMs & database tools
		{Name: "Prisma", Category: "database"},
		{Name: "Drizzle", Category: "database"},
		{Name: "TypeORM", Category: "database"},
		{Name: "Sequelize", Category: "database"},
		{Name: "Knex", Category: "database"},
		{Name: "Mongoose", Category: "database"},
		{Name: "GORM", Category: "database"},
		{Name: "SQLAlchemy", Category: "database"},
		{Name: "Liquibase", Category: "database"},
		{Name: "Flyway", Category: "database"},

		// mobile
		{Name: "React Native", Category: "mobile"},
		{Name: "Flutter", Category: "mobile"},
		{Name: "SwiftUI", Category: "mobile"},
		{Name: "Jetpack Compose", Category: "mobile"},
		{Name: "Ionic", Category: "mobile"},
		{Name: "Capacitor", Category: "mobile"},
		{Name: "Xamarin", Category: "mobile"},
		{Name: ".NET MAUI", Category: "mobile"},
		{Name: "Expo", Category: "mobile"},
		{Name: "Unity", Category: "mobile"},

		// tools & editors
		{Name: "Git", Category: "tool"},
		{Name: "VS Code", Category: "tool"},
		{Name: "IntelliJ IDEA", Category: "tool"},
		{Name: "WebStorm", Category: "tool"},
		{Name: "Vim", Category: "tool"},
		{Name: "Neovim", Category: "tool"},
		{Name: "Emacs", Category: "tool"},
		{Name: "Sublime Text", Category: "tool"},
		{Name: "Zed", Category: "tool"},
		{Name: "Vite", Category: "tool"},
		{Name: "Webpack", Category: "tool"},
		{Name: "Rollup", Category: "tool"},
		{Name: "Parcel", Category: "tool"},
		{Name: "esbuild", Category: "tool"},
		{Name: "Babel", Category: "tool"},
		{Name: "ESLint", Category: "tool"},
		{Name: "Prettier", Category: "tool"},
		{Name: "Husky", Category: "tool"},
		{Name: "pnpm", Category: "tool"},
		{Name: "Yarn", Category: "tool"},
		{Name: "npm", Category: "tool"},
		{Name: "Postman", Category: "tool"},
		{Name: "Insomnia", Category: "tool"},
		{Name: "Swagger", Category: "tool"},
		{Name: "Figma", Category: "tool"},
		{Name: "Storybook", Category: "tool"},
		{Name: "Cypress", Category: "tool"},
		{Name: "Playwright", Category: "tool"},
		{Name: "Vitest", Category: "tool"},
		{Name: "Jest", Category: "tool"},
		{Name: "Mocha", Category: "tool"},
		{Name: "Selenium", Category: "tool"},
		{Name: "Puppeteer", Category: "tool"},
		{Name: "Homebrew", Category: "tool"},
		{Name: "curl", Category: "tool"},
		{Name: "jq", Category: "tool"},
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
