FROM golang:1.18-alpine3.15 as build
WORKDIR /src
COPY .. .
RUN apk update && apk upgrade && apk add --no-cache cmake bash git openssh alpine-sdk libgit2-dev postgresql-dev postgresql-libs
RUN cd /src/libgit2-backends && \
    rm -rf build && \
    mkdir build && cd build && \
    cmake /src/libgit2-backends/postgres && \
    cmake --build . && \
    cp /src/libgit2-backends/build/libgit2-postgres.so /usr/local/lib/libgit2-postgres.so
WORKDIR /src
RUN GOOS=linux CGO_ENABLED=1 go build
RUN ls -al
RUN echo $PWD

FROM alpine:3.15.0
WORKDIR /src
COPY --from=build src/libgit2-backends/build/libgit2-postgres.so /usr/local/lib/libgit2-postgres.so
COPY --from=build /src/.env ./.env
COPY --from=build src/bipgit ./bipgit
RUN apk update && apk upgrade && apk add --no-cache alpine-sdk libgit2-dev postgresql-dev postgresql-libs

RUN echo $PWD
LABEL Name=bip-git-backend Version=0.0.1
EXPOSE 9004

RUN cd /src && ls -al
CMD ["/src/bipgit"]
