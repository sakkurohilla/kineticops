# Production Development Rules

## Core Principles
1. **NO MOCK DATA**: Never add static, fake, or mock data - all components must use real backend APIs
2. **PRODUCTION ONLY**: All changes are for production environment, no demo/testing implementations
3. **CLEAN CODEBASE**: Always delete testing files and unused code
4. **ROOT LEVEL THINKING**: Consider system architecture and enterprise-level implications
5. **NO CODE PUSH**: Don't attempt to push code to repositories
6. **NO FAKE INTEGRATIONS**: Avoid demo or placeholder integrations

## Database & Authentication
- **PostgreSQL Credentials**: Use .env file values (username: akash, password: akash)
- **System Credentials**: Username and password are both "akash"
- **UTC Timestamps**: Backend uses UTC, frontend displays local time with formatTimestamp utility
- **Docker Services**: Always check docker-compose for PostgreSQL, MongoDB, Redis and other services

## Auto-Accept Policy
- Automatically accept all prompts and proceed with implementation
- No confirmation requests for standard operations
- Execute changes immediately when requirements are clear

## Architecture Requirements
- Enterprise-ready components with proper error handling
- LRU caching for performance optimization
- Bulk operations for database efficiency
- Kafka pipeline for data processing
- Real-time WebSocket connections
- 30-day data retention policies
- 10,000 connection limits

## UI Standards
- Paytm-style colorful gradients and modern animations
- Dark gradient sidebars with hover effects
- No duplicate headers or components
- Professional enterprise design patterns