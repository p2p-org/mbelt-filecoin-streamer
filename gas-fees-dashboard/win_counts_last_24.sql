-- Win Counts Last 24h
--
-- {"type":"COUNTER","name":"Win Counts Last 24h","description":"","options":{"counterLabel":"","counterColName":"win_counts","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":21}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":4,"row":87}},"text":"","width":1,"dashboard_id":1,"visualization_id":24}
SELECT sum(win_count) AS win_counts
FROM filecoin.blocks AS blks
LEFT JOIN filecoin.reward_actor_states AS rwrd ON blks.height = rwrd.epoch
WHERE block_time >
    (SELECT max(block_time)
     FROM filecoin.blocks) - '24 hours'::interval;