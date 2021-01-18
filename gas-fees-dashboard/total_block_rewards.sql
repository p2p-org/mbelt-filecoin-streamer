-- Total Block Rewards
--
-- {"type":"COUNTER","name":"Total Block Rewards (FIL)","description":"","options":{"counterLabel":"","counterColName":"total_block_rewards","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":14}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":0,"row":92}},"text":"","width":1,"dashboard_id":1,"visualization_id":28}
SELECT sum(this_epoch_reward) * 0.000000000000000001 AS total_block_rewards
FROM filecoin.reward_actor_states;