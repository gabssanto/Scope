package completions

import (
	"fmt"
	"strings"
)

// Bash generates bash completion script
func Bash() string {
	return `# Scope bash completion script
# Add to ~/.bashrc: eval "$(scope completions bash)"

_scope_completions() {
    local cur prev commands tags
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="tag bulk untag tags list start scan go pick open edit each status pull rename merge clone-tag remove-tag prune doctor export import update debug help version completions"

    # Get tags dynamically
    if command -v scope &> /dev/null; then
        tags=$(scope list 2>/dev/null | grep -E '^\s+\S+' | awk '{print $1}')
    fi

    case "${prev}" in
        scope)
            COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
            return 0
            ;;
        tag|untag|tags)
            # Complete with directories
            COMPREPLY=( $(compgen -d -- "${cur}") )
            return 0
            ;;
        list|start|go|open|edit|each|status|pull|remove-tag|pick)
            # Complete with tag names
            COMPREPLY=( $(compgen -W "${tags}" -- "${cur}") )
            return 0
            ;;
        rename|merge|clone-tag)
            # Complete with tag names
            COMPREPLY=( $(compgen -W "${tags}" -- "${cur}") )
            return 0
            ;;
        import)
            # Complete with yaml files
            COMPREPLY=( $(compgen -f -X '!*.yml' -- "${cur}") $(compgen -f -X '!*.yaml' -- "${cur}") )
            return 0
            ;;
        bulk)
            # Complete with files, then tags
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -f -- "${cur}") )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=( $(compgen -W "${tags}" -- "${cur}") )
            elif [[ ${COMP_CWORD} -eq 4 ]]; then
                COMPREPLY=( $(compgen -W "--dry-run" -- "${cur}") )
            fi
            return 0
            ;;
        completions)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "${cur}") )
            return 0
            ;;
        prune)
            COMPREPLY=( $(compgen -W "--dry-run" -- "${cur}") )
            return 0
            ;;
        update)
            COMPREPLY=( $(compgen -W "--check" -- "${cur}") )
            return 0
            ;;
        each)
            # After 'each', complete with tags, then commands
            if [[ ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=( $(compgen -W "${tags}" -- "${cur}") )
            elif [[ ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=( $(compgen -W "-p --parallel" -- "${cur}") )
            fi
            return 0
            ;;
        *)
            ;;
    esac

    # Default to commands if nothing else matches
    COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
}

complete -F _scope_completions scope
`
}

// Zsh generates zsh completion script
func Zsh() string {
	return `#compdef scope
# Scope zsh completion script
# Add to ~/.zshrc: eval "$(scope completions zsh)"

_scope() {
    local -a commands tags

    commands=(
        'tag:Tag a folder'
        'bulk:Bulk tag paths from file'
        'untag:Remove a tag from a folder'
        'tags:Show all tags for a folder'
        'list:List all tags or folders with a tag'
        'start:Start a scoped session'
        'scan:Scan for .scope files'
        'go:Jump to a tagged folder'
        'pick:Interactive folder picker'
        'open:Open folder in file manager'
        'edit:Open folder in editor'
        'each:Run command in each folder'
        'status:Git status across folders'
        'pull:Git pull across folders'
        'rename:Rename a tag'
        'merge:Merge source tag into destination'
        'clone-tag:Copy tag associations to new tag'
        'remove-tag:Delete a tag entirely'
        'prune:Remove non-existent folders'
        'doctor:Check database health'
        'export:Export tags to YAML'
        'import:Import tags from YAML'
        'update:Update to latest version'
        'debug:Show debug information'
        'completions:Generate shell completions'
        'help:Show help'
        'version:Show version'
    )

    # Get tags dynamically
    if (( $+commands[scope] )); then
        tags=(${(f)"$(scope list 2>/dev/null | grep -E '^\s+\S+' | awk '{print $1}')"})
    fi

    _arguments -C \
        '1: :->command' \
        '*: :->args'

    case $state in
        command)
            _describe -t commands 'scope commands' commands
            ;;
        args)
            case $words[2] in
                tag|untag|tags)
                    _files -/
                    ;;
                list|start|go|open|edit|status|pull|remove-tag|pick)
                    _describe -t tags 'tags' tags
                    ;;
                rename|merge|clone-tag)
                    _describe -t tags 'tags' tags
                    ;;
                each)
                    if [[ $CURRENT -eq 3 ]]; then
                        _describe -t tags 'tags' tags
                    elif [[ $CURRENT -eq 4 ]]; then
                        _values 'flags' '-p[parallel]' '--parallel[parallel]'
                    fi
                    ;;
                import)
                    _files -g '*.y(a|)ml'
                    ;;
                bulk)
                    if [[ $CURRENT -eq 3 ]]; then
                        _files
                    elif [[ $CURRENT -eq 4 ]]; then
                        _describe -t tags 'tags' tags
                    elif [[ $CURRENT -eq 5 ]]; then
                        _values 'flags' '--dry-run[preview changes]'
                    fi
                    ;;
                completions)
                    _values 'shells' 'bash' 'zsh' 'fish'
                    ;;
                prune)
                    _values 'flags' '--dry-run[preview changes]'
                    ;;
                update)
                    _values 'flags' '--check[check only]'
                    ;;
            esac
            ;;
    esac
}

_scope "$@"
`
}

