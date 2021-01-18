-- Transactions count
--
-- {"type":"COUNTER","name":"Transactions","description":"","options":{"counterLabel":"","counterColName":"transactions","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000","targetColName":""},"query_id":9}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":0,"row":97}},"text":"","width":1,"dashboard_id":1,"visualization_id":8}
SELECT count(*) AS all, count(cid) filter (WHERE method = 0 AND value > 0) AS transactions, count(cid) filter (WHERE method > 0) AS method_calls FROM filecoin.messages;