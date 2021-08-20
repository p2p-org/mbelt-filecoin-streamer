-- Rewards Last 24h
--
-- {"type":"COUNTER","name":"Rewards Last 24h (FIL)","description":"","options":{"counterLabel":"","counterColName":"rewards","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":22}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":4,"row":97}},"text":"","width":1,"dashboard_id":1,"visualization_id":38}
SELECT sum(this_epoch_reward) * 0.000000000000000001 AS rewards
FROM filecoin.reward_actor_states AS rwrd
LEFT JOIN filecoin.tipsets AS tps ON tps.height = rwrd.epoch
WHERE min_timestamp >
    (SELECT max(min_timestamp)
     FROM filecoin.tipsets) - '24 hours'::interval;