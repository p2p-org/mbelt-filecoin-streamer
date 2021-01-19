-- Blocks Mined Last 24h
--
-- {"type":"COUNTER","name":"Blocks Mined Last 24h","description":"","options":{"counterLabel":"","counterColName":"blocks","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":20}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":2,"row":87}},"text":"","width":1,"dashboard_id":1,"visualization_id":22}
SELECT count(cid) AS blocks,
       sum(win_count) AS win_counts,
       sum(this_epoch_reward) * 0.000000000000000001 AS rewards
FROM filecoin.blocks AS blks
LEFT JOIN filecoin.reward_actor_states AS rwrd ON blks.height = rwrd.epoch
WHERE block_time >
    (SELECT max(block_time)
     FROM filecoin.blocks) - '24 hours'::interval;