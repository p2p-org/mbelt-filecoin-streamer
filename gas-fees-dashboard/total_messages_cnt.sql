-- Total Messages Count
--
-- {"type":"COUNTER","name":"Total Messages Count","description":"","options":{"counterLabel":"","counterColName":"count","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":27}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"maxSizeY":1000,"maxSizeX":6,"minSizeY":1,"minSizeX":1,"col":2,"row":102}},"text":"","id":30,"width":1,"dashboard_id":1,"visualization_id":46}
SELECT count(*)
FROM filecoin.messages;