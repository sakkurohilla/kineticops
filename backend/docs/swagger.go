package docs

// @title KineticOps API
// @version 1.0
// @description Comprehensive infrastructure monitoring and observability platform
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@kineticops.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Health
// @tag.description Health check and system status endpoints

// @tag.name Authentication
// @tag.description User authentication and authorization

// @tag.name Hosts
// @tag.description Host management and monitoring

// @tag.name Metrics
// @tag.description Metrics collection and aggregation

// @tag.name Logs
// @tag.description Log management and search

// @tag.name Agents
// @tag.description Agent deployment and management

// @tag.name Alerts
// @tag.description Alert rules and notifications

// @tag.name APM
// @tag.description Application Performance Monitoring

// @tag.name Synthetic
// @tag.description Synthetic monitoring and uptime checks

// @tag.name Dashboards
// @tag.description Custom dashboards and visualizations

// @tag.name Users
// @tag.description User management

// @tag.name Tenants
// @tag.description Multi-tenancy management

// SwaggerInfo holds exported Swagger Info
var SwaggerInfo = struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}{
	Version:     "1.0",
	Host:        "localhost:8080",
	BasePath:    "/api/v1",
	Schemes:     []string{"http", "https"},
	Title:       "KineticOps API",
	Description: "Comprehensive infrastructure monitoring and observability platform",
}
