#!/bin/bash

if [ -f schema.db ]
then
  rm schema.db
fi
sqlite3 schema.db < schema.sql
chmod 777 schema.db

if [ -d files ]
then
  true
else
  mkdir files
  chmod 777 files
fi
