# Istruzioni per installare e configurare mpm (My Project Manager)

# 1. Assicurati di avere Go installato
# Se non lo hai, puoi installarlo da https://golang.org/doc/install

# 2. Crea una directory per il progetto
mkdir -p ~/go/src/mpm
cd ~/go/src/mpm

# 3. Copia il codice del file main.go nella directory

# 4. Installa le dipendenze
go mod init mpm
go get github.com/charmbracelet/bubbles/list
go get github.com/charmbracelet/bubbles/textinput
go get github.com/charmbracelet/bubbles/key
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/spf13/cobra

# 5. Compila il progetto
go build -o mpm

# 6. Sposta l'eseguibile in una directory nel PATH
sudo mv mpm /usr/local/bin/

# 7. Crea un wrapper script per consentire la navigazione tra directory
# Aggiungi questo al tuo ~/.bashrc o ~/.zshrc:

echo '
# Simple mpm wrapper function for interactive navigation
mpm() {
  if [[ "$1" = "go" ]]; then
    # For direct cd commands
    local output
    output=$(command mpm "$@")
    if [[ "$output" == cd* ]]; then
      eval "$output"
    else
      echo "$output"
    fi
  elif [[ "$1" = "i" ]]; then
    # For interactive mode
    command mpm i
    
    # After exiting, check for the temporary cd command file
    local cdfile="/tmp/mpm_cd_command"
    if [[ -f "$cdfile" ]]; then
      source "$cdfile"
      rm "$cdfile"
    fi
  else
    command mpm "$@"
  fi
}
' >> ~/.zshrc

# 8. Ricarica il tuo shell
source ~/.zshrc

echo "mpm è stato installato con successo!"
echo "Usa 'mpm i' per avviare la modalità interattiva."


# 9. Comandi

mpm add -n nome_progetto -p /path/to/project/folder -c categoria
mpm list
mpm go nome_progetto
mpm remove nome_progetto
mpm i

# Instructions for using the interactive navigation:
# 1. Run "mpm i" to open the interactive mode
# 2. Navigate to a project and press "g"
# 3. The application will create a temporary file with the cd command
# 4. When you exit, the shell function will automatically source this file
# 5. You will be navigated to the selected project directory