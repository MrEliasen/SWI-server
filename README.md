# Street Wars Inc - Server

SWI is made as a spiritual successor to the original Street Wars Online 2, from 2000, and was created as a means to learn about GO and protobufs.

[Also see the  SWI Client](https://github.com/MrEliasen/SWI-client) 

## Overview

Server is written in Go, uses Protobufs and WebSockets for server/client communication, and SQLite as the DB (supports [Turso](https://turso.tech/)). The client is a very basic site written in Svelte/TS. I cannot speak on performance or how optimised the server is VS could be, but I did make some decisions based on what would be more performant.

### Setup/Install

In order to run the server we need 2 things:

##### 1. TLS Certificate.

If you are running locally you can generate a selfsigned certificate for "localhost" by running:

```sh
bash ./generate-certificate.sh
```

For production environments it ... should... automatically generate a cert via LetsEncrypt. If you are using Cloudflare, you can generate an Origin Certificate for your domain, and put it in the `./certs` directory at the swi-server root. 

##### 2. Database

You can sign up for free for a database at Turso.tech or create a new .db file at the swi-server root.

Then import the `internal/databate/migration.sql` file, you can use a tool like [Atlast](https://atlasgo.io/) for migrations if you wish, there are [Turso support](https://blog.turso.tech/database-migrations-made-easy-with-atlas-df2b259862db) as well.
I have included an `atlas.hcl` config file.


### Run

#### Development

run: `go run . --dburl "libsql://...." --env dev --domain localhost`    
`--dburl` Is the url / path to your database    
`--env` choose between `dev` and `prod`, it only changes the log level.    
`--domain` tells the server which domain cert it should load in.    

#### Production

Build for your platform: `env GOOS=linux GOARCH=arm64 go build` chaning `GOOS` and `GOARCH` with your platform.

run: `swi-server --dburl "libsql://...." --env prod --domain <domain>`    
