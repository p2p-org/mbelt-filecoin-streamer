async function getQuery(postgres, queryString) {
  return await postgres
    .query(queryString)
    .catch((err) => {
      console.log(err.stack)
      throw new Error(`Error executing query`)
    })
    .then((res) => {
      return res
    })
}


async function getFileCoinTopBlock(postgres) {
  const topBlock = await getQuery(postgres, 'select max(height) as top_block from filecoin.blocks')
  return topBlock.rows[0].top_block
}

async function getFileCoinMissedTipsets(postgres) {
  const missedTipsets = await getQuery(postgres, 'select count(*)-max(height) as missed_tipsets from filecoin.tipsets')
  return missedTipsets.rows[0].missed_tipsets
}

async function getFileCoinMissedBlocks(postgres) {
  const missedBlocks = await getQuery(
    postgres,
    '\n' +
    'select count(*) as missed_blocks from(\n' +
    'select filecoin.tipsets.height,filecoin.tipsets.blocks,count(filecoin.blocks.height)\n' +
    'from filecoin.tipsets\n' +
    'FULL OUTER JOIN filecoin.blocks on filecoin.blocks.height=filecoin.tipsets.height\n' +
    'where filecoin.tipsets.state!=1\n' +
    'group by filecoin.tipsets.height,filecoin.tipsets.blocks\n' +
    'HAVING cardinality(filecoin.tipsets.blocks) <> count(filecoin.blocks.height)\n' +
    ')t'
  )
  return missedBlocks.rows[0].missed_blocks
}

module.exports = {
  getFileCoinTopBlock,
  getFileCoinMissedTipsets,
  getFileCoinMissedBlocks
}
