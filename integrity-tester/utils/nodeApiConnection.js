const { HttpJsonRpcConnector, LotusClient } = require('filecoin.js')
require('dotenv').config({ path: '../.env' })
let connection = null
async function getConnection() {
    try {
        if (connection == null) {
            connection = new HttpJsonRpcConnector({ url: `${process.env.NODE_API_URL}`, token: `${process.env.NODE_API_TOKEN}` });
        }
    } catch (err) {
        console.log(err.stack)
        throw new Error(`Connection error`)
    }

    return connection
}


async function getLotusClient(apiConnector) {
    return new LotusClient(apiConnector);
}

async function disconnect() {
    connection.disconnect()
}
module.exports = {
    getConnection,
    getLotusClient,
    disconnect
}