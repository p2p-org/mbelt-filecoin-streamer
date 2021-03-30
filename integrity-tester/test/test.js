require('dotenv').config()
const { equal } = require('assert')
const nodeApi = require('../utils/nodeApiConnection')
const postgres = require('../utils/postgres/postgres')
const postgresQueries = require('../utils/postgres/postgresQueries')

let nodeApiConnector
let postgresConnector
let jsonRpcProvider

describe('Init', () => {
  before(async () => {
    try {
      postgresConnector = postgres.pool
      nodeApiConnector = await nodeApi.getConnection()
      jsonRpcProvider = await nodeApi.getLotusClient(nodeApiConnector)
    } catch (err) {
      console.log(err.stack)
      throw new Error(`Connection error`)
    }
  })

  it('Check node top block with indexator top block', async () => {
    const nodeTopBlock = await jsonRpcProvider.chain.getHead()
    const indexatorTopBlock = await postgresQueries.getFileCoinTopBlock(postgresConnector)
    const topBlockDifference = nodeTopBlock.Height - indexatorTopBlock
    equal(topBlockDifference < 20, true, `Indexator top block is lag behind of node top block for ${topBlockDifference} blocks`)
  })

  it('Check indexator missed blocks', async () => {
    const missedBlocks = await postgresQueries.getFileCoinMissedBlocks(postgresConnector)
    equal(missedBlocks < 10, true, `Indexator has  ${missedBlocks} missed blocks`)
  })

  it('Check indexator missed tipsets', async () => {
    const missedTipsets = await postgresQueries.getFileCoinMissedTipsets(postgresConnector)
    equal(missedTipsets < 10, true, `Indexator has  ${missedTipsets} missed tipsets`)
  })




  after(function () {
    nodeApi.disconnect()
  })
})