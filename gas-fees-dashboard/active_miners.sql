-- Active Miners
-- Miners that were active (any activity, not only block mining) last 24 hours.
-- {"type":"COUNTER","name":"Active Miners","description":"","options":{"counterLabel":"","counterColName":"active_miners","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000","targetColName":""},"query_id":4}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":2,"row":102}},"text":"","width":1,"dashboard_id":1,"visualization_id":14}
SELECT count(DISTINCT miner) AS active_miners
FROM filecoin.miner_infos
WHERE height IN
    (SELECT height
     FROM filecoin.blocks
     WHERE block_time >
         (SELECT max(block_time)
          FROM filecoin.blocks) - '24 hours'::interval);