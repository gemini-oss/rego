#!/bin/bash

##########
: <<'END'
This script initializes/updates the content directory for Hugo documentation.

:Copyright: (c) 2024 by Gemini Space Station, LLC., see AUTHORS for more info
:License: See the LICENSE file for details
:Author: Anthony Dardano <anthony.dardano@gemini.com>
END
##########

DOCS_DIR="hugo/content/docs"
SRC_DIR="pkg"

# Function to create directories and generate _index.md files
create_dir_and_index() {
    local dir_path="$1"
    local dir_name=$(basename "${dir_path}")

    # Create directory if it doesn't exist
    mkdir -p "${dir_path}"

    # Path for the _index.md file
    local index_file="${dir_path}/_index.md"

    # Create _index.md if it doesn't exist
    if [[ ! -f "${index_file}" ]]; then
        # Use heredoc to write content to _index.md
        cat << EOF > "${index_file}"
---
title: ${dir_name^}
weight: 1
---
Explore the following sections to learn more:

{{< cards >}}
EOF
    fi
}

# Function to add cards to _index.md files
add_cards_to_index() {
    local dir_path="$1"
    local index_file="${dir_path}/_index.md"

    # Check if _index.md exists
    if [[ -f "${index_file}" ]]; then
        # Temp file for storing new content
        local temp_file=$(mktemp)

        # Copy content before the cards section
        awk '/{{< \/cards >}}/{exit} {print}' "${index_file}" > "${temp_file}"

        # Add cards for each subdirectory or markdown file
        for item in "${dir_path}"/*; do
            if [[ -d "${item}" ]] && [[ "${item}" != "${dir_path}" ]]; then
                local name=$(basename "${item}" | sed -e "s/.md\$//")
                local link="${name}"
                echo "  {{< card link=\"${link}\" title=\"${name^}\" icon=\"document-duplicate\" >}}" >> "${temp_file}"
            fi
        done

        # Add the closing cards tag and embed tags
        cat << EOF >> "${temp_file}"
{{< /cards >}}

<!-- gomarkdoc:embed:start -->
<!-- gomarkdoc:embed:end -->
EOF

        # Replace the original file with the new content
        mv "${temp_file}" "${index_file}"
    fi
}

# Create the _index.md file for the root directory
cp README.md hugo/content/_index.md

# Traverse the source directory structure and mirror it in the docs directory
find "${SRC_DIR}" -type d | while read -r src_dir; do
    docs_dir_path="${DOCS_DIR}/${src_dir}"
    create_dir_and_index "${docs_dir_path}"
done

# Add cards to each _index.md file
find "${DOCS_DIR}" -type d | while read -r docs_dir; do
    add_cards_to_index "${docs_dir}"
done

# Generate documentation last
if which gomarkdoc >/dev/null 2>&1; then
    $(go env GOPATH)/bin/gomarkdoc ./... --output 'hugo/content/docs/{{.Dir}}/_index.md' --exclude-dirs ./pkg/internal/tests/... --embed
else
    go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
fi
