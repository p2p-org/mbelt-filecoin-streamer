-- Miners Info
-- Miners Info Table
-- {"type":"TABLE","name":"Miners Info","description":"","options":{"itemsPerPage":10,"columns":[{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner","type":"string","displayAs":"string","visible":true,"order":100000,"title":"miner","allowSearch":true,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"owner","type":"string","displayAs":"string","visible":true,"order":100001,"title":"owner","allowSearch":true,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"worker","type":"string","displayAs":"string","visible":true,"order":100002,"title":"worker","allowSearch":true,"alignContent":"left","allowHTML":false,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner_raw_byte_power","type":"integer","displayAs":"number","visible":true,"order":100003,"title":"miner_raw_byte_power","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"miner_quality_adj_power","type":"integer","displayAs":"number","visible":true,"order":100004,"title":"miner_quality_adj_power","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"height","type":"integer","displayAs":"number","visible":true,"order":100005,"title":"height","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0.00","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"balance","type":"float","displayAs":"number","visible":true,"order":100006,"title":"balance","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false},{"numberFormat":"0,0","booleanValues":["false","true"],"imageUrlTemplate":"{{ @ }}","imageTitleTemplate":"{{ @ }}","imageWidth":"","imageHeight":"","linkUrlTemplate":"{{ @ }}","linkTextTemplate":"{{ @ }}","linkTitleTemplate":"{{ @ }}","linkOpenInNewTab":true,"name":"hours_running","type":"integer","displayAs":"number","visible":true,"order":100007,"title":"hours_running","allowSearch":false,"alignContent":"right","allowHTML":true,"highlightLinks":false}]},"query_id":26}
-- {"options":{"parameterMappings":{},"isHidden":false,"position":{"autoHeight":true,"sizeX":6,"sizeY":12,"maxSizeY":1000,"maxSizeX":6,"minSizeY":1,"minSizeX":2,"col":0,"row":142}},"text":"","width":1,"dashboard_id":1,"visualization_id":42}
SELECT T.miner,
       T.owner,
       T.worker,
       T.miner_raw_byte_power,
       T.miner_quality_adj_power,
       T.height,
       act.balance * 0.000000000000000001 AS balance,
       hours_running
FROM
  (SELECT miner,
          OWNER,
          worker,
          miner_raw_byte_power,
          miner_quality_adj_power,
          height,
          row_number() over(PARTITION BY miner
                            ORDER BY height DESC) AS rn
   FROM filecoin.miner_infos) AS T
LEFT JOIN
  (SELECT addr,
          balance,
          height
   FROM
     (SELECT addr,
             balance,
             height,
             row_number() over(PARTITION BY addr
                               ORDER BY height DESC) AS act_rn
      FROM filecoin.actor_states) AS D
   WHERE act_rn = 1) AS act ON T.miner = act.addr
LEFT JOIN
  (SELECT miner_blocks.miner,
          miner_blocks.blocks,
          miner_blocks.win_counts,
          last_tipset,
          (
             (SELECT max(height) AS latest_epoch
              FROM filecoin.tipsets) - min(inf.height)) / 120 AS hours_running
   FROM
     (SELECT miner,
             count(*) AS blocks,
             sum(win_count) AS win_counts,
             max(height) AS last_tipset
      FROM filecoin.blocks
      GROUP BY miner) AS miner_blocks
   LEFT JOIN filecoin.miner_infos AS inf ON miner_blocks.miner = inf.miner
   GROUP BY miner_blocks.miner,
            miner_blocks.blocks,
            miner_blocks.win_counts,
            last_tipset) AS min_blks ON T.miner = min_blks.miner
WHERE rn = 1;