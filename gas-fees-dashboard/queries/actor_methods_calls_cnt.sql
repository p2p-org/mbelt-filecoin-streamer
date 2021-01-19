-- Actor Method Calls Count
--
-- {"type":"COUNTER","name":"Actor Methods Calls","description":"","options":{"counterLabel":"","counterColName":"method_calls","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000","targetColName":""},"query_id":10}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":4,"row":102}},"text":"","width":1,"dashboard_id":1,"visualization_id":6}
SELECT count(cid) filter (WHERE method > 0) AS method_calls FROM filecoin.messages;