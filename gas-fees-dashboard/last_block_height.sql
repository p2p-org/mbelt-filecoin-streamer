-- Latest Block Height
--
-- {"type":"COUNTER","name":"Latest Block Height","description":"","options":{"counterLabel":"","counterColName":"max","rowNumber":1,"targetRowNumber":1,"stringDecimal":0,"stringDecChar":".","stringThouSep":",","tooltipFormat":"0,0.000"},"query_id":17}
--
SELECT max(height)
FROM filecoin.blocks;