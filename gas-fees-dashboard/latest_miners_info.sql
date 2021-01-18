-- Latest Miners Info
--
-- {"type":"TABLE","name":"Latest Miners Info","description":"","options":{"itemsPerPage":10,"columns":[{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner","type":"string","displayAs":"string","visible":true,"order":100000,"title":"miner","allowSearch":true,"alignContent":"left","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner_raw_byte_power","type":"integer","displayAs":"number","visible":true,"order":100001,"title":"miner_raw_byte_power","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner_quality_adj_power","type":"integer","displayAs":"number","visible":true,"order":100002,"title":"miner_quality_adj_power","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"total_raw_byte_power","type":"integer","displayAs":"number","visible":true,"order":100003,"title":"total_raw_byte_power","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"total_quality_adj_power","type":"integer","displayAs":"number","visible":true,"order":100004,"title":"total_quality_adj_power","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false}]},"query_id":18}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":true,"sizeX":6,"sizeY":12,"maxSizeY":1000,"maxSizeX":6,"minSizeY":1,"minSizeX":2,"col":0,"row":130}},"text":"","width":1,"dashboard_id":1,"visualization_id":4}
SELECT miner,
       miner_raw_byte_power,
       miner_quality_adj_power,
       total_raw_byte_power,
       total_quality_adj_power
FROM
  (SELECT miner,
          miner_raw_byte_power,
          miner_quality_adj_power,
          total_raw_byte_power,
          total_quality_adj_power,
          row_number() over(PARTITION BY miner
                            ORDER BY height DESC) AS rn
   FROM filecoin.miner_infos) AS T
WHERE rn = 1;