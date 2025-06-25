#!/usr/bin/env bash
cd sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" $1 $2
cd ../..
