-- Mined Miners
-- Miners that mined blocks in last 24h
-- {"type":"COUNTER","name":"Mined Miners (Last 24h)","description":"","options":{"counterLabel":"","counterColName":"count","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":12}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":2,"row":97}},"text":"","width":1,"dashboard_id":1,"visualization_id":2}
SELECT count(DISTINCT miner)
FROM filecoin.blocks
WHERE block_time > (
                      (SELECT max(block_time)
                       FROM filecoin.blocks) - '24 hours'::interval);

