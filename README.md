## Git Server Code Base

This repository contains the code base for our Git server. It needs to be hosted and made available for the bip-api.
We need the endpoint for this server for the git-api. 

Also please note the secret you use here should be present in bip-api `.env` file. 

## Prerequisites

Before running the Git server, make sure to build the binary for the `lib-git` library. We have provided a `lib-git` configuration for Postgres.

To build the `libgit2-backend`, execute the following commands:

```
cd libgit2-backend
mkdir build && cd build
cmake ../postgres
cmake --build .
```

Additionally, for the Postgres database, create two tables using the following SQL statements:

```
CREATE TABLE IF NOT EXISTS odb (
    oid TEXT NOT NULL,
    type INTEGER NOT NULL,
    size INTEGER NOT NULL,
    data bytea,
    repo TEXT NOT NULL,
    PRIMARY KEY (repo, oid)
);
```

```
CREATE TABLE IF NOT EXISTS refdb (
    repo    TEXT NOT NULL,
    refname TEXT NOT NULL,
    target  TEXT NOT NULL,
    type INTEGER NOT NULL,
    PRIMARY KEY (repo, refname)
);
```

## Running the Server

Before running the server, make sure to update the `sample.env` file with your credentials. The code requires a `.env` file specific to your environment settings.

Once the necessary configurations are in place, you can start the server by executing `go run main.go`.