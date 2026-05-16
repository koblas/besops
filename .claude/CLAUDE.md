~/.claude/CLAUDE.md

# Claude Code Instructions

## Go Development

Always use LSP for all Go code navigation and analysis. Never fall back to grep or find for tasks that LSP can handle.

### Required LSP usage for Go:
- **Symbol lookup**: Use `hover` to get type info, never read the file manually
- **Navigation**: Use `goToDefinition` instead of searching for declarations
- **References**: Use `findReferences` instead of grep for usages
- **Diagnostics**: Use `getDiagnostics` to check for errors after edits
- **Symbols**: Use `documentSymbols` to explore a file's structure

### Rules:
- Wait for gopls to be ready before starting Go tasks if LSP shows "server is starting"
- Never use `grep`, `find`, or `rg` to locate Go type definitions, function declarations, or references
- Always prefer semantic LSP navigation over text search for `.go` files
