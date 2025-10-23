#!/bin/bash
set -e

echo "🚀 KineticOps Project Initialization Started..."
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create backend directory structure
echo -e "${BLUE}Creating backend directory structure...${NC}"
mkdir -p backend/{cmd/server,internal/{api/{handlers,routes,middleware},models,repository/{postgres,mongodb,redis},services,auth,websocket,messaging,utils},config,migrations/postgres}

# Create frontend directory structure
echo -e "${BLUE}Creating frontend directory structure...${NC}"
mkdir -p frontend/src/{pages/{Login,Register,Dashboard,Metrics,Logs,Infrastructure,Alerts,Settings},components/{common,layout,auth,dashboard},services/{api,auth,websocket},hooks,context,utils,types,i18n}
mkdir -p frontend/public

# Create scripts and docs
echo -e "${BLUE}Creating scripts and documentation directories...${NC}"
mkdir -p scripts docs

# Backend initialization
echo -e "${BLUE}Initializing Go backend...${NC}"
cd backend

# Initialize Go module
go mod init github.com/sakkurohilla/kineticops 2>/dev/null || true

# Add dependencies (using stable public packages)
echo -e "${BLUE}Installing Go dependencies...${NC}"
go get github.com/gofiber/fiber/v2@latest
go get github.com/golang-jwt/jwt/v5@latest
go get gorm.io/gorm@latest
go get gorm.io/driver/postgres@latest
go get go.mongodb.org/mongo-driver@latest
go get github.com/redis/go-redis/v9@latest
go get github.com/joho/godotenv@latest
go get github.com/google/uuid@latest
go get golang.org/x/crypto@latest
go get github.com/gofiber/contrib/websocket@latest
go get github.com/twmb/franz-go@latest  # Kafka/Redpanda compatible client

# Tidy modules
go mod tidy

cd ..

# Frontend initialization
echo -e "${BLUE}Initializing React frontend...${NC}"
cd frontend

# Initialize npm
npm init -y > /dev/null 2>&1

# Install dependencies
echo -e "${BLUE}Installing npm dependencies...${NC}"
npm install --legacy-peer-deps \
  react@latest \
  react-dom@latest \
  typescript \
  --save

npm install --save-dev \
  @types/react \
  @types/react-dom \
  @types/node \
  tailwindcss \
  postcss \
  autoprefixer

npm install --save \
  axios \
  react-router-dom \
  zustand \
  swr \
  lucide-react \
  recharts \
  react-i18next \
  i18next \
  js-cookie \
  @types/js-cookie

# Initialize Tailwind
npx tailwindcss init -p > /dev/null 2>&1

cd ..

echo ""
echo -e "${GREEN}✅ Project structure created successfully!${NC}"
echo ""
echo -e "${BLUE}📁 Project Structure:${NC}"
echo "kineticops/"
echo "├── backend/          (Go Fiber server)"
echo "├── frontend/         (React TypeScript)"
echo "├── scripts/          (Utility scripts)"
echo "└── docs/             (Documentation)"
echo ""
echo -e "${BLUE}📋 Next Steps:${NC}"
echo "1. Copy .env.example to .env in backend directory"
echo "2. Start Docker services: docker-compose up"
echo "3. Run backend: cd backend && go run cmd/server/main.go"
echo "4. Run frontend: cd frontend && npm start"
echo ""
echo -e "${GREEN}Ready to start building KineticOps! 🎉${NC}"
