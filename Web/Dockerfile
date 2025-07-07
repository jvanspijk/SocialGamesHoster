FROM node:20-alpine as build

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

EXPOSE 51144

# Note: You might need to adjust the host setting in vite.config.js to '0.0.0.0'
# for it to be accessible from outside the container.
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "51144"]