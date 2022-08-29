#!/usr/bin/env bash
sleep 1
for f in /zeromigrations/common/*.SQL ; do
    psql postgresql://postgres:postgres@postgres:5432/common --file=$f
done