# Opencode Templates

This directory contains templates, examples, and boilerplates for common tasks and workflows.

## Template Categories

### 📝 **Documentation Templates**
- **API Documentation** - Structured API docs with examples
- **README Templates** - Project README with badges and sections
- **Technical Specs** - Feature specification templates
- **User Guides** - Step-by-step user documentation

### 🛠️ **Code Templates**
- **Project Scaffolding** - Boilerplate for new projects
- **Component Templates** - Reusable component patterns
- **API Endpoints** - RESTful API templates
- **Database Schemas** - Common database patterns

### ⚙️ **Configuration Templates**
- **CI/CD Pipelines** - GitHub Actions, GitLab CI
- **Docker Setups** - Multi-stage Dockerfiles
- **Infrastructure** - Terraform, Kubernetes manifests
- **Environment Configs** - Development, staging, production

### 📊 **Project Management**
- **Project Plans** - Timeline and milestone templates
- **Issue Templates** - Bug reports, feature requests
- **Sprint Planning** - Agile sprint templates
- **Meeting Notes** - Structured meeting formats

## Template Structure

Each template includes:
- `template.md` - Main template file
- `variables.yaml` - Template variables and defaults
- `examples/` - Sample usage and outputs
- `README.md` - Template documentation

## Template Variables

Templates use variable substitution with this syntax:
```markdown
{{variable_name}}
{{variable_name|default_value}}
{{variable_name|filter}}
```

### Common Variables
- `project_name` - Project identifier
- `author_name` - Author information
- `created_date` - Creation timestamp
- `description` - Project description
- `license` - License type

### Filters
- `upper` - Uppercase text
- `lower` - Lowercase text
- `title` - Title case
- `snake_case` - Convert to snake_case
- `kebab-case` - Convert to kebab-case

## Template Usage

### CLI Usage
```bash
# List available templates
opencode template list

# Create from template
opencode template create api-endpoint --name "user-auth" --method "POST"

# Preview template
opencode template preview readme --project "My App"

# Install custom template
opencode template install ./my-template
```

### Interactive Mode
```bash
opencode template create
# Interactive prompt for template selection and variables
```

## Available Templates

### API Documentation
```yaml
# api-documentation.yaml
name: "API Documentation"
description: "Comprehensive API documentation template"
variables:
  api_name: "My API"
  base_url: "https://api.example.com"
  version: "v1"
  auth_type: "Bearer Token"
sections:
  - "overview"
  - "authentication"
  - "endpoints"
  - "errors"
  - "examples"
```

### Project README
```yaml
# project-readme.yaml
name: "Project README"
description: "Complete project README with badges"
variables:
  project_name: "My Project"
  description: "Project description"
  license: "MIT"
  build_status: true
  coverage_badge: true
sections:
  - "header"
  - "badges"
  - "description"
  - "installation"
  - "usage"
  - "contributing"
  - "license"
```

### Component Template
```yaml
# react-component.yaml
name: "React Component"
description: "React functional component with hooks"
variables:
  component_name: "MyComponent"
  props: "title:string,onClick:function"
  hooks: "useState,useEffect"
exports:
  - "Component.tsx"
  - "Component.test.tsx"
  - "Component.stories.tsx"
  - "index.ts"
```

## Template Development

### Creating Templates
1. Create directory with template name
2. Add template files with variable placeholders
3. Create variables.yaml for configuration
4. Add examples and documentation
5. Test template generation

### Template Syntax
Templates support:
- Variable substitution: `{{variable}}`
- Conditional blocks: `{{#if condition}}...{{/if}}`
- Loops: `{{#each items}}...{{/each}}`
- Includes: `{{> include-template}}`
- Custom functions: `{{function_name arg1 arg2}}`

### Example Template
```markdown
# {{project_name|title}}

{{description}}

## Installation

```bash
npm install {{package_name}}
```

## Usage

{{#if has_examples}}
### Examples

{{#each examples}}
- {{name}}: {{description}}
{{/each}}
{{/if}}

## Author

{{author_name}}

## License

{{license|upper}}
```

## Template Gallery

### Web Development
- **Express API** - RESTful API with middleware
- **React App** - Modern React with TypeScript
- **Vue Component** - Vue 3 composition API
- **Next.js Site** - Full-stack Next.js application

### Backend Services
- **Microservice** - Node.js microservice template
- **GraphQL Server** - Apollo GraphQL setup
- **Database Service** - Database abstraction layer
- **Message Queue** - Event-driven service template

### DevOps
- **Docker Compose** - Multi-service application
- **Kubernetes Deployment** - Production-ready manifests
- **CI/CD Pipeline** - Automated testing and deployment
- **Monitoring Setup** - Logging and metrics configuration

### Documentation
- **API Spec** - OpenAPI/Swagger documentation
- **Technical Guide** - Technical writing template
- **User Manual** - End-user documentation
- **Architecture Decision Record** - ADR template

## Template Distribution

### Official Templates
Available in the Opencode template registry:
```bash
opencode template search official
```

### Community Templates
Community-contributed templates:
```bash
opencode template search community
```

### Submit Templates
Share your templates with the community:
```bash
opencode template publish ./my-template
```

## Template Customization

### Custom Variables
Define project-specific variables:
```yaml
# .opencode/templates/custom.yaml
project_specific:
  company_name: "My Company"
  coding_standards: "company-standards.md"
  review_process: "team-process.md"
```

### Template Inheritance
Extend existing templates:
```yaml
# my-custom-api.yaml
extends: "api-documentation"
variables:
  api_name: "Custom API"
  custom_section: true
additional_sections:
  - "rate-limiting"
  - "webhooks"
```