#!/bin/bash

# Script to replace alert() and confirm() with modern UI components

echo "Replacing alerts and confirms in remaining files..."
echo ""

# Function to replace in a file
replace_in_file() {
    local file=$1
    echo "Processing $file..."

    # Replace success alerts
    sed -i '' "s/alert('\([^']*\)successfully\([^']*\)');/Toast.success('\1successfully\2');/g" "$file"
    sed -i '' 's/alert("\([^"]*\)successfully\([^"]*\)");/Toast.success("\1successfully\2");/g' "$file"

    # Replace error/failed alerts
    sed -i '' "s/alert('Failed\([^']*\)');/Toast.error('Failed\1');/g" "$file"
    sed -i '' 's/alert("Failed\([^"]*\)");/Toast.error("Failed\1");/g' "$file"
    sed -i '' "s/alert('\([^']*\)failed\([^']*\)');/Toast.error('\1failed\2');/g" "$file"

    # Replace No/Invalid alerts
    sed -i '' "s/alert('No \([^']*\)');/Toast.warning('No \1');/g" "$file"
    sed -i '' "s/alert('Invalid\([^']*\)');/Toast.error('Invalid\1');/g" "$file"

    # Replace remaining alerts with Toast.info
    sed -i '' "s/alert(\`\([^\`]*\)\`);/Toast.info(\`\1\`);/g" "$file"

    echo "✓ $file alerts replaced"
}

# Replace in employee-detail.html
replace_in_file "static/employee-detail.html"

# Replace in employee-agent-configs.html
replace_in_file "static/employee-agent-configs.html"

# Replace in team-detail.html
replace_in_file "static/team-detail.html"

echo ""
echo "✅ All alerts replaced!"
echo ""
echo "Note: confirm() calls need manual review for proper Modal.confirm() conversion"
