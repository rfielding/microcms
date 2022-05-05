#!/bin/bash

if [ -f schema.db ]
then
  rm schema.db
fi
sqlite3 schema.db < schema.sql

