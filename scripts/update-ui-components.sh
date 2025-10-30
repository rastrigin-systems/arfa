#!/bin/bash

# Script to add UI components include to remaining HTML files

FILES="employee-detail.html employee-agent-configs.html team-detail.html"

for file in $FILES; do
    filepath="static/$file"

    # Check if file already has components include
    if grep -q "components.html" "$filepath"; then
        echo "✓ $file already has components include"
        continue
    fi

    # Add components include after <body> tag
    echo "Adding components include to $file..."

    # Use sed to add the script block after the opening body tag
    sed -i '' '/<body[^>]*>/a\
    <!-- Include UI Components -->\
    <script>\
        fetch('\''/components.html'\'')\
            .then(r => r.text())\
            .then(html => {\
                const temp = document.createElement('\''div'\'');\
                temp.innerHTML = html;\
                document.body.insertBefore(temp, document.body.firstChild);\
            });\
    </script>
' "$filepath"

    echo "✓ $file updated"
done

echo ""
echo "✅ All files updated!"
