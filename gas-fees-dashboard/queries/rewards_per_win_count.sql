-- Rewards Per WinCount
--
-- {"type":"COUNTER","name":"Rewards per win_count (FIL)","description":"","options":{"counterLabel":"","counterColName":"rewards_per_win_count","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":16}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":0,"row":92}},"text":"","width":1,"dashboard_id":1,"visualization_id":28}
SELECT this_epoch_reward / sum(win_count) * 0.000000000000000001 AS rewards_per_win_count
FROM filecoin.reward_actor_states ras
LEFT JOIN filecoin.blocks blks ON ras.epoch = blks.height
GROUP BY this_epoch_reward,
         ras.epoch
ORDER BY ras.epoch DESC
LIMIT 1;