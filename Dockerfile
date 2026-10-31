# Imagen del servidor para el despliegue en Azure Container Apps (TEC-09).
#
# Etapa 1: compila ./cmd/server con la versión de Go de go.mod. El binario es
# estático (CGO_ENABLED=0) e incluye la interfaz web, embebida con embed.
ARG GO_VERSION=1.27.1
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src
# Primero las dependencias, para que Docker reutilice esta capa si solo
# cambia el código.
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Etapa 2: imagen final mínima, sin el código fuente ni el compilador, que
# corre con un usuario sin privilegios (nonroot).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
# Sin DATABASE_URL, los datos se guardan en memoria (TEC-05): se pierden al
# reiniciar o volver a desplegar el contenedor.
ENV PORT=8080 STORAGE=memory
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
