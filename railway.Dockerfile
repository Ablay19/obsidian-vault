# Railway.app Deployment
FROM node:18-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci --only=production

# Copy your bot code (assuming it's Node.js based)
COPY . .

# Build your application
RUN npm run build

# Expose port
EXPOSE 8080

# Start command
CMD ["npm", "start"]