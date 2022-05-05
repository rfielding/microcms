#!/bin/bash

if [ -f schema.db ]
then
  rm schema.db
fi
sqlite3 schema.db < schema.sql

if [ -d files ]
then
  true
else
  mkdir files
  mkdir meta
  mkdir permissions
fi
