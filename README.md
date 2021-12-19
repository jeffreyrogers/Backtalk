# Backtalk: Comments system for delimitedoptions.com

This is a work in progress. It is not ready for use yet.

## Usage

You need a domain name, e.g. example.com, that points to the server you are going to run Backtalk on. I will assume you
know how to get a domain name and point it at a server that you've provisioned. You also need a postgres database that your app can connect to.

1. Run migrations to configure the DB
2. Build/cross-compile the Backtalk application and copy it to your server
3. Run the application with the proper environment variables

## Migrations

I use [migrate](https://github.com/golang-migrate/migrate) to manage migrations. Before your run the application for the first time you must apply
the migrations to the database so the schema matches what the application expects. The `migrate` command adds a table to the database that tracks
which migrations have been run so far. This way if any new migrations are added to extend functionality or fix bugs the same command can be run to
bring the database to the latest version.

Running migrations locally:

    migrate -path migrations -database "postgres://<dbuser>@localhost/backtalk?sslmode=disable" up

Running migrations on the production database is the same, but just replace the connection string with whatever is appropriate for your database.

## Environment Variables

TODO

## Developer Setup

You do not need to follow these steps to use Backtalk. These steps are only required if you are developer.

0. (Optional) Install [pre-commit](https://pre-commit.com/#install) and then run `pre-commit install` to setup git-hooks that help prevent checking
in bad code.
1. Install [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html). sqlc is a code generator. It lets us write SQL queries in SQL and generates
Go code that interfaces with those queries.
2. Run `sqlc generate` to generate the interface code. It will be put in internal/sqlc.
