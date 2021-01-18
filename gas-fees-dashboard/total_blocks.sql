-- Total Blocks
--
-- {"type":"COUNTER","name":"Total Blocks","description":"","options":{"counterLabel":"","counterColName":"count","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":15}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":0,"row":87}},"text":"","width":1,"dashboard_id":1,"visualization_id":34}
SELECT count(*)
FROM filecoin.blocks;