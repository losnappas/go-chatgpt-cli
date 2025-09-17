# go-chatgpt-cli

CLI interface for openai (chat)gpt (and gemini).

## Features

- Streaming output
- Pretty markdown render with https://github.com/charmbracelet/glamour
- History saved in a markdown file, easy to edit
- Neat way of keeping conversation history:
  
  ```bash
  # Put these in bashrc, and each new terminal will have its own conversation history.
  LOCAL_CHATGPT_CONTEXT="chatgpt-$RANDOM"
  alias ca='chatgpt_cli -history-file=$XDG_STATE_HOME/chatgpt/$LOCAL_CHATGPT_CONTEXT.md -model=openai/gpt-5-chat-latest -api-key="openai=$(pass show openai/api-key)" -system-prompt="Keep your answers concise."'
  
  # Optional ctrl-q keybind for bash prompt rewrites
  _rewrite_prompt_with_ai() {
    if [[ -n "$READLINE_LINE" ]]; then
        READLINE_LINE=$(chatgpt_cli -history-file="$XDG_STATE_HOME/chatgpt/${LOCAL_CHATGPT_CONTEXT}-cli.md" -model=openai/gpt-5-chat-latest -api-key="openai=$(pass show openai/api-key)" -system-prompt="Reply with a bash command. Output plain text, no markdown. E.g. <output>ls -la</output>" "$READLINE_LINE")
        READLINE_POINT=${#READLINE_LINE}
    fi
  }
  bind -x '"\C-q": _rewrite_prompt_with_ai'
  ```
