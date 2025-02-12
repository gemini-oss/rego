#!/bin/bash

##########
: <<'END'
This script initializes/updates the content directory for Hugo documentation.

:Copyright: (c) 2025 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
END
##########

# -------------------------------------------------------------------------------------------
# CONFIG & VARIABLES
# -------------------------------------------------------------------------------------------
DOCS_DIR="hugo/content/docs"
SRC_DIR="pkg"
LOG_FILE="./gen_hugo_index.log"

# -------------------------------------------------------------------------------------------
# LOG FUNCTION
# -------------------------------------------------------------------------------------------
log() {
    local level="$1"
    local message="$2"

    # Color Escape Codes
    local Black="\033[30m"
    local Red="\033[31m"
    local Green="\033[32m"
    local Yellow="\033[33m"
    local Blue="\033[34m"
    local Magenta="\033[35m"
    local Cyan="\033[36m"
    local White="\033[37m"
    local Reset="\033[0m"

    if [[ "$level" == "DEBUG" && "$DEBUG" != "true" ]]; then
        return
    fi

    local datetime
    datetime=$(date "+%Y/%m/%d %I:%M:%S %p")
    local funcname="${FUNCNAME[1]}"
    local lineno="${BASH_LINENO[0]}"

    echo "[$datetime] {$funcname} {${BASH_SOURCE[1]}:$lineno} $level $message" >> "$LOG_FILE"

    # Determine color for log level
    case $level in
        DEBUG) level="${Blue}DEBUG${Reset}" ;;
        INFO) level="${Green}INFO${Reset}" ;;
        WARN) level="${Yellow}WARN${Reset}" ;;
        ERROR) level="${Red}ERROR${Reset}" ;;
    esac
    echo -e "[${Green}${datetime}${Reset}] {${Blue}${funcname}${Reset}} {${Magenta}${BASH_SOURCE[1]}${Reset}:${Yellow}${lineno}${Reset}} ${level} ${message}"
}



# -------------------------------------------------------------------------------------------
# Function to create directories and generate _index.md files
# -------------------------------------------------------------------------------------------
create_dir_and_index() {
    local dir_path="$1"
    local dir_name
    dir_name=$(basename "${dir_path}")

    log INFO "Creating directory: ${dir_path}"
    mkdir -p "${dir_path}"

    local index_file="${dir_path}/_index.md"

    if [[ ! -f "${index_file}" ]]; then
        log INFO "Generating _index.md in: ${dir_path}"
        cat << EOF > "${index_file}"
---
title: ${dir_name}
weight: 1
---
Explore the following sections to learn more:

{{< cards >}}
EOF
    fi
}

# -------------------------------------------------------------------------------------------
# Function to add cards to _index.md files
# -------------------------------------------------------------------------------------------
add_cards_to_index() {
    local dir_path="$1"
    local index_file="${dir_path}/_index.md"

    if [[ -f "${index_file}" ]]; then
        log INFO "Adding cards to: ${index_file}"
        local temp_file
        temp_file=$(mktemp)

        awk '/{{< \/cards >}}/{exit} {print}' "${index_file}" > "${temp_file}"

        for item in "${dir_path}"/*; do
            if [[ -d "${item}" ]]; then
                local name
                name=$(basename "${item}" | sed -e "s/.md\$//")
                local link="${name}"
                echo "  {{< card link=\"${link}\" title=\"${name}\" icon=\"document-duplicate\" >}}" >> "${temp_file}"
            fi
        done

        cat << EOF >> "${temp_file}"
{{< /cards >}}

<!-- gomarkdoc:embed:start -->
<!-- gomarkdoc:embed:end -->
EOF

        mv "${temp_file}" "${index_file}"
    fi
}

# -------------------------------------------------------------------------------------------
# Traverse the source directory structure and mirror it in the docs directory
# -------------------------------------------------------------------------------------------
log INFO "Mirroring source directories..."
find "${SRC_DIR}" -type d | while read -r src_dir; do
    docs_dir_path="${DOCS_DIR}/${src_dir}"
    create_dir_and_index "${docs_dir_path}"
done

# -------------------------------------------------------------------------------------------
# Add cards to each _index.md file
# -------------------------------------------------------------------------------------------
log INFO "Adding cards to _index.md files..."
find "${DOCS_DIR}" -type d | while read -r docs_dir; do
    add_cards_to_index "${docs_dir}"
done

# -------------------------------------------------------------------------------------------
# Generate documentation with gomarkdoc
# -------------------------------------------------------------------------------------------
if which gomarkdoc >/dev/null 2>&1; then
    log INFO "Generating documentation with gomarkdoc..."
    "$(go env GOPATH)/bin/gomarkdoc" ./... --output 'hugo/content/docs/{{.Dir}}/_index.md' --exclude-dirs ./pkg/internal/tests/... --embed
else
    log WARN "gomarkdoc not found. Installing..."
    go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
    log INFO "Generating documentation with gomarkdoc..."
    "$(go env GOPATH)/bin/gomarkdoc" ./... --output 'hugo/content/docs/{{.Dir}}/_index.md' --exclude-dirs ./pkg/internal/tests/... --embed
fi

# -------------------------------------------------------------------------------------------
# Create the _index.md file for the root directory
# -------------------------------------------------------------------------------------------
log INFO "Copying README.md to hugo/content/_index.md"
cp README.md hugo/content/_index.md
