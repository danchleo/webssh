FROM node:18.19.1-alpine AS build-frontend
WORKDIR /app
ADD frontend ./
RUN npm config set registry https://registry.npmmirror.com
RUN npm config set strict-ssl false
RUN npm install
RUN NODE_OPTIONS=--openssl-legacy-provider npm run build

FROM golang:1.22-alpine AS build-backend
WORKDIR /app
ADD webssh ./webssh
ADD go.mod go.sum ./
ADD main.go ./
RUN GOPROXY=https://goproxy.io CGO_ENABLED=0 go build -o ./webssh-server main.go

FROM alpine
COPY --from=build-frontend /app/dist /app/public
COPY --from=build-backend /app/webssh-server /app/
RUN ln -s /app/webssh-server /usr/bin/webssh-server

CMD ["webssh-server"]