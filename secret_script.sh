#!/bin/bash

# Create secrets directory
echo "Creating secrets directory..."
mkdir -p secrets
chmod 700 secrets

# Create secret files
echo "Creating secret files..."

# SMTP User
if [ ! -f secrets/smtp_user ]; then
    echo -n "Enter SMTP_USER: "
    read -r smtp_user
    echo -n "$smtp_user" > secrets/smtp_user
    chmod 600 secrets/smtp_user
    echo "✓ Created secrets/smtp_user"
else
    echo "✓ secrets/smtp_user already exists"
fi

# SMTP Password
if [ ! -f secrets/smtp_pass ]; then
    echo -n "Enter SMTP_PASS: "
    read -rs smtp_pass
    echo
    echo -n "$smtp_pass" > secrets/smtp_pass
    chmod 600 secrets/smtp_pass
    echo "✓ Created secrets/smtp_pass"
else
    echo "✓ secrets/smtp_pass already exists"
fi

# Add secrets directory to .gitignore
if [ -f .gitignore ]; then
    if ! grep -q "^secrets/" .gitignore; then
        echo "secrets/" >> .gitignore
        echo "✓ Added secrets/ to .gitignore"
    fi
else
    echo "secrets/" > .gitignore
    echo "✓ Created .gitignore with secrets/"
fi

echo
echo "Setup complete! Your secrets are stored in:"
echo "  - secrets/smtp_user"
echo "  - secrets/smtp_pass"
echo
echo "Remember: Never commit these files to version control!"