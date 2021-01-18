-- Total Quality Adjusted Power
-- Total Quality Adjusted Power in gigabytes
-- {"type":"COUNTER","name":"Total Quality Adjusted Power (GB)","description":"","options":{"counterLabel":"","counterColName":"power","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":13}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":false,"sizeX":2,"sizeY":5,"minSizeX":1,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":4,"row":92}},"text":"","width":1,"dashboard_id":1,"visualization_id":40}
SELECT total_quality_adj_power / (1024 * 1024 * 1024) AS power
FROM filecoin.miner_infos
ORDER BY height DESC
LIMIT 1;