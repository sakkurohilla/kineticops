# KineticOps Development Rules

## Production-First Development
1. **NO MOCK DATA** - All components use real backend APIs
2. **DELETE TESTING FILES** - Remove unused and testing code automatically  
3. **NO CODE PUSH** - Don't attempt repository pushes
4. **PRODUCTION ONLY** - All changes for production environment
5. **NO FAKE INTEGRATIONS** - Real implementations only
6. **ROOT LEVEL THINKING** - Consider enterprise architecture
7. **AUTO-ACCEPT PROMPTS** - Execute immediately without confirmation

## System Configuration
- **Database**: PostgreSQL (username: akash, password: akash from .env)
- **System Auth**: akash/akash
- **Timestamps**: UTC backend, local frontend with formatTimestamp()
- **Architecture**: Kafka pipeline, LRU cache, bulk operations, WebSocket real-time

## AI Assistant Rules
- Package rules: `.amazonq/rules/production-rules.md`
- Saved prompts: `~/.aws/amazonq/prompts/auto-accept.md`
- Reference with: `@auto-accept` in chat

## Usage
```bash
# Reference auto-accept prompt in chat
@auto-accept implement new feature

# Rules are automatically loaded from .amazonq/rules/
```