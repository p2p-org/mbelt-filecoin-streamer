const { Pool } = require('pg')
require('dotenv').config({ path: '../.env' })

/** @type {Pool} */
const pool = new Pool({
  host: process.env.POSTGRES_HOST,
  user: process.env.POSTGRES_USER,
  database: process.env.POSTGRES_DB,
  password: process.env.POSTGRES_PASSWORD,
  port: process.env.POSTGRES_PORT
})

module.exports = {
  pool
}
