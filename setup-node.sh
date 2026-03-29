#!/bin/bash

export PATH=/tmp/node-v20.19.0-linux-x64/bin:$PATH

echo "Node.js version: $(node --version)"
echo "npm version: $(npm --version)"

cd frontend/app
rm -rf node_modules package-lock.json
npm install

echo "Setup complete!"
echo "To use this Node.js version, run:"
echo "  export PATH=/tmp/node-v20.19.0-linux-x64/bin:\$PATH"