// Fish generates fish completion script
func Fish() string {
	return `# Scope fish completion script
# Add to ~/.config/fish/completions/scope.fish

# Disable file completion by default
complete -c scope -f

# Commands
complete -c scope -n "__fish_use_subcommand" -a "tag" -d "Tag a folder"
complete -c scope -n "__fish_use_subcommand" -a "bulk" -d "Bulk tag paths from file"
complete -c scope -n "__fish_use_subcommand" -a "untag" -d "Remove a tag from a folder"
complete -c scope -n "__fish_use_subcommand" -a "tags" -d "Show all tags for a folder"
complete -c scope -n "__fish_use_subcommand" -a "list" -d "List all tags or folders"
complete -c scope -n "__fish_use_subcommand" -a "start" -d "Start a scoped session"
complete -c scope -n "__fish_use_subcommand" -a "scan" -d "Scan for .scope files"
complete -c scope -n "__fish_use_subcommand" -a "go" -d "Jump to a tagged folder"
complete -c scope -n "__fish_use_subcommand" -a "pick" -d "Interactive folder picker"
complete -c scope -n "__fish_use_subcommand" -a "open" -d "Open folder in file manager"
complete -c scope -n "__fish_use_subcommand" -a "edit" -d "Open folder in editor"
complete -c scope -n "__fish_use_subcommand" -a "each" -d "Run command in each folder"
complete -c scope -n "__fish_use_subcommand" -a "status" -d "Git status across folders"
complete -c scope -n "__fish_use_subcommand" -a "pull" -d "Git pull across folders"
complete -c scope -n "__fish_use_subcommand" -a "rename" -d "Rename a tag"
complete -c scope -n "__fish_use_subcommand" -a "merge" -d "Merge source tag into destination"
complete -c scope -n "__fish_use_subcommand" -a "clone-tag" -d "Copy tag associations"
complete -c scope -n "__fish_use_subcommand" -a "remove-tag" -d "Delete a tag entirely"
complete -c scope -n "__fish_use_subcommand" -a "prune" -d "Remove non-existent folders"
complete -c scope -n "__fish_use_subcommand" -a "doctor" -d "Check database health"
complete -c scope -n "__fish_use_subcommand" -a "export" -d "Export tags to YAML"
complete -c scope -n "__fish_use_subcommand" -a "import" -d "Import tags from YAML"
complete -c scope -n "__fish_use_subcommand" -a "update" -d "Update to latest version"
complete -c scope -n "__fish_use_subcommand" -a "debug" -d "Show debug information"
complete -c scope -n "__fish_use_subcommand" -a "completions" -d "Generate shell completions"
complete -c scope -n "__fish_use_subcommand" -a "help" -d "Show help"
complete -c scope -n "__fish_use_subcommand" -a "version" -d "Show version"

# Helper function to get tags
function __scope_tags
    scope list 2>/dev/null | string match -r '^\s+\S+' | string trim | string split ' ' | head -1
end

# Tag completions for commands that take tags
complete -c scope -n "__fish_seen_subcommand_from list start go open edit status pull remove-tag pick" -a "(__scope_tags)" -d "Tag"
complete -c scope -n "__fish_seen_subcommand_from rename merge clone-tag" -a "(__scope_tags)" -d "Tag"
complete -c scope -n "__fish_seen_subcommand_from each" -a "(__scope_tags)" -d "Tag"

# Directory completion for tag/untag/tags
complete -c scope -n "__fish_seen_subcommand_from tag untag tags" -a "(__fish_complete_directories)"

# Flags
complete -c scope -n "__fish_seen_subcommand_from prune" -l dry-run -d "Preview changes"
complete -c scope -n "__fish_seen_subcommand_from bulk" -l dry-run -d "Preview changes"
complete -c scope -n "__fish_seen_subcommand_from update" -l check -d "Check only"
complete -c scope -n "__fish_seen_subcommand_from each" -s p -l parallel -d "Run in parallel"

# Shell completion for completions command
complete -c scope -n "__fish_seen_subcommand_from completions" -a "bash zsh fish" -d "Shell"

# File completion for import
complete -c scope -n "__fish_seen_subcommand_from import" -a "(__fish_complete_suffix .yml .yaml)"
`
}

// Generate returns the completion script for the given shell
func Generate(shell string) (string, error) {
	switch strings.ToLower(shell) {
	case "bash":
		return Bash(), nil
	case "zsh":
		return Zsh(), nil
	case "fish":
		return Fish(), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}
}
