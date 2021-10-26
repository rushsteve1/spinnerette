# This is a more complex example involving the native SQLite3 and JSON libraries

# Import everything we need
(import html)
(import spin/responses :as r)
(import sqlite3 :as sql)

# Open an SQLite3 database at the file "test.db"
(def dbfile "./test.db")
(def db (sql/open dbfile))

# Insert some data into the database and query it back out
(sql/eval db `CREATE TABLE IF NOT EXISTS customers(id INTEGER PRIMARY KEY, name TEXT);`)
(sql/eval db `INSERT INTO customers VALUES(:id, :name);` {:name "John" :id 12345})
(def res (sql/eval db `SELECT * FROM customers;`))

# Close the database
(sql/close db)

# Also delete the database (ONLY FOR TESTING, DON'T DO THIS FOR REAL)
(os/rm dbfile)

# Respond with the JSON of the queried data
(r/json res)
