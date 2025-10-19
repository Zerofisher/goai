## Tool Usage Best Practices

### File Operations
- Always use `read_file` before `write_file` or `edit_file`
- Validate file paths are within the working directory
- Create backups for important edits

### Search Operations
- Use specific patterns for better results
- Combine with file type filters when possible
- Review search results before making changes

### Edit Operations
- Choose the right strategy (replace, insert, anchored, apply_patch)
- Use conflict detection for safety
- Review diffs before confirming changes

### Bash Commands
- Prefer read-only operations when possible
- Always check command output for errors
- Use timeouts for potentially long-running commands
