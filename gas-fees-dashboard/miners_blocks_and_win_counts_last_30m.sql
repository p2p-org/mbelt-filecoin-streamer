-- Miners Blocks And Win Counts Last 30m
--
-- {"type":"TABLE","name":"Miners Blocks And Win Counts Last 30m","description":"","options":{"itemsPerPage":25,"columns":[{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner","type":"string","displayAs":"string","visible":true,"order":100000,"title":"miner","allowSearch":true,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"mined_blocks","type":"integer","displayAs":"number","visible":true,"order":100001,"title":"mined_blocks","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"win_counts","type":"integer","displayAs":"number","visible":true,"order":100002,"title":"win_counts","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false}]},"query_id":23}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":true,"sizeX":3,"sizeY":14,"minSizeX":2,"maxSizeX":6,"minSizeY":1,"maxSizeY":1000,"col":3,"row":123}},"text":"","width":1,"dashboard_id":1,"visualization_id":48}
SELECT miner,
       count(*) AS mined_blocks,
       sum(win_count) AS win_counts
FROM filecoin.blocks
WHERE miner IN
    (SELECT miner
     FROM filecoin.miner_infos)
  AND block_time >
    (SELECT max(block_time)
     FROM filecoin.blocks) - '30 MINUTES'::interval
GROUP BY miner;