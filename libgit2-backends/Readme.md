# To build libgit2-backend run the following commands:
- `cd libgit2-backend`
- `mkdir build && cd build`
- `cmake ../postgres`
- `cmake --build .`


# For Postgres database create two tables:

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